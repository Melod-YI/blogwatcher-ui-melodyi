# Article Favorites Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add article favorites/bookmarks functionality — users can star/unstar articles via UI and CLI, and view a filtered list of favorited articles.

**Architecture:** Add `is_favorited BOOLEAN DEFAULT FALSE` column to the `articles` table, following the exact same pattern as `is_read` and `has_note`. Two API endpoints (favorite/unfavorite), two CLI subcommands, a star button on article cards, and a sidebar navigation entry.

**Tech Stack:** Go 1.25+, SQLite (modernc.org/sqlite), HTMX, Go html/template

---

## File Structure

| File | Action | Responsibility |
|------|--------|---------------|
| `internal/model/model.go` | Modify | Add `IsFavorited` field to `Article`, `ArticleWithBlog`, `SearchOptions` |
| `internal/storage/database.go` | Modify | Migration, `FavoriteArticle`/`UnfavoriteArticle` methods, update all `scan*` functions and SELECT queries |
| `internal/server/routes.go` | Modify | Register favorite/unfavorite routes |
| `internal/server/handlers.go` | Modify | Add `handleFavorite`/`handleUnfavorite` handlers, update `parseSearchOptions` |
| `internal/cli/commands/article.go` | Modify | Add `favorite`/`unfavorite` subcommands, add `--favorited` flag to `list` |
| `internal/cli/output/json.go` | Modify | Add `is_favorited` field to JSON output |
| `internal/cli/output/simple.go` | Modify | Add favorite status to simple output |
| `internal/cli/output/table.go` | Modify | Add favorite status to table output |
| `assets/templates/partials/article-items.gohtml` | Modify | Add star button to article cards |
| `assets/templates/partials/article-list.gohtml` | Modify | Add star button to article cards, update header for favorites filter |
| `assets/templates/partials/sidebar.gohtml` | Modify | Add Favorites nav entry |
| `assets/static/styles.css` | Modify | Add favorite button styles |
| `internal/storage/database_test.go` | Modify | Add favorite-related tests |
| `internal/server/handlers_test.go` | Modify | Add handler tests |

---

### Task 1: Model — Add IsFavorited fields

**Files:**
- Modify: `internal/model/model.go:24-36` (Article struct)
- Modify: `internal/model/model.go:40-54` (ArticleWithBlog struct)
- Modify: `internal/model/model.go:58-66` (SearchOptions struct)

- [ ] **Step 1: Add IsFavorited to Article struct**

In `internal/model/model.go`, add `IsFavorited bool` to the `Article` struct after `HasNote`:

```go
type Article struct {
	ID             int64
	BlogID         int64
	Title          string
	URL            string
	ThumbnailURL   string
	PublishedDate  *time.Time
	DiscoveredDate *time.Time
	IsRead         bool
	HasNote        bool
	IsFavorited    bool     // 文章收藏状态
	HNURL          string
	HNStatus       HNStatus
}
```

- [ ] **Step 2: Add IsFavorited to ArticleWithBlog struct**

Same field after `HasNote`:

```go
type ArticleWithBlog struct {
	ID             int64
	BlogID         int64
	Title          string
	URL            string
	ThumbnailURL   string
	PublishedDate  *time.Time
	DiscoveredDate *time.Time
	IsRead         bool
	BlogName       string
	BlogURL        string
	HasNote        bool
	IsFavorited    bool     // 文章收藏状态
	HNURL          string
	HNStatus       HNStatus
}
```

- [ ] **Step 3: Add IsFavorited to SearchOptions struct**

```go
type SearchOptions struct {
	SearchQuery string
	IsRead      *bool
	IsFavorited *bool      // nil = all, true = favorited only, false = non-favorited only
	BlogID      *int64
	DateFrom    *time.Time
	DateTo      *time.Time
	Limit       int
	Offset      int
}
```

- [ ] **Step 4: Add IsFavorited to ListFilterOptions**

In `internal/storage/database.go`, add `IsFavorited *bool` to `ListFilterOptions` struct (around line 1391):

```go
type ListFilterOptions struct {
	BlogName     string
	CategoryName string
	IsRead       *bool
	HasNote      *bool
	IsFavorited  *bool      // 收藏状态筛选（nil 表示所有状态）
	AfterDate    *time.Time
	Limit        int
	Offset       int
}
```

- [ ] **Step 5: Verify compilation**

