// ABOUTME: Provides database connection and query methods for reading blog and article data.
// ABOUTME: Supports both reading and writing for scanner operations (sync) and UI interactions.
package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
)

const sqliteTimeLayout = time.RFC3339Nano

func DefaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".blogwatcher", "blogwatcher.db"), nil
}

type Database struct {
	path string
	conn *sql.DB
}

func OpenDatabase(path string) (*Database, error) {
	if path == "" {
		var err error
		path, err = DefaultDBPath()
		if err != nil {
			return nil, err
		}
	}

	// Create directory if it doesn't exist (instead of failing when DB is missing)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", path)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	// Set SQLite single-writer constraint
	conn.SetMaxOpenConns(1)

	db := &Database{path: path, conn: conn}

	// Verify connection works
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, err
	}

	// Initialize schema (idempotent - safe for existing databases)
	if err := db.initSchema(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("schema initialization failed: %w", err)
	}

	// Run migrations for new columns (idempotent)
	if err := db.ensureMigrations(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return db, nil
}

// initSchema creates the base database schema if it doesn't exist.
// This is idempotent - safe to call on existing databases.
// Schema matches the blogwatcher CLI for full compatibility.
func (db *Database) initSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS blogs (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			url TEXT NOT NULL UNIQUE,
			feed_url TEXT,
			scrape_selector TEXT,
			last_scanned TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS articles (
			id INTEGER PRIMARY KEY,
			blog_id INTEGER,
			title TEXT NOT NULL,
			url TEXT NOT NULL UNIQUE,
			published_date TIMESTAMP,
			discovered_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			is_read BOOLEAN DEFAULT FALSE,
			thumbnail_url TEXT,
			FOREIGN KEY (blog_id) REFERENCES blogs(id)
		);
	`
	_, err := db.conn.Exec(schema)
	return err
}

// ensureMigrations runs idempotent schema migrations for new columns and tables.
// Checks column/table existence before adding.
func (db *Database) ensureMigrations() error {
	// Add thumbnail_url column if it doesn't exist
	if !db.columnExists("articles", "thumbnail_url") {
		if _, err := db.conn.Exec(`ALTER TABLE articles ADD COLUMN thumbnail_url TEXT`); err != nil {
			return err
		}
	}

	// Migrate articles.blog_id to nullable (for blog deletion without cascade)
	// SQLite does not support ALTER COLUMN, so we must recreate the table
	if db.columnIsNotNull("articles", "blog_id") {
		if err := db.migrateBlogIDToNullable(); err != nil {
			return fmt.Errorf("failed to migrate articles.blog_id to nullable: %w", err)
		}
	}

	// Add FTS5 virtual table and sync triggers for title search
	if !db.tableExists("articles_fts") {
		// Create FTS5 virtual table with external content pattern
		if _, err := db.conn.Exec(`CREATE VIRTUAL TABLE articles_fts USING fts5(
			title,
			content='articles',
			content_rowid='id'
		)`); err != nil {
			return fmt.Errorf("failed to create articles_fts: %w", err)
		}

		// Create INSERT trigger
		if _, err := db.conn.Exec(`CREATE TRIGGER articles_ai AFTER INSERT ON articles BEGIN
			INSERT INTO articles_fts(rowid, title) VALUES (new.id, new.title);
		END`); err != nil {
			return fmt.Errorf("failed to create articles_ai trigger: %w", err)
		}

		// Create UPDATE trigger (delete old entry first, then insert new)
		if _, err := db.conn.Exec(`CREATE TRIGGER articles_au AFTER UPDATE ON articles BEGIN
			INSERT INTO articles_fts(articles_fts, rowid, title) VALUES('delete', old.id, old.title);
			INSERT INTO articles_fts(rowid, title) VALUES (new.id, new.title);
		END`); err != nil {
			return fmt.Errorf("failed to create articles_au trigger: %w", err)
		}

		// Create DELETE trigger
		if _, err := db.conn.Exec(`CREATE TRIGGER articles_ad AFTER DELETE ON articles BEGIN
			INSERT INTO articles_fts(articles_fts, rowid, title) VALUES('delete', old.id, old.title);
		END`); err != nil {
			return fmt.Errorf("failed to create articles_ad trigger: %w", err)
		}

		// Populate FTS5 from existing articles
		if _, err := db.conn.Exec(`INSERT INTO articles_fts(rowid, title) SELECT id, title FROM articles`); err != nil {
			return fmt.Errorf("failed to populate articles_fts: %w", err)
		}
	}

	// Add has_note column if it doesn't exist
	if !db.columnExists("articles", "has_note") {
		if _, err := db.conn.Exec(`ALTER TABLE articles ADD COLUMN has_note BOOLEAN DEFAULT FALSE`); err != nil {
			return fmt.Errorf("failed to add has_note column: %w", err)
		}
	}

	return nil
}

// columnExists checks if a column exists in a table using PRAGMA table_info.
func (db *Database) columnExists(table, column string) bool {
	rows, err := db.conn.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue any
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			continue
		}
		if name == column {
			return true
		}
	}
	return false
}

// columnIsNotNull checks if a column has the NOT NULL constraint.
func (db *Database) columnIsNotNull(table, column string) bool {
	rows, err := db.conn.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue any
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			continue
		}
		if name == column {
			return notnull == 1
		}
	}
	return false
}

// migrateBlogIDToNullable recreates articles table with nullable blog_id column.
// SQLite does not support ALTER COLUMN, so we must recreate the table.
func (db *Database) migrateBlogIDToNullable() error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Create new table with nullable blog_id (matching existing schema order)
	_, err = tx.Exec(`CREATE TABLE articles_new (
		id INTEGER PRIMARY KEY,
		blog_id INTEGER,
		title TEXT NOT NULL,
		url TEXT NOT NULL UNIQUE,
		published_date TIMESTAMP,
		discovered_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		is_read BOOLEAN DEFAULT FALSE,
		thumbnail_url TEXT,
		FOREIGN KEY (blog_id) REFERENCES blogs(id)
	)`)
	if err != nil {
		return fmt.Errorf("create articles_new: %w", err)
	}

	// Copy data from old table using explicit column list
	_, err = tx.Exec(`INSERT INTO articles_new
		(id, blog_id, title, url, published_date, discovered_date, is_read, thumbnail_url)
		SELECT id, blog_id, title, url, published_date, discovered_date, is_read, thumbnail_url
		FROM articles`)
	if err != nil {
		return fmt.Errorf("copy data: %w", err)
	}

	// Drop FTS5 triggers (they reference the old table)
	_, _ = tx.Exec(`DROP TRIGGER IF EXISTS articles_ai`)
	_, _ = tx.Exec(`DROP TRIGGER IF EXISTS articles_au`)
	_, _ = tx.Exec(`DROP TRIGGER IF EXISTS articles_ad`)

	// Drop FTS5 table (will be recreated by ensureMigrations)
	_, _ = tx.Exec(`DROP TABLE IF EXISTS articles_fts`)

	// Drop old table and rename new table
	_, err = tx.Exec(`DROP TABLE articles`)
	if err != nil {
		return fmt.Errorf("drop old table: %w", err)
	}

	_, err = tx.Exec(`ALTER TABLE articles_new RENAME TO articles`)
	if err != nil {
		return fmt.Errorf("rename table: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// tableExists checks if a table exists using sqlite_master.
func (db *Database) tableExists(tableName string) bool {
	var name string
	err := db.conn.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&name)
	return err == nil && name == tableName
}

func (db *Database) Path() string {
	return db.path
}

func (db *Database) Close() error {
	if db.conn == nil {
		return nil
	}
	return db.conn.Close()
}

func (db *Database) ListBlogs() ([]model.Blog, error) {
	rows, err := db.conn.Query(`SELECT id, name, url, feed_url, scrape_selector, last_scanned FROM blogs ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blogs []model.Blog
	for rows.Next() {
		blog, err := scanBlog(rows)
		if err != nil {
			return nil, err
		}
		if blog != nil {
			blogs = append(blogs, *blog)
		}
	}
	return blogs, rows.Err()
}

