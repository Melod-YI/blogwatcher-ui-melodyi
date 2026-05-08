// ABOUTME: Tests for database storage layer operations.
// ABOUTME: Covers schema initialization, blog CRUD, and migration scenarios.
package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
)

func TestOpenDatabaseCreatesDirectoryAndSchema(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "subdir", "blogwatcher.db")

	db, err := OpenDatabase(path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	// Verify file was created
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected db file to exist: %v", err)
	}

	// Verify schema by inserting a blog
	blog, err := db.AddBlog(model.Blog{Name: "Test", URL: "https://example.com"})
	if err != nil {
		t.Fatalf("add blog: %v", err)
	}
	if blog.ID == 0 {
		t.Fatal("expected blog ID")
	}
}

func TestOpenDatabaseWorksWithExistingDB(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "blogwatcher.db")

	// Open and close to create database
	db, err := OpenDatabase(path)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}

	// Add a blog before closing
	blog, err := db.AddBlog(model.Blog{Name: "Test", URL: "https://example.com"})
	if err != nil {
		t.Fatalf("add blog: %v", err)
	}
	db.Close()

	// Re-open should work (idempotent)
	db, err = OpenDatabase(path)
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	defer db.Close()

	// Verify data persisted
	fetched, err := db.GetBlogByID(blog.ID)
	if err != nil {
		t.Fatalf("get blog: %v", err)
	}
	if fetched == nil || fetched.Name != "Test" {
		t.Fatalf("expected blog to persist, got: %+v", fetched)
	}
}