Run: `go build ./...`
Expected: Compiles with no errors (fields are unused but valid)

- [ ] **Step 6: Commit**

```bash
git add internal/model/model.go internal/storage/database.go
git commit -m "feat(model): add IsFavorited field to Article, ArticleWithBlog, SearchOptions, ListFilterOptions"
```

---

### Task 2: Storage — Migration and CRUD methods

**Files:**
- Modify: `internal/storage/database.go` (ensureMigrations, new methods)

- [ ] **Step 1: Add migration for is_favorited column**

In `ensureMigrations()` method, add after the `hn_status` migration block (after line 203):

```go
	// Add is_favorited column if it doesn't exist
	if !db.columnExists("articles", "is_favorited") {
		if _, err := db.conn.Exec(`ALTER TABLE articles ADD COLUMN is_favorited BOOLEAN DEFAULT FALSE`); err != nil {
			return fmt.Errorf("failed to add is_favorited column: %w", err)
		}
	}
```

- [ ] **Step 2: Add FavoriteArticle and UnfavoriteArticle methods**

Add these two methods after the existing `UpdateArticleHasNote` method (around line 787):

```go
// FavoriteArticle marks an article as favorited.
// Returns error if the article does not exist.
func (db *Database) FavoriteArticle(id int64) error {
	result, err := db.conn.Exec(`UPDATE articles SET is_favorited = 1 WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to favorite article: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("article not found: %d", id)
	}
	return nil
}

