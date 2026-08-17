// ABOUTME: HTTP request handlers with HTMX detection support
// ABOUTME: Handlers return full pages or partial fragments based on HX-Request header
package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/processor"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/rss"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/scanner"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/service"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
)

// renderTemplate executes a named template with the given data.
// Logs errors and returns 500 status on failure. If the template has already
// started writing to the response, the 500 is skipped to avoid a superfluous
// WriteHeader call.
func (s *Server) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	buf := &bytes.Buffer{}
	if err := s.templates.ExecuteTemplate(buf, name, data); err != nil {
		log.Printf("Error rendering template %s: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if _, err := buf.WriteTo(w); err != nil {
		log.Printf("Error writing template %s to response: %v", name, err)
	}
}

// handleIndex serves the main index page
// Fetches both blogs and articles for initial render
// Supports filter, blog, search, and date query params for direct URL access
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	blogs, err := s.db.ListBlogs()
	if err != nil {
		log.Printf("Error fetching blogs: %v", err)
		blogs = nil
	}

	// Build search options from query parameters
	opts, filter, currentBlogID := parseSearchOptions(r)

	// Fetch articles using SearchArticles for all filter combinations
	articles, articleCount, err := s.db.SearchArticles(opts)
	if err != nil {
		log.Printf("Error fetching articles: %v", err)
		articles = nil
		articleCount = 0
	}

	// Calculate if there are more articles
	pageSize := opts.Limit
	if pageSize <= 0 {
		pageSize = model.DefaultPageSize
	}
	hasMore := len(articles) == pageSize && opts.Offset+len(articles) < articleCount
	nextOffset := opts.Offset + len(articles)
	displayedCount := opts.Offset + len(articles)

	data := map[string]interface{}{
		"Title":           "BlogWatcher",
		"Blogs":           blogs,
		"Articles":        articles,
		"ArticleCount":    articleCount,
		"DisplayedCount":  displayedCount,
		"CurrentFilter":   filter,
		"CurrentBlogID":   currentBlogID, // 0 means no blog filter active
		"CurrentBlogName": s.blogNameForID(currentBlogID),
		"SearchQuery":     opts.SearchQuery,
		"DateFrom":        r.URL.Query().Get("date_from"),
		"DateTo":          r.URL.Query().Get("date_to"),
		"Version":         s.version,
		"HasMore":         hasMore,
		"NextOffset":      nextOffset,
	}
	s.renderTemplate(w, "index.gohtml", data)
}

