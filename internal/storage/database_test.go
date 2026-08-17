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

func TestCreateCategory(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	category, err := db.CreateCategory("Tech")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	if category.ID == 0 {
		t.Fatal("expected category ID")
	}

	if category.Name != "Tech" {
		t.Errorf("Name = %q, want %q", category.Name, "Tech")
	}

	// Verify CreatedAt is recent (within 1 minute)
	if time.Since(category.CreatedAt) > time.Minute {
		t.Errorf("CreatedAt = %v, want recent time", category.CreatedAt)
	}
}

func TestCreateCategoryEmptyName(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	_, err := db.CreateCategory("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestListCategoriesWithBlogCount(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	// Create categories
	_, err := db.CreateCategory("News")
	if err != nil {
		t.Fatalf("create category News: %v", err)
	}

	techCat, err := db.CreateCategory("Tech")
	if err != nil {
		t.Fatalf("create category Tech: %v", err)
	}

	// Create blogs and assign to Tech category
	blog1, err := db.AddBlog(model.Blog{Name: "Blog 1", URL: "https://1.example.com"})
	if err != nil {
		t.Fatalf("create blog 1: %v", err)
	}
	blog2, err := db.AddBlog(model.Blog{Name: "Blog 2", URL: "https://2.example.com"})
	if err != nil {
		t.Fatalf("create blog 2: %v", err)
	}

	// Assign blogs to Tech category
	if err := db.UpdateBlogCategory(blog1.ID, &techCat.ID); err != nil {
		t.Fatalf("assign blog 1 to Tech: %v", err)
	}
	if err := db.UpdateBlogCategory(blog2.ID, &techCat.ID); err != nil {
		t.Fatalf("assign blog 2 to Tech: %v", err)
	}

	// List categories with blog count
	categories, err := db.ListCategoriesWithBlogCount()
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}

	if len(categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(categories))
	}

	// Should be ordered by name (News, Tech)
	if categories[0].Name != "News" {
		t.Errorf("first category = %q, want %q", categories[0].Name, "News")
	}
	if categories[1].Name != "Tech" {
		t.Errorf("second category = %q, want %q", categories[1].Name, "Tech")
	}

	// Verify blog counts
	if categories[0].BlogCount != 0 {
		t.Errorf("News BlogCount = %d, want 0", categories[0].BlogCount)
	}
	if categories[1].BlogCount != 2 {
		t.Errorf("Tech BlogCount = %d, want 2", categories[1].BlogCount)
	}
}

func TestUpdateCategoryName(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	category, err := db.CreateCategory("Tech")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	// Update name
	err = db.UpdateCategoryName(category.ID, "Technology")
	if err != nil {
		t.Fatalf("update category name: %v", err)
	}

	// Verify update
	categories, err := db.ListCategoriesWithBlogCount()
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}

	if len(categories) != 1 {
		t.Fatalf("expected 1 category, got %d", len(categories))
	}

	if categories[0].Name != "Technology" {
		t.Errorf("Name = %q, want %q", categories[0].Name, "Technology")
	}
}

func TestUpdateCategoryNameEmpty(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	category, err := db.CreateCategory("Tech")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	err = db.UpdateCategoryName(category.ID, "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestUpdateCategoryNameNotFound(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	err := db.UpdateCategoryName(999, "NewName")
	if err == nil {
		t.Fatal("expected error for non-existent category")
	}
}

func TestDeleteCategory(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	// Create category
	category, err := db.CreateCategory("Tech")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	// Create blogs and assign to category
	blog1, err := db.AddBlog(model.Blog{Name: "Blog 1", URL: "https://1.example.com"})
	if err != nil {
		t.Fatalf("create blog 1: %v", err)
	}
	blog2, err := db.AddBlog(model.Blog{Name: "Blog 2", URL: "https://2.example.com"})
	if err != nil {
		t.Fatalf("create blog 2: %v", err)
	}

	if err := db.UpdateBlogCategory(blog1.ID, &category.ID); err != nil {
		t.Fatalf("assign blog 1: %v", err)
	}
	if err := db.UpdateBlogCategory(blog2.ID, &category.ID); err != nil {
		t.Fatalf("assign blog 2: %v", err)
	}

	// Delete category
	err = db.DeleteCategory(category.ID)
	if err != nil {
		t.Fatalf("delete category: %v", err)
	}

	// Verify category deleted
	categories, err := db.ListCategoriesWithBlogCount()
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}
	if len(categories) != 0 {
		t.Fatalf("expected 0 categories after delete, got %d", len(categories))
	}

	// Verify blog.category_id set to NULL
	fetchedBlog1, err := db.GetBlogByID(blog1.ID)
	if err != nil {
		t.Fatalf("get blog 1: %v", err)
	}
	if fetchedBlog1.CategoryID != nil {
		t.Errorf("blog 1 CategoryID = %d, want nil", *fetchedBlog1.CategoryID)
	}

	fetchedBlog2, err := db.GetBlogByID(blog2.ID)
	if err != nil {
		t.Fatalf("get blog 2: %v", err)
	}
	if fetchedBlog2.CategoryID != nil {
		t.Errorf("blog 2 CategoryID = %d, want nil", *fetchedBlog2.CategoryID)
	}
}