// UnfavoriteArticle removes the favorite mark from an article.
// Returns error if the article does not exist.
func (db *Database) UnfavoriteArticle(id int64) error {
	result, err := db.conn.Exec(`UPDATE articles SET is_favorited = 0 WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to unfavorite article: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("article not found: %d", id)
	}
	return nil
}
```

- [ ] **Step 3: Verify compilation**

Run: `go build ./...`
Expected: Compiles successfully

- [ ] **Step 4: Commit**

```bash
git add internal/storage/database.go
git commit -m "feat(storage): add is_favorited migration and FavoriteArticle/UnfavoriteArticle methods"
```

---

### Task 3: Storage — Update all scan functions and SELECT queries

**Files:**
- Modify: `internal/storage/database.go` (scan functions, SELECT queries)

This is the most tedious task — every function that reads from the `articles` table must include `is_favorited`.

- [ ] **Step 1: Update scanArticle function (line ~1143)**

Add `isFavorited bool` variable and update the Scan call. The SELECT columns order must match:

```go
func scanArticle(scanner interface{ Scan(dest ...any) error }) (*model.Article, error) {
	var (
		id            int64
		blogID        int64
		title         string
		url           string
		thumbnailURL  sql.NullString
		publishedDate sql.NullString
		discovered    sql.NullString
		isRead        bool
		hasNote       bool
		hnURL         sql.NullString
		hnStatus      sql.NullString
		isFavorited   bool
	)
	if err := scanner.Scan(&id, &blogID, &title, &url, &thumbnailURL, &publishedDate, &discovered, &isRead, &hasNote, &hnURL, &hnStatus, &isFavorited); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	article := &model.Article{
		ID:           id,
		BlogID:       blogID,
		Title:        title,
		URL:          url,
		ThumbnailURL: thumbnailURL.String,
		IsRead:       isRead,
		HasNote:      hasNote,
		IsFavorited:  isFavorited,
		HNURL:        hnURL.String,
		HNStatus:     model.HNStatus(hnStatus.String),
	}
	// ... rest unchanged
```

- [ ] **Step 2: Update scanArticleWithBlog function (line ~1215)**

Add `isFavorited bool` variable:

```go
func scanArticleWithBlog(scanner interface{ Scan(dest ...any) error }) (*model.ArticleWithBlog, error) {
	var (
		id            int64
		blogID        int64
		title         string
		url           string
		thumbnailURL  sql.NullString
		publishedDate sql.NullString
		discovered    sql.NullString
		isRead        bool
		hasNote       bool
		hnURL         sql.NullString
		hnStatus      sql.NullString
		blogName      string
		blogURL       string
		isFavorited   bool
	)
	if err := scanner.Scan(&id, &blogID, &title, &url, &thumbnailURL, &publishedDate, &discovered, &isRead, &hasNote, &hnURL, &hnStatus, &blogName, &blogURL, &isFavorited); err != nil {
```

And set it in the struct:

```go
	article := &model.ArticleWithBlog{
		// ... existing fields ...
		HasNote:      hasNote,
		IsFavorited:  isFavorited,
		BlogName:     blogName,
		BlogURL:      blogURL,
		HNURL:        hnURL.String,
		HNStatus:     model.HNStatus(hnStatus.String),
	}
```

- [ ] **Step 3: Update scanArticleWithBlogAndCount function (line ~1265)**

Same pattern — add `isFavorited bool` before `totalCount int` in the var block:

```go
	var (
		// ... existing fields through hnStatus ...
		blogName      string
		blogURL       string
		isFavorited   bool
		totalCount    int
	)
	if err := scanner.Scan(&id, &blogID, &title, &url, &thumbnailURL, &publishedDate, &discovered, &isRead, &hasNote, &hnURL, &hnStatus, &blogName, &blogURL, &isFavorited, &totalCount); err != nil {
```

And set `IsFavorited: isFavorited` in the struct.

- [ ] **Step 4: Update all SELECT queries to include is_favorited**

Every query that SELECTs article columns must add `is_favorited` after `hn_status`. Update these queries:

1. **`ListArticles`** (line ~561): Change SELECT to include `is_favorited`:
```go
query := `SELECT id, blog_id, title, url, thumbnail_url, published_date, discovered_date, is_read, has_note, hn_url, hn_status, is_favorited FROM articles WHERE 1=1`
```

2. **`ListArticlesByReadStatus`** (line ~595):
```go
query := `SELECT id, blog_id, title, url, thumbnail_url, published_date, discovered_date, is_read, has_note, hn_url, hn_status, is_favorited FROM articles WHERE is_read = ?`
```

3. **`ListArticlesWithBlog`** (line ~627):
```go
query := `SELECT a.id, a.blog_id, a.title, a.url, a.thumbnail_url, a.published_date, a.discovered_date, a.is_read, a.has_note, a.hn_url, a.hn_status, a.is_favorited, b.name, b.url
```

4. **`SearchArticles`** (line ~664):
```go
query.WriteString(`SELECT a.id, a.blog_id, a.title, a.url, a.thumbnail_url, a.published_date, a.discovered_date, a.is_read, a.has_note, a.hn_url, a.hn_status, a.is_favorited, b.name, b.url, COUNT(*) OVER() as total_count
```

5. **`GetArticleByID`** (line ~830):
```go
row := db.conn.QueryRow(`SELECT id, blog_id, title, url, thumbnail_url, published_date, discovered_date, is_read, has_note, hn_url, hn_status, is_favorited FROM articles WHERE id = ?`, id)
```

6. **`ListArticlesWithFilters`** (line ~1405):
```go
query := `SELECT a.id, a.blog_id, a.title, a.url, a.thumbnail_url, a.published_date, a.discovered_date, a.is_read, a.has_note, a.hn_url, a.hn_status, a.is_favorited, b.name, b.url
```

- [ ] **Step 5: Add is_favorited filter to SearchArticles**

In `SearchArticles` method, after the `IsRead` condition block (around line 684), add:

```go
	// Add favorite filter if IsFavorited is not nil
	if opts.IsFavorited != nil {
		conditions = append(conditions, "a.is_favorited = ?")
		args = append(args, *opts.IsFavorited)
	}
```

- [ ] **Step 6: Add is_favorited filter to buildFilterConditions**

In `buildFilterConditions` function (around line 1768), after the `HasNote` condition block, add:

```go
	// 收藏状态筛选
	if opts.IsFavorited != nil {
		conditions = append(conditions, "a.is_favorited = ?")
		args = append(args, *opts.IsFavorited)
	}
```

- [ ] **Step 7: Verify compilation and run existing tests**

Run: `go build ./...`
Expected: Compiles successfully

Run: `go test ./internal/storage/...`
Expected: All existing tests pass

- [ ] **Step 8: Commit**

```bash
git add internal/storage/database.go
git commit -m "feat(storage): add is_favorited to all scan functions and SELECT queries"
```

---

### Task 4: Storage — Write tests

**Files:**
- Modify: `internal/storage/database_test.go`

- [ ] **Step 1: Add TestFavoriteArticle**

```go
func TestFavoriteArticle(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	// Create a blog and article
	blog, err := db.AddBlog(model.Blog{Name: "Test", URL: "https://example.com"})
	if err != nil {
		t.Fatalf("add blog: %v", err)
	}

	articles := []model.Article{
		{BlogID: blog.ID, Title: "Test Article", URL: "https://example.com/1", HNStatus: model.HNStatusNotSearch},
	}
	inserted, _, err := db.AddArticlesBulk(articles)
	if err != nil || inserted != 1 {
		t.Fatalf("add articles: inserted=%d, skipped=%d, err=%v", inserted, 0, err)
	}

	// Get the article ID
	allArticles, err := db.ListArticles(false, nil)
	if err != nil || len(allArticles) != 1 {
		t.Fatalf("list articles: %v, count=%d", err, len(allArticles))
	}
	articleID := allArticles[0].ID

	// Verify default is not favorited
	if allArticles[0].IsFavorited {
		t.Fatal("expected article to not be favorited by default")
	}

	// Favorite the article
	if err := db.FavoriteArticle(articleID); err != nil {
		t.Fatalf("favorite article: %v", err)
	}

	// Verify it is now favorited
	article, err := db.GetArticleByID(articleID)
	if err != nil {
		t.Fatalf("get article: %v", err)
	}
	if !article.IsFavorited {
		t.Fatal("expected article to be favorited")
	}

	// Unfavorite the article
	if err := db.UnfavoriteArticle(articleID); err != nil {
		t.Fatalf("unfavorite article: %v", err)
	}

	// Verify it is no longer favorited
	article, err = db.GetArticleByID(articleID)
	if err != nil {
		t.Fatalf("get article: %v", err)
	}
	if article.IsFavorited {
		t.Fatal("expected article to not be favorited after unfavorite")
	}
}

func TestFavoriteArticleNotFound(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	err := db.FavoriteArticle(99999)
	if err == nil {
		t.Fatal("expected error for non-existent article")
	}

	err = db.UnfavoriteArticle(99999)
	if err == nil {
		t.Fatal("expected error for non-existent article")
	}
}

func TestSearchArticlesWithFavoriteFilter(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	// Create blog and articles
	blog, err := db.AddBlog(model.Blog{Name: "Test", URL: "https://example.com"})
	if err != nil {
		t.Fatalf("add blog: %v", err)
	}

	articles := []model.Article{
		{BlogID: blog.ID, Title: "Article 1", URL: "https://example.com/1", HNStatus: model.HNStatusNotSearch},
		{BlogID: blog.ID, Title: "Article 2", URL: "https://example.com/2", HNStatus: model.HNStatusNotSearch},
		{BlogID: blog.ID, Title: "Article 3", URL: "https://example.com/3", HNStatus: model.HNStatusNotSearch},
	}
	inserted, _, err := db.AddArticlesBulk(articles)
	if err != nil || inserted != 3 {
		t.Fatalf("add articles: inserted=%d, err=%v", inserted, err)
	}

	// Favorite article 1 and 2
	allArticles, _ := db.ListArticles(false, nil)
	if err := db.FavoriteArticle(allArticles[0].ID); err != nil {
		t.Fatalf("favorite: %v", err)
	}
	if err := db.FavoriteArticle(allArticles[1].ID); err != nil {
		t.Fatalf("favorite: %v", err)
	}

	// Search for favorited only
	isFav := true
	opts := model.SearchOptions{IsFavorited: &isFav, Limit: 100}
	results, count, err := db.SearchArticles(opts)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 favorited articles, got %d", count)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.IsFavorited {
			t.Fatalf("expected all results to be favorited, got %+v", r)
		}
	}

	// Search for non-favorited only
	isFav = false
	opts = model.SearchOptions{IsFavorited: &isFav, Limit: 100}
	results, count, err = db.SearchArticles(opts)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 non-favorited article, got %d", count)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
```

- [ ] **Step 2: Run the tests**

Run: `go test ./internal/storage/... -run "TestFavorite|TestSearchArticlesWithFavorite" -v`
Expected: All 3 new tests PASS

- [ ] **Step 3: Run all tests**

Run: `go test ./...`
Expected: All tests pass

- [ ] **Step 4: Commit**

```bash
git add internal/storage/database_test.go
git commit -m "test(storage): add tests for favorite/unfavorite and favorite filter"
```

---

### Task 5: API — Register routes and add handlers

**Files:**
- Modify: `internal/server/routes.go:23-25`
- Modify: `internal/server/handlers.go`

- [ ] **Step 1: Register routes**

In `internal/server/routes.go`, add after the existing article management actions (after line 25):

```go
	s.mux.HandleFunc("POST /articles/{id}/favorite", s.handleFavorite)
	s.mux.HandleFunc("POST /articles/{id}/unfavorite", s.handleUnfavorite)
```

- [ ] **Step 2: Add handleFavorite handler**

In `internal/server/handlers.go`, add after `handleMarkUnread` (after line 263):

```go
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
// Uses a dedicated single-card template to avoid requiring list-level data (DisplayedCount, etc.).
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

	data := map[string]interface{}{
		"Articles":       []model.ArticleWithBlog{articleWithBlog},
		"DisplayedCount": 0, // Suppress results-info div rendering
		"ArticleCount":   0, // Suppress results-info div rendering
	}
	s.renderTemplate(w, "article-items.gohtml", data)
}
```

- [ ] **Step 3: Update parseSearchOptions to support favorites filter**

In `parseSearchOptions` function (around line 476), update the filter switch to handle "favorites":

```go
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
	default:
		isRead := false
		opts.IsRead = &isRead
		filter = "unread"
	}