// handleArticleList serves the article list
// Returns partial fragment for HTMX requests, full page otherwise
// Supports filter, blog, search, and date query parameters
func (s *Server) handleArticleList(w http.ResponseWriter, r *http.Request) {
	// Build search options from query parameters
	opts, filter, currentBlogID := parseSearchOptions(r)

	// Fetch articles using SearchArticles for all filter combinations
	articles, articleCount, err := s.db.SearchArticles(opts)
	if err != nil {
		log.Printf("Error fetching articles: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Calculate if there are more articles
	pageSize := opts.Limit
	if pageSize <= 0 {
		pageSize = model.DefaultPageSize
	}
	hasMore := len(articles) == pageSize && opts.Offset+len(articles) < articleCount
	nextOffset := opts.Offset + len(articles)
	displayedCount := opts.Offset + len(articles)

	data := map[string]interface{}{
		"Articles":        articles,
		"ArticleCount":    articleCount,
		"DisplayedCount":  displayedCount,
		"CurrentFilter":   filter,
		"CurrentBlogID":   currentBlogID, // 0 means no blog filter active
		"CurrentBlogName": s.blogNameForID(currentBlogID),
		"SearchQuery":     opts.SearchQuery,
		"DateFrom":        r.URL.Query().Get("date_from"),
		"DateTo":          r.URL.Query().Get("date_to"),
		"HasMore":         hasMore,
		"NextOffset":      nextOffset,
		"IsLoadMore":      opts.Offset > 0,
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// If this is a "load more" request (offset > 0), return just the articles
		if opts.Offset > 0 {
			s.renderTemplate(w, "article-items.gohtml", data)
			return
		}
		// Return partial fragment for HTMX
		s.renderTemplate(w, "article-list.gohtml", data)
		return
	}

	// Return full page for direct navigation
	data["Title"] = "BlogWatcher"
	data["Version"] = s.version
	blogs, err := s.db.ListBlogs()
	if err != nil {
		log.Printf("Error fetching blogs for sidebar: %v", err)
		data["Blogs"] = []model.Blog{} // 设置空数组避免模板错误
	} else {
		data["Blogs"] = blogs
	}
	s.renderTemplate(w, "index.gohtml", data)
}

// handleBlogList serves the blog list
// Returns partial fragment for HTMX requests, full page otherwise
func (s *Server) handleBlogList(w http.ResponseWriter, r *http.Request) {
	blogs, err := s.db.ListBlogs()
	if err != nil {
		log.Printf("Error fetching blogs: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Blogs": blogs,
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// Return partial fragment for HTMX
		s.renderTemplate(w, "blog-list.gohtml", data)
	} else {
		// Return full page for direct navigation
		data["Title"] = "BlogWatcher"
		data["Version"] = s.version
		articles, err := s.db.ListArticles(false, nil)
		if err != nil {
			log.Printf("Error fetching articles: %v", err)
		} else {
			data["Articles"] = articles
		}
		s.renderTemplate(w, "index.gohtml", data)
	}
}

// handleBlogListGrouped serves the blog list grouped by category
// Returns partial fragment for HTMX requests
func (s *Server) handleBlogListGrouped(w http.ResponseWriter, r *http.Request) {
	// 获取当前选中的 blog ID（从 query 参数）
	currentBlogIDStr := r.URL.Query().Get("blog")
	var currentBlogID int64
	if currentBlogIDStr != "" {
		id, err := strconv.ParseInt(currentBlogIDStr, 10, 64)
		if err == nil {
			currentBlogID = id
		}
	}

	// 获取分组数据
	grouped, err := s.db.ListBlogsGroupedByCategory(currentBlogID)
	if err != nil {
		log.Printf("Error fetching grouped blogs: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Grouped": grouped,
	}

	// 返回 partial fragment
	s.renderTemplate(w, "category-group.gohtml", data)
}

// handleMarkRead marks an article as read.
// On the favorites page, returns the updated card to keep it visible.
// On other pages (e.g. unread), returns empty response so HTMX removes the card.
func (s *Server) handleMarkRead(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	log.Printf("handleMarkRead: marking article %d as read", id)

	found, err := s.db.MarkArticleRead(id)
	if err != nil {
		log.Printf("Error marking article %d as read: %v", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !found {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	log.Printf("handleMarkRead: article %d marked as read successfully", id)

	// On favorites page, re-render the card so it stays visible
	if strings.Contains(r.Referer(), "filter=favorites") {
		s.renderUpdatedArticleCard(w, id)
		return
	}

	// On other pages (e.g. unread), return empty body so HTMX outerHTML swap removes the card
	w.WriteHeader(http.StatusOK)
}

// handleMarkUnread marks an article as unread.
// On the favorites page, returns the updated card to keep it visible.
// On other pages, returns empty response so HTMX removes the card.
func (s *Server) handleMarkUnread(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	log.Printf("handleMarkUnread: marking article %d as unread", id)

	found, err := s.db.MarkArticleUnread(id)
	if err != nil {
		log.Printf("Error marking article %d as unread: %v", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !found {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	log.Printf("handleMarkUnread: article %d marked as unread successfully", id)

	// On favorites page, re-render the card so it stays visible
	if strings.Contains(r.Referer(), "filter=favorites") {
		s.renderUpdatedArticleCard(w, id)
		return
	}

	// On other pages, return empty body so HTMX outerHTML swap removes the card
	w.WriteHeader(http.StatusOK)
}

// handleFavorite marks an article as favorited and returns the updated article card
func (s *Server) handleFavorite(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	log.Printf("handleFavorite: favoriting article %d", id)

	if err := s.db.FavoriteArticle(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Article not found", http.StatusNotFound)
			return
		}
		log.Printf("Error favoriting article %d: %v", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	log.Printf("handleFavorite: article %d favorited successfully", id)

	// Return the updated article card for HTMX swap
	s.renderUpdatedArticleCard(w, id)
}

// handleUnfavorite removes the favorite mark from an article and returns the updated card
func (s *Server) handleUnfavorite(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	log.Printf("handleUnfavorite: unfavoriting article %d", id)

	if err := s.db.UnfavoriteArticle(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Article not found", http.StatusNotFound)
			return
		}
		log.Printf("Error unfavoriting article %d: %v", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	log.Printf("handleUnfavorite: article %d unfavorited successfully", id)

	// Return the updated article card for HTMX swap
	s.renderUpdatedArticleCard(w, id)
}

// renderUpdatedArticleCard fetches an article with blog info and renders just the card partial.
func (s *Server) renderUpdatedArticleCard(w http.ResponseWriter, id int64) {
	article, err := s.db.GetArticleByID(id)
	if err != nil || article == nil {
		log.Printf("Error fetching article %d for card re-render: %v", id, err)
		w.WriteHeader(http.StatusOK)
		return
	}

	blog, err := s.db.GetBlogByID(article.BlogID)
	if err != nil || blog == nil {
		log.Printf("Error fetching blog %d for article card: %v", article.BlogID, err)
		w.WriteHeader(http.StatusOK)
		return
	}

	articleWithBlog := model.ArticleWithBlog{
		ID:             article.ID,
		BlogID:         article.BlogID,
		Title:          article.Title,
		URL:            article.URL,
		ThumbnailURL:   article.ThumbnailURL,
		PublishedDate:  article.PublishedDate,
		DiscoveredDate: article.DiscoveredDate,
		IsRead:         article.IsRead,
		HasNote:        article.HasNote,
		IsFavorited:    article.IsFavorited,
		HNURL:          article.HNURL,
		HNStatus:       article.HNStatus,
		BlogName:       blog.Name,
		BlogURL:        blog.URL,
	}

	// 装配标签用于卡片 chips 渲染
	if tags, err := s.db.GetArticleTags(id); err == nil {
		articleWithBlog.Tags = tags
	}

	data := map[string]interface{}{
		"Articles":       []model.ArticleWithBlog{articleWithBlog},
		"DisplayedCount": 0,
		"ArticleCount":   0,
	}
	s.renderTemplate(w, "article-items.gohtml", data)
}

// handleMarkAllRead marks all unread articles as read and returns refreshed article list
func (s *Server) handleMarkAllRead(w http.ResponseWriter, r *http.Request) {
	// Parse optional blog filter from query params
	blogParam := r.URL.Query().Get("blog")
	var blogID *int64
	if blogParam != "" && blogParam != "0" {
		if id, err := strconv.ParseInt(blogParam, 10, 64); err == nil {
			blogID = &id
		}
	}

	// Mark all unread as read
	if err := s.db.MarkAllUnreadArticlesRead(blogID); err != nil {
		log.Printf("Error marking all articles as read: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Build search options from query parameters (preserves search/date filters)
	opts, filter, currentBlogID := parseSearchOptions(r)

	// Return refreshed article list with current filters
	articles, articleCount, err := s.db.SearchArticles(opts)
	if err != nil {
		log.Printf("Error fetching articles: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Articles":        articles,
		"ArticleCount":    articleCount,
		"CurrentFilter":   filter,
		"CurrentBlogID":   currentBlogID,
		"CurrentBlogName": s.blogNameForID(currentBlogID),
		"SearchQuery":     opts.SearchQuery,
		"DateFrom":        r.URL.Query().Get("date_from"),
		"DateTo":          r.URL.Query().Get("date_to"),
	}
	s.renderTemplate(w, "article-list.gohtml", data)
}

// handleSync triggers a scan of all blogs and returns refreshed article list.
// Uses a 3-minute timeout to prevent hanging on slow external sites.
func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Minute)
	defer cancel()

	results, err := scanner.ScanAllBlogs(ctx, s.db)
	if err != nil {
		log.Printf("Sync failed: %v", err)
		http.Error(w, "Sync failed", http.StatusInternalServerError)
		return
	}

	// Log results
	totalNew := 0
	for _, result := range results {
		if result.Error != "" {
			log.Printf("Sync error for %s: %s", result.BlogName, result.Error)
		} else {
			log.Printf("Synced %s: %d new articles (source: %s)", result.BlogName, result.NewArticles, result.Source)
			totalNew += result.NewArticles
		}
	}
	log.Printf("Sync complete: %d blogs scanned, %d new articles total", len(results), totalNew)

	// Bail out early if the client disconnected during scanning
	if ctx.Err() != nil {
		log.Printf("Sync client disconnected before rendering response")
		return
	}

	// Build search options from query parameters (preserves all filters)
	opts, filter, currentBlogID := parseSearchOptions(r)

	// Return refreshed article list with current filters
	articles, articleCount, err := s.db.SearchArticles(opts)
	if err != nil {
		log.Printf("Error fetching articles: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Articles":        articles,
		"ArticleCount":    articleCount,
		"CurrentFilter":   filter,
		"CurrentBlogID":   currentBlogID,
		"CurrentBlogName": s.blogNameForID(currentBlogID),
		"SearchQuery":     opts.SearchQuery,
		"DateFrom":        r.URL.Query().Get("date_from"),
		"DateTo":          r.URL.Query().Get("date_to"),
	}
	s.renderTemplate(w, "article-list.gohtml", data)
}

// handleAPISync triggers a scan of all blogs and returns JSON stats for programmatic use.
// Intended for cronjob or API consumers that need structured data instead of HTML.
// Uses a 3-minute timeout to prevent hanging on slow external sites.
// Thumbnail extraction happens during scanning via Open Graph fallback, so a separate
// SyncThumbnails pass is not needed here.
func (s *Server) handleAPISync(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Minute)
	defer cancel()

	results, err := scanner.ScanAllBlogs(ctx, s.db)
	if err != nil {
		log.Printf("API sync failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Aggregate scan results
	totalNew := 0
	var scanErrors []string
	for _, result := range results {
		if result.Error != "" {
			scanErrors = append(scanErrors, result.BlogName+": "+result.Error)
		} else {
			totalNew += result.NewArticles
		}
	}

	log.Printf("API sync complete: %d blogs scanned, %d new articles total", len(results), totalNew)

	resp := map[string]interface{}{
		"blogs_scanned": len(results),
		"new_articles":  totalNew,
	}
	if len(scanErrors) > 0 {
		resp["errors"] = scanErrors
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleSettings serves the settings page showing all blogs with article counts
// Returns partial fragment for HTMX requests, full page otherwise
// Supports query params: preview_success, blog_name, feed_url for success message
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	blogsWithCounts, err := s.db.ListBlogsWithCounts()
	if err != nil {
		log.Printf("Error fetching blogs with counts: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Fetch categories with blog counts for the category management section
	categories, err := s.db.ListCategoriesWithBlogCount()
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		categories = nil // Non-blocking - page still works without categories
	}

	data := map[string]interface{}{
		"SettingsBlogs":  blogsWithCounts,
		"Categories":    categories,
		"IsSettingsPage": true,
	}

	// Check for preview success query params
	if r.URL.Query().Get("preview_success") == "true" {
		data["PreviewSuccess"] = true
		data["PreviewBlogName"] = r.URL.Query().Get("blog_name")
		data["PreviewFeedURL"] = r.URL.Query().Get("feed_url")
		log.Printf("handleSettings: showing preview success message for blog '%s'", data["PreviewBlogName"])
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// Return partial fragment for HTMX
		s.renderTemplate(w, "settings-page.gohtml", data)
		return
	}

	// Return full page for direct navigation - need regular Blogs for sidebar
	blogs, err := s.db.ListBlogs()
	if err != nil {
		log.Printf("Error fetching blogs for sidebar: %v", err)
	} else {
		data["Blogs"] = blogs
	}
	data["Title"] = "Settings - BlogWatcher"
	data["Version"] = s.version
	s.renderTemplate(w, "settings.gohtml", data)
}

// blogNameForID returns the blog name for the given ID, or empty string if not found.
func (s *Server) blogNameForID(id int64) string {
	if id <= 0 {
		return ""
	}
	blog, err := s.db.GetBlogByID(id)
	if err != nil || blog == nil {
		return ""
	}
	return blog.Name
}

// parseSearchOptions extracts all search and filter parameters from the request.
// Returns SearchOptions, the filter string (for template), and currentBlogID.
func parseSearchOptions(r *http.Request) (model.SearchOptions, string, int64) {
	opts := model.SearchOptions{
		SearchQuery: r.URL.Query().Get("search"),
	}

	// Parse status filter
	filter := r.URL.Query().Get("filter")
	switch filter {
	case "read":
		isRead := true
		opts.IsRead = &isRead
	case "unread", "":
		isRead := false
		opts.IsRead = &isRead
		filter = "unread" // Default
	case "favorites":
		isFav := true
		opts.IsFavorited = &isFav
		// Don't set IsRead filter — show both read and unread favorited articles
		opts.IsRead = nil
	case "tag":
		opts.TagName = r.URL.Query().Get("tag")
		opts.IsRead = nil // 标签筛选不强制已读状态
	default:
		isRead := false
		opts.IsRead = &isRead
		filter = "unread"
	}

	// Parse blog filter
	var currentBlogID int64
	if blogParam := r.URL.Query().Get("blog"); blogParam != "" && blogParam != "0" {
		if id, err := strconv.ParseInt(blogParam, 10, 64); err == nil {
			opts.BlogID = &id
			currentBlogID = id
		}
	}

	// Parse date filters (format: 2006-01-02)
	if dateFrom := r.URL.Query().Get("date_from"); dateFrom != "" {
		if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
			opts.DateFrom = &t
		}
	}
	if dateTo := r.URL.Query().Get("date_to"); dateTo != "" {
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			opts.DateTo = &t
		}
	}

	// Parse pagination
	if offsetParam := r.URL.Query().Get("offset"); offsetParam != "" {
		if offset, err := strconv.Atoi(offsetParam); err == nil && offset >= 0 {
			opts.Offset = offset
		}
	}
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if limit, err := strconv.Atoi(limitParam); err == nil && limit > 0 {
			opts.Limit = limit
		}
	}

	return opts, filter, currentBlogID
}

// handleAddBlog handles blog addition with auto feed discovery and sync.
// Uses BlogService for validation and creation instead of external CLI.
func (s *Server) handleAddBlog(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))
	url := strings.TrimSpace(r.FormValue("url"))
	feedURL := strings.TrimSpace(r.FormValue("feed_url"))
	// Basic validation
	if name == "" || url == "" {
		s.renderAddBlogError(w, "Blog name and URL are required", name, url, feedURL)
		return
	}

	// Use service layer for business logic
	input := service.AddBlogInput{
		Name: name,
		URL:  url,
		FeedURL: feedURL,
	}

	result, err := s.blogService.AddBlog(r.Context(), input)
	if err != nil {
		// Check for domain errors
		var dupErr service.BlogAlreadyExistsError
		if errors.As(err, &dupErr) {
			s.renderAddBlogError(w, dupErr.Error(), name, url, feedURL)
			return
		}
		// Unexpected error
		log.Printf("Error adding blog: %v", err)
		s.renderAddBlogError(w, "Failed to add blog", name, url, feedURL)
		return
	}

	log.Printf("Added blog '%s' with feed %s", result.Blog.Name, result.Blog.FeedURL)

	// Auto-sync the new blog in background
	go s.autoSyncNewBlog(result.Blog.Name)

	s.renderAddBlogSuccess(w, result.Blog.Name, result.Blog.FeedURL)
}

// autoSyncNewBlog syncs a single blog by name in the background
func (s *Server) autoSyncNewBlog(blogName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	result, err := scanner.ScanBlogByName(ctx, s.db, blogName)
	if err != nil {
		log.Printf("Auto-sync failed for %s: %v", blogName, err)
		return
	}
	if result != nil {
		log.Printf("Auto-synced %s: %d new articles", blogName, result.NewArticles)
	}
}

// renderAddBlogError renders the add blog form with an error message
func (s *Server) renderAddBlogError(w http.ResponseWriter, message, name, url, feedURL string) {
	data := map[string]interface{}{
		"Error":   message,
		"Name":    name,    // Pre-populate form
		"URL":     url,     // Pre-populate form
		"FeedURL": feedURL, // Pre-populate form
	}
	s.renderTemplate(w, "add-blog-form.gohtml", data)
}

// renderAddBlogSuccess renders the add blog form with a success message
func (s *Server) renderAddBlogSuccess(w http.ResponseWriter, name, feedURL string) {
	data := map[string]interface{}{
		"Success":  true,
		"BlogName": name,
		"FeedURL":  feedURL,
	}
	s.renderTemplate(w, "add-blog-form.gohtml", data)
}

// handleGetBlog returns the blog display row partial for HTMX swap (used by cancel button)
func (s *Server) handleGetBlog(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid blog ID", http.StatusBadRequest)
		return
	}

	blog, err := s.db.GetBlogByID(id)
	if err != nil {
		log.Printf("Error fetching blog %d: %v", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if blog == nil {
		http.Error(w, "Blog not found", http.StatusNotFound)
		return
	}

	articleCount, err := s.db.GetArticleCountForBlog(id)
	if err != nil {
		log.Printf("Error fetching article count for blog %d: %v", id, err)
		articleCount = 0
	}

	// Get category name for display
	var categoryName string
	if blog.CategoryID != nil {
		category, err := s.db.GetCategoryByID(*blog.CategoryID)
		if err != nil {
			log.Printf("Error fetching category %d: %v", *blog.CategoryID, err)
		} else if category != nil {
			categoryName = category.Name
		}
	}

	data := map[string]interface{}{
		"Blog":         blog,
		"ArticleCount": articleCount,
		"CategoryName": categoryName,
	}
	s.renderTemplate(w, "blog-display-row.gohtml", data)
}

// handleEditBlog returns the blog edit form partial for HTMX swap
func (s *Server) handleEditBlog(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid blog ID", http.StatusBadRequest)
		return
	}

	blog, err := s.db.GetBlogByID(id)
	if err != nil {
		log.Printf("Error fetching blog %d: %v", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if blog == nil {
		http.Error(w, "Blog not found", http.StatusNotFound)
		return
	}

	// Fetch categories for the dropdown selection
	categories, err := s.db.ListCategoriesWithBlogCount()
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		categories = nil
	}

	data := map[string]interface{}{
		"Blog":       blog,
		"Categories": categories,
	}
	s.renderTemplate(w, "blog-edit-form.gohtml", data)
}

// handleUpdateBlogName updates the blog name and returns the display row partial
func (s *Server) handleUpdateBlogName(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid blog ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" || len(name) > 100 {
		http.Error(w, "Blog name must be 1-100 characters", http.StatusBadRequest)
		return
	}

	// 解析 URL 参数（per SETT-05）
	url := strings.TrimSpace(r.FormValue("url"))
	feedURL := strings.TrimSpace(r.FormValue("feed_url"))

	// 验证 URL 格式（per SETT-04 和 D-03）
	if err := validateURL(url); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateURL(feedURL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 获取当前 blog 数据（保留其他字段不变）
	blog, err := s.db.GetBlogByID(id)
	if err != nil {
		log.Printf("Error fetching blog %d: %v", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if blog == nil {
		http.Error(w, "Blog not found", http.StatusNotFound)
		return
	}

	// 记录旧值用于日志对比
	oldName := blog.Name
	oldURL := blog.URL
	oldFeedURL := blog.FeedURL
	oldCategoryID := blog.CategoryID

	// 更新字段（per SETT-05）
	blog.Name = name
	blog.URL = url
	blog.FeedURL = feedURL

	// 更新分类（per Phase 17）
	categoryIDStr := r.FormValue("category_id")
	var categoryID *int64
	if categoryIDStr != "" {
		catID, err := strconv.ParseInt(categoryIDStr, 10, 64)
		if err != nil {
			log.Printf("Invalid category ID '%s': %v", categoryIDStr, err)
			// Non-blocking - ignore invalid category IDs
		} else {
			categoryID = &catID
		}
	}
	blog.CategoryID = categoryID

	// 使用 UpdateBlog 一次性更新所有字段（per SETT-05）
	if err := s.db.UpdateBlog(*blog); err != nil {
		log.Printf("Error updating blog %d: %v", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	log.Printf("Updated blog %d: name='%s' (was '%s'), url='%s' (was '%s'), feed_url='%s' (was '%s'), category=%v (was %v)",
		id, name, oldName, url, oldURL, feedURL, oldFeedURL, categoryID, oldCategoryID)

	articleCount, err := s.db.GetArticleCountForBlog(id)
	if err != nil {
		log.Printf("Error fetching article count for blog %d: %v", id, err)
		articleCount = 0
	}

	// Get category name for display
	var categoryName string
	if blog.CategoryID != nil {
		category, err := s.db.GetCategoryByID(*blog.CategoryID)
		if err != nil {
			log.Printf("Error fetching category %d: %v", *blog.CategoryID, err)
		} else if category != nil {
			categoryName = category.Name
		}
	}

	// Trigger sidebar and category list refresh via HTMX events
	w.Header().Set("HX-Trigger", "blogListUpdated, categoryListUpdated")

	data := map[string]interface{}{
		"Blog":         blog,
		"ArticleCount": articleCount,
		"CategoryName": categoryName,
	}
	s.renderTemplate(w, "blog-display-row.gohtml", data)
}

// handleDeleteBlog deletes a blog and all its articles
func (s *Server) handleDeleteBlog(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid blog ID", http.StatusBadRequest)
		return
	}

	// Delete blog and all its articles
	err = s.db.DeleteBlogWithArticles(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Blog not found", http.StatusNotFound)
			return
		}
		log.Printf("Error deleting blog %d: %v", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	log.Printf("Deleted blog %d with articles", id)

	// Trigger sidebar and category list refresh via HTMX events
	w.Header().Set("HX-Trigger", "blogListUpdated, categoryListUpdated")

	// Return empty response - HTMX will remove the blog card via outerHTML swap
	w.WriteHeader(http.StatusOK)
}

// handleCategoriesList returns all categories with blog counts for HTMX list refresh
func (s *Server) handleCategoriesList(w http.ResponseWriter, r *http.Request) {
	categories, err := s.db.ListCategoriesWithBlogCount()
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Categories": categories,
	}
	s.renderTemplate(w, "category-list.gohtml", data)
}

// handleGetCategory returns a single category item for HTMX swap (used by cancel button)
func (s *Server) handleGetCategory(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	categories, err := s.db.ListCategoriesWithBlogCount()
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var found *storage.CategoryWithBlogCount
	for _, c := range categories {
		if c.ID == id {
			found = &c
			break
		}
	}
	if found == nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{
		"Category":  found,
		"BlogCount": found.BlogCount,
	}
	s.renderTemplate(w, "category-item.gohtml", data)
}

// handleCategoriesNew returns the add category form partial for HTMX swap
func (s *Server) handleCategoriesNew(w http.ResponseWriter, r *http.Request) {
	s.renderTemplate(w, "category-add-form.gohtml", nil)
}

// handleCategoriesCreate creates a new category and returns the category item partial
func (s *Server) handleCategoriesCreate(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		http.Error(w, "Category name required", http.StatusBadRequest)
		return
	}

	category, err := s.db.CreateCategory(name)
	if err != nil {
		log.Printf("Error creating category: %v", err)
		http.Error(w, "Failed to create category", http.StatusInternalServerError)
		return
	}

	log.Printf("Created category '%s' (id=%d)", category.Name, category.ID)

	data := map[string]interface{}{
		"Category":  category,
		"BlogCount": 0,
	}
	w.Header().Set("HX-Trigger", "categoryListUpdated")
	s.renderTemplate(w, "category-item.gohtml", data)
}

// handleCategoryEdit returns the category edit form partial for HTMX swap
func (s *Server) handleCategoryEdit(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	categories, err := s.db.ListCategoriesWithBlogCount()
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var found *storage.CategoryWithBlogCount
	for _, c := range categories {
		if c.ID == id {
			found = &c
			break
		}
	}
	if found == nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{
		"Category":  found,
		"BlogCount": found.BlogCount,
	}
	s.renderTemplate(w, "category-edit-form.gohtml", data)
}

// handleCategoryUpdate updates a category name and returns the category item partial
func (s *Server) handleCategoryUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		http.Error(w, "Category name required", http.StatusBadRequest)
		return
	}

	err = s.db.UpdateCategoryName(id, name)
	if err != nil {
		log.Printf("Error updating category %d: %v", id, err)
		http.Error(w, "Failed to update category", http.StatusInternalServerError)
		return
	}

	log.Printf("Updated category %d name to '%s'", id, name)

	// Get updated category with blog count
	categories, err := s.db.ListCategoriesWithBlogCount()
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var found *storage.CategoryWithBlogCount
	for _, c := range categories {
		if c.ID == id {
			found = &c
			break
		}
	}
	if found == nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{
		"Category":  found,
		"BlogCount": found.BlogCount,
	}
	w.Header().Set("HX-Trigger", "categoryListUpdated")
	s.renderTemplate(w, "category-item.gohtml", data)
}

// handleCategoryDelete deletes a category and returns empty response for HTMX swap
func (s *Server) handleCategoryDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	err = s.db.DeleteCategory(id)
	if err != nil {
		log.Printf("Error deleting category %d: %v", id, err)
		http.Error(w, "Failed to delete category", http.StatusInternalServerError)
		return
	}

	log.Printf("Deleted category %d", id)

	w.Header().Set("HX-Trigger", "categoryListUpdated, blogListUpdated")
	w.WriteHeader(http.StatusOK)
}

// handleNote serves the note page for an article
// Fetches article info from database and reads note content from file
func (s *Server) handleNote(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	// Fetch article from database
	article, err := s.db.GetArticleByID(id)
	if err != nil {
		log.Printf("Error fetching article %d: %v", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if article == nil {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	// Read note content from file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Error getting home directory: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	notePath := filepath.Join(homeDir, ".blogwatcher", "notes", fmt.Sprintf("%d.md", id))

	var content string
	noteBytes, err := os.ReadFile(notePath)
	if err != nil {
		// File not found or unreadable - return empty content
		// Template will show "备注内容为空" message
		content = ""
		if os.IsNotExist(err) {
			log.Printf("Note file not found for article %d: %s", id, notePath)
		} else {
			log.Printf("Error reading note file for article %d: %v (path: %s)", id, err, notePath)
		}
	} else {
		content = string(noteBytes)
	}

	data := map[string]interface{}{
		"ID":      id,
		"Title":   article.Title,
		"URL":     article.URL,
		"Content": content,
	}
	s.renderTemplate(w, "note.gohtml", data)
}

// handleBlogPreview handles the Preview button, parses Feed URL and returns preview page.
// PREV-02: 点击预览触发临时 feed 解析
// PREV-03: 预览页面显示解析的文章列表（最多 20 条）
// PREV-04: 预览失败显示错误信息
func (s *Server) handleBlogPreview(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))
	blogURL := strings.TrimSpace(r.FormValue("url"))

	log.Printf("handleBlogPreview: entry - name='%s', url='%s'", name, blogURL)

	// Basic validation
	if name == "" || blogURL == "" {
		log.Printf("handleBlogPreview: validation failed - name or url empty")
		data := map[string]interface{}{
			"Error":    "Blog name and URL are required",
			"BlogName": name,
			"BlogURL":  blogURL,
		}
		s.renderTemplate(w, "preview-page.gohtml", data)
		return
	}

	// URL format validation (per D-03)
	if err := validateURL(blogURL); err != nil {
		log.Printf("handleBlogPreview: URL validation failed - %v", err)
		data := map[string]interface{}{
			"Error":    err.Error(),
			"BlogName": name,
			"BlogURL":  blogURL,
		}
		s.renderTemplate(w, "preview-page.gohtml", data)
		return
	}

	// Create context with timeout for external HTTP requests
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Discover Feed URL (same logic as handleAddBlog)
	log.Printf("handleBlogPreview: discovering feed URL for '%s'", blogURL)
	feedURL, _ := rss.DiscoverFeedURL(ctx, blogURL)
	if feedURL == "" {
		log.Printf("handleBlogPreview: no feed URL discovered for '%s'", blogURL)
		data := map[string]interface{}{
			"Error":    "No RSS/Atom feed found at this URL. Please check the URL or provide a valid feed URL.",
			"BlogName": name,
			"BlogURL":  blogURL,
			"FeedURL":  "",
		}
		s.renderTemplate(w, "preview-page.gohtml", data)
		return
	}

	log.Printf("handleBlogPreview: discovered feed URL '%s' for blog '%s' at '%s'", feedURL, name, blogURL)

	// Parse Feed (PREV-02)
	proc := processor.DefaultRegistry.Get(feedURL)
	articles, err := rss.ParseFeed(ctx, feedURL, proc)
	if err != nil {
		log.Printf("handleBlogPreview: failed to parse feed '%s': %v", feedURL, err)
		var errorMsg string
		if rss.IsFeedError(err) {
			var parseErr rss.FeedParseError
			if errors.As(err, &parseErr) {
				errorMsg = parseErr.Message
			} else {
				errorMsg = err.Error()
			}
		} else {
			errorMsg = fmt.Sprintf("Failed to parse feed: %v", err)
		}
		data := map[string]interface{}{
			"Error":    errorMsg,
			"BlogName": name,
			"BlogURL":  blogURL,
			"FeedURL":  feedURL,
		}
		s.renderTemplate(w, "preview-page.gohtml", data)
		return
	}

	// Limit to 20 articles (PREV-03)
	totalCount := len(articles)
	displayedCount := totalCount
	if totalCount > 20 {
		articles = articles[:20]
		displayedCount = 20
	}

	log.Printf("handleBlogPreview: parsed %d articles from feed '%s' (showing %d)", totalCount, feedURL, displayedCount)

	// Success: render preview page
	data := map[string]interface{}{
		"BlogName":       name,
		"BlogURL":        blogURL,
		"FeedURL":        feedURL,
		"Articles":       articles,
		"TotalCount":     totalCount,
		"DisplayedCount": displayedCount,
	}
	s.renderTemplate(w, "preview-page.gohtml", data)
	log.Printf("handleBlogPreview: completed successfully")
}

// handleBlogPreviewSave handles the Save as Blog button from preview page.
// PREV-05: 预览页面有保存按钮（保存为正式 blog）
// PREV-06: 预览页面有返回修改按钮（返回添加表单）
// D-04: 保存成功后跳转到 Settings 页面显示成功消息
func (s *Server) handleBlogPreviewSave(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))
	blogURL := strings.TrimSpace(r.FormValue("url"))
	feedURL := strings.TrimSpace(r.FormValue("feed_url"))

	log.Printf("handleBlogPreviewSave: name=%s, url=%s, feed_url=%s", name, blogURL, feedURL)

	// Basic validation
	if name == "" || blogURL == "" {
		log.Printf("handleBlogPreviewSave: validation failed - name or url empty")
		data := map[string]interface{}{
			"Error":    "Blog name and URL are required",
			"BlogName": name,
			"BlogURL":  blogURL,
			"FeedURL":  feedURL,
		}
		s.renderTemplate(w, "preview-page.gohtml", data)
		return
	}

	// URL format validation
	if err := validateURL(blogURL); err != nil {
		log.Printf("handleBlogPreviewSave: URL validation failed - %v", err)
		data := map[string]interface{}{
			"Error":    err.Error(),
			"BlogName": name,
			"BlogURL":  blogURL,
			"FeedURL":  feedURL,
		}
		s.renderTemplate(w, "preview-page.gohtml", data)
		return
	}

	// Use service layer for business logic (same as handleAddBlog)
	input := service.AddBlogInput{
		Name:    name,
		URL:     blogURL,
		FeedURL: feedURL, // May be empty if preview failed
	}

	result, err := s.blogService.AddBlog(r.Context(), input)
	if err != nil {
		var dupErr service.BlogAlreadyExistsError
		if errors.As(err, &dupErr) {
			log.Printf("handleBlogPreviewSave: duplicate blog - %v", dupErr)
			data := map[string]interface{}{
				"Error":    dupErr.Error(),
				"BlogName": name,
				"BlogURL":  blogURL,
				"FeedURL":  feedURL,
			}
			s.renderTemplate(w, "preview-page.gohtml", data)
			return
		}
		// Unexpected error
		log.Printf("handleBlogPreviewSave: error saving blog - %v", err)
		data := map[string]interface{}{
			"Error":    "Failed to save blog",
			"BlogName": name,
			"BlogURL":  blogURL,
			"FeedURL":  feedURL,
		}
		s.renderTemplate(w, "preview-page.gohtml", data)
		return
	}

	log.Printf("handleBlogPreviewSave: saved blog '%s' with feed %s (ID=%d)", result.Blog.Name, result.Blog.FeedURL, result.Blog.ID)

	// Auto-sync the new blog in background (same as handleAddBlog)
	go s.autoSyncNewBlog(result.Blog.Name)

	// D-04: 跳转到 Settings 页面，使用 HX-Redirect 避免嵌套渲染
	// Trigger sidebar and category list refresh via HTMX events
	w.Header().Set("HX-Trigger", "blogListUpdated, categoryListUpdated")
	// Use HX-Redirect to navigate to settings page with success message
	redirectURL := fmt.Sprintf("/settings?preview_success=true&blog_name=%s&feed_url=%s",
		url.QueryEscape(result.Blog.Name),
		url.QueryEscape(result.Blog.FeedURL))
	w.Header().Set("HX-Redirect", redirectURL)
	w.WriteHeader(http.StatusOK)
	log.Printf("handleBlogPreviewSave: completed successfully, redirecting to Settings")
}

// validateURL 验证 URL 格式（HTTP/HTTPS），空值允许（per D-03a）
// Uses net/url for comprehensive URL validation including scheme and host checks.
func validateURL(urlStr string) error {
	if urlStr == "" {
		return nil // 空值允许（nullable 字段）
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("URL 格式无效: %v", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL 必须使用 http 或 https 协议")
	}
	if u.Host == "" {
		return fmt.Errorf("URL 必须包含主机名")
	}
	return nil
}