// BlogWithCount extends Blog with article count for settings display.
type BlogWithCount struct {
	model.Blog
	ArticleCount int
}

// ListBlogsWithCounts returns all blogs with their article counts.
// Uses LEFT JOIN to include blogs with zero articles.
func (db *Database) ListBlogsWithCounts() ([]BlogWithCount, error) {
	query := `SELECT
		b.id,
		b.name,
		b.url,
		b.feed_url,
		b.scrape_selector,
		b.last_scanned,
		COUNT(a.id) as article_count
	FROM blogs b
	LEFT JOIN articles a ON b.id = a.blog_id
	GROUP BY b.id, b.name, b.url, b.feed_url, b.scrape_selector, b.last_scanned
	ORDER BY b.name`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blogs []BlogWithCount
	for rows.Next() {
		var (
			id             int64
			name           string
			url            string
			feedURL        sql.NullString
			scrapeSelector sql.NullString
			lastScanned    sql.NullString
			articleCount   int
		)
		if err := rows.Scan(&id, &name, &url, &feedURL, &scrapeSelector, &lastScanned, &articleCount); err != nil {
			return nil, err
		}

		blog := BlogWithCount{
			Blog: model.Blog{
				ID:             id,
				Name:           name,
				URL:            url,
				FeedURL:        feedURL.String,
				ScrapeSelector: scrapeSelector.String,
			},
			ArticleCount: articleCount,
		}
		if lastScanned.Valid {
			if parsed, err := parseTime(lastScanned.String); err == nil {
				blog.LastScanned = &parsed
			}
		}
		blogs = append(blogs, blog)
	}
	return blogs, rows.Err()
}