```

- [ ] **Step 4: Verify compilation**

Run: `go build ./...`
Expected: Compiles successfully

- [ ] **Step 5: Commit**

```bash
git add internal/server/routes.go internal/server/handlers.go
git commit -m "feat(api): add favorite/unfavorite endpoints and favorites filter"
```

---

### Task 6: API — Write handler tests

**Files:**
- Modify: `internal/server/handlers_test.go`

- [ ] **Step 1: Add handler tests**

```go
func TestHandleFavoriteAndUnfavorite(t *testing.T) {
	srv := createTestServer(t)

	// Setup: create blog and article
	form := url.Values{}
	form.Set("name", "Test Blog")
	form.Set("url", "https://example.com")
	req := httptest.NewRequest(http.MethodPost, "/blogs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	// Get article ID by querying
	db := srv.(*Server).db
	articles, err := db.ListArticles(false, nil)
	if err != nil || len(articles) == 0 {
		t.Skip("No articles available for testing (blog needs feed)")
	}
	// Manually insert a test article for deterministic testing
	blog, _ := db.GetBlogByName("Test Blog")
	if blog == nil {
		t.Fatal("blog not found")
	}
	inserted, _, err := db.AddArticlesBulk([]model.Article{
		{BlogID: blog.ID, Title: "Test Article", URL: "https://example.com/test-fav", HNStatus: model.HNStatusNotSearch},
	})
	if err != nil || inserted == 0 {
		t.Fatalf("insert article: %v", err)
	}
	articles, _ = db.ListArticles(false, nil)
	articleID := articles[0].ID

	// Test favorite
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/articles/%d/favorite", articleID), nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("favorite: status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	// Verify in DB
	article, _ := db.GetArticleByID(articleID)
	if !article.IsFavorited {
		t.Error("expected article to be favorited after favorite endpoint")
	}

	// Test unfavorite
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/articles/%d/unfavorite", articleID), nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("unfavorite: status = %d, want %d", rec.Code, http.StatusOK)
	}

	article, _ = db.GetArticleByID(articleID)
	if article.IsFavorited {
		t.Error("expected article to not be favorited after unfavorite endpoint")
	}
}