func TestDeleteCategoryNotFound(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	err := db.DeleteCategory(999)
	if err == nil {
		t.Fatal("expected error for non-existent category")
	}
}

func TestUpdateBlogCategory(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	// Create category
	category, err := db.CreateCategory("Tech")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	// Create blog (uncategorized)
	blog, err := db.AddBlog(model.Blog{Name: "Blog", URL: "https://example.com"})
	if err != nil {
		t.Fatalf("create blog: %v", err)
	}

	// Verify initially uncategorized
	fetched, err := db.GetBlogByID(blog.ID)
	if err != nil {
		t.Fatalf("get blog: %v", err)
	}
	if fetched.CategoryID != nil {
		t.Fatalf("initial CategoryID = %d, want nil", *fetched.CategoryID)
	}

	// Assign to category
	err = db.UpdateBlogCategory(blog.ID, &category.ID)
	if err != nil {
		t.Fatalf("assign to category: %v", err)
	}

	// Verify category assigned
	fetched, err = db.GetBlogByID(blog.ID)
	if err != nil {
		t.Fatalf("get blog: %v", err)
	}
	if fetched.CategoryID == nil || *fetched.CategoryID != category.ID {
		t.Errorf("CategoryID = %v, want %d", fetched.CategoryID, category.ID)
	}

	// Set back to uncategorized (nil)
	err = db.UpdateBlogCategory(blog.ID, nil)
	if err != nil {
		t.Fatalf("set uncategorized: %v", err)
	}

	// Verify uncategorized again
	fetched, err = db.GetBlogByID(blog.ID)
	if err != nil {
		t.Fatalf("get blog: %v", err)
	}
	if fetched.CategoryID != nil {
		t.Errorf("CategoryID = %d, want nil", *fetched.CategoryID)
	}
}

func TestUpdateBlogCategoryNotFound(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	category, err := db.CreateCategory("Tech")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	err = db.UpdateBlogCategory(999, &category.ID)
	if err == nil {
		t.Fatal("expected error for non-existent blog")
	}
}