func (db *Database) ListArticles(unreadOnly bool, blogID *int64) ([]model.Article, error) {
	query := `SELECT id, blog_id, title, url, thumbnail_url, published_date, discovered_date, is_read, has_note FROM articles WHERE 1=1`
	var args []interface{}
	if unreadOnly {
		query += " AND is_read = 0"
	}
	if blogID != nil {
		query += " AND blog_id = ?"
		args = append(args, *blogID)
	}
	query += " ORDER BY COALESCE(published_date, discovered_date) DESC"

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []model.Article
	for rows.Next() {
		article, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		if article != nil {
			articles = append(articles, *article)
		}
	}
	return articles, rows.Err()
}

// ListArticlesByReadStatus returns articles filtered by explicit read status.
// isRead=true returns read articles, isRead=false returns unread articles.
// blogID filters to a specific blog if provided.
func (db *Database) ListArticlesByReadStatus(isRead bool, blogID *int64) ([]model.Article, error) {
	query := `SELECT id, blog_id, title, url, thumbnail_url, published_date, discovered_date, is_read, has_note FROM articles WHERE is_read = ?`
	args := []interface{}{isRead}

	if blogID != nil {
		query += " AND blog_id = ?"
		args = append(args, *blogID)
	}
	query += " ORDER BY COALESCE(published_date, discovered_date) DESC"

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []model.Article
	for rows.Next() {
		article, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		if article != nil {
			articles = append(articles, *article)
		}
	}
	return articles, rows.Err()
}

// ListArticlesWithBlog returns articles with blog metadata (name, URL) for display.
// Uses INNER JOIN to fetch blog info alongside article data.
// isRead filters by read status, blogID optionally filters to a specific blog.
func (db *Database) ListArticlesWithBlog(isRead bool, blogID *int64) ([]model.ArticleWithBlog, error) {
	query := `SELECT a.id, a.blog_id, a.title, a.url, a.thumbnail_url, a.published_date, a.discovered_date, a.is_read, a.has_note, b.name, b.url
		FROM articles a
		INNER JOIN blogs b ON a.blog_id = b.id
		WHERE a.is_read = ?`
	args := []interface{}{isRead}

	if blogID != nil {
		query += " AND a.blog_id = ?"
		args = append(args, *blogID)
	}
	query += " ORDER BY COALESCE(a.published_date, a.discovered_date) DESC"

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []model.ArticleWithBlog
	for rows.Next() {
		article, err := scanArticleWithBlog(rows)
		if err != nil {
			return nil, err
		}
		if article != nil {
			articles = append(articles, *article)
		}
	}
	return articles, rows.Err()
}

