# 文章收藏/已读时间戳与排序 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为文章收藏/已读状态记录时间戳（`favorited_at`、`read_at`），并在 Web UI 收藏页/已读页提供按对应时间排序的能力（可切回发布时间），收件箱仅发布时间。

**Architecture:** 在 `articles` 表加两列 nullable timestamp；收藏/已读写方法在置位时写入当前时间、复位时置 NULL；`SearchOptions` 增 `Sort` 字段，`SearchArticles` 一处 ORDER BY 分支（NULL 用 `discovered_date` 回退）；`parseSearchOptions` 按 `filter` 设默认 sort 并解析 `sort` 查询参数；模板在收藏/已读页渲染排序分段按钮。

**Tech Stack:** Go 1.25+、modernc.org/sqlite、FTS5、gohtml 模板、HTMX、Tailwind/CSS。

参考规格：`docs/superpowers/specs/2026-08-17-article-sort-timestamps-design.md`

---

## 文件结构

| 文件 | 责任 |
|------|------|
| `internal/model/model.go` | Article/ArticleWithBlog 加 `FavoritedAt`/`ReadAt`；SearchOptions 加 `Sort` 与常量 |
| `internal/storage/database.go` | migration 两列；状态写方法更新时间戳；scan 函数与所有 SELECT 列表同步两列；SearchArticles ORDER BY 分支 |
| `internal/storage/database_test.go` | 时间戳写入/清空、排序、migration 幂等测试 |
| `internal/server/handlers.go` | parseSearchOptions 解析 sort+默认；各 handler 注入 `CurrentSort` |
| `internal/server/handlers_test.go` | parseSearchOptions 默认值测试 |
| `assets/templates/partials/article-list.gohtml` | 排序分段控件、load-more 带 sort、search/date include `#sort-hidden` |
| `assets/static/styles.css` | `.sort-toggle` 样式 |

无新文件。

---

## Task 1: Model 结构变更

**Files:**
- Modify: `internal/model/model.go:24-69`

- [ ] **Step 1: 为 Article 与 ArticleWithBlog 增加时间字段，为 SearchOptions 增 Sort 与常量**

编辑 `internal/model/model.go`。在 `Article` 结构体 `IsFavorited bool` 行后加：

```go
	FavoritedAt  *time.Time // 收藏时间，未收藏为 nil
	ReadAt       *time.Time // 已读时间，未读为 nil
```

在 `ArticleWithBlog` 结构体 `IsFavorited bool` 行后加同样两行。

在 `SearchOptions` 结构体内（`Limit`/`Offset` 之后）加：

```go
	Sort string // "published"(默认) | "favorited" | "read"
```

在文件末尾（HN 常量块之后）加 Sort 取值常量：

```go
// Sort 取值常量
const (
	SortPublished = "published"
	SortFavorited = "favorited"
	SortRead      = "read"
)
```

- [ ] **Step 2: 编译验证**

Run: `go build ./internal/model/`
Expected: 编译通过，无报错。

- [ ] **Step 3: Commit**

```bash
git add internal/model/model.go
git commit -m "feat(model): 文章增加 FavoritedAt/ReadAt 字段与 Sort 选项"
```

---

## Task 2: 数据库 Migration

**Files:**
- Modify: `internal/storage/database.go:111-213` (`ensureMigrations`)
- Test: `internal/storage/database_test.go`

- [ ] **Step 1: 写失败测试——新列存在且幂等**

在 `internal/storage/database_test.go` 末尾加：

```go
func TestMigrationsAddFavoritedAtAndReadAt(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "blogwatcher.db")

	db, err := OpenDatabase(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db.Close()

	// Re-open: idempotent, must not error on already-existing columns
	db, err = OpenDatabase(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer db.Close()

	if !db.columnExists("articles", "favorited_at") {
		t.Fatal("expected favorited_at column to exist")
	}
	if !db.columnExists("articles", "read_at") {
		t.Fatal("expected read_at column to exist")
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

Run: `go test ./internal/storage/ -run TestMigrationsAddFavoritedAtAndReadAt -v`
Expected: FAIL —— `expected favorited_at column to exist`（列尚未添加）。

- [ ] **Step 3: 实现 migration**

在 `internal/storage/database.go` 的 `ensureMigrations()` 内，紧接 `is_favorited` 列添加块之后（`return nil` 之前）加：

```go
	// Add favorited_at column if it doesn't exist
	if !db.columnExists("articles", "favorited_at") {
		if _, err := db.conn.Exec(`ALTER TABLE articles ADD COLUMN favorited_at TIMESTAMP`); err != nil {
			return fmt.Errorf("failed to add favorited_at column: %w", err)
		}
	}

	// Add read_at column if it doesn't exist
	if !db.columnExists("articles", "read_at") {
		if _, err := db.conn.Exec(`ALTER TABLE articles ADD COLUMN read_at TIMESTAMP`); err != nil {
			return fmt.Errorf("failed to add read_at column: %w", err)
		}
	}
