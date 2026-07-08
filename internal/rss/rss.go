// ABOUTME: Provides RSS/Atom feed parsing and autodiscovery functionality.
// ABOUTME: Used by scanner to fetch articles from blogs with RSS feeds.
package rss

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/processor"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/thumbnail"
	"github.com/mmcdole/gofeed"
	"github.com/mmcdole/gofeed/rss"
)

type FeedArticle struct {
	Title         string
	URL           string
	ThumbnailURL  string
	PublishedDate *time.Time
	HNURL         string // 如果 RSS comments 字段是 HN 链接，则直接提取
}

type FeedParseError struct {
	Message string
}

func (e FeedParseError) Error() string {
	return e.Message
}

func ParseFeed(ctx context.Context, feedURL string, proc processor.BlogProcessor) ([]FeedArticle, error) {
	// 抓取前按 BLOGWATCHER_FEED_HOSTMAP 改写主机（如 rsshub:1200 -> localhost:19998），
	// 使 web（容器）与 CLI（宿主）可共用同一份逻辑地址的 feed URL。
	// 注意：仅改写用于 HTTP 请求的地址，不影响调用方按原始 URL 选择 processor。
	fetchURL := RewriteFeedURL(feedURL)
	if fetchURL != feedURL {
		log.Printf("[RSS] feed URL 主机已重写: %s -> %s", feedURL, fetchURL)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fetchURL, nil)
	if err != nil {
		return nil, FeedParseError{Message: fmt.Sprintf("failed to build request: %v", err)}
	}
	client := &http.Client{Transport: http.DefaultTransport}
	response, err := client.Do(req)
	if err != nil {
		return nil, FeedParseError{Message: fmt.Sprintf("failed to fetch feed: %v", err)}
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, FeedParseError{Message: fmt.Sprintf("failed to fetch feed: status %d", response.StatusCode)}
	}

	// 使用 TeeReader 来同时检测类型和解析
	var buf bytes.Buffer
	tee := io.TeeReader(response.Body, &buf)
	feedType := gofeed.DetectFeedType(tee)

	// 重新组装 reader 以供解析
	r := io.MultiReader(&buf, response.Body)

	var articles []FeedArticle

	switch feedType {
	case gofeed.FeedTypeRSS:
		// 使用底层 RSS Parser 来获取 Comments 字段
		rp := &rss.Parser{}
		rssFeed, err := rp.Parse(r)
		if err != nil {
			return nil, FeedParseError{Message: fmt.Sprintf("failed to parse RSS feed: %v", err)}
		}

		// 使用默认翻译器获取通用 Feed（用于 thumbnail 等信息）
		translator := &gofeed.DefaultRSSTranslator{}
		genericFeed, err := translator.Translate(rssFeed)
		if err != nil {
			return nil, FeedParseError{Message: fmt.Sprintf("failed to translate RSS feed: %v", err)}
		}

		// 建立 URL 到 RSS Item 的映射（用于获取 Comments）
		rssItemMap := make(map[string]*rss.Item)
		for _, item := range rssFeed.Items {
			link := strings.TrimSpace(item.Link)
			if link != "" {
				link = proc.NormalizeArticleURL(link)
				rssItemMap[link] = item
			}
		}

		// 处理通用 Feed 的 Items
		for _, item := range genericFeed.Items {
			title := strings.TrimSpace(item.Title)
			link := strings.TrimSpace(item.Link)
			if title == "" || link == "" {
				continue
			}
			if proc.ShouldSkipArticle(title) {
				continue
			}
			link = proc.NormalizeArticleURL(link)

			// 从原始 RSS Item 提取 HN 链接：优先 <comments>，缺失时回退到 <description>
			hnURL := ""
			if rssItem, ok := rssItemMap[link]; ok {
				hnURL = extractHNURLFromComments(rssItem.Comments)
				if hnURL == "" {
					hnURL = extractHNURLFromDescription(rssItem.Description)
				}
			}

			thumbnailURL := proc.NormalizeThumbnailURL(thumbnail.ExtractFromRSS(item))
			articles = append(articles, FeedArticle{
				Title:         title,
				URL:           link,
				ThumbnailURL:  thumbnailURL,
				PublishedDate: pickPublishedDate(item),
				HNURL:         hnURL,
			})
		}

	default:
		// 对于 Atom 和 JSON Feed，使用通用 Parser
		parser := gofeed.NewParser()
		feed, err := parser.Parse(r)
		if err != nil {
			return nil, FeedParseError{Message: fmt.Sprintf("failed to parse feed: %v", err)}
		}

		for _, item := range feed.Items {
			title := strings.TrimSpace(item.Title)
			link := strings.TrimSpace(item.Link)
			if title == "" || link == "" {
				continue
			}
			if proc.ShouldSkipArticle(title) {
				continue
			}
			link = proc.NormalizeArticleURL(link)
			thumbnailURL := proc.NormalizeThumbnailURL(thumbnail.ExtractFromRSS(item))
			articles = append(articles, FeedArticle{
				Title:         title,
				URL:           link,
				ThumbnailURL:  thumbnailURL,
				PublishedDate: pickPublishedDate(item),
				HNURL:         "", // Atom/JSON feed 没有 Comments 字段
			})
		}
	}

	return articles, nil
}

