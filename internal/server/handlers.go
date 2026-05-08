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
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/scanner"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/service"
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
		log.Printf("Error fetching blogs: %v", err)
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

// handleMarkRead marks an article as read and returns empty response for HTMX swap
func (s *Server) handleMarkRead(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

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

	// Return 200 OK with empty body - HTMX outerHTML swap will remove the card
	w.WriteHeader(http.StatusOK)
}

// handleMarkUnread marks an article as unread and returns empty response for HTMX swap
func (s *Server) handleMarkUnread(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

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

	// Return 200 OK with empty body - HTMX outerHTML swap will remove the card
	w.WriteHeader(http.StatusOK)
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
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	blogsWithCounts, err := s.db.ListBlogsWithCounts()
	if err != nil {
		log.Printf("Error fetching blogs with counts: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"SettingsBlogs":  blogsWithCounts,
		"IsSettingsPage": true,
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

	// Basic validation
	if name == "" || url == "" {
		s.renderAddBlogError(w, "Blog name and URL are required", name, url)
		return
	}

	// Use service layer for business logic
	input := service.AddBlogInput{
		Name: name,
		URL:  url,
	}

	result, err := s.blogService.AddBlog(r.Context(), input)
	if err != nil {
		// Check for domain errors
		var dupErr service.BlogAlreadyExistsError
		if errors.As(err, &dupErr) {
			s.renderAddBlogError(w, dupErr.Error(), name, url)
			return
		}
		// Unexpected error
		log.Printf("Error adding blog: %v", err)
		s.renderAddBlogError(w, "Failed to add blog", name, url)
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
func (s *Server) renderAddBlogError(w http.ResponseWriter, message, name, url string) {
	data := map[string]interface{}{
		"Error": message,
		"Name":  name, // Pre-populate form
		"URL":   url,  // Pre-populate form
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

	data := map[string]interface{}{
		"Blog":         blog,
		"ArticleCount": articleCount,
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

	data := map[string]interface{}{
		"Blog": blog,
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

	if err := s.db.UpdateBlogName(id, name); err != nil {
		log.Printf("Error updating blog %d name: %v", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	blog, err := s.db.GetBlogByID(id)
	if err != nil {
		log.Printf("Error fetching updated blog %d: %v", id, err)
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

	// Trigger sidebar refresh via HTMX event
	w.Header().Set("HX-Trigger", "blogListUpdated")

	data := map[string]interface{}{
		"Blog":         blog,
		"ArticleCount": articleCount,
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

	// Trigger sidebar refresh via HTMX event
	w.Header().Set("HX-Trigger", "blogListUpdated")

	// Return empty response - HTMX will remove the blog card via outerHTML swap
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
		log.Printf("Note file not found for article %d: %s", id, notePath)
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