// SearchArticles returns articles matching the given search options with total count.
// Uses FTS5 for title search when SearchQuery is non-empty.
// Returns (articles, totalCount, error).
func (db *Database) SearchArticles(opts model.SearchOptions) ([]model.ArticleWithBlog, int, error) {
	// Build base query - conditionally add FTS5 JOIN only when searching
	var query strings.Builder
	query.WriteString(`SELECT a.id, a.blog_id, a.title, a.url, a.thumbnail_url, a.published_date, a.discovered_date, a.is_read, a.has_note, b.name, b.url, COUNT(*) OVER() as total_count
		FROM articles a`)

	var conditions []string
	var args []interface{}

	// Add FTS5 JOIN only if search query provided
	if opts.SearchQuery != "" {
		query.WriteString(` JOIN articles_fts ON a.id = articles_fts.rowid`)
		conditions = append(conditions, "articles_fts MATCH ?")
		args = append(args, opts.SearchQuery)
	}

	query.WriteString(` INNER JOIN blogs b ON a.blog_id = b.id`)

	// Add status condition only if IsRead is not nil
	if opts.IsRead != nil {
		conditions = append(conditions, "a.is_read = ?")
		args = append(args, *opts.IsRead)
	}

	// Add blog filter if provided
	if opts.BlogID != nil {
		conditions = append(conditions, "a.blog_id = ?")
		args = append(args, *opts.BlogID)
	}

	// Add date range using COALESCE for published_date fallback to discovered_date
	if opts.DateFrom != nil {
		conditions = append(conditions, "COALESCE(a.published_date, a.discovered_date) >= ?")
		args = append(args, opts.DateFrom.Format("2006-01-02"))
	}
	if opts.DateTo != nil {
		// Include entire end date by comparing to next day
		endDate := opts.DateTo.AddDate(0, 0, 1)
		conditions = append(conditions, "COALESCE(a.published_date, a.discovered_date) < ?")
		args = append(args, endDate.Format("2006-01-02"))
	}

	// Build WHERE clause
	if len(conditions) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(conditions, " AND "))
	}

	query.WriteString(" ORDER BY COALESCE(a.published_date, a.discovered_date) DESC")

	// Add pagination
	limit := opts.Limit
	if limit <= 0 {
		limit = model.DefaultPageSize
	}
	query.WriteString(fmt.Sprintf(" LIMIT %d", limit))
	if opts.Offset > 0 {
		query.WriteString(fmt.Sprintf(" OFFSET %d", opts.Offset))
	}

	rows, err := db.conn.Query(query.String(), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var articles []model.ArticleWithBlog
	var totalCount int
	for rows.Next() {
		article, count, err := scanArticleWithBlogAndCount(rows)
		if err != nil {
			return nil, 0, err
		}
		if article != nil {
			articles = append(articles, *article)
			totalCount = count
		}
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return articles, totalCount, nil
}

func (db *Database) MarkArticleRead(id int64) (bool, error) {
	result, err := db.conn.Exec(`UPDATE articles SET is_read = 1 WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (db *Database) MarkArticleUnread(id int64) (bool, error) {
	result, err := db.conn.Exec(`UPDATE articles SET is_read = 0 WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

// UpdateArticleHasNote updates the has_note field for an article.
// Returns error if the article does not exist.
func (db *Database) UpdateArticleHasNote(id int64, hasNote bool) error {
	result, err := db.conn.Exec(`UPDATE articles SET has_note = ? WHERE id = ?`, hasNote, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("article not found: %d", id)
	}
	return nil
}

// MarkAllUnreadArticlesRead marks all unread articles as read.
// If blogID is provided, only marks articles from that blog.
func (db *Database) MarkAllUnreadArticlesRead(blogID *int64) error {
	query := `UPDATE articles SET is_read = 1 WHERE is_read = 0`
	var args []interface{}

	if blogID != nil {
		query += " AND blog_id = ?"
		args = append(args, *blogID)
	}

	_, err := db.conn.Exec(query, args...)
	return err
}

// GetBlogByName returns a blog by its name, or nil if not found.
func (db *Database) GetBlogByName(name string) (*model.Blog, error) {
	row := db.conn.QueryRow(`SELECT id, name, url, feed_url, scrape_selector, last_scanned FROM blogs WHERE name = ?`, name)
	return scanBlog(row)
}

// GetBlogByID returns a blog by its ID, or nil if not found.
func (db *Database) GetBlogByID(id int64) (*model.Blog, error) {
	row := db.conn.QueryRow(`SELECT id, name, url, feed_url, scrape_selector, last_scanned FROM blogs WHERE id = ?`, id)
	return scanBlog(row)
}

// GetArticleByID returns an article by its ID, or nil if not found.
func (db *Database) GetArticleByID(id int64) (*model.Article, error) {
	row := db.conn.QueryRow(`SELECT id, blog_id, title, url, thumbnail_url, published_date, discovered_date, is_read, has_note FROM articles WHERE id = ?`, id)
	return scanArticle(row)
}

// GetBlogByURL returns a blog by its URL, or nil if not found.
func (db *Database) GetBlogByURL(url string) (*model.Blog, error) {
	row := db.conn.QueryRow(`SELECT id, name, url, feed_url, scrape_selector, last_scanned FROM blogs WHERE url = ?`, url)
	return scanBlog(row)
}

// AddBlog inserts a new blog and returns it with the assigned ID.
func (db *Database) AddBlog(blog model.Blog) (model.Blog, error) {
	result, err := db.conn.Exec(
		`INSERT INTO blogs (name, url, feed_url, scrape_selector, last_scanned)
		VALUES (?, ?, ?, ?, ?)`,
		blog.Name,
		blog.URL,
		nullIfEmpty(blog.FeedURL),
		nullIfEmpty(blog.ScrapeSelector),
		formatTimePtr(blog.LastScanned),
	)
	if err != nil {
		return blog, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return blog, err
	}
	blog.ID = id
	return blog, nil
}

// UpdateBlogName updates the display name of a blog by ID.
func (db *Database) UpdateBlogName(id int64, name string) error {
	_, err := db.conn.Exec(`UPDATE blogs SET name = ? WHERE id = ?`, name, id)
	return err
}

// GetArticleCountForBlog returns the number of articles for a specific blog.
func (db *Database) GetArticleCountForBlog(blogID int64) (int, error) {
	var count int
	err := db.conn.QueryRow(`SELECT COUNT(*) FROM articles WHERE blog_id = ?`, blogID).Scan(&count)
	return count, err
}

// DeleteBlogOnly deletes a blog but keeps its articles (sets blog_id to NULL).
// Uses a transaction to ensure atomicity.
func (db *Database) DeleteBlogOnly(id int64) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}

	// Orphan articles by setting blog_id to NULL
	if _, err := tx.Exec(`UPDATE articles SET blog_id = NULL WHERE blog_id = ?`, id); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("orphan articles: %w", err)
	}

	// Delete the blog
	result, err := tx.Exec(`DELETE FROM blogs WHERE id = ?`, id)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("delete blog: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		_ = tx.Rollback()
		return fmt.Errorf("blog not found")
	}

	return tx.Commit()
}