func TestHandleFavoriteInvalidID(t *testing.T) {
	srv := createTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/articles/notanumber/favorite", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleFavoriteNotFound(t *testing.T) {
	srv := createTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/articles/99999/favorite", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
```

- [ ] **Step 2: Run the tests**

Run: `go test ./internal/server/... -run "TestHandleFavorite|TestHandleUnfavorite" -v`
Expected: All tests PASS

- [ ] **Step 3: Commit**

```bash
git add internal/server/handlers_test.go
git commit -m "test(api): add tests for favorite/unfavorite handlers"
```

---

### Task 7: CLI — Add favorite/unfavorite subcommands

**Files:**
- Modify: `internal/cli/commands/article.go`

- [ ] **Step 1: Register new subcommands in NewArticleCmd**

In `NewArticleCmd()` function, add after `cmd.AddCommand(NewMarkUnreadCmd())`:

```go
	cmd.AddCommand(NewFavoriteCmd())
	cmd.AddCommand(NewUnfavoriteCmd())
```

Also update the Long description to include the new subcommands:

```go
		Long: `文章管理命令，提供文章列表和状态管理功能。

子命令：
  list        列出文章，支持筛选和多种输出格式
  mark-read   标记文章已读（单篇或全部）
  mark-unread 标记文章未读
  favorite    收藏文章
  unfavorite  取消收藏文章`,
```

- [ ] **Step 2: Add NewFavoriteCmd function**

```go
// NewFavoriteCmd creates the favorite subcommand
func NewFavoriteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "favorite <id>",
		Short: "收藏文章",
		Long: `收藏指定文章。

示例：
  blogwatcher article favorite 1  # 收藏文章 ID 1`,
		Args: cobra.ExactArgs(1),
		Run:  runFavorite,
	}
	return cmd
}