```

- [ ] **Step 4: 运行测试，确认通过**

Run: `go test ./internal/storage/ -run TestMigrationsAddFavoritedAtAndReadAt -v`
Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add internal/storage/database.go internal/storage/database_test.go
git commit -m "feat(storage): migration 新增 favorited_at/read_at 列"
```

---

## Task 3: scan 函数与所有 SELECT 列表同步两列

此任务把 `favorited_at, read_at` 加入 scan 与所有文章查询的 SELECT，保证后续 favorite/read/sort 能读到时间。注意列顺序：两列统一追加在 `is_favorited` 之后（SearchArticles 中在 `COUNT(*) OVER() as total_count` 之前）。

**Files:**
- Modify: `internal/storage/database.go`（scanArticle / scanArticleWithBlog / scanArticleWithBlogAndCount；ListArticles / ListArticlesByReadStatus / ListArticlesWithBlog / SearchArticles / GetArticleByID / ListArticlesWithFilters 的 SELECT）
- Test: `internal/storage/database_test.go`

- [ ] **Step 1: 写失败测试——三种 scan 路径都能读回时间**

在 `internal/storage/database_test.go` 末尾加：

```go
func TestScanReadsFavoritedAtAndReadAt(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	blog, err := db.AddBlog(model.Blog{Name: "T", URL: "https://example.com"})
	if err != nil {
		t.Fatalf("add blog: %v", err)
	}

	pub := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	inserted, _, err := db.AddArticlesBulk([]model.Article{
		{BlogID: blog.ID, Title: "A", URL: "https://example.com/a", PublishedDate: &pub, HNStatus: model.HNStatusNotSearch},
	})
	if err != nil || inserted != 1 {
		t.Fatalf("add articles: inserted=%d err=%v", inserted, err)
	}

	all, _ := db.ListArticles(false, nil)
	id := all[0].ID

	favTime := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	readTime := time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC)
	if _, err := db.conn.Exec(
		`UPDATE articles SET favorited_at=?, read_at=? WHERE id=?`,
		favTime.Format(time.RFC3339Nano), readTime.Format(time.RFC3339Nano), id,
	); err != nil {
		t.Fatalf("raw set timestamps: %v", err)
	}

	// scanArticle path (GetArticleByID)
	a, err := db.GetArticleByID(id)
	if err != nil || a == nil {
		t.Fatalf("get by id: %v %v", a, err)
	}
	if a.FavoritedAt == nil || !a.FavoritedAt.Equal(favTime) {
		t.Errorf("GetArticleByID FavoritedAt = %v, want %v", a.FavoritedAt, favTime)
	}
	if a.ReadAt == nil || !a.ReadAt.Equal(readTime) {
		t.Errorf("GetArticleByID ReadAt = %v, want %v", a.ReadAt, readTime)
	}

	// scanArticleWithBlog path (ListArticlesWithBlog)
	wb, err := db.ListArticlesWithBlog(false, nil)
	if err != nil || len(wb) != 1 {
		t.Fatalf("ListArticlesWithBlog: %v count=%d", err, len(wb))
	}
	if wb[0].FavoritedAt == nil || !wb[0].FavoritedAt.Equal(favTime) {
		t.Errorf("ListArticlesWithBlog FavoritedAt = %v", wb[0].FavoritedAt)
	}

	// scanArticleWithBlogAndCount path (SearchArticles)
	isRead := false
	res, _, err := db.SearchArticles(model.SearchOptions{IsRead: &isRead})
	if err != nil || len(res) != 1 {
		t.Fatalf("SearchArticles: %v count=%d", err, len(res))
	}
	if res[0].FavoritedAt == nil || !res[0].FavoritedAt.Equal(favTime) {
		t.Errorf("SearchArticles FavoritedAt = %v", res[0].FavoritedAt)
	}
	if res[0].ReadAt == nil || !res[0].ReadAt.Equal(readTime) {
		t.Errorf("SearchArticles ReadAt = %v", res[0].ReadAt)
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

Run: `go test ./internal/storage/ -run TestScanReadsFavoritedAtAndReadAt -v`
Expected: FAIL —— `FavoritedAt` 为 nil（scan 尚未读取）。

- [ ] **Step 3: 更新 scanArticle**

在 `scanArticle` 中，变量声明区在 `isFavorited bool` 后加：

```go
	favoritedAt sql.NullString
	readAt      sql.NullString