// DeleteBlogWithArticles deletes a blog and all its articles (cascade delete).
// Uses a transaction to ensure atomicity.
func (db *Database) DeleteBlogWithArticles(id int64) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}

	// Delete articles first (FTS5 trigger handles articles_fts cleanup)
	if _, err := tx.Exec(`DELETE FROM articles WHERE blog_id = ?`, id); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("delete articles: %w", err)
	}

	// Delete the blog
	result, err := tx.Exec(`DELETE FROM blogs WHERE id = ?`, id)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("delete blog: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		_ = tx.Rollback()
		return fmt.Errorf("blog not found")
	}

	return tx.Commit()
}

// UpdateBlog updates all fields of a blog by ID.
func (db *Database) UpdateBlog(blog model.Blog) error {
	_, err := db.conn.Exec(
		`UPDATE blogs SET name = ?, url = ?, feed_url = ?, scrape_selector = ?, last_scanned = ? WHERE id = ?`,
		blog.Name,
		blog.URL,
		nullIfEmpty(blog.FeedURL),
		nullIfEmpty(blog.ScrapeSelector),
		formatTimePtr(blog.LastScanned),
		blog.ID,
	)
	return err
}

// UpdateBlogLastScanned updates just the last_scanned timestamp for a blog.
func (db *Database) UpdateBlogLastScanned(id int64, lastScanned time.Time) error {
	_, err := db.conn.Exec(`UPDATE blogs SET last_scanned = ? WHERE id = ?`, lastScanned.Format(sqliteTimeLayout), id)
	return err
}