func runFavorite(cmd *cobra.Command, args []string) {
	dbPath := flags.DBPath()
	if dbPath == "" {
		var err error
		dbPath, err = storage.DefaultDBPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取默认数据库路径失败: %v\n", err)
			os.Exit(1)
		}
	}

	db, err := storage.OpenDatabase(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "文章 ID 格式错误: %v\n", err)
		os.Exit(1)
	}

	// Verify article exists
	article, err := db.GetArticleByID(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查询文章失败: %v\n", err)
		os.Exit(1)
	}
	if article == nil {
		fmt.Fprintf(os.Stderr, "文章 %d 不存在\n", id)
		os.Exit(1)
	}

	if err := db.FavoriteArticle(id); err != nil {
		fmt.Fprintf(os.Stderr, "收藏文章失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("已收藏: %s\n", article.Title)
}
```

- [ ] **Step 3: Add NewUnfavoriteCmd function**

```go
// NewUnfavoriteCmd creates the unfavorite subcommand
func NewUnfavoriteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unfavorite <id>",
		Short: "取消收藏文章",
		Long: `取消收藏指定文章。

示例：
  blogwatcher article unfavorite 1  # 取消收藏文章 ID 1`,
		Args: cobra.ExactArgs(1),
		Run:  runUnfavorite,
	}
	return cmd
}

func runUnfavorite(cmd *cobra.Command, args []string) {
	dbPath := flags.DBPath()
	if dbPath == "" {
		var err error
		dbPath, err = storage.DefaultDBPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取默认数据库路径失败: %v\n", err)
			os.Exit(1)
		}
	}

	db, err := storage.OpenDatabase(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "文章 ID 格式错误: %v\n", err)
		os.Exit(1)
	}

	article, err := db.GetArticleByID(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查询文章失败: %v\n", err)
		os.Exit(1)
	}
	if article == nil {
		fmt.Fprintf(os.Stderr, "文章 %d 不存在\n", id)
		os.Exit(1)
	}

	if err := db.UnfavoriteArticle(id); err != nil {
		fmt.Fprintf(os.Stderr, "取消收藏失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("已取消收藏: %s\n", article.Title)
}
```

- [ ] **Step 4: Add --favorited flag to list command**

In `NewListCmd()`, add the flag after the existing flags:

```go
	cmd.Flags().Bool("favorited", false, "仅收藏文章")
```

Add to the Long description:
```go
  --favorited      仅显示收藏文章
```

- [ ] **Step 5: Handle --favorited in runList**

In `runList` function, after the `HasNote` filter section (around line 226), add:

```go
	// 设置 IsFavorited 状态筛选
	favorited, _ := cmd.Flags().GetBool("favorited")
	if favorited {
		isFav := true
		opts.IsFavorited = &isFav
	}
```

- [ ] **Step 6: Verify compilation**

Run: `go build ./...`
Expected: Compiles successfully

- [ ] **Step 7: Commit**

```bash
git add internal/cli/commands/article.go
git commit -m "feat(cli): add favorite/unfavorite subcommands and --favorited list flag"
```

---

### Task 8: CLI — Update output formats

**Files:**
- Modify: `internal/cli/output/json.go`
- Modify: `internal/cli/output/simple.go`
- Modify: `internal/cli/output/table.go`

- [ ] **Step 1: Add is_favorited to JSON output**

In `internal/cli/output/json.go`, add field to `ArticleJSONOutput`:

```go
type ArticleJSONOutput struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Blog        string `json:"blog"`
	Read        bool   `json:"read"`
	HasNote     bool   `json:"has_note"`
	IsFavorited bool   `json:"is_favorited"`
	Published   string `json:"published,omitempty"`
	HNURL       string `json:"hn_url,omitempty"`
	HNStatus    string `json:"hn_status,omitempty"`
}
```

And in `FormatJSON`, set the field:

```go
		data[i] = ArticleJSONOutput{
			ID:          article.ID,
			Title:       article.Title,
			URL:         article.URL,
			Blog:        article.BlogName,
			Read:        article.IsRead,
			HasNote:     article.HasNote,
			IsFavorited: article.IsFavorited,
			HNURL:       article.HNURL,
			HNStatus:    string(article.HNStatus),
		}