```

把 `scanner.Scan(...)` 的参数列表末尾（`&isFavorited` 之后）追加：

```go
		, &favoritedAt, &readAt
```

在结构体构造 `article := &model.Article{...}` 内 `IsFavorited: isFavorited,` 之后追加解析：

```go
	}
	if favoritedAt.Valid {
		if parsed, err := parseTime(favoritedAt.String); err == nil {
			article.FavoritedAt = &parsed
		}
	}
	if readAt.Valid {
		if parsed, err := parseTime(readAt.String); err == nil {
			article.ReadAt = &parsed
		}
	}
```

（注意：原 `article := &model.Article{...}` 的闭合 `}` 之后已有 `if publishedDate.Valid {...}` 与 `if discovered.Valid {...}` 块；把上面的两块追加到这两块之后、`return article, nil` 之前。）

- [ ] **Step 4: 更新 scanArticleWithBlog**

同样在该函数变量声明区 `isFavorited bool` 后加 `favoritedAt`、`readAt` 两个 `sql.NullString`；`scanner.Scan(...)` 参数末尾追加 `&favoritedAt, &readAt`；结构体构造闭合后追加两个 `if ...Valid` 解析块赋值给 `article.FavoritedAt`/`article.ReadAt`（结构与 scanArticle 相同）。

- [ ] **Step 5: 更新 scanArticleWithBlogAndCount**

变量声明区 `isFavorited bool` 后加 `favoritedAt`、`readAt`（`totalCount int` 保持最后扫描）。`scanner.Scan(...)` 参数列表中，把 `&isFavorited` 与 `&totalCount` 之间插入 `&favoritedAt, &readAt`，即顺序变为 `..., &isFavorited, &favoritedAt, &readAt, &totalCount`。结构体构造后追加两个解析块。

- [ ] **Step 6: 更新所有 SELECT 列表**

按以下精确替换（每处都是把 `favorited_at, read_at` 追加在 `is_favorited` 之后；SearchArticles 中追加在 `COUNT(*) OVER() as total_count` 之前）：

`ListArticles`：
```go
query := `SELECT id, blog_id, title, url, thumbnail_url, published_date, discovered_date, is_read, has_note, hn_url, hn_status, is_favorited, favorited_at, read_at FROM articles WHERE 1=1`
```

`ListArticlesByReadStatus`：
```go
query := `SELECT id, blog_id, title, url, thumbnail_url, published_date, discovered_date, is_read, has_note, hn_url, hn_status, is_favorited, favorited_at, read_at FROM articles WHERE is_read = ?`
```

`ListArticlesWithBlog`：
```go
query := `SELECT a.id, a.blog_id, a.title, a.url, a.thumbnail_url, a.published_date, a.discovered_date, a.is_read, a.has_note, a.hn_url, a.hn_status, b.name, b.url, a.is_favorited, a.favorited_at, a.read_at
	FROM articles a
	INNER JOIN blogs b ON a.blog_id = b.id
	WHERE a.is_read = ?`
```

`SearchArticles`（注意 `a.favorited_at, a.read_at` 在 `COUNT(*) OVER()` 之前）：
```go
	query.WriteString(`SELECT a.id, a.blog_id, a.title, a.url, a.thumbnail_url, a.published_date, a.discovered_date, a.is_read, a.has_note, a.hn_url, a.hn_status, b.name, b.url, a.is_favorited, a.favorited_at, a.read_at, COUNT(*) OVER() as total_count
		FROM articles a`)
```

`GetArticleByID`：
```go
row := db.conn.QueryRow(`SELECT id, blog_id, title, url, thumbnail_url, published_date, discovered_date, is_read, has_note, hn_url, hn_status, is_favorited, favorited_at, read_at FROM articles WHERE id = ?`, id)
```

`ListArticlesWithFilters`：
```go
query := `SELECT a.id, a.blog_id, a.title, a.url, a.thumbnail_url, a.published_date, a.discovered_date, a.is_read, a.has_note, a.hn_url, a.hn_status, b.name, b.url, a.is_favorited, a.favorited_at, a.read_at
	FROM articles a
	INNER JOIN blogs b ON a.blog_id = b.id
	WHERE 1=1`
