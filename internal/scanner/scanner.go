// ABOUTME: Orchestrates blog scanning using RSS feeds or HTML scraping as fallback.
// ABOUTME: Used by the web UI sync feature to discover new articles from tracked blogs.
package scanner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/rss"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/scraper"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/thumbnail"
)

type ScanResult struct {
	BlogName     string
	NewArticles  int
	SkippedCount int // 因重复 URL 跳过的文章数量
	TotalFound   int
	Source       string
	Error        string
}

// ScanBlog scans a single blog for articles. It performs an incremental sync:
// only articles whose URLs are not already in the database are processed, and
// expensive operations like Open Graph thumbnail extraction only run for those
// genuinely new articles.
func ScanBlog(ctx context.Context, db *storage.Database, blog model.Blog) ScanResult {
	var (
		source  = "none"
		errText string
	)

	feedURL := blog.FeedURL
	if feedURL == "" {
		if discovered, err := rss.DiscoverFeedURL(ctx, blog.URL); err == nil && discovered != "" {
			feedURL = discovered
			blog.FeedURL = discovered
			_ = db.UpdateBlog(blog)
		}
	}

	// Phase 1: Collect lightweight article stubs (no OG fetching yet)
	type articleStub struct {
		BlogID        int64
		Title         string
		URL           string
		ThumbnailURL  string // only populated if RSS already provided one
		PublishedDate *time.Time
	}

	var stubs []articleStub

	if feedURL != "" {
		feedArticles, err := rss.ParseFeed(ctx, feedURL)
		if err != nil {
			errText = err.Error()
		} else {
			for _, a := range feedArticles {
				stubs = append(stubs, articleStub{
					BlogID:        blog.ID,
					Title:         a.Title,
					URL:           a.URL,
					ThumbnailURL:  a.ThumbnailURL, // from RSS only, no OG
					PublishedDate: a.PublishedDate,
				})
			}
			source = "rss"
		}
	}

	if len(stubs) == 0 && blog.ScrapeSelector != "" {
		scrapedArticles, err := scraper.ScrapeBlog(ctx, blog.URL, blog.ScrapeSelector)
		if err != nil {
			if errText != "" {
				errText = fmt.Sprintf("RSS: %s; Scraper: %s", errText, err.Error())
			} else {
				errText = err.Error()
			}
		} else {
			for _, a := range scrapedArticles {
				stubs = append(stubs, articleStub{
					BlogID:        blog.ID,
					Title:         a.Title,
					URL:           a.URL,
					PublishedDate: a.PublishedDate,
				})
			}
			source = "scraper"
			errText = ""
		}
	}

	// Phase 2: Deduplicate within the current batch
	seenURLs := make(map[string]struct{})
	uniqueStubs := make([]articleStub, 0, len(stubs))
	for _, stub := range stubs {
		if _, exists := seenURLs[stub.URL]; exists {
			continue
		}
		seenURLs[stub.URL] = struct{}{}
		uniqueStubs = append(uniqueStubs, stub)
	}

	// Phase 3: Filter out articles that already exist in the database
	urlList := make([]string, 0, len(seenURLs))
	for url := range seenURLs {
		urlList = append(urlList, url)
	}

	existing, err := db.GetExistingArticleURLs(urlList)
	if err != nil {
		errText = err.Error()
	}

	discoveredAt := time.Now()
	var newStubs []articleStub
	for _, stub := range uniqueStubs {
		if _, exists := existing[stub.URL]; exists {
			continue
		}
		newStubs = append(newStubs, stub)
	}

	// Phase 4: Only for genuinely new articles, fetch OG thumbnails if needed
	newArticles := make([]model.Article, 0, len(newStubs))
	for _, stub := range newStubs {
		thumbURL := stub.ThumbnailURL
		if thumbURL == "" {
			thumbURL = thumbnail.ExtractFromOpenGraph(ctx, stub.URL)
		}
		newArticles = append(newArticles, model.Article{
			BlogID:        stub.BlogID,
			Title:         stub.Title,
			URL:           stub.URL,
			ThumbnailURL:  thumbURL,
			PublishedDate: stub.PublishedDate,
			DiscoveredDate: &discoveredAt,
			IsRead:        false,
		})
	}

	// Phase 5: Persist new articles
	newCount := 0
	if len(newArticles) > 0 {
		count, err := db.AddArticlesBulk(newArticles)
		if err != nil {
			errText = err.Error()
		} else {
			newCount = count
		}
	}

	_ = db.UpdateBlogLastScanned(blog.ID, time.Now())

	return ScanResult{
		BlogName:    blog.Name,
		NewArticles: newCount,
		TotalFound:  len(seenURLs),
		Source:      source,
		Error:       errText,
	}
}