```

- [ ] **Step 2: Update simple output**

In `internal/cli/output/simple.go`, add favorite indicator:

```go
		// 状态（中文）
		status := "[未读]"
		if article.IsRead {
			status = "[已读]"
		}

		// 收藏标记
		favMark := ""
		if article.IsFavorited {
			favMark = " ★"
		}
```

Update the line format:

```go
		line := fmt.Sprintf("%s %s%s (%s) - %s",
			status,
			article.Title,
			favMark,
			article.BlogName,
			published)
```

- [ ] **Step 3: Update table output**

In `internal/cli/output/table.go`, add a Fav column. Update column widths:

```go
	idWidth := 8
	titleWidth := 50
	blogWidth := 20
	statusWidth := 8
	favWidth := 5
	timeWidth := 20
```

Update header:

```go
	header := fmt.Sprintf("| %-*s | %-*s | %-*s | %-*s | %-*s | %-*s |",
		idWidth, "ID",
		titleWidth, "Title",
		blogWidth, "Blog",
		statusWidth, "Status",
		favWidth, "Fav",
		timeWidth, "Published")
```

Update separator:

```go
	separator := fmt.Sprintf("|-%s-|-%s-|-%s-|-%s-|-%s-|-%s-|",
		strings.Repeat("-", idWidth),
		strings.Repeat("-", titleWidth),
		strings.Repeat("-", blogWidth),
		strings.Repeat("-", statusWidth),
		strings.Repeat("-", favWidth),
		strings.Repeat("-", timeWidth))
```

Update row:

```go
		fav := ""
		if article.IsFavorited {
			fav = "★"
		}

		row := fmt.Sprintf("| %-*d | %-*s | %-*s | %-*s | %-*s | %-*s |",
			idWidth, article.ID,
			titleWidth, title,
			blogWidth, blogName,
			statusWidth, status,
			favWidth, fav,
			timeWidth, published)
```

- [ ] **Step 4: Verify compilation**

Run: `go build ./...`
Expected: Compiles successfully

- [ ] **Step 5: Commit**

```bash
git add internal/cli/output/
git commit -m "feat(cli-output): add is_favorited to JSON, simple, and table output formats"
```

---

### Task 9: UI — Add star button to article cards

**Files:**
- Modify: `assets/templates/partials/article-items.gohtml`
- Modify: `assets/templates/partials/article-list.gohtml`

- [ ] **Step 1: Guard results-info div and add star button to article-items.gohtml**

In `article-items.gohtml`, wrap the `results-info` div with a condition so it only renders during list loads (not single-card re-renders):

```html
{{if .DisplayedCount}}
<div id="results-info" hx-swap-oob="true" class="results-info">
    Showing {{.DisplayedCount}} of {{.ArticleCount}} article{{if ne .ArticleCount 1}}s{{end}}
</div>
{{end}}
```

Then add the favorite button inside `<div class="article-actions">`, **before** the HasNote block:

```html
    <div class="article-actions">
        {{if .IsFavorited}}
        <button class="action-btn action-btn-favorite active"
                hx-post="/articles/{{.ID}}/unfavorite"
                hx-target="#article-{{.ID}}"
                hx-swap="outerHTML swap:300ms"
                title="取消收藏"
                onclick="event.stopPropagation();">
            <svg viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="1" style="width:16px;height:16px;flex-shrink:0;"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
        </button>
        {{else}}
        <button class="action-btn action-btn-favorite"
                hx-post="/articles/{{.ID}}/favorite"
                hx-target="#article-{{.ID}}"
                hx-swap="outerHTML swap:300ms"
                title="收藏"
                onclick="event.stopPropagation();">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="width:16px;height:16px;flex-shrink:0;"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>
        </button>
        {{end}}
        {{if .HasNote}}