func TestAddBlogAndRetrieval(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	blog, err := db.AddBlog(model.Blog{
		Name:    "Test Blog",
		URL:     "https://test.example.com",
		FeedURL: "https://test.example.com/feed",
	})
	if err != nil {
		t.Fatalf("add blog: %v", err)
	}
	if blog.ID == 0 {
		t.Fatal("expected blog ID")
	}

	// Verify by name
	byName, err := db.GetBlogByName("Test Blog")
	if err != nil {
		t.Fatalf("get by name: %v", err)
	}
	if byName == nil || byName.ID != blog.ID {
		t.Fatalf("expected blog by name, got: %+v", byName)
	}

	// Verify by URL
	byURL, err := db.GetBlogByURL("https://test.example.com")
	if err != nil {
		t.Fatalf("get by url: %v", err)
	}
	if byURL == nil || byURL.ID != blog.ID {
		t.Fatalf("expected blog by url, got: %+v", byURL)
	}

	// Verify by ID
	byID, err := db.GetBlogByID(blog.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if byID == nil || byID.Name != "Test Blog" {
		t.Fatalf("expected blog by id, got: %+v", byID)
	}
}

func TestAddBlogDuplicateURLFails(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	_, err := db.AddBlog(model.Blog{Name: "First", URL: "https://example.com"})
	if err != nil {
		t.Fatalf("first add: %v", err)
	}

	// SQLite UNIQUE constraint should fail on duplicate URL
	_, err = db.AddBlog(model.Blog{Name: "Second", URL: "https://example.com"})
	if err == nil {
		t.Fatal("expected duplicate URL error")
	}
}

func TestGetBlogByURLNotFound(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	blog, err := db.GetBlogByURL("https://nonexistent.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if blog != nil {
		t.Fatalf("expected nil for non-existent URL, got: %+v", blog)
	}
}

func TestGetBlogByNameNotFound(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	blog, err := db.GetBlogByName("NonExistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if blog != nil {
		t.Fatalf("expected nil for non-existent name, got: %+v", blog)
	}
}

func TestAddBlogWithAllFields(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	now := time.Now().UTC().Truncate(time.Nanosecond)
	blog, err := db.AddBlog(model.Blog{
		Name:           "Full Blog",
		URL:            "https://full.example.com",
		FeedURL:        "https://full.example.com/rss",
		ScrapeSelector: "article.content",
		LastScanned:    &now,
	})
	if err != nil {
		t.Fatalf("add blog: %v", err)
	}

	fetched, err := db.GetBlogByID(blog.ID)
	if err != nil {
		t.Fatalf("get blog: %v", err)
	}
	if fetched == nil {
		t.Fatal("expected blog")
	}
	if fetched.FeedURL != "https://full.example.com/rss" {
		t.Errorf("FeedURL = %q, want %q", fetched.FeedURL, "https://full.example.com/rss")
	}
	if fetched.ScrapeSelector != "article.content" {
		t.Errorf("ScrapeSelector = %q, want %q", fetched.ScrapeSelector, "article.content")
	}
	if fetched.LastScanned == nil {
		t.Error("expected LastScanned to be set")
	}
}

func TestAddBlogWithEmptyOptionalFields(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	blog, err := db.AddBlog(model.Blog{
		Name: "Minimal Blog",
		URL:  "https://minimal.example.com",
	})
	if err != nil {
		t.Fatalf("add blog: %v", err)
	}

	fetched, err := db.GetBlogByID(blog.ID)
	if err != nil {
		t.Fatalf("get blog: %v", err)
	}
	if fetched == nil {
		t.Fatal("expected blog")
	}
	if fetched.FeedURL != "" {
		t.Errorf("FeedURL = %q, want empty", fetched.FeedURL)
	}
	if fetched.ScrapeSelector != "" {
		t.Errorf("ScrapeSelector = %q, want empty", fetched.ScrapeSelector)
	}
	if fetched.LastScanned != nil {
		t.Errorf("LastScanned = %v, want nil", fetched.LastScanned)
	}
}

func TestListBlogs(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	// Add multiple blogs
	_, err := db.AddBlog(model.Blog{Name: "Blog A", URL: "https://a.example.com"})
	if err != nil {
		t.Fatalf("add blog A: %v", err)
	}
	_, err = db.AddBlog(model.Blog{Name: "Blog B", URL: "https://b.example.com"})
	if err != nil {
		t.Fatalf("add blog B: %v", err)
	}

	blogs, err := db.ListBlogs()
	if err != nil {
		t.Fatalf("list blogs: %v", err)
	}
	if len(blogs) != 2 {
		t.Fatalf("expected 2 blogs, got %d", len(blogs))
	}

	// Should be ordered by name
	if blogs[0].Name != "Blog A" || blogs[1].Name != "Blog B" {
		t.Errorf("expected blogs ordered by name, got: %v, %v", blogs[0].Name, blogs[1].Name)
	}
}

func TestSchemaIncludesArticlesTable(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	// Add a blog first (required for foreign key)
	blog, err := db.AddBlog(model.Blog{Name: "Test", URL: "https://example.com"})
	if err != nil {
		t.Fatalf("add blog: %v", err)
	}

	// Add articles via bulk insert (tests articles table exists)
	articles := []model.Article{
		{BlogID: blog.ID, Title: "Article 1", URL: "https://example.com/1"},
		{BlogID: blog.ID, Title: "Article 2", URL: "https://example.com/2"},
	}
	inserted, skipped, err := db.AddArticlesBulk(articles)
	if err != nil {
		t.Fatalf("add articles: %v", err)
	}
	if inserted != 2 {
		t.Fatalf("expected 2 articles inserted, got %d", inserted)
	}
	if skipped != 0 {
		t.Fatalf("expected 0 articles skipped, got %d", skipped)
	}

	// Verify articles can be listed
	listed, err := db.ListArticles(false, nil)
	if err != nil {
		t.Fatalf("list articles: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("expected 2 articles, got %d", len(listed))
	}
}

func openTestDB(t *testing.T) *Database {
	t.Helper()
	path := filepath.Join(t.TempDir(), "blogwatcher.db")
	db, err := OpenDatabase(path)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	return db
}