// ScanAllBlogs scans all blogs concurrently using goroutines and channels.
// Each blog gets its own goroutine for network I/O, but database writes are
// serialized through the single db connection to avoid SQLite write conflicts.
func ScanAllBlogs(ctx context.Context, db *storage.Database) ([]ScanResult, error) {
	blogs, err := db.ListBlogs()
	if err != nil {
		return nil, err
	}

	if len(blogs) == 0 {
		return nil, nil
	}

	results := make([]ScanResult, len(blogs))
	resultCh := make(chan struct {
		Index  int
		Result ScanResult
	}, len(blogs))

	var wg sync.WaitGroup
	for i, blog := range blogs {
		wg.Add(1)
		go func(index int, b model.Blog) {
			defer wg.Done()
			result := ScanBlog(ctx, db, b)
			resultCh <- struct {
				Index  int
				Result ScanResult
			}{Index: index, Result: result}
		}(i, blog)
	}

	// Close the channel once all goroutines complete
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for item := range resultCh {
		results[item.Index] = item.Result
	}

	return results, nil
}

func ScanBlogByName(ctx context.Context, db *storage.Database, name string) (*ScanResult, error) {
	blog, err := db.GetBlogByName(name)
	if err != nil {
		return nil, err
	}
	if blog == nil {
		return nil, nil
	}
	result := ScanBlog(ctx, db, *blog)
	return &result, nil
}

// ThumbnailSyncResult holds the result of a thumbnail sync operation.
type ThumbnailSyncResult struct {
	Total   int
	Updated int
	Errors  int
}

// SyncThumbnails re-fetches thumbnails for articles that have empty thumbnail_url.
// For each article, it re-parses the RSS feed to find the matching item and extract thumbnail.
// Falls back to Open Graph if RSS doesn't provide a thumbnail.
func SyncThumbnails(ctx context.Context, db *storage.Database) (ThumbnailSyncResult, error) {
	articles, err := db.GetArticlesMissingThumbnails()
	if err != nil {
		return ThumbnailSyncResult{}, err
	}

	result := ThumbnailSyncResult{Total: len(articles)}

	// Group articles by feed URL to avoid re-fetching the same feed multiple times
	feedArticles := make(map[string][]storage.ArticleForThumbnailSync)
	for _, a := range articles {
		feedArticles[a.FeedURL] = append(feedArticles[a.FeedURL], a)
	}

	// Process each feed
	for feedURL, articleList := range feedArticles {
		feedItems, err := rss.ParseFeed(ctx, feedURL)
		if err != nil {
			result.Errors += len(articleList)
			continue
		}

		// Build URL to thumbnail map from feed
		feedThumbnails := make(map[string]string)
		for _, item := range feedItems {
			if item.ThumbnailURL != "" {
				feedThumbnails[item.URL] = item.ThumbnailURL
			}
		}

		// Update each article
		for _, article := range articleList {
			thumbnailURL := feedThumbnails[article.URL]

			// Fallback to Open Graph if RSS didn't provide thumbnail
			if thumbnailURL == "" {
				thumbnailURL = thumbnail.ExtractFromOpenGraph(ctx, article.URL)
			}

			if thumbnailURL != "" {
				if err := db.UpdateArticleThumbnail(article.ID, thumbnailURL); err != nil {
					result.Errors++
				} else {
					result.Updated++
				}
			}
		}
	}

	return result, nil
}