```

- [ ] **Step 2: Add the same star button to article-list.gohtml**

In `article-list.gohtml`, find the `<div class="article-actions">` block (around line 171) and add the same favorite button block before the `{{if .HasNote}}` line.

- [ ] **Step 3: Update article-list.gohtml header title**

In the `<h1>` tag (line 8), add favorites case:

```html
  <h1>{{if .CurrentBlogName}}{{.CurrentBlogName}}{{else if eq .CurrentFilter "read"}}Archived{{else if eq .CurrentFilter "favorites"}}Favorites{{else}}Inbox{{end}}</h1>
```

- [ ] **Step 4: Commit**

```bash
git add assets/templates/partials/article-items.gohtml assets/templates/partials/article-list.gohtml
git commit -m "feat(ui): add star button to article cards and favorites header title"
```

---

### Task 10: UI — Add sidebar navigation entry

**Files:**
- Modify: `assets/templates/partials/sidebar.gohtml`

- [ ] **Step 1: Add Favorites nav link**

In `sidebar.gohtml`, add the Favorites link between Inbox and Archived (between the Inbox `<a>` tag ending at line 36 and the Archived `<a>` tag starting at line 37):

```html
        <a href="/articles?filter=favorites"
           hx-get="/articles?filter=favorites"
           hx-target="#main-content"
           hx-push-url="true"
           hx-on:click="document.querySelectorAll('.sidebar-nav .nav-link, .blog-item').forEach(el => el.classList.remove('active')); this.classList.add('active'); document.getElementById('sidebar-toggle').checked = false;"
           class="nav-link{{if eq .CurrentFilter "favorites"}} active{{end}}">
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"></polygon>
            </svg>
            <span>Favorites</span>
        </a>
```

- [ ] **Step 2: Commit**

```bash
git add assets/templates/partials/sidebar.gohtml
git commit -m "feat(ui): add Favorites navigation entry in sidebar"
```

---

### Task 11: UI — Add favorite button CSS styles

**Files:**
- Modify: `assets/static/styles.css`

- [ ] **Step 1: Add favorite button styles**

Add these CSS rules after the existing `.action-btn-muted` styles (around line 2141):

```css
/* Favorite button styles */
.action-btn-favorite {
  color: var(--text-secondary);
  transition: color 0.2s ease;
}

.action-btn-favorite:hover {
  color: #f59e0b;
  background-color: transparent;
  border-color: #f59e0b;
}

.action-btn-favorite.active {
  color: #f59e0b;
  border-color: #f59e0b;
}

.action-btn-favorite.active:hover {
  color: var(--text-secondary);
  border-color: var(--border);
}
```

- [ ] **Step 2: Commit**

```bash
git add assets/static/styles.css
git commit -m "feat(ui): add favorite button CSS styles with gold highlight"
```

---

### Task 12: Final verification — Build, test, deploy

**Files:** None (verification only)

- [ ] **Step 1: Run all tests**

Run: `go test ./...`
Expected: All tests pass

- [ ] **Step 2: Build and install CLI**

Run: `go install ./cmd/blogwatcher`
Expected: Builds successfully

- [ ] **Step 3: Rebuild Docker**

Run: `docker compose build blogwatcher-ui --no-cache && docker compose up -d blogwatcher-ui`
Expected: Container starts successfully

- [ ] **Step 4: Manual verification checklist**

- [ ] Open Web UI → verify sidebar shows "Favorites" entry between Inbox and Archived
- [ ] Click an article's star icon → verify it fills with gold color
- [ ] Click the star again → verify it returns to outline
- [ ] Click "Favorites" in sidebar → verify only favorited articles show
- [ ] Run `blogwatcher article favorite <id>` → verify CLI output "已收藏: <title>"
- [ ] Run `blogwatcher article unfavorite <id>` → verify CLI output "已取消收藏: <title>"
- [ ] Run `blogwatcher article list --favorited` → verify only favorited articles show
- [ ] Run `blogwatcher article list --format json` → verify `is_favorited` field present

- [ ] **Step 5: Final commit**

```bash
git add -A
git status
# If there are any uncommitted changes, commit them
```