```

- [ ] **Step 7: 运行测试，确认通过**

Run: `go test ./internal/storage/ -v`
Expected: 全部 PASS（含新测试与既有测试；既有测试若因列顺序失败需回查 SELECT/Scan 是否一致）。

- [ ] **Step 8: Commit**

```bash
git add internal/storage/database.go internal/storage/database_test.go
git commit -m "feat(storage): scan 与查询同步 favorited_at/read_at 列"
```

---

## Task 4: 收藏写方法写入/清空 favorited_at

**Files:**
- Modify: `internal/storage/database.go:804-834` (FavoriteArticle/UnfavoriteArticle)
- Test: `internal/storage/database_test.go`

- [ ] **Step 1: 写失败测试**

在 `internal/storage/database_test.go` 末尾加：

```go
func TestFavoriteArticleSetsFavoritedAt(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	blog, _ := db.AddBlog(model.Blog{Name: "T", URL: "https://example.com"})
	if _, _, err := db.AddArticlesBulk([]model.Article{
		{BlogID: blog.ID, Title: "A", URL: "https://example.com/a", HNStatus: model.HNStatusNotSearch},
	}); err != nil {
		t.Fatalf("add: %v", err)
	}
	all, _ := db.ListArticles(false, nil)
	id := all[0].ID

	if err := db.FavoriteArticle(id); err != nil {
		t.Fatalf("favorite: %v", err)
	}
	a, _ := db.GetArticleByID(id)
	if a.FavoritedAt == nil {
		t.Fatal("expected favorited_at to be set after favorite")
	}

	if err := db.UnfavoriteArticle(id); err != nil {
		t.Fatalf("unfavorite: %v", err)
	}
	a, _ = db.GetArticleByID(id)
	if a.FavoritedAt != nil {
		t.Fatalf("expected favorited_at to be nil after unfavorite, got %v", a.FavoritedAt)
	}
	if a.IsFavorited {
		t.Fatal("expected is_favorited=false after unfavorite")
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

Run: `go test ./internal/storage/ -run TestFavoriteArticleSetsFavoritedAt -v`
Expected: FAIL —— `expected favorited_at to be set`（当前 FavoriteArticle 不写时间）。

- [ ] **Step 3: 实现 FavoriteArticle/UnfavoriteArticle**

替换 `FavoriteArticle` 方法体为：

```go
func (db *Database) FavoriteArticle(id int64) error {
	now := time.Now().Format(sqliteTimeLayout)
	log.Printf("[storage] FavoriteArticle: id=%d time=%s", id, now)
	result, err := db.conn.Exec(`UPDATE articles SET is_favorited = 1, favorited_at = ? WHERE id = ?`, now, id)
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
```

替换 `UnfavoriteArticle` 方法体为：

```go
func (db *Database) UnfavoriteArticle(id int64) error {
	log.Printf("[storage] UnfavoriteArticle: id=%d", id)
	result, err := db.conn.Exec(`UPDATE articles SET is_favorited = 0, favorited_at = NULL WHERE id = ?`, id)
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

- [ ] **Step 4: 确认 log 包已导入**

Run: `grep -n '"log"' internal/storage/database.go`
Expected: 输出一行 import。若 database.go 尚未导入 `log`，在 import 块加 `"log"`。

- [ ] **Step 5: 运行测试，确认通过**

Run: `go test ./internal/storage/ -run TestFavoriteArticle -v`
Expected: PASS（含既有 `TestFavoriteArticle`/`TestFavoriteArticleNotFound` 与新测试）。

- [ ] **Step 6: Commit**

```bash
git add internal/storage/database.go internal/storage/database_test.go
git commit -m "feat(storage): 收藏/取消收藏写入与清空 favorited_at"
```

---

## Task 5: 已读写方法写入/清空 read_at

**Files:**
- Modify: `internal/storage/database.go:761-849` (MarkArticleRead/MarkArticleUnread/MarkAllUnreadArticlesRead)
- Test: `internal/storage/database_test.go`

- [ ] **Step 1: 写失败测试**

在 `internal/storage/database_test.go` 末尾加：

```go
func TestMarkArticleReadSetsReadAt(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	blog, _ := db.AddBlog(model.Blog{Name: "T", URL: "https://example.com"})
	if _, _, err := db.AddArticlesBulk([]model.Article{
		{BlogID: blog.ID, Title: "A", URL: "https://example.com/a", HNStatus: model.HNStatusNotSearch},
	}); err != nil {
		t.Fatalf("add: %v", err)
	}
	all, _ := db.ListArticles(false, nil)
	id := all[0].ID

	if _, err := db.MarkArticleRead(id); err != nil {
		t.Fatalf("mark read: %v", err)
	}
	a, _ := db.GetArticleByID(id)
	if a.ReadAt == nil {
		t.Fatal("expected read_at set after mark read")
	}

	if _, err := db.MarkArticleUnread(id); err != nil {
		t.Fatalf("mark unread: %v", err)
	}
	a, _ = db.GetArticleByID(id)
	if a.ReadAt != nil {
		t.Fatalf("expected read_at nil after unread, got %v", a.ReadAt)
	}
	if a.IsRead {
		t.Fatal("expected is_read=false after unread")
	}
}

func TestMarkAllUnreadArticlesReadSetsReadAt(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	blog, _ := db.AddBlog(model.Blog{Name: "T", URL: "https://example.com"})
	if _, _, err := db.AddArticlesBulk([]model.Article{
		{BlogID: blog.ID, Title: "A", URL: "https://example.com/a", HNStatus: model.HNStatusNotSearch},
		{BlogID: blog.ID, Title: "B", URL: "https://example.com/b", HNStatus: model.HNStatusNotSearch},
	}); err != nil {
		t.Fatalf("add: %v", err)
	}

	if err := db.MarkAllUnreadArticlesRead(nil); err != nil {
		t.Fatalf("mark all: %v", err)
	}
	all, _ := db.ListArticlesByReadStatus(true, nil)
	for _, a := range all {
		if a.ReadAt == nil {
			t.Errorf("article %d: expected read_at set", a.ID)
		}
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

Run: `go test ./internal/storage/ -run 'TestMarkArticleReadSetsReadAt|TestMarkAllUnreadArticlesReadSetsReadAt' -v`
Expected: FAIL —— `expected read_at set`。

- [ ] **Step 3: 实现三个写方法**

替换 `MarkArticleRead`：

```go
func (db *Database) MarkArticleRead(id int64) (bool, error) {
	now := time.Now().Format(sqliteTimeLayout)
	log.Printf("[storage] MarkArticleRead: id=%d time=%s", id, now)
	result, err := db.conn.Exec(`UPDATE articles SET is_read = 1, read_at = ? WHERE id = ?`, now, id)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}
```

替换 `MarkArticleUnread`：

```go
func (db *Database) MarkArticleUnread(id int64) (bool, error) {
	log.Printf("[storage] MarkArticleUnread: id=%d", id)
	result, err := db.conn.Exec(`UPDATE articles SET is_read = 0, read_at = NULL WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}
```

替换 `MarkAllUnreadArticlesRead`：

```go
func (db *Database) MarkAllUnreadArticlesRead(blogID *int64) error {
	now := time.Now().Format(sqliteTimeLayout)
	log.Printf("[storage] MarkAllUnreadArticlesRead: blog=%v time=%s", blogID, now)
	query := `UPDATE articles SET is_read = 1, read_at = ? WHERE is_read = 0`
	var args []interface{}
	args = append(args, now)

	if blogID != nil {
		query += " AND blog_id = ?"
		args = append(args, *blogID)
	}

	_, err := db.conn.Exec(query, args...)
	return err
}
```

- [ ] **Step 4: 运行测试，确认通过**

Run: `go test ./internal/storage/ -v`
Expected: 全部 PASS。

- [ ] **Step 5: Commit**

```bash
git add internal/storage/database.go internal/storage/database_test.go
git commit -m "feat(storage): 已读/未读/批量已读写入与清空 read_at"
```

---

## Task 6: SearchArticles 按 Sort 排序

**Files:**
- Modify: `internal/storage/database.go:668-759` (SearchArticles ORDER BY)
- Test: `internal/storage/database_test.go`

- [ ] **Step 1: 写失败测试——按 favorited 与 read 排序，含 NULL 回退**

在 `internal/storage/database_test.go` 末尾加：

```go
func TestSearchArticlesSortByFavorited(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	blog, _ := db.AddBlog(model.Blog{Name: "T", URL: "https://example.com"})

	// 三个文章，discovered_date 不同；其中两个收藏，favorited_at 顺序与 discovered 相反以验证排序按 favorited_at
	d1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	d3 := time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)
	favLate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	favEarly := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	arts := []model.Article{
		{BlogID: blog.ID, Title: "no-fav", URL: "https://example.com/1", DiscoveredDate: &d1, HNStatus: model.HNStatusNotSearch},
		{BlogID: blog.ID, Title: "fav-early", URL: "https://example.com/2", DiscoveredDate: &d2, HNStatus: model.HNStatusNotSearch},
		{BlogID: blog.ID, Title: "fav-late", URL: "https://example.com/3", DiscoveredDate: &d3, HNStatus: model.HNStatusNotSearch},
	}
	if _, _, err := db.AddArticlesBulk(arts); err != nil {
		t.Fatalf("add: %v", err)
	}
	all, _ := db.ListArticles(false, nil)
	idMap := map[string]int64{}
	for _, a := range all {
		idMap[a.Title] = a.ID
	}
	db.conn.Exec(`UPDATE articles SET is_favorited=1, favorited_at=? WHERE id=?`, favEarly.Format(time.RFC3339Nano), idMap["fav-early"])
	db.conn.Exec(`UPDATE articles SET is_favorited=1, favorited_at=? WHERE id=?`, favLate.Format(time.RFC3339Nano), idMap["fav-late"])

	isFav := true
	res, _, err := db.SearchArticles(model.SearchOptions{IsFavorited: &isFav, Sort: model.SortFavorited})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 favorited, got %d", len(res))
	}
	if res[0].Title != "fav-late" || res[1].Title != "fav-early" {
		t.Errorf("order = %s, %s; want fav-late, fav-early", res[0].Title, res[1].Title)
	}
}

func TestSearchArticlesSortByReadWithNullFallback(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	blog, _ := db.AddBlog(model.Blog{Name: "T", URL: "https://example.com"})

	d1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) // older discovered, read_at NULL -> falls back to d1
	d2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	readLate := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	arts := []model.Article{
		{BlogID: blog.ID, Title: "read-null", URL: "https://example.com/1", DiscoveredDate: &d1, HNStatus: model.HNStatusNotSearch},
		{BlogID: blog.ID, Title: "read-set", URL: "https://example.com/2", DiscoveredDate: &d2, HNStatus: model.HNStatusNotSearch},
	}
	if _, _, err := db.AddArticlesBulk(arts); err != nil {
		t.Fatalf("add: %v", err)
	}
	all, _ := db.ListArticles(false, nil)
	for _, a := range all {
		if a.Title == "read-set" {
			db.conn.Exec(`UPDATE articles SET is_read=1, read_at=? WHERE id=?`, readLate.Format(time.RFC3339Nano), a.ID)
		} else {
			db.conn.Exec(`UPDATE articles SET is_read=1 WHERE id=?`, a.ID) // read_at stays NULL
		}
	}

	isRead := true
	res, _, err := db.SearchArticles(model.SearchOptions{IsRead: &isRead, Sort: model.SortRead})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 read, got %d", len(res))
	}
	// read-set has read_at=readLate (newest) -> first; read-null falls back to d1 (older) -> second
	if res[0].Title != "read-set" || res[1].Title != "read-null" {
		t.Errorf("order = %s, %s; want read-set, read-null", res[0].Title, res[1].Title)
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

Run: `go test ./internal/storage/ -run 'TestSearchArticlesSortByFavorited|TestSearchArticlesSortByReadWithNullFallback' -v`
Expected: FAIL —— 排序仍按 published_date，顺序不符。

- [ ] **Step 3: 实现 ORDER BY 分支**

在 `SearchArticles` 中，找到：

```go
	query.WriteString(" ORDER BY COALESCE(a.published_date, a.discovered_date) DESC")
```

替换为：

```go
	var orderBy string
	switch opts.Sort {
	case model.SortFavorited:
		orderBy = "COALESCE(a.favorited_at, a.discovered_date) DESC"
	case model.SortRead:
		orderBy = "COALESCE(a.read_at, a.discovered_date) DESC"
	default:
		orderBy = "COALESCE(a.published_date, a.discovered_date) DESC"
	}
	query.WriteString(" ORDER BY " + orderBy)
```

- [ ] **Step 4: 运行测试，确认通过**

Run: `go test ./internal/storage/ -v`
Expected: 全部 PASS。

- [ ] **Step 5: Commit**

```bash
git add internal/storage/database.go internal/storage/database_test.go
git commit -m "feat(storage): SearchArticles 支持按 favorited/read 时间排序"
```

---

## Task 7: parseSearchOptions 解析 sort + 注入 CurrentSort

**Files:**
- Modify: `internal/server/handlers.go:590-650` (parseSearchOptions) 与各 handler data map
- Test: `internal/server/handlers_test.go`

- [ ] **Step 1: 写失败测试——默认值随 filter 变化**

在 `internal/server/handlers_test.go` 末尾加：

```go
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
	}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodGet, c.query, nil)
		opts, _, _ := parseSearchOptions(req)
		if opts.Sort != c.want {
			t.Errorf("query=%s: Sort=%q want %q", c.query, opts.Sort, c.want)
		}
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

Run: `go test ./internal/server/ -run TestParseSearchOptionsSortDefaults -v`
Expected: FAIL —— `Sort` 为空字符串。

- [ ] **Step 3: 实现 parseSearchOptions 的 sort 解析**

在 `parseSearchOptions` 中，`filter` switch 之后、解析 blog 之前，插入：

```go
	// Parse sort (default depends on filter)
	sortParam := r.URL.Query().Get("sort")
	switch sortParam {
	case model.SortFavorited, model.SortRead, model.SortPublished:
		opts.Sort = sortParam
	default:
		switch filter {
		case "favorites":
			opts.Sort = model.SortFavorited
		case "read":
			opts.Sort = model.SortRead
		default:
			opts.Sort = model.SortPublished
		}
	}
```

- [ ] **Step 4: 各 handler data map 注入 CurrentSort**

在 `handleIndex` 的 data map 加：`"CurrentSort": opts.Sort,`（与 `"CurrentFilter": filter,` 相邻）。

`handleArticleList` 的 data map（HTMX 分支与全页分支共用同一个 data，在 `"CurrentFilter": filter,` 相邻处）加：`"CurrentSort": opts.Sort,`。

`handleMarkAllRead` 的 data map 加：`"CurrentSort": opts.Sort,`。

`handleSync` 的 data map 加：`"CurrentSort": opts.Sort,`。

- [ ] **Step 5: 运行测试，确认通过**

Run: `go test ./internal/server/ -v`
Expected: 全部 PASS（含新测试与既有测试）。

- [ ] **Step 6: Commit**

```bash
git add internal/server/handlers.go internal/server/handlers_test.go
git commit -m "feat(server): parseSearchOptions 解析 sort 并按 filter 设默认"
```

---

## Task 8: 模板排序控件与 load-more 携带 sort

**Files:**
- Modify: `assets/templates/partials/article-list.gohtml`

- [ ] **Step 1: 在 filter-bar 内加排序控件与隐藏 input**

在 `article-list.gohtml` 的 `.filter-bar` 内，紧接 `<input type="hidden" name="filter" id="filter-hidden" ...>` 那一行之后插入：

```html
    {{if or (eq .CurrentFilter "favorites") (eq .CurrentFilter "read")}}
    <div class="sort-toggle" role="radiogroup" aria-label="Sort">
        {{if eq .CurrentFilter "favorites"}}
        <input type="radio" name="sort" id="sort-favorited" value="favorited"
               {{if eq .CurrentSort "favorited"}}checked{{end}}
               hx-get="/articles"
               hx-trigger="change"
               hx-target="#main-content"
               hx-include="#filter-hidden, #blog-hidden, #search-input, #date_from, #date_to"
               hx-push-url="true">
        <label for="sort-favorited">收藏时间</label>
        {{else}}
        <input type="radio" name="sort" id="sort-read" value="read"
               {{if eq .CurrentSort "read"}}checked{{end}}
               hx-get="/articles"
               hx-trigger="change"
               hx-target="#main-content"
               hx-include="#filter-hidden, #blog-hidden, #search-input, #date_from, #date_to"
               hx-push-url="true">
        <label for="sort-read">已读时间</label>
        {{end}}
        <input type="radio" name="sort" id="sort-published" value="published"
               {{if eq .CurrentSort "published"}}checked{{end}}
               hx-get="/articles"
               hx-trigger="change"
               hx-target="#main-content"
               hx-include="#filter-hidden, #blog-hidden, #search-input, #date_from, #date_to"
               hx-push-url="true">
        <label for="sort-published">发布时间</label>
    </div>
    {{end}}
    <input type="hidden" name="sort" id="sort-hidden" value="{{.CurrentSort}}">
```

- [ ] **Step 2: search/date 输入的 hx-include 追加 #sort-hidden**

把 search-input 的 `hx-include="#filter-hidden, #blog-hidden, #date_from, #date_to"` 改为：
`hx-include="#filter-hidden, #blog-hidden, #date_from, #date_to, #sort-hidden"`

把 date_from 的 `hx-include="#filter-hidden, #blog-hidden, #search-input, #date_to"` 改为：
`hx-include="#filter-hidden, #blog-hidden, #search-input, #date_to, #sort-hidden"`

把 date_to 的 `hx-include="#filter-hidden, #blog-hidden, #search-input, #date_from"` 改为：
`hx-include="#filter-hidden, #blog-hidden, #search-input, #date_from, #sort-hidden"`

- [ ] **Step 3: load-more 链接携带 sort**

把 `#load-more-trigger` 的 `hx-get` URL 末尾 `&offset={{.NextOffset}}"` 改为 `&amp;sort={{.CurrentSort}}&amp;offset={{.NextOffset}}"`，即完整一行：

```html
<div id="load-more-trigger"
     hx-get="/articles?filter={{.CurrentFilter}}{{if .CurrentBlogID}}&amp;blog={{.CurrentBlogID}}{{end}}{{if .SearchQuery}}&amp;search={{.SearchQuery}}{{end}}{{if .DateFrom}}&amp;date_from={{.DateFrom}}{{end}}{{if .DateTo}}&amp;date_to={{.DateTo}}{{end}}&amp;sort={{.CurrentSort}}&amp;offset={{.NextOffset}}"
     hx-trigger="intersect once threshold:0.1"
     hx-swap="outerHTML"
     hx-indicator="#loading-indicator">
```

- [ ] **Step 4: 写 handler 测试——收藏页响应含排序控件**

在 `internal/server/handlers_test.go` 末尾加：

```go
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
```

- [ ] **Step 5: 运行测试，确认通过**

Run: `go test ./internal/server/ -run 'TestArticleListRendersSortControlOnFavorites|TestArticleListNoSortControlOnInbox' -v`
Expected: PASS。

- [ ] **Step 6: Commit**

```bash
git add assets/templates/partials/article-list.gohtml internal/server/handlers_test.go
git commit -m "feat(ui): 收藏/已读页排序分段控件与 load-more 携带 sort"
```

---

## Task 9: sort-toggle 样式

**Files:**
- Modify: `assets/static/styles.css`（在 `.view-toggle` 块之后）

- [ ] **Step 1: 追加 .sort-toggle 样式**

在 `styles.css` 的 `.view-toggle svg { ... }` 块之后插入（复用 view-toggle 视觉，仅字号适配文字 label）：

```css
/* ============================================
   Sort Toggle (favorites / read pages)
   ============================================ */
.sort-toggle {
  display: inline-flex;
  background-color: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 4px;
  gap: 4px;
}

.sort-toggle input[type="radio"] {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
  pointer-events: none;
}

.sort-toggle label {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 6px 12px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  transition: background-color 0.2s ease, color 0.2s ease;
  color: var(--text-secondary);
}

.sort-toggle label:hover {
  background-color: var(--bg-elevated);
}

.sort-toggle input[type="radio"]:checked + label {
  background-color: var(--bg-elevated);
  color: var(--text-primary);
}

.sort-toggle input[type="radio"]:focus-visible + label {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}
```

- [ ] **Step 2: Commit**

```bash
git add assets/static/styles.css
git commit -m "style(ui): sort-toggle 排序控件样式"
```

---

## Task 10: 全量测试、构建与部署

按项目 CLAUDE.md 流程要求。

- [ ] **Step 1: 全量测试**

Run: `go test ./...`
Expected: 全部 PASS。

- [ ] **Step 2: 编译并安装 CLI 到全局**

Run: `go install ./cmd/blogwatcher`
Expected: 无报错，`C:\Users\Melodyi\go\bin\blogwatcher.exe` 更新。

- [ ] **Step 3: 重新构建并部署 Docker 服务**

Run: `docker compose build blogwatcher-ui --no-cache && docker compose up -d blogwatcher-ui`
Expected: 构建并启动成功。

- [ ] **Step 4: 手动验证 UI**

在浏览器访问 Web UI：
- 打开收藏页（`?filter=favorites`）：默认按收藏时间排序，排序控件「收藏时间」选中；切换到「发布时间」后顺序变化、URL 含 `sort=published`、刷新后保持。
- 打开已读页（`?filter=read`）：默认按已读时间排序，控件「已读时间」选中。
- 收件箱：无排序控件。
- 收藏一篇文章后，收藏页顶部出现该文章（取消即清空后从收藏页消失）。
- load-more 加载更多时排序保持。

- [ ] **Step 5: 最终 Commit（如有验证修复）**

若手动验证发现需修补，按 TDD 补测试并提交；否则无需额外 commit。

---

## Self-Review 结果

**Spec coverage：** 规格各节均有对应任务——migration（Task2）、model（Task1）、状态写方法含时间戳与清空（Task4/5）、SearchArticles ORDER BY 与 NULL 回退（Task6）、parseSearchOptions 解析+默认（Task7）、模板控件+load-more+include（Task8）、CSS（Task9）、测试计划全部覆盖（Task2-8 各含测试）、日志（Task4/5 写方法含 log.Printf）、文件清单一致。CLI 与卡片时间显示按规格不在范围内，计划未涉及——一致。

**Placeholder scan：** 无 TBD/TODO；每个代码步骤均给出完整代码；测试均含可运行代码。

**Type consistency：** `model.SortFavorited/SortRead/SortPublished` 在 Task1 定义，Task6/7 引用一致；`FavoritedAt`/`ReadAt` 字段名与 scan 解析、测试断言一致；`CurrentSort` 模板键在 Task7 注入、Task8 引用一致。
