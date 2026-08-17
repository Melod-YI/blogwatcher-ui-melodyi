// ABOUTME: Tests for HTTP handler functions.
// ABOUTME: Covers blog addition, validation, and error handling via HTTP endpoints.
package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/esttorhe/blogwatcher-ui/v2/assets"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
)

func TestHandleAddBlogSuccess(t *testing.T) {
	srv := createTestServer(t)

	form := url.Values{}
	form.Set("name", "Test Blog")
	form.Set("url", "https://example.com")

	req := httptest.NewRequest(http.MethodPost, "/blogs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Check response contains success indicator
	body := rec.Body.String()
	if !strings.Contains(body, "Test Blog") {
		t.Errorf("response should contain blog name, got: %s", body)
	}
}

func TestHandleAddBlogDuplicateName(t *testing.T) {
	srv := createTestServer(t)

	// Add first blog
	form := url.Values{}
	form.Set("name", "Duplicate")
	form.Set("url", "https://first.example.com")

	req := httptest.NewRequest(http.MethodPost, "/blogs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("first add: status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Try to add blog with same name
	form = url.Values{}
	form.Set("name", "Duplicate")
	form.Set("url", "https://second.example.com")

	req = httptest.NewRequest(http.MethodPost, "/blogs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("duplicate add: status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Should contain error message about duplicate
	body := rec.Body.String()
	if !strings.Contains(body, "already exists") {
		t.Errorf("response should contain 'already exists', got: %s", body)
	}
}

func TestHandleAddBlogDuplicateURL(t *testing.T) {
	srv := createTestServer(t)

	// Add first blog
	form := url.Values{}
	form.Set("name", "First Blog")
	form.Set("url", "https://example.com")

	req := httptest.NewRequest(http.MethodPost, "/blogs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("first add: status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Try to add blog with same URL
	form = url.Values{}
	form.Set("name", "Second Blog")
	form.Set("url", "https://example.com")

	req = httptest.NewRequest(http.MethodPost, "/blogs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("duplicate add: status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Should contain error message about duplicate
	body := rec.Body.String()
	if !strings.Contains(body, "already exists") {
		t.Errorf("response should contain 'already exists', got: %s", body)
	}
}

func TestHandleAddBlogValidationEmptyName(t *testing.T) {
	srv := createTestServer(t)

	form := url.Values{}
	form.Set("name", "")
	form.Set("url", "https://example.com")

	req := httptest.NewRequest(http.MethodPost, "/blogs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "required") {
		t.Errorf("response should contain 'required', got: %s", body)
	}
}

func TestHandleAddBlogValidationEmptyURL(t *testing.T) {
	srv := createTestServer(t)

	form := url.Values{}
	form.Set("name", "Test Blog")
	form.Set("url", "")

	req := httptest.NewRequest(http.MethodPost, "/blogs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "required") {
		t.Errorf("response should contain 'required', got: %s", body)
	}
}

func TestHandleAddBlogValidationBothEmpty(t *testing.T) {
	srv := createTestServer(t)

	form := url.Values{}
	form.Set("name", "   ") // Whitespace only
	form.Set("url", "   ")  // Whitespace only

	req := httptest.NewRequest(http.MethodPost, "/blogs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "required") {
		t.Errorf("response should contain 'required', got: %s", body)
	}
}

func TestArticleListHeaderShowsBlogName(t *testing.T) {
	srv := createTestServer(t)

	// Add a blog first
	form := url.Values{}
	form.Set("name", "My Cool Blog")
	form.Set("url", "https://coolblog.example.com")

	req := httptest.NewRequest(http.MethodPost, "/blogs/add", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("add blog: status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Request articles filtered by blog=1 via HTMX
	req = httptest.NewRequest(http.MethodGet, "/articles?blog=1", nil)
	req.Header.Set("HX-Request", "true")
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("articles: status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "My Cool Blog") {
		t.Errorf("header should contain blog name 'My Cool Blog', got: %s", body)
	}
}

func TestArticleListHeaderShowsInboxWithoutBlogFilter(t *testing.T) {
	srv := createTestServer(t)

	// Request articles without blog filter via HTMX
	req := httptest.NewRequest(http.MethodGet, "/articles?filter=unread", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("articles: status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Inbox") {
		t.Errorf("header should contain 'Inbox' when no blog filter, got: %s", body)
	}
}

func TestHandleAPISync_Success(t *testing.T) {
	srv := createTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/sync", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Verify Content-Type is JSON
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	// Verify response is valid JSON with expected fields
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	// Check required top-level fields exist
	for _, field := range []string{"blogs_scanned", "new_articles"} {
		if _, ok := resp[field]; !ok {
			t.Errorf("response missing field %q", field)
		}
	}
}

func TestHandleAPISync_GetDoesNotReturnJSON(t *testing.T) {
	srv := createTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/sync", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	// GET /api/sync should not hit the API handler (falls through to index)
	ct := rec.Header().Get("Content-Type")
	if ct == "application/json" {
		t.Errorf("GET /api/sync should not return JSON, got Content-Type = %q", ct)
	}
}

func TestHandleFavoriteAndUnfavorite(t *testing.T) {
	srv := createTestServer(t)
	db := srv.(*Server).db

	// Create blog and article
	blog, err := db.AddBlog(model.Blog{Name: "Test Blog", URL: "https://example.com"})
	if err != nil {
		t.Fatalf("add blog: %v", err)
	}
	inserted, _, err := db.AddArticlesBulk([]model.Article{
		{BlogID: blog.ID, Title: "Test Article", URL: "https://example.com/test-fav", HNStatus: model.HNStatusNotSearch},
	})
	if err != nil || inserted == 0 {
		t.Fatalf("insert article: inserted=%d, err=%v", inserted, err)
	}
	articles, err := db.ListArticles(false, nil)
	if err != nil || len(articles) == 0 {
		t.Fatalf("list articles: err=%v, len=%d", err, len(articles))
	}
	articleID := articles[0].ID

	// Test favorite
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/articles/%d/favorite", articleID), nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("favorite: status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	// Verify in DB
	article, err := db.GetArticleByID(articleID)
	if err != nil {
		t.Fatalf("get article after favorite: %v", err)
	}
	if !article.IsFavorited {
		t.Error("expected article to be favorited after favorite endpoint")
	}

	// Test unfavorite
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/articles/%d/unfavorite", articleID), nil)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("unfavorite: status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	// Verify in DB
	article, err = db.GetArticleByID(articleID)
	if err != nil {
		t.Fatalf("get article after unfavorite: %v", err)
	}
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
		t.Errorf("status = %d, want %d, body: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestHandleFavoriteNotFound(t *testing.T) {
	srv := createTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/articles/99999/favorite", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d, body: %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestHandleMarkReadFromFavoritesReturnsCard(t *testing.T) {
	srv := createTestServer(t)
	db := srv.(*Server).db

	// Create blog and article
	blog, err := db.AddBlog(model.Blog{Name: "Test Blog", URL: "https://example.com"})
	if err != nil {
		t.Fatalf("add blog: %v", err)
	}
	inserted, _, err := db.AddArticlesBulk([]model.Article{
		{BlogID: blog.ID, Title: "Fav Article", URL: "https://example.com/fav-read", HNStatus: model.HNStatusNotSearch},
	})
	if err != nil || inserted == 0 {
		t.Fatalf("insert article: inserted=%d, err=%v", inserted, err)
	}
	articles, err := db.ListArticles(false, nil)
	if err != nil || len(articles) == 0 {
		t.Fatalf("list articles: err=%v, len=%d", err, len(articles))
	}
	articleID := articles[0].ID

	// Favorite the article first
	if err := db.FavoriteArticle(articleID); err != nil {
		t.Fatalf("favorite article: %v", err)
	}

	// Mark as read from favorites page
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/articles/%d/read", articleID), nil)
	req.Header.Set("Referer", "http://localhost:8080/articles?filter=favorites")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	// Should return the article card (not empty)
	body := rec.Body.String()
	if !strings.Contains(body, "Fav Article") {
		t.Errorf("expected article card in response on favorites page, got empty or wrong body: %s", body)
	}
}

func TestHandleMarkReadFromUnreadReturnsEmpty(t *testing.T) {
	srv := createTestServer(t)
	db := srv.(*Server).db

	// Create blog and article
	blog, err := db.AddBlog(model.Blog{Name: "Test Blog", URL: "https://example.com"})
	if err != nil {
		t.Fatalf("add blog: %v", err)
	}
	inserted, _, err := db.AddArticlesBulk([]model.Article{
		{BlogID: blog.ID, Title: "Unread Article", URL: "https://example.com/unread-test", HNStatus: model.HNStatusNotSearch},
	})
	if err != nil || inserted == 0 {
		t.Fatalf("insert article: inserted=%d, err=%v", inserted, err)
	}
	articles, err := db.ListArticles(false, nil)
	if err != nil || len(articles) == 0 {
		t.Fatalf("list articles: err=%v, len=%d", err, len(articles))
	}
	articleID := articles[0].ID

	// Mark as read from unread page (no favorites filter in Referer)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/articles/%d/read", articleID), nil)
	req.Header.Set("Referer", "http://localhost:8080/articles?filter=unread")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Should return empty body so HTMX removes the card
	body := rec.Body.String()
	if body != "" {
		t.Errorf("expected empty body on unread page, got: %s", body)
	}
}

func TestHandleMarkUnreadFromFavoritesReturnsCard(t *testing.T) {
	srv := createTestServer(t)
	db := srv.(*Server).db

	// Create blog and article
	blog, err := db.AddBlog(model.Blog{Name: "Test Blog", URL: "https://example.com"})
	if err != nil {
		t.Fatalf("add blog: %v", err)
	}
	inserted, _, err := db.AddArticlesBulk([]model.Article{
		{BlogID: blog.ID, Title: "Read Fav Article", URL: "https://example.com/read-fav", HNStatus: model.HNStatusNotSearch},
	})
	if err != nil || inserted == 0 {
		t.Fatalf("insert article: inserted=%d, err=%v", inserted, err)
	}
	articles, err := db.ListArticles(false, nil)
	if err != nil || len(articles) == 0 {
		t.Fatalf("list articles: err=%v, len=%d", err, len(articles))
	}
	articleID := articles[0].ID

	// Favorite and mark as read
	if err := db.FavoriteArticle(articleID); err != nil {
		t.Fatalf("favorite article: %v", err)
	}
	if _, err := db.MarkArticleRead(articleID); err != nil {
		t.Fatalf("mark read: %v", err)
	}

	// Mark as unread from favorites page
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/articles/%d/unread", articleID), nil)
	req.Header.Set("Referer", "http://localhost:8080/articles?filter=favorites")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	// Should return the article card (not empty)
	body := rec.Body.String()
	if !strings.Contains(body, "Read Fav Article") {
		t.Errorf("expected article card in response on favorites page, got: %s", body)
	}
}

func createTestServer(t *testing.T) http.Handler {
	t.Helper()

	// Create temp database
	path := filepath.Join(t.TempDir(), "blogwatcher.db")
	db, err := storage.OpenDatabase(path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Extract embedded filesystems
	staticFiles, err := fs.Sub(assets.StaticFS, "static")
	if err != nil {
		t.Fatalf("extract static: %v", err)
	}
	templateFiles, err := fs.Sub(assets.TemplateFS, "templates")
	if err != nil {
		t.Fatalf("extract templates: %v", err)
	}

	// Create server
	srv, err := NewServerWithFS(db, templateFiles, staticFiles, "test")
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	return srv
}

func TestParseSearchOptions_TagFilter(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/articles?filter=tag&tag=Go", nil)
	opts, filter, _ := parseSearchOptions(req)
	if filter != "tag" {
		t.Fatalf("filter = %q, want 'tag'", filter)
	}
	if opts.TagName != "Go" {
		t.Fatalf("TagName = %q, want 'Go'", opts.TagName)
	}
	// tag 筛选不应强制 IsRead
	if opts.IsRead != nil {
		t.Fatalf("IsRead should be nil for tag filter, got %v", *opts.IsRead)
	}
}

func TestHandleTagCreate(t *testing.T) {
	srv := createTestServer(t)
	body := strings.NewReader("name=Go")
	req := httptest.NewRequest(http.MethodPost, "/tags", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("create: status=%d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("HX-Trigger"); got != "articleListUpdated" {
		t.Fatalf("HX-Trigger = %q, want articleListUpdated", got)
	}
	db := srv.(*Server).db
	tag, err := db.GetTagByName("Go")
	if err != nil || tag == nil {
		t.Fatalf("tag not persisted: %v", err)
	}
}

func TestHandleTagRename(t *testing.T) {
	srv := createTestServer(t)
	db := srv.(*Server).db
	tag, _ := db.CreateTag("old")
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/tags/%d", tag.ID), strings.NewReader("name=new"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("rename: status=%d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("HX-Trigger"); got != "articleListUpdated" {
		t.Fatalf("HX-Trigger = %q, want articleListUpdated", got)
	}
	got, _ := db.GetTagByID(tag.ID)
	if got.Name != "new" {
		t.Fatalf("expected 'new', got %q", got.Name)
	}
}

func TestHandleTagRename_Conflict(t *testing.T) {
	srv := createTestServer(t)
	db := srv.(*Server).db
	a, _ := db.CreateTag("a")
	db.CreateTag("b")
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/tags/%d", a.ID), strings.NewReader("name=b"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestHandleTagDelete(t *testing.T) {
	srv := createTestServer(t)
	db := srv.(*Server).db
	tag, _ := db.CreateTag("tmp")
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/tags/%d", tag.ID), nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete: status=%d", rec.Code)
	}
	if got := rec.Header().Get("HX-Trigger"); got != "articleListUpdated" {
		t.Fatalf("HX-Trigger = %q, want articleListUpdated", got)
	}
	if got, _ := db.GetTagByID(tag.ID); got != nil {
		t.Fatal("tag should be deleted")
	}
}

func TestHandleTagDelete_NotFound(t *testing.T) {
	srv := createTestServer(t)
	req := httptest.NewRequest(http.MethodDelete, "/tags/99999", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandleArticleTagAdd(t *testing.T) {
	srv := createTestServer(t)
	db := srv.(*Server).db
	blog, _ := db.AddBlog(model.Blog{Name: "B", URL: "https://example.com"})
	db.AddArticlesBulk([]model.Article{{BlogID: blog.ID, Title: "A", URL: "https://example.com/hta", HNStatus: model.HNStatusNotSearch}})
	arts, _ := db.ListArticles(false, nil)
	id := arts[0].ID

	body := strings.NewReader("name=Go")
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/articles/%d/tags", id), body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("add: status=%d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("HX-Trigger") == "" {
		t.Fatal("expected HX-Trigger set")
	}
	tags, _ := db.GetArticleTags(id)
	if len(tags) != 1 || tags[0].Name != "Go" {
		t.Fatalf("expected tag Go persisted, got %+v", tags)
	}
}

func TestHandleArticleTagRemove(t *testing.T) {
	srv := createTestServer(t)
	db := srv.(*Server).db
	blog, _ := db.AddBlog(model.Blog{Name: "B", URL: "https://example.com"})
	db.AddArticlesBulk([]model.Article{{BlogID: blog.ID, Title: "A", URL: "https://example.com/htr", HNStatus: model.HNStatusNotSearch}})
	arts, _ := db.ListArticles(false, nil)
	id := arts[0].ID
	tag, _ := db.CreateTag("Go")
	db.AddArticleTag(id, tag.ID)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/articles/%d/tags/%d", id, tag.ID), nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("remove: status=%d", rec.Code)
	}
	tags, _ := db.GetArticleTags(id)
	if len(tags) != 0 {
		t.Fatalf("expected 0 tags, got %d", len(tags))
	}
	if rec.Header().Get("HX-Trigger") == "" {
		t.Fatal("expected HX-Trigger set")
	}
}

func TestHandleArticleTags_NonExistentArticle(t *testing.T) {
	srv := createTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/articles/99999/tags", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandleArticleTagSave(t *testing.T) {
	srv := createTestServer(t)
	db := srv.(*Server).db
	blog, _ := db.AddBlog(model.Blog{Name: "B", URL: "https://example.com"})
	db.AddArticlesBulk([]model.Article{{BlogID: blog.ID, Title: "A", URL: "https://example.com/hts", HNStatus: model.HNStatusNotSearch}})
	arts, _ := db.ListArticles(false, nil)
	id := arts[0].ID
	t1, _ := db.CreateTag("a")
	t2, _ := db.CreateTag("b")

	form := fmt.Sprintf("tag_ids=%d&tag_ids=%d", t1.ID, t2.ID)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/articles/%d/tags/save", id), strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("save: status=%d", rec.Code)
	}
	tags, _ := db.GetArticleTags(id)
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
	if rec.Header().Get("HX-Trigger") == "" {
		t.Fatal("expected HX-Trigger set")
	}
}

func TestParseSearchOptionsSortDefaults(t *testing.T) {
	cases := []struct {
		query string
		want  string
	}{
		{"/articles?filter=favorites", model.SortFavorited},
		{"/articles?filter=read", model.SortRead},
		{"/articles", model.SortPublished},
		{"/articles?filter=unread", model.SortPublished},
		// 显式 sort 覆盖默认
		{"/articles?filter=favorites&sort=published", model.SortPublished},
		{"/articles?filter=read&sort=favorited", model.SortFavorited},
		// 未知 sort 值回退到 filter 默认
		{"/articles?filter=favorites&sort=bogus", model.SortFavorited},
		{"/articles?sort=bogus", model.SortPublished},
	}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodGet, c.query, nil)
		opts, _, _ := parseSearchOptions(req)
		if opts.Sort != c.want {
			t.Errorf("query=%s: Sort=%q want %q", c.query, opts.Sort, c.want)
		}
	}
}

func TestArticleListRendersSortControlOnFavorites(t *testing.T) {
	srv := createTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/articles?filter=favorites", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, `id="sort-favorited"`) {
		t.Error("favorites page should render sort-favorited radio")
	}
	if !strings.Contains(body, `id="sort-published"`) {
		t.Error("favorites page should render sort-published radio")
	}
	if !strings.Contains(body, `name="sort" id="sort-hidden"`) {
		t.Error("page should render hidden sort input")
	}
}

func TestArticleListNoSortControlOnInbox(t *testing.T) {
	srv := createTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/articles?filter=unread", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	body := rec.Body.String()
	if strings.Contains(body, `class="sort-toggle"`) {
		t.Error("inbox page should not render sort control")
	}
}