// extractHNURLFromComments 从 RSS Item 的 Comments 字段提取 HN 链接
// 如果 Comments 是 HN 链接格式（https://news.ycombinator.com/item?id=...），则返回该链接
func extractHNURLFromComments(comments string) string {
	comments = strings.TrimSpace(comments)
	if comments == "" {
		return ""
	}

	// 检查是否是 HN 链接
	const hnPrefix = "https://news.ycombinator.com/item?id="
	if strings.HasPrefix(comments, hnPrefix) {
		return comments
	}

	return ""
}

// hnDiscussionURLRe 匹配 RSSHub HN 路由在 <description> 中渲染的讨论帖锚点。
// RSSHub 格式: <a href="https://news.ycombinator.com/item?id=xxx">Comments on Hacker News</a>
// 仅匹配锚点文本为 "Comments on Hacker News" 的链接，避免误提取正文里偶然出现的 HN 链接。
var hnDiscussionURLRe = regexp.MustCompile(`href="(https://news\.ycombinator\.com/item\?id=\d+)"[^>]*>Comments on Hacker News<`)

// extractHNURLFromDescription 从 RSS Item 的 Description 字段提取 HN 讨论帖链接。
// RSSHub 的 HN 路由（如 /hackernews/best、/hackernews/news）不输出 <comments> 元素，
// 而是把讨论帖地址放在 <description> 的 HTML 锚点里（锚点文本固定为 "Comments on Hacker News"）。
// gofeed 已对 XML 实体反转义，此处直接在 HTML 文本中匹配该锚点的 href。
func extractHNURLFromDescription(description string) string {
	m := hnDiscussionURLRe.FindStringSubmatch(description)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}

func DiscoverFeedURL(ctx context.Context, blogURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, blogURL, nil)
	if err != nil {
		return "", nil
	}
	client := &http.Client{Transport: http.DefaultTransport}
	response, err := client.Do(req)
	if err != nil {
		return "", nil
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", nil
	}

	base, err := url.Parse(blogURL)
	if err != nil {
		return "", nil
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return "", nil
	}

	feedTypes := []string{
		"application/rss+xml",
		"application/atom+xml",
		"application/feed+json",
		"application/xml",
		"text/xml",
	}

	for _, feedType := range feedTypes {
		selection := doc.Find(fmt.Sprintf("link[rel='alternate'][type='%s']", feedType)).First()
		if selection.Length() == 0 {
			continue
		}
		href, exists := selection.Attr("href")
		if !exists {
			continue
		}
		resolved := resolveURL(base, href)
		if resolved != "" {
			return resolved, nil
		}
	}

	commonPaths := []string{
		"/feed",
		"/feed/",
		"/rss",
		"/rss/",
		"/feed.xml",
		"/rss.xml",
		"/atom.xml",
		"/index.xml",
	}

	for _, path := range commonPaths {
		resolved := resolveURL(base, path)
		if resolved == "" {
			continue
		}
		ok, err := isValidFeed(ctx, resolved)
		if err == nil && ok {
			return resolved, nil
		}
	}

	return "", nil
}

func isValidFeed(ctx context.Context, feedURL string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return false, err
	}
	client := &http.Client{Transport: http.DefaultTransport}
	response, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return false, nil
	}

	parser := gofeed.NewParser()
	feed, err := parser.Parse(response.Body)
	if err != nil {
		return false, err
	}

	return len(feed.Items) > 0 || strings.TrimSpace(feed.Title) != "", nil
}

func resolveURL(base *url.URL, href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}
	parsed, err := url.Parse(href)
	if err != nil {
		return ""
	}
	return base.ResolveReference(parsed).String()
}

func pickPublishedDate(item *gofeed.Item) *time.Time {
	if item == nil {
		return nil
	}
	if item.PublishedParsed != nil {
		return item.PublishedParsed
	}
	if item.UpdatedParsed != nil {
		return item.UpdatedParsed
	}
	return nil
}

func IsFeedError(err error) bool {
	var parseErr FeedParseError
	return errors.As(err, &parseErr)
}