func TestSanitizeFTS5Query(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "simple word no special chars",
			input: "anthropic",
			want:  "anthropic",
		},
		{
			name:  "hyphenated phrase",
			input: "the-anthropic-economic-index",
			want:  `"the-anthropic-economic-index"`,
		},
		{
			name:  "phrase with space",
			input: "anthropic economic",
			want:  `"anthropic economic"`,
		},
		{
			name:  "phrase with asterisk",
			input: "anthropic*",
			want:  `"anthropic*"`,
		},
		{
			name:  "phrase with parentheses",
			input: "(anthropic)",
			want:  `"(anthropic)"`,
		},
		{
			name:  "contains double quote",
			input: `test"query`,
			want:  `"test""query"`,
		},
		{
			name:  "phrase with colon",
			input: "title:anthropic",
			want:  `"title:anthropic"`,
		},
		{
			name:  "phrase with caret",
			input: "anthropic^5",
			want:  `"anthropic^5"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeFTS5Query(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeFTS5Query(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

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

func TestTagsTablesMigration(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	for _, tbl := range []string{"tags", "article_tags"} {
		if !db.tableExists(tbl) {
			t.Fatalf("expected table %s to exist after migration", tbl)
		}
	}
}

func TestCreateTag(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	tag, err := db.CreateTag("Go")
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}
	if tag.Name != "Go" || tag.ID == 0 {
		t.Fatalf("unexpected tag: %+v", tag)
	}

	// 重名幂等：返回已存在的标签
	dup, err := db.CreateTag("Go")
	if err != nil {
		t.Fatalf("create duplicate tag: %v", err)
	}
	if dup.ID != tag.ID {
		t.Fatalf("expected same ID for duplicate, got %d vs %d", dup.ID, tag.ID)
	}
}

func TestCreateTag_EmptyName(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	if _, err := db.CreateTag(""); err == nil {
		t.Fatal("expected error for empty tag name")
	}
}

func TestGetTagByName(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	created, _ := db.CreateTag("数据库")
	got, err := db.GetTagByName("数据库")
	if err != nil || got == nil || got.ID != created.ID {
		t.Fatalf("get by name: got=%+v err=%v", got, err)
	}
	missing, err := db.GetTagByName("不存在")
	if err != nil || missing != nil {
		t.Fatalf("expected nil for missing tag, got=%+v err=%v", missing, err)
	}
}

func TestListTags_WithCount(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	// 空列表
	tags, err := db.ListTags()
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	if len(tags) != 0 {
		t.Fatalf("expected empty, got %d", len(tags))
	}

	db.CreateTag("Go")
	db.CreateTag("DB")
	tags, err = db.ListTags()
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2, got %d", len(tags))
	}
	// 未绑文章，计数应为 0
	for _, tg := range tags {
		if tg.ArticleCount != 0 {
			t.Fatalf("expected 0 count, got %d for %s", tg.ArticleCount, tg.Name)
		}
	}
}

func TestRenameTag(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	tag, _ := db.CreateTag("old")
	if err := db.RenameTag(tag.ID, "new"); err != nil {
		t.Fatalf("rename: %v", err)
	}
	got, _ := db.GetTagByID(tag.ID)
	if got.Name != "new" {
		t.Fatalf("expected name 'new', got %q", got.Name)
	}
}

func TestRenameTag_Conflict(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	a, _ := db.CreateTag("a")
	db.CreateTag("b")
	if err := db.RenameTag(a.ID, "b"); err == nil {
		t.Fatal("expected conflict error when renaming to existing name")
	}
}

func TestDeleteTag(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	tag, _ := db.CreateTag("solo")
	affected, err := db.DeleteTag(tag.ID)
	if err != nil || affected != 0 {
		t.Fatalf("delete tag: affected=%d err=%v", affected, err)
	}
	if got, _ := db.GetTagByID(tag.ID); got != nil {
		t.Fatal("tag should be deleted")
	}
	// 删不存在
	if _, err := db.DeleteTag(99999); err == nil {
		t.Fatal("expected error deleting missing tag")
	}
}

func TestAddArticleTag_Idempotent(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	blog, _ := db.AddBlog(model.Blog{Name: "B", URL: "https://example.com"})
	db.AddArticlesBulk([]model.Article{{BlogID: blog.ID, Title: "A", URL: "https://example.com/idem", HNStatus: model.HNStatusNotSearch}})
	arts, _ := db.ListArticles(false, nil)
	articleID := arts[0].ID
	tag, _ := db.CreateTag("Go")

	if err := db.AddArticleTag(articleID, tag.ID); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := db.AddArticleTag(articleID, tag.ID); err != nil {
		t.Fatalf("add duplicate: %v", err)
	}
	tags, _ := db.GetArticleTags(articleID)
	if len(tags) != 1 {
		t.Fatalf("expected 1 tag after idempotent add, got %d", len(tags))
	}
}

func TestRemoveArticleTag(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	blog, _ := db.AddBlog(model.Blog{Name: "B", URL: "https://example.com"})
	db.AddArticlesBulk([]model.Article{{BlogID: blog.ID, Title: "A", URL: "https://example.com/rm", HNStatus: model.HNStatusNotSearch}})
	arts, _ := db.ListArticles(false, nil)
	articleID := arts[0].ID
	tag, _ := db.CreateTag("Go")
	db.AddArticleTag(articleID, tag.ID)

	if err := db.RemoveArticleTag(articleID, tag.ID); err != nil {
		t.Fatalf("remove: %v", err)
	}
	tags, _ := db.GetArticleTags(articleID)
	if len(tags) != 0 {
		t.Fatalf("expected 0 tags after remove, got %d", len(tags))
	}
	if err := db.RemoveArticleTag(articleID, tag.ID); err != nil {
		t.Fatalf("remove absent: %v", err)
	}
}

func TestSetArticleTags_FullReplace(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	blog, _ := db.AddBlog(model.Blog{Name: "B", URL: "https://example.com"})
	db.AddArticlesBulk([]model.Article{{BlogID: blog.ID, Title: "A", URL: "https://example.com/set", HNStatus: model.HNStatusNotSearch}})
	arts, _ := db.ListArticles(false, nil)
	articleID := arts[0].ID
	t1, _ := db.CreateTag("a")
	t2, _ := db.CreateTag("b")
	t3, _ := db.CreateTag("c")

	if err := db.SetArticleTags(articleID, []int64{t1.ID, t2.ID}); err != nil {
		t.Fatalf("set1: %v", err)
	}
	tags, _ := db.GetArticleTags(articleID)
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
	if err := db.SetArticleTags(articleID, []int64{t2.ID, t3.ID}); err != nil {
		t.Fatalf("set2: %v", err)
	}
	names := map[string]bool{}
	for _, tg := range mustGetArticleTags(t, db, articleID) {
		names[tg.Name] = true
	}
	if !names["b"] || !names["c"] || names["a"] {
		t.Fatalf("unexpected tags after replace: %+v", names)
	}
	if err := db.SetArticleTags(articleID, nil); err != nil {
		t.Fatalf("set3: %v", err)
	}
	if tags, _ := db.GetArticleTags(articleID); len(tags) != 0 {
		t.Fatalf("expected 0 after clear, got %d", len(tags))
	}
}

func TestGetTagsForArticles_Batch(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	blog, _ := db.AddBlog(model.Blog{Name: "B", URL: "https://example.com"})
	db.AddArticlesBulk([]model.Article{
		{BlogID: blog.ID, Title: "A1", URL: "https://example.com/b1", HNStatus: model.HNStatusNotSearch},
		{BlogID: blog.ID, Title: "A2", URL: "https://example.com/b2", HNStatus: model.HNStatusNotSearch},
	})
	arts, _ := db.ListArticles(false, nil)
	id1, id2 := arts[0].ID, arts[1].ID
	t1, _ := db.CreateTag("Go")
	t2, _ := db.CreateTag("DB")
	db.AddArticleTag(id1, t1.ID)
	db.AddArticleTag(id2, t2.ID)
	db.AddArticleTag(id2, t1.ID)

	m, err := db.GetTagsForArticles([]int64{id1, id2})
	if err != nil {
		t.Fatalf("batch: %v", err)
	}
	if len(m[id1]) != 1 || m[id1][0].Name != "Go" {
		t.Fatalf("id1 tags wrong: %+v", m[id1])
	}
	if len(m[id2]) != 2 {
		t.Fatalf("id2 expected 2 tags, got %d", len(m[id2]))
	}
}

func TestDeleteTag_CascadesAssociations(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	blog, _ := db.AddBlog(model.Blog{Name: "B", URL: "https://example.com"})
	db.AddArticlesBulk([]model.Article{{BlogID: blog.ID, Title: "A", URL: "https://example.com/ca", HNStatus: model.HNStatusNotSearch}})
	arts, _ := db.ListArticles(false, nil)
	articleID := arts[0].ID

	tag, _ := db.CreateTag("tmp")
	db.AddArticleTag(articleID, tag.ID)

	tags, _ := db.GetArticleTags(articleID)
	if len(tags) != 1 {
		t.Fatalf("expected 1 tag before delete, got %d", len(tags))
	}

	affected, err := db.DeleteTag(tag.ID)
	if err != nil || affected == 0 {
		t.Fatalf("delete tag: affected=%d err=%v", affected, err)
	}

	tags, _ = db.GetArticleTags(articleID)
	if len(tags) != 0 {
		t.Fatalf("expected 0 tags after delete, got %d", len(tags))
	}
	if art, _ := db.GetArticleByID(articleID); art == nil {
		t.Fatal("article should still exist after tag deletion")
	}
}

// mustGetArticleTags 测试辅助：断言获取成功并返回标签。
func mustGetArticleTags(t *testing.T, db *Database, articleID int64) []model.Tag {
	t.Helper()
	tags, err := db.GetArticleTags(articleID)
	if err != nil {
		t.Fatalf("get tags: %v", err)
	}
	return tags
}