// AddArticlesBulk inserts multiple articles in a single transaction.
// Returns the count of inserted articles.
func (db *Database) AddArticlesBulk(articles []model.Article) (int, error) {
	if len(articles) == 0 {
		return 0, nil
	}
	tx, err := db.conn.Begin()
	if err != nil {
		return 0, err
	}
	stmt, err := tx.Prepare(`INSERT INTO articles (blog_id, title, url, thumbnail_url, published_date, discovered_date, is_read) VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	defer stmt.Close()

	for _, article := range articles {
		_, err := stmt.Exec(
			article.BlogID,
			article.Title,
			article.URL,
			nullIfEmpty(article.ThumbnailURL),
			formatTimePtr(article.PublishedDate),
			formatTimePtr(article.DiscoveredDate),
			article.IsRead,
		)
		if err != nil {
			_ = tx.Rollback()
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return len(articles), nil
}

// GetExistingArticleURLs returns a set of article URLs that already exist in the database.
// Used for deduplication during scanning. Handles chunking for large URL lists.
func (db *Database) GetExistingArticleURLs(urls []string) (map[string]struct{}, error) {
	result := make(map[string]struct{})
	if len(urls) == 0 {
		return result, nil
	}

	chunkSize := 900
	for start := 0; start < len(urls); start += chunkSize {
		end := start + chunkSize
		if end > len(urls) {
			end = len(urls)
		}
		chunk := urls[start:end]
		placeholders := strings.TrimRight(strings.Repeat("?,", len(chunk)), ",")
		query := fmt.Sprintf("SELECT url FROM articles WHERE url IN (%s)", placeholders)
		rows, err := db.conn.Query(query, interfaceSlice(chunk)...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var url string
			if err := rows.Scan(&url); err != nil {
				rows.Close()
				return nil, err
			}
			result[url] = struct{}{}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}
	return result, nil
}

func scanBlog(scanner interface{ Scan(dest ...any) error }) (*model.Blog, error) {
	var (
		id             int64
		name           string
		url            string
		feedURL        sql.NullString
		scrapeSelector sql.NullString
		lastScanned    sql.NullString
	)
	if err := scanner.Scan(&id, &name, &url, &feedURL, &scrapeSelector, &lastScanned); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	blog := &model.Blog{
		ID:             id,
		Name:           name,
		URL:            url,
		FeedURL:        feedURL.String,
		ScrapeSelector: scrapeSelector.String,
	}
	if lastScanned.Valid {
		if parsed, err := parseTime(lastScanned.String); err == nil {
			blog.LastScanned = &parsed
		}
	}
	return blog, nil
}

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
	)
	if err := scanner.Scan(&id, &blogID, &title, &url, &thumbnailURL, &publishedDate, &discovered, &isRead, &hasNote); err != nil {
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
	}
	if publishedDate.Valid {
		if parsed, err := parseTime(publishedDate.String); err == nil {
			article.PublishedDate = &parsed
		}
	}
	if discovered.Valid {
		if parsed, err := parseTime(discovered.String); err == nil {
			article.DiscoveredDate = &parsed
		}
	}

	return article, nil
}

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
		blogName      string
		blogURL       string
		hasNote       bool
	)
	if err := scanner.Scan(&id, &blogID, &title, &url, &thumbnailURL, &publishedDate, &discovered, &isRead, &blogName, &blogURL, &hasNote); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	article := &model.ArticleWithBlog{
		ID:           id,
		BlogID:       blogID,
		Title:        title,
		URL:          url,
		ThumbnailURL: thumbnailURL.String,
		IsRead:       isRead,
		BlogName:     blogName,
		BlogURL:      blogURL,
		HasNote:      hasNote,
	}
	if publishedDate.Valid {
		if parsed, err := parseTime(publishedDate.String); err == nil {
			article.PublishedDate = &parsed
		}
	}
	if discovered.Valid {
		if parsed, err := parseTime(discovered.String); err == nil {
			article.DiscoveredDate = &parsed
		}
	}

	return article, nil
}

func scanArticleWithBlogAndCount(scanner interface{ Scan(dest ...any) error }) (*model.ArticleWithBlog, int, error) {
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
		blogName      string
		blogURL       string
		totalCount    int
	)
	if err := scanner.Scan(&id, &blogID, &title, &url, &thumbnailURL, &publishedDate, &discovered, &isRead, &hasNote, &blogName, &blogURL, &totalCount); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, nil
		}
		return nil, 0, err
	}

	article := &model.ArticleWithBlog{
		ID:           id,
		BlogID:       blogID,
		Title:        title,
		URL:          url,
		ThumbnailURL: thumbnailURL.String,
		IsRead:       isRead,
		HasNote:      hasNote,
		BlogName:     blogName,
		BlogURL:      blogURL,
	}
	if publishedDate.Valid {
		if parsed, err := parseTime(publishedDate.String); err == nil {
			article.PublishedDate = &parsed
		}
	}
	if discovered.Valid {
		if parsed, err := parseTime(discovered.String); err == nil {
			article.DiscoveredDate = &parsed
		}
	}

	return article, totalCount, nil
}

func parseTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, errors.New("empty time")
	}
	parsed, err := time.Parse(sqliteTimeLayout, value)
	if err == nil {
		return parsed, nil
	}
	return time.Parse("2006-01-02 15:04:05", value)
}

func nullIfEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func formatTimePtr(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format(sqliteTimeLayout)
	return &formatted
}

func interfaceSlice(values []string) []interface{} {
	result := make([]interface{}, len(values))
	for i, value := range values {
		result[i] = value
	}
	return result
}

// ArticleForThumbnailSync holds minimal article data needed for thumbnail sync.
type ArticleForThumbnailSync struct {
	ID      int64
	URL     string
	FeedURL string
}

// GetArticlesMissingThumbnails returns articles that have empty thumbnail_url.
// Includes feed_url from the blog for RSS re-parsing.
func (db *Database) GetArticlesMissingThumbnails() ([]ArticleForThumbnailSync, error) {
	rows, err := db.conn.Query(`
		SELECT a.id, a.url, b.feed_url
		FROM articles a
		INNER JOIN blogs b ON a.blog_id = b.id
		WHERE (a.thumbnail_url IS NULL OR a.thumbnail_url = '')
		AND b.feed_url IS NOT NULL AND b.feed_url != ''
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []ArticleForThumbnailSync
	for rows.Next() {
		var a ArticleForThumbnailSync
		if err := rows.Scan(&a.ID, &a.URL, &a.FeedURL); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, rows.Err()
}

// UpdateArticleThumbnail updates the thumbnail_url for a single article.
func (db *Database) UpdateArticleThumbnail(id int64, thumbnailURL string) error {
	_, err := db.conn.Exec(`UPDATE articles SET thumbnail_url = ? WHERE id = ?`, nullIfEmpty(thumbnailURL), id)
	return err
}

// ListFilterOptions 筛选文章的选项参数
// 用于 CLI article list 命令，支持按博客名称、已读状态、备注状态、日期筛选
type ListFilterOptions struct {
	BlogName  string     // 博客名称筛选（空表示所有博客）
	IsRead    *bool      // 已读状态筛选（nil 表示所有状态）
	HasNote   *bool      // 备注状态筛选（nil 表示所有状态，false 表示无备注）
	AfterDate *time.Time // 日期筛选（nil 表示无限制）
	Limit     int        // 结果数量限制（0 表示无限制）
}

// ListArticlesWithFilters 根据筛选选项列出文章（带博客信息）
// 支持按博客名称、已读状态、日期筛选，用于 CLI article list 命令
func (db *Database) ListArticlesWithFilters(opts ListFilterOptions) ([]model.ArticleWithBlog, error) {
	// 构建基础查询
	query := `SELECT a.id, a.blog_id, a.title, a.url, a.thumbnail_url, a.published_date, a.discovered_date, a.is_read, a.has_note, b.name, b.url
		FROM articles a
		INNER JOIN blogs b ON a.blog_id = b.id
		WHERE 1=1`

	var conditions []string
	var args []interface{}

	// 博客名称筛选（通过子查询获取 blog_id）
	if opts.BlogName != "" {
		conditions = append(conditions, "a.blog_id = (SELECT id FROM blogs WHERE name = ?)")
		args = append(args, opts.BlogName)
	}

	// 已读状态筛选
	if opts.IsRead != nil {
		conditions = append(conditions, "a.is_read = ?")
		args = append(args, *opts.IsRead)
	}

	// 备注状态筛选
	if opts.HasNote != nil {
		conditions = append(conditions, "a.has_note = ?")
		args = append(args, *opts.HasNote)
	}

	// 日期筛选（使用 published_date 或 discovered_date）
	if opts.AfterDate != nil {
		conditions = append(conditions, "COALESCE(a.published_date, a.discovered_date) >= ?")
		args = append(args, opts.AfterDate.Format("2006-01-02"))
	}

	// 添加 WHERE 条件
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	// 排序：按发布日期或发现日期降序
	query += " ORDER BY COALESCE(a.published_date, a.discovered_date) DESC"

	// 结果数量限制
	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", opts.Limit)
	}

	// 执行查询
	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 扫描结果
	var articles []model.ArticleWithBlog
	for rows.Next() {
		article, err := scanArticleWithBlog(rows)
		if err != nil {
			return nil, err
		}
		if article != nil {
			articles = append(articles, *article)
		}
	}

	return articles, rows.Err()
}
