# 文章标签功能 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为文章新增可自定义的多对多"标签"维度，支持 UI（卡片弹层 + 单文管理页 + 侧边栏/工具栏筛选）与 CLI（tag 命令组 + article tag/untag + list --tag）。

**Architecture:** 规范多对多：`tags` 表 + `article_tags` 关联表。标签为独立维度（不限收藏文章）。即用即建 + 独立管理页（重命名/删除）。Storage 事务级联删除关联；列表查询用 `EXISTS` 子查询按标签过滤、用批量查询装配卡片 chips（避免 N+1）。

**Tech Stack:** Go 1.25+ / modernc.org/sqlite / cobra / gohtml 模板 + HTMX / Tailwind（嵌入式）

**Spec:** `docs/superpowers/specs/2026-08-17-article-tags-design.md`

**关键既有模式（实现时对照）：**
- Migration：`internal/storage/database.go:111` `ensureMigrations()`，用 `db.tableExists("...")`（无白名单）建表。
- 查询过滤：`database.go:1457` `ListArticlesWithFilters`（CLI 用）与 `database.go:1852` 附近的 `SearchArticles`/`CountArticlesWithFilters`（Web 用）——两处 WHERE 构建都要加标签过滤。
- Handler：`internal/server/handlers.go:291` `handleFavorite` / `:345` `renderUpdatedArticleCard`（HTMX 重渲染卡片 + 触发 `articleListUpdated`）。
- 页面 handler：`handlers.go:1138` `handleNote` + `assets/templates/pages/note.gohtml`。
- CLI 命令：`internal/cli/commands/article.go:424` `NewFavoriteCmd`/`:458` `runFavorite`；`category.go:17` 命令组风格；`root.go:40` 注册。
- output：`internal/cli/output/category.go` 格式化范式。
- 测试：`internal/storage/database_test.go:649` `TestFavoriteArticle`（`openTestDB(t)`）；`internal/server/handlers_test.go:287` `TestHandleFavoriteAndUnfavorite`（`createTestServer(t)`，`srv.ServeHTTP(rec, req)`）。

---

## 文件结构

| 文件 | 责任 |
|------|------|
| `internal/model/model.go` | `Tag` 结构；`ArticleWithBlog.Tags`；`SearchOptions`/`ListFilterOptions.TagName` |
| `internal/storage/database.go` | 两表 migration；Tag/ArticleTag 全套 storage 方法；两个过滤查询加 TagName + 批量标签装配 |
| `internal/cli/commands/tag.go` | 新增：`tag` 命令组 list/rename/delete |
| `internal/cli/commands/article.go` | 新增 `tag`/`untag` 子命令；`list` 加 `--tag` |
| `internal/cli/commands/root.go` | 注册 `tag` 命令组 |
| `internal/cli/output/tag.go` | 新增：`FormatTagTable`/`FormatTagJSON` |
| `internal/server/routes.go` | 注册 10 条新路由 |
| `internal/server/handlers.go` | 标签 CRUD + 文章标签 handler；`parseSearchOptions` 加 tag 分支；列表 handler 装配 Tags |
| `assets/templates/partials/sidebar.gohtml` | Tags 分区 |
| `assets/templates/partials/article-items.gohtml` | 标签按钮 + chips |
| `assets/templates/partials/article-list.gohtml` | 工具栏标签下拉 |
| `assets/templates/pages/tags.gohtml` | 新增：全局标签管理页 |
| `assets/templates/pages/article-tags.gohtml` | 新增：单文标签管理页 |
| `assets/templates/partials/article-tags-edit.gohtml` | 新增：卡片弹层片段 |
| `assets/static/styles.css` | tag-chip / popover / tag-btn 样式 |

---

## Task 1: Model 与 Migration

**Files:**
- Modify: `internal/model/model.go`
- Modify: `internal/storage/database.go:111-213`（`ensureMigrations` 末尾追加）
- Test: `internal/storage/database_test.go`

- [ ] **Step 1: 在 model.go 新增 Tag 结构与字段**

`internal/model/model.go`，在文件末尾（HNStatus 常量之后）新增：

```go
// Tag 表示一个用户自定义标签，用于对文章标注与分类。
type Tag struct {
	ID           int64
	Name         string
	CreatedAt    time.Time
	ArticleCount int64 // 仅 ListTags 时填充，其余方法为 0
}
```

在 `ArticleWithBlog` 结构体中（`model.go:41-56`）`HNStatus` 字段后追加：

```go
	Tags []Tag // 文章标签，列表渲染时由批量查询装配
```

在 `SearchOptions`（`model.go:60-69`）`Offset` 字段后追加：

```go
	TagName string // 空字符串 = 不按标签筛选；非空 = 该标签下的文章
```

- [ ] **Step 2: 在 ListFilterOptions 加 TagName**

`internal/storage/database.go:1444-1453` `ListFilterOptions` 末尾（`Offset` 之后）追加：

```go
	TagName      string     // 标签名称筛选（空表示不按标签筛选）
```

- [ ] **Step 3: 在 ensureMigrations 末尾追加两表**

`internal/storage/database.go`，在 `ensureMigrations()` 的 `return nil` 之前（`is_favorited` migration 之后）追加：

```go
	// Create tags table if it doesn't exist
	if !db.tableExists("tags") {
		if _, err := db.conn.Exec(`CREATE TABLE IF NOT EXISTS tags (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`); err != nil {
			return fmt.Errorf("failed to create tags table: %w", err)
		}
	}

	// Create article_tags join table if it doesn't exist
	if !db.tableExists("article_tags") {
		if _, err := db.conn.Exec(`CREATE TABLE IF NOT EXISTS article_tags (
			tag_id INTEGER NOT NULL,
			article_id INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (tag_id, article_id)
		)`); err != nil {
			return fmt.Errorf("failed to create article_tags table: %w", err)
		}
	}
```

- [ ] **Step 4: 写 migration 测试**

`internal/storage/database_test.go`，新增：

```go
func TestTagsTablesMigration(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	for _, tbl := range []string{"tags", "article_tags"} {
		if !db.tableExists(tbl) {
			t.Fatalf("expected table %s to exist after migration", tbl)
		}
	}
}
```

- [ ] **Step 5: 运行测试验证通过**

Run: `go test ./internal/storage/ -run TestTagsTablesMigration -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/model/model.go internal/storage/database.go internal/storage/database_test.go
git commit -m "feat(tags): 新增 tags/article_tags 表与 Tag model"
```

---

## Task 2: Storage — 标签 CRUD（Create/Get/List/Rename/Delete）

**Files:**
- Modify: `internal/storage/database.go`（在 `CreateCategory` 附近，`:1543` 之后新增一组方法）
- Test: `internal/storage/database_test.go`

- [ ] **Step 1: 写失败测试**

`internal/storage/database_test.go`，新增：

```go
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

	blog, _ := db.AddBlog(model.Blog{Name: "B", URL: "https://example.com"})
	db.AddArticlesBulk([]model.Article{{BlogID: blog.ID, Title: "A1", URL: "https://example.com/a1", HNStatus: model.HNStatusNotSearch}})
	arts, _ := db.ListArticles(false, nil)
	articleID := arts[0].ID

	t1, _ := db.CreateTag("Go")
	t2, _ := db.CreateTag("DB")
	db.AddArticleTag(articleID, t1.ID)
	db.AddArticleTag(articleID, t2.ID)

	tags, err := db.ListTags()
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	// 两个标签，各计数 1
	byName := map[string]int64{}
	for _, tg := range tags {
		byName[tg.Name] = tg.ArticleCount
	}
	if byName["Go"] != 1 || byName["DB"] != 1 {
		t.Fatalf("unexpected counts: %+v", byName)
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
	if _, _ = db.CreateTag("b"); false {
	}
	if err := db.RenameTag(a.ID, "b"); err == nil {
		t.Fatal("expected conflict error when renaming to existing name")
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

	// 删除前有关联
	tags, _ := db.GetArticleTags(articleID)
	if len(tags) != 1 {
		t.Fatalf("expected 1 tag before delete, got %d", len(tags))
	}

	affected, err := db.DeleteTag(tag.ID)
	if err != nil || affected == 0 {
		t.Fatalf("delete tag: affected=%d err=%v", affected, err)
	}

	// 删除后关联清空，文章仍在
	tags, _ = db.GetArticleTags(articleID)
	if len(tags) != 0 {
		t.Fatalf("expected 0 tags after delete, got %d", len(tags))
	}
	if art, _ := db.GetArticleByID(articleID); art == nil {
		t.Fatal("article should still exist after tag deletion")
	}
}
```

- [ ] **Step 2: 运行测试验证失败**

Run: `go test ./internal/storage/ -run "TestCreateTag|TestGetTagByName|TestListTags|TestRenameTag|TestDeleteTag" -v`
Expected: FAIL（方法未定义）

- [ ] **Step 3: 实现 storage 方法**

`internal/storage/database.go`，在 `CreateCategory` 方法（`:1543`）之后新增：

```go
// CreateTag 创建标签，重名幂等：已存在则返回现有标签。
func (db *Database) CreateTag(name string) (model.Tag, error) {
	if name == "" {
		return model.Tag{}, errors.New("tag name cannot be empty")
	}

	// 先尝试插入
	result, err := db.conn.Exec(`INSERT INTO tags (name) VALUES (?)`, name)
	if err != nil {
		// UNIQUE 冲突 → 回查现有
		existing, qerr := db.GetTagByName(name)
		if qerr != nil || existing == nil {
			return model.Tag{}, err
		}
		return *existing, nil
	}
	id, _ := result.LastInsertId()
	tag, _ := db.GetTagByID(id)
	return *tag, nil
}

// GetTagByName 按名查标签，不存在返回 (nil, nil)。
func (db *Database) GetTagByName(name string) (*model.Tag, error) {
	row := db.conn.QueryRow(`SELECT id, name, created_at FROM tags WHERE name = ?`, name)
	return scanTag(row)
}

// GetTagByID 按 ID 查标签，不存在返回 (nil, nil)。
func (db *Database) GetTagByID(id int64) (*model.Tag, error) {
	row := db.conn.QueryRow(`SELECT id, name, created_at FROM tags WHERE id = ?`, id)
	return scanTag(row)
}

// scanTag 从单行扫描标签。
func scanTag(row *sql.Row) (*model.Tag, error) {
	var tag model.Tag
	var created sql.NullString
	if err := row.Scan(&tag.ID, &tag.Name, &created); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if created.Valid {
		if parsed, err := parseTime(created.String); err == nil {
			tag.CreatedAt = parsed
		}
	}
	return &tag, nil
}

// ListTags 列出所有标签及关联文章计数，按名称排序。
func (db *Database) ListTags() ([]model.Tag, error) {
	rows, err := db.conn.Query(`SELECT t.id, t.name, t.created_at, COUNT(at.article_id) AS cnt
		FROM tags t
		LEFT JOIN article_tags at ON at.tag_id = t.id
		GROUP BY t.id
		ORDER BY t.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var tag model.Tag
		var created sql.NullString
		if err := rows.Scan(&tag.ID, &tag.Name, &created, &tag.ArticleCount); err != nil {
			return nil, err
		}
		if created.Valid {
			if parsed, err := parseTime(created.String); err == nil {
				tag.CreatedAt = parsed
			}
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

// RenameTag 重命名标签。新名与现有标签撞名时返回错误。
func (db *Database) RenameTag(id int64, newName string) error {
	if newName == "" {
		return errors.New("tag name cannot be empty")
	}
	result, err := db.conn.Exec(`UPDATE tags SET name = ? WHERE id = ?`, newName, id)
	if err != nil {
		return fmt.Errorf("failed to rename tag: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("tag not found: %d", id)
	}
	return nil
}

// DeleteTag 删除标签，事务内先解除所有文章关联再删除标签本身。
// 返回被删除的关联行数。
func (db *Database) DeleteTag(id int64) (int64, error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`DELETE FROM article_tags WHERE tag_id = ?`, id)
	if err != nil {
		return 0, fmt.Errorf("failed to clear tag associations: %w", err)
	}
	affected, _ := res.RowsAffected()

	res, err = tx.Exec(`DELETE FROM tags WHERE id = ?`, id)
	if err != nil {
		return 0, fmt.Errorf("failed to delete tag: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return 0, fmt.Errorf("tag not found: %d", id)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return affected, nil
}
```

> 注：需确认文件顶部已 `import "database/sql"`。`parseTime` 为本文件已有辅助函数（`database.go:1612` 处已用）。如 `database.go` 未引入 `database/sql`，则在 import 块补 `"database/sql"`。

- [ ] **Step 4: 运行测试验证通过**

Run: `go test ./internal/storage/ -run "TestCreateTag|TestGetTagByName|TestListTags|TestRenameTag|TestDeleteTag" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/storage/database.go internal/storage/database_test.go
git commit -m "feat(tags): storage 标签 CRUD（Create/Get/List/Rename/Delete）"
```

---

## Task 3: Storage — 文章-标签关联方法

**Files:**
- Modify: `internal/storage/database.go`（紧接 Task 2 方法之后）
- Test: `internal/storage/database_test.go`

- [ ] **Step 1: 写失败测试**

```go
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
	// 重复加幂等，不报错
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
	// 再删无影响行，仍成功
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

	// 初始空
	if err := db.SetArticleTags(articleID, []int64{t1.ID, t2.ID}); err != nil {
		t.Fatalf("set1: %v", err)
	}
	tags, _ := db.GetArticleTags(articleID)
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
	// 全量替换为 t2,t3
	if err := db.SetArticleTags(articleID, []int64{t2.ID, t3.ID}); err != nil {
		t.Fatalf("set2: %v", err)
	}
	names := map[string]bool{}
	for _, tg := range mustGetTags(t, db, articleID) {
		names[tg.Name] = true
	}
	if !names["b"] || !names["c"] || names["a"] {
		t.Fatalf("unexpected tags after replace: %+v", names)
	}
	// 清空
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

// 辅助：测试用，断言获取成功
func mustGetTags(t *testing.T, db *storage.Database, articleID int64) []model.Tag {
	t.Helper()
	tags, err := db.GetArticleTags(articleID)
	if err != nil {
		t.Fatalf("get tags: %v", err)
	}
	return tags
}
```

> 注：`storage.Database` 引用需在测试文件中已 import（同包测试用 `*Database` 即可，本辅助函数放在同包 `database_test.go` 内，签名改为 `func mustGetTags(t *testing.T, db *Database, articleID int64) []model.Tag`）。

- [ ] **Step 2: 运行测试验证失败**

Run: `go test ./internal/storage/ -run "TestAddArticleTag|TestRemoveArticleTag|TestSetArticleTags|TestGetTagsForArticles" -v`
Expected: FAIL（方法未定义）

- [ ] **Step 3: 实现关联方法**

紧接 Task 2 方法之后新增：

```go
// AddArticleTag 给文章加标签，幂等（INSERT OR IGNORE）。
func (db *Database) AddArticleTag(articleID, tagID int64) error {
	_, err := db.conn.Exec(`INSERT OR IGNORE INTO article_tags (tag_id, article_id) VALUES (?, ?)`, tagID, articleID)
	return err
}

// RemoveArticleTag 移除文章标签关联，无影响行也算成功。
func (db *Database) RemoveArticleTag(articleID, tagID int64) error {
	_, err := db.conn.Exec(`DELETE FROM article_tags WHERE tag_id = ? AND article_id = ?`, tagID, articleID)
	return err
}

// SetArticleTags 全量替换文章标签：事务内删旧关联、插新关联。
func (db *Database) SetArticleTags(articleID int64, tagIDs []int64) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM article_tags WHERE article_id = ?`, articleID); err != nil {
		return fmt.Errorf("failed to clear article tags: %w", err)
	}
	for _, tid := range tagIDs {
		if _, err := tx.Exec(`INSERT OR IGNORE INTO article_tags (tag_id, article_id) VALUES (?, ?)`, tid, articleID); err != nil {
			return fmt.Errorf("failed to insert article tag: %w", err)
		}
	}
	return tx.Commit()
}

// GetArticleTags 取单篇文章的所有标签。
func (db *Database) GetArticleTags(articleID int64) ([]model.Tag, error) {
	rows, err := db.conn.Query(`SELECT t.id, t.name, t.created_at FROM tags t
		INNER JOIN article_tags at ON at.tag_id = t.id
		WHERE at.article_id = ? ORDER BY t.name`, articleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var tag model.Tag
		var created sql.NullString
		if err := rows.Scan(&tag.ID, &tag.Name, &created); err != nil {
			return nil, err
		}
		if created.Valid {
			if parsed, err := parseTime(created.String); err == nil {
				tag.CreatedAt = parsed
			}
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

// GetTagsForArticles 批量取多文章的标签，按 article_id 聚合，避免列表渲染 N+1。
// 返回 map[articleID][]Tag。空入参返回空 map。
func (db *Database) GetTagsForArticles(articleIDs []int64) (map[int64][]model.Tag, error) {
	result := map[int64][]model.Tag{}
	if len(articleIDs) == 0 {
		return result, nil
	}
	// 构造 IN (?, ?, ...)
	placeholders := make([]string, len(articleIDs))
	args := make([]interface{}, len(articleIDs))
	for i, id := range articleIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	query := fmt.Sprintf(`SELECT at.article_id, t.id, t.name, t.created_at FROM tags t
		INNER JOIN article_tags at ON at.tag_id = t.id
		WHERE at.article_id IN (%s) ORDER BY t.name`, strings.Join(placeholders, ","))
	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var articleID int64
		var tag model.Tag
		var created sql.NullString
		if err := rows.Scan(&articleID, &tag.ID, &tag.Name, &created); err != nil {
			return nil, err
		}
		if created.Valid {
			if parsed, err := parseTime(created.String); err == nil {
				tag.CreatedAt = parsed
			}
		}
		result[articleID] = append(result[articleID], tag)
	}
	return result, rows.Err()
}
```

- [ ] **Step 4: 运行测试验证通过**

Run: `go test ./internal/storage/ -run "TestAddArticleTag|TestRemoveArticleTag|TestSetArticleTags|TestGetTagsForArticles" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/storage/database.go internal/storage/database_test.go
git commit -m "feat(tags): storage 文章-标签关联（Add/Remove/Set/Get/批量Get）"
```

---

## Task 4: Storage — 列表查询接入标签过滤与装配

**Files:**
- Modify: `internal/storage/database.go:1457`（`ListArticlesWithFilters`）
- Modify: `internal/storage/database.go:1852` 附近（`SearchArticles` 与 `CountArticlesWithFilters`）
- Test: `internal/storage/database_test.go`

> 实现前先用 `grep -n "func (db \*Database) SearchArticles\|func (db \*Database) CountArticlesWithFilters" internal/storage/database.go` 定位这两个函数的确切位置与当前 WHERE 构建结构，按下面同样的 `EXISTS` 模式注入。

- [ ] **Step 1: 写失败测试**

```go
func TestListArticlesWithFilters_TagFilter(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	blog, _ := db.AddBlog(model.Blog{Name: "B", URL: "https://example.com"})
	db.AddArticlesBulk([]model.Article{
		{BlogID: blog.ID, Title: "With", URL: "https://example.com/w", HNStatus: model.HNStatusNotSearch},
		{BlogID: blog.ID, Title: "Without", URL: "https://example.com/wo", HNStatus: model.HNStatusNotSearch},
	})
	arts, _ := db.ListArticles(false, nil)
	taggedID := arts[0].ID // "With"
	tag, _ := db.CreateTag("Go")
	db.AddArticleTag(taggedID, tag.ID)

	// 按标签筛选：只返回带 "Go" 的那篇
	got, err := db.ListArticlesWithFilters(storage.ListFilterOptions{TagName: "Go"})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 || got[0].ID != taggedID {
		t.Fatalf("expected only tagged article, got %+v", got)
	}

	// 装配 Tags：被筛中的文章应带标签
	if len(got[0].Tags) != 1 || got[0].Tags[0].Name != "Go" {
		t.Fatalf("expected assembled Tags, got %+v", got[0].Tags)
	}
}

func TestSearchArticles_TagFilter(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	blog, _ := db.AddBlog(model.Blog{Name: "B", URL: "https://example.com"})
	db.AddArticlesBulk([]model.Article{
		{BlogID: blog.ID, Title: "Tagged", URL: "https://example.com/s1", HNStatus: model.HNStatusNotSearch},
		{BlogID: blog.ID, Title: "Plain", URL: "https://example.com/s2", HNStatus: model.HNStatusNotSearch},
	})
	arts, _ := db.ListArticles(false, nil)
	taggedID := arts[0].ID
	tag, _ := db.CreateTag("Go")
	db.AddArticleTag(taggedID, tag.ID)

	opts := model.SearchOptions{TagName: "Go", Limit: 20}
	got, err := db.SearchArticles(opts)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(got) != 1 || got[0].ID != taggedID {
		t.Fatalf("expected only tagged article, got %+v", got)
	}
	if len(got[0].Tags) != 1 {
		t.Fatalf("expected assembled tags, got %d", len(got[0].Tags))
	}
}
```

> 实现前确认 `SearchArticles` 返回类型与签名（很可能是 `([]model.ArticleWithBlog, error)`），并在该函数内部对结果批量装配 Tags。

- [ ] **Step 2: 运行测试验证失败**

Run: `go test ./internal/storage/ -run "TestListArticlesWithFilters_TagFilter|TestSearchArticles_TagFilter" -v`
Expected: FAIL

- [ ] **Step 3: 在 ListArticlesWithFilters 注入标签过滤与装配**

`database.go:1496`（`IsFavorited` 筛选块之后、日期筛选之前）新增标签过滤块：

```go
	// 标签筛选（EXISTS 子查询）
	if opts.TagName != "" {
		conditions = append(conditions, `EXISTS(SELECT 1 FROM article_tags at JOIN tags t ON t.id=at.tag_id WHERE at.article_id=a.id AND t.name=?)`)
		args = append(args, opts.TagName)
	}
```

在该函数 `return articles, rows.Err()` 之前（`articles` 扫描循环之后）装配 Tags：

```go
	// 批量装配标签，避免 N+1
	if len(articles) > 0 {
		ids := make([]int64, len(articles))
		for i, a := range articles {
			ids[i] = a.ID
		}
		tagMap, err := db.GetTagsForArticles(ids)
		if err != nil {
			return nil, err
		}
		for i := range articles {
			articles[i].Tags = tagMap[articles[i].ID]
		}
	}
```

- [ ] **Step 4: 在 SearchArticles 与 CountArticlesWithFilters 注入同样过滤**

定位两个函数（`grep`），在各自 WHERE 条件构建处加同样的 `EXISTS` 块。`SearchArticles` 在结果返回前加与 Step 3 相同的批量装配逻辑。

`CountArticlesWithFilters` 只需加过滤条件，无需装配 Tags（计数查询）。

- [ ] **Step 5: 运行测试验证通过**

Run: `go test ./internal/storage/ -run "TestListArticlesWithFilters_TagFilter|TestSearchArticles_TagFilter" -v`
Expected: PASS

- [ ] **Step 6: 运行全量 storage 测试确保无回归**

Run: `go test ./internal/storage/ -v`
Expected: PASS（含既有测试）

- [ ] **Step 7: Commit**

```bash
git add internal/storage/database.go internal/storage/database_test.go
git commit -m "feat(tags): 列表查询支持标签过滤与 Tags 批量装配"
```

---

## Task 5: CLI — output 格式化

**Files:**
- Create: `internal/cli/output/tag.go`
- Test: `internal/cli/output/tag_test.go`

> 对照 `internal/cli/output/category.go` 的 `FormatCategoryTable` / `FormatCategoryJSON` 与 `CategoryJSONOutput` 结构。

- [ ] **Step 1: 写失败测试**

```go
package output

import (
	"strings"
	"testing"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
)

func TestFormatTagTable(t *testing.T) {
	tags := []model.Tag{
		{ID: 1, Name: "Go", ArticleCount: 3},
		{ID: 2, Name: "DB", ArticleCount: 0},
	}
	out := FormatTagTable(tags, PaginationMeta{Total: 2, Count: 2})
	if !strings.Contains(out, "Go") || !strings.Contains(out, "DB") {
		t.Fatalf("expected names in table, got: %s", out)
	}
}

func TestFormatTagJSON(t *testing.T) {
	tags := []model.Tag{
		{ID: 1, Name: "Go", ArticleCount: 3},
	}
	out := FormatTagJSON(tags, PaginationMeta{Total: 1, Count: 1})
	if !strings.Contains(out, `"name":"Go"`) || !strings.Contains(out, `"article_count":3`) {
		t.Fatalf("unexpected json: %s", out)
	}
}
```

- [ ] **Step 2: 运行测试验证失败**

Run: `go test ./internal/cli/output/ -run "TestFormatTag" -v`
Expected: FAIL（函数未定义）

- [ ] **Step 3: 实现**

```go
// ABOUTME: Tag output formatters
// ABOUTME: Provides table and JSON formatting for tag list command
package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
)

// TagJSONOutput 用于 JSON 输出的简化标签结构
type TagJSONOutput struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	ArticleCount int64  `json:"article_count"`
}

// FormatTagTable 将标签列表格式化为表格输出
func FormatTagTable(tags []model.Tag, meta PaginationMeta) string {
	if len(tags) == 0 {
		return "没有标签\n"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("ID\t名称\t文章数\n"))
	for _, t := range tags {
		b.WriteString(fmt.Sprintf("%d\t%s\t%d\n", t.ID, t.Name, t.ArticleCount))
	}
	b.WriteString(fmt.Sprintf("\n共 %d 个标签\n", meta.Total))
	return b.String()
}

// FormatTagJSON 将标签列表格式化为 JSON 输出
func FormatTagJSON(tags []model.Tag, meta PaginationMeta) string {
	out := make([]TagJSONOutput, 0, len(tags))
	for _, t := range tags {
		out = append(out, TagJSONOutput{
			ID:           t.ID,
			Name:         t.Name,
			ArticleCount: t.ArticleCount,
		})
	}
	result := map[string]interface{}{
		"tags":  out,
		"total": meta.Total,
		"count": meta.Count,
	}
	b, _ := json.MarshalIndent(result, "", "  ")
	return string(b)
}
```

> 注：`PaginationMeta` 类型在 `output/types.go`（既有）。对照 category.go 确认其字段名与签名风格，若 category 的格式化函数不接收 meta 参数，则此处也去掉 meta 参数并相应调整测试。实现时以 `category.go` 实际签名为准对齐。

- [ ] **Step 4: 运行测试验证通过**

Run: `go test ./internal/cli/output/ -run "TestFormatTag" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/cli/output/tag.go internal/cli/output/tag_test.go
git commit -m "feat(tags): CLI output 标签格式化（table/json）"
```

---

## Task 6: CLI — tag 命令组（list/rename/delete）

**Files:**
- Create: `internal/cli/commands/tag.go`
- Modify: `internal/cli/commands/root.go:40`（注册 `NewTagCmd()`）
- Test: `internal/cli/commands/tag_test.go`

> 对照 `category.go:17` 命令组结构、`article.go:458` `runFavorite` 的 db 打开与错误处理模式。

- [ ] **Step 1: 写失败测试**

```go
package commands

import (
	"testing"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/cli/flags"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
)

func newTagTestDB(t *testing.T) *storage.Database {
	t.Helper()
	dbPath := t.TempDir() + "\\test.db"
	flags.SetDBPath(dbPath) // 若 flags 包提供覆盖入口；否则用环境变量
	db, err := storage.OpenDatabase(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
```

> 注：实现前确认 `flags` 包是否提供测试用 `SetDBPath` 入口（看 `internal/cli/flags`）。若无可覆盖入口，则测试直接调用 `runTagList` 等函数前手动设置 flag，或改为构造 cobra 命令并设置 `--db` flag（参照 `flags.DBPath()` 读取逻辑）。以实际 flags 包能力为准。下面的命令级测试以 `execute` 方式断言输出与副作用：

```go
func TestTagListCmd(t *testing.T) {
	db := newTagTestDB(t)
	db.CreateTag("Go")
	db.CreateTag("DB")

	out := captureStdout(t, func() {
		cmd := NewTagListCmd()
		runTagList(cmd, nil)
	})
	if !strings.Contains(out, "Go") || !strings.Contains(out, "DB") {
		t.Fatalf("expected both tags, got: %s", out)
	}
}
```

> `captureStdout` 为测试辅助（重定向 `os.Stdout` 或抽 `runTagList` 输出到 `io.Writer`）。若现有 CLI 测试已有捕获范式（见 `internal/cli/commands/*_test.go`），沿用之；若没有，则把 `runTagList` 改为返回 `(string, error)` 由测试直接断言——以既有 article 测试范式为准。

- [ ] **Step 2: 运行测试验证失败**

Run: `go test ./internal/cli/commands/ -run "TestTagList" -v`
Expected: FAIL

> 鉴于 CLI 测试范式需对照既有 `article_test.go`，实现 Step 3 前先读 `internal/cli/commands/article_test.go` 确认断言范式（是否用 `execute`/`Run`、是否捕获 stdout），并据此调整上面的测试骨架。这是本计划唯一需要实现时校对的范式点——以既有测试实际写法为准。

- [ ] **Step 3: 实现 tag.go**

```go
// ABOUTME: tag 子命令定义
// ABOUTME: 提供标签列表、重命名、删除功能
package commands

import (
	"fmt"
	"os"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/cli/flags"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/cli/output"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
	"github.com/spf13/cobra"
)

// NewTagCmd 创建 tag 命令组
func NewTagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "标签管理命令",
		Long: `标签管理命令组。

子命令：
  list           列出所有标签及文章计数
  rename <old> <new>  重命名标签
  delete <name>  删除标签（级联解除关联）`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	cmd.AddCommand(NewTagListCmd())
	cmd.AddCommand(NewTagRenameCmd())
	cmd.AddCommand(NewTagDeleteCmd())
	return cmd
}

// NewTagListCmd 创建 list 子命令
func NewTagListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有标签",
		Run:   runTagList,
	}
	cmd.Flags().String("format", "table", "输出格式 (table/json)")
	return cmd
}

func runTagList(cmd *cobra.Command, args []string) {
	db := openTagCmdDB()
	defer db.Close()

	tags, err := db.ListTags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取标签列表失败: %v\n", err)
		os.Exit(1)
	}
	meta := output.PaginationMeta{Total: int64(len(tags)), Count: len(tags)}
	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "json":
		fmt.Println(output.FormatTagJSON(tags, meta))
	default:
		fmt.Println(output.FormatTagTable(tags, meta))
	}
}

// NewTagRenameCmd 创建 rename 子命令
func NewTagRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <oldName> <newName>",
		Short: "重命名标签",
		Args:  cobra.ExactArgs(2),
		Run:   runTagRename,
	}
}

func runTagRename(cmd *cobra.Command, args []string) {
	db := openTagCmdDB()
	defer db.Close()

	oldName, newName := args[0], args[1]
	tag, err := db.GetTagByName(oldName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查询标签失败: %v\n", err)
		os.Exit(1)
	}
	if tag == nil {
		fmt.Fprintf(os.Stderr, "标签 '%s' 不存在\n", oldName)
		os.Exit(1)
	}
	if err := db.RenameTag(tag.ID, newName); err != nil {
		fmt.Fprintf(os.Stderr, "重命名标签失败: %v\n", err)
		os.Exit(1)
	}
	log.Printf("tag rename: '%s' -> '%s'", oldName, newName)
	fmt.Printf("已重命名标签: %s -> %s\n", oldName, newName)
}

// NewTagDeleteCmd 创建 delete 子命令
func NewTagDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "删除标签（级联解除关联）",
		Args:  cobra.ExactArgs(1),
		Run:   runTagDelete,
	}
}

func runTagDelete(cmd *cobra.Command, args []string) {
	db := openTagCmdDB()
	defer db.Close()

	name := args[0]
	tag, err := db.GetTagByName(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查询标签失败: %v\n", err)
		os.Exit(1)
	}
	if tag == nil {
		fmt.Fprintf(os.Stderr, "标签 '%s' 不存在\n", name)
		os.Exit(1)
	}
	affected, err := db.DeleteTag(tag.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "删除标签失败: %v\n", err)
		os.Exit(1)
	}
	log.Printf("tag delete: '%s' (解除 %d 关联)", name, affected)
	fmt.Printf("已删除标签: %s（解除 %d 篇文章关联）\n", name, affected)
}

// openTagCmdDB 复用统一的 db 打开逻辑（与 article.go 一致）
func openTagCmdDB() *storage.Database {
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
	return db
}
```

> 注：`log` 包需 import（`log`）。若既有 CLI 命令不使用 `log.Printf` 而是仅 `fmt.Fprintf`，去掉 log 调用以对齐风格——以 `article.go` 实际是否用 log 为准（article.go 未见 log，故此处去掉 log.Printf 行，仅保留输出）。

- [ ] **Step 4: 注册 tag 命令组**

`internal/cli/commands/root.go`，在 `rootCmd.AddCommand(NewCategoryCmd())`（`:40`）之后追加：

```go
	rootCmd.AddCommand(NewTagCmd())
```

- [ ] **Step 5: 运行测试验证通过**

Run: `go test ./internal/cli/commands/ -run "TestTagList|TestTagRename|TestTagDelete" -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/cli/commands/tag.go internal/cli/commands/tag_test.go internal/cli/commands/root.go
git commit -m "feat(tags): CLI tag 命令组（list/rename/delete）"
```

---

## Task 7: CLI — article tag/untag 子命令与 list --tag

**Files:**
- Modify: `internal/cli/commands/article.go`（`NewArticleCmd` 加子命令；`NewListCmd` 加 flag；`runList` 读 flag；新增两个 run 函数）
- Test: `internal/cli/commands/article_test.go`

> 对照 `article.go:424` `NewFavoriteCmd` / `:458` `runFavorite` 模式逐字复刻，仅改 storage 调用。

- [ ] **Step 1: 写失败测试**

> 沿用 Task 6 确认的 CLI 测试范式。核心断言：`article tag <id> Go` 后该文带 "Go" 标签且标签存在；`article untag <id> Go` 后该文无此标签但标签仍在；`article list --tag Go` 只返回带该标签的文章。

- [ ] **Step 2: 运行测试验证失败**

Run: `go test ./internal/cli/commands/ -run "TestArticleTag|TestArticleUntag|TestArticleListCmd_TagFilter" -v`
Expected: FAIL

- [ ] **Step 3: 在 NewArticleCmd 注册子命令**

`article.go:44-45`（`NewFavoriteCmd`/`NewUnfavoriteCmd` 之后）追加：

```go
	cmd.AddCommand(NewArticleTagCmd())
	cmd.AddCommand(NewArticleUntagCmd())
```

同步更新 `article.go:30-37` 的 Long 帮助文本，追加 `tag`/`untag` 说明与 `list --tag`。

- [ ] **Step 4: 在 NewListCmd 加 --tag flag**

`article.go` `NewListCmd` 的 flags 区（`:96` 附近）追加：

```go
	cmd.Flags().String("tag", "", "标签名称筛选")
```

并在 `NewListCmd` 的 Long 帮助文本补 `--tag <name>` 说明。

- [ ] **Step 5: 在 runList 读取 --tag**

`article.go:234-239`（`--favorited` 处理之后）追加：

```go
	// 标签名称筛选
	tagName, _ := cmd.Flags().GetString("tag")
	if tagName != "" {
		opts.TagName = tagName
	}
```

- [ ] **Step 6: 实现 article tag / untag 子命令**

`article.go` 末尾新增（逐字仿 `NewFavoriteCmd`/`runFavorite`，`:422-504`）：

```go
// NewArticleTagCmd 创建 article tag 子命令
func NewArticleTagCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tag <id> <name>",
		Short: "给文章打标签",
		Long: `给指定文章打标签，标签不存在则自动创建。

示例：
  blogwatcher article tag 1 Go  # 给文章 1 打标签 "Go"`,
		Args: cobra.ExactArgs(2),
		Run:  runArticleTag,
	}
}

// NewArticleUntagCmd 创建 article untag 子命令
func NewArticleUntagCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "untag <id> <name>",
		Short: "移除文章标签",
		Long: `移除指定文章上的某标签（不影响标签本身）。

示例：
  blogwatcher article untag 1 Go  # 移除文章 1 的 "Go" 标签`,
		Args: cobra.ExactArgs(2),
		Run:  runArticleUntag,
	}
}

func runArticleTag(cmd *cobra.Command, args []string) {
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
	name := args[1]
	if len(name) > MaxInputLength {
		fmt.Fprintf(os.Stderr, "标签名称长度超过最大值 %d\n", MaxInputLength)
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

	tag, err := db.CreateTag(name) // 幂等
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建标签失败: %v\n", err)
		os.Exit(1)
	}
	if err := db.AddArticleTag(id, tag.ID); err != nil {
		fmt.Fprintf(os.Stderr, "打标签失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("已为文章 %d 添加标签: %s\n", id, tag.Name)
}

func runArticleUntag(cmd *cobra.Command, args []string) {
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
	name := args[1]

	article, err := db.GetArticleByID(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查询文章失败: %v\n", err)
		os.Exit(1)
	}
	if article == nil {
		fmt.Fprintf(os.Stderr, "文章 %d 不存在\n", id)
		os.Exit(1)
	}

	tag, err := db.GetTagByName(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查询标签失败: %v\n", err)
		os.Exit(1)
	}
	if tag == nil {
		fmt.Fprintf(os.Stderr, "标签 '%s' 不存在\n", name)
		os.Exit(1)
	}
	if err := db.RemoveArticleTag(id, tag.ID); err != nil {
		fmt.Fprintf(os.Stderr, "移除标签失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("已移除文章 %d 的标签: %s\n", id, tag.Name)
}
```

- [ ] **Step 7: 运行测试验证通过**

Run: `go test ./internal/cli/commands/ -run "TestArticleTag|TestArticleUntag|TestArticleListCmd_TagFilter" -v`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add internal/cli/commands/article.go internal/cli/commands/article_test.go
git commit -m "feat(tags): CLI article tag/untag 子命令与 list --tag 筛选"
```

---

## Task 8: Server — 路由、parseSearchOptions、列表装配

**Files:**
- Modify: `internal/server/routes.go`
- Modify: `internal/server/handlers.go:588`（`parseSearchOptions`）
- Modify: `internal/server/handlers.go`（`handleArticleList` 与 `renderUpdatedArticleCard` 装配 Tags）
- Test: `internal/server/handlers_test.go`

- [ ] **Step 1: 写失败测试**

```go
func TestParseSearchOptions_TagFilter(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/articles?filter=tag&tag=Go", nil)
	opts, filter, _ := parseSearchOptions(req)
	if filter != "tag" {
		t.Fatalf("filter = %q, want 'tag'", filter)
	}
	if opts.TagName != "Go" {
		t.Fatalf("TagName = %q, want 'Go'", opts.TagName)
	}
}
```

> 注：`parseSearchOptions` 为包内函数，测试在同包 `handlers_test.go` 可直接调用。

- [ ] **Step 2: 运行测试验证失败**

Run: `go test ./internal/server/ -run TestParseSearchOptions_TagFilter -v`
Expected: FAIL

- [ ] **Step 3: 改造 parseSearchOptions**

`handlers.go:597` `switch filter` 中，在 `default` 之前新增：

```go
	case "tag":
		opts.TagName = r.URL.Query().Get("tag")
		opts.IsRead = nil // 标签筛选不强制已读状态
```

同时在该函数末尾 `return opts, filter, currentBlogID` 之前，把 `CurrentTag` 透出——但函数签名只返回三值。改为在调用处（`handleArticleList`）直接从 `r.URL.Query().Get("tag")` 取标签名放进模板数据。无需改 `parseSearchOptions` 签名。

- [ ] **Step 4: 注册路由**

`internal/server/routes.go`，在 category 管理路由之后（文件末尾）追加：

```go
	// Tag management
	s.mux.HandleFunc("GET /tags", s.handleTagsPage)
	s.mux.HandleFunc("GET /tags/list", s.handleTagsListPartial)
	s.mux.HandleFunc("POST /tags", s.handleTagCreate)
	s.mux.HandleFunc("PUT /tags/{id}", s.handleTagRename)
	s.mux.HandleFunc("DELETE /tags/{id}", s.handleTagDelete)

	// Article tags
	s.mux.HandleFunc("GET /articles/{id}/tags", s.handleArticleTagsPage)
	s.mux.HandleFunc("GET /articles/{id}/tags/edit", s.handleArticleTagsEditPartial)
	s.mux.HandleFunc("POST /articles/{id}/tags", s.handleArticleTagAdd)
	s.mux.HandleFunc("DELETE /articles/{id}/tags/{tagID}", s.handleArticleTagRemove)
	s.mux.HandleFunc("POST /articles/{id}/tags/save", s.handleArticleTagSave)
```

- [ ] **Step 5: 在列表渲染装配 Tags（handleArticleList）**

定位 `handleArticleList`（`handlers.go:55` / `:98` / `:404` / `:459` 等多处调用 `parseSearchOptions`）。`SearchArticles`（Task 4 已在其内部装配 Tags）返回的 `ArticleWithBlog` 已带 `Tags`，故列表页无需额外处理——确认调用链确实经过 Task 4 改造后的 `SearchArticles` 即可。

`renderUpdatedArticleCard`（`handlers.go:345-383`）需为单文卡片装配 Tags：在构造 `articleWithBlog` 之后、渲染之前，调 `GetArticleTags(id)` 填充：

```go
	// 装配标签用于卡片 chips 渲染
	if tags, err := s.db.GetArticleTags(id); err == nil {
		articleWithBlog.Tags = tags
	}
```

插入位置：`handlers.go:375`（`BlogURL: blog.URL,` 之后、`data := map[string]interface{}{` 之前）。

- [ ] **Step 6: 运行测试验证通过**

Run: `go test ./internal/server/ -run TestParseSearchOptions_TagFilter -v`
Expected: PASS

- [ ] **Step 7: 运行全量 server 测试确保无回归**

Run: `go test ./internal/server/ -v`
Expected: PASS（编译会因新 handler 未实现而失败——本 Task 仅注册路由则编译失败。改为：本 Task 只改 parseSearchOptions + renderUpdatedArticleCard 装配，**路由注册放到 Task 9/10 实现 handler 之后**。）

> 调整：将 Step 4（路由注册）移到 Task 10 末尾，随 handler 实现一起提交，避免编译失败。本 Task 只完成 Step 3（parseSearchOptions）与 Step 5（装配 Tags）。

- [ ] **Step 8: Commit**

```bash
git add internal/server/handlers.go internal/server/handlers_test.go
git commit -m "feat(tags): parseSearchOptions 支持 filter=tag 并装配卡片 Tags"
```

---

## Task 9: Server — 标签 CRUD Handler

**Files:**
- Modify: `internal/server/handlers.go`（新增 5 个 handler）
- Modify: `internal/server/routes.go`（注册标签 CRUD 5 条路由）
- Test: `internal/server/handlers_test.go`

- [ ] **Step 1: 写失败测试**

```go
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
	// 验证落库
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
	got, _ := db.GetTagByID(tag.ID)
	if got.Name != "new" {
		t.Fatalf("expected 'new', got %q", got.Name)
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
	if got, _ := db.GetTagByID(tag.ID); got != nil {
		t.Fatal("tag should be deleted")
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
```

- [ ] **Step 2: 运行测试验证失败**

Run: `go test ./internal/server/ -run "TestHandleTagCreate|TestHandleTagRename|TestHandleTagDelete" -v`
Expected: FAIL（handler 未实现）

- [ ] **Step 3: 实现 CRUD handler**

`handlers.go` 末尾新增：

```go
// handleTagsPage 渲染全局标签管理页。
func (s *Server) handleTagsPage(w http.ResponseWriter, r *http.Request) {
	tags, err := s.db.ListTags()
	if err != nil {
		log.Printf("Error listing tags: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{
		"Tags": tags,
	}
	s.renderTemplate(w, "tags.gohtml", data)
}

// handleTagsListPartial 渲染标签列表片段（侧边栏 Tags 分区 / 工具栏下拉用）。
func (s *Server) handleTagsListPartial(w http.ResponseWriter, r *http.Request) {
	tags, err := s.db.ListTags()
	if err != nil {
		log.Printf("Error listing tags partial: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	currentTag := r.URL.Query().Get("tag")
	data := map[string]interface{}{
		"Tags":        tags,
		"CurrentTag":  currentTag,
	}
	s.renderTemplate(w, "tags-list-partial.gohtml", data) // 见 Task 11：用 sidebar 内嵌片段或独立 partial
}
```

> 注：`handleTagsListPartial` 渲染的片段模板名以 Task 11 实际创建的 partial 名为准（建议 `tags-list.gohtml`），此处与 Task 11 对齐。

```go
// handleTagCreate 创建标签并返回管理页片段。
func (s *Server) handleTagCreate(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" || len(name) > 50 {
		http.Error(w, "Invalid tag name", http.StatusBadRequest)
		return
	}
	if _, err := s.db.CreateTag(name); err != nil {
		log.Printf("Error creating tag '%s': %v", name, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Trigger", "articleListUpdated")
	// 重渲染管理页
	s.handleTagsPage(w, r)
}

// handleTagRename 重命名标签。
func (s *Server) handleTagRename(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" || len(name) > 50 {
		http.Error(w, "Invalid tag name", http.StatusBadRequest)
		return
	}
	if err := s.db.RenameTag(id, name); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "constraint") {
			http.Error(w, "Tag name already exists", http.StatusConflict)
			return
		}
		log.Printf("Error renaming tag %d: %v", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Trigger", "articleListUpdated")
	s.handleTagsPage(w, r)
}

// handleTagDelete 删除标签（级联解除关联）。
func (s *Server) handleTagDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}
	if _, err := s.db.DeleteTag(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Tag not found", http.StatusNotFound)
			return
		}
		log.Printf("Error deleting tag %d: %v", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Trigger", "articleListUpdated")
	s.handleTagsPage(w, r)
}
```

- [ ] **Step 4: 注册标签 CRUD 路由**

`routes.go` 末尾追加：

```go
	// Tag management
	s.mux.HandleFunc("GET /tags", s.handleTagsPage)
	s.mux.HandleFunc("GET /tags/list", s.handleTagsListPartial)
	s.mux.HandleFunc("POST /tags", s.handleTagCreate)
	s.mux.HandleFunc("PUT /tags/{id}", s.handleTagRename)
	s.mux.HandleFunc("DELETE /tags/{id}", s.handleTagDelete)
```

- [ ] **Step 5: 运行测试验证通过**

Run: `go test ./internal/server/ -run "TestHandleTagCreate|TestHandleTagRename|TestHandleTagDelete|TestHandleTagRename_Conflict" -v`
Expected: PASS

> `handleTagsPage`/`handleTagsListPartial` 依赖 `tags.gohtml`/`tags-list.gohtml` 模板（Task 11）。若测试触发模板渲染失败，先把这两个 handler 的渲染临时改为 `w.WriteHeader(200); return`，待 Task 11 模板就绪后恢复——或调整 Task 顺序：先做 Task 11 模板再做本 Task。**建议实现时将 Task 9 与 Task 11 顺序对调**（先建空模板骨架再实现 handler），以避免编译/渲染缺口。下文以"模板与 handler 同一提交"为准。

- [ ] **Step 6: Commit**

```bash
git add internal/server/handlers.go internal/server/routes.go internal/server/handlers_test.go
git commit -m "feat(tags): server 标签 CRUD handler 与路由"
```

---

## Task 10: Server — 文章标签 Handler

**Files:**
- Modify: `internal/server/handlers.go`（5 个 handler）
- Modify: `internal/server/routes.go`（注册 5 条路由）
- Test: `internal/server/handlers_test.go`

- [ ] **Step 1: 写失败测试**

```go
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
	// 触发了 articleListUpdated
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
}
```

- [ ] **Step 2: 运行测试验证失败**

Run: `go test ./internal/server/ -run "TestHandleArticleTagAdd|TestHandleArticleTagRemove|TestHandleArticleTagSave|TestHandleArticleTags_NonExistentArticle" -v`
Expected: FAIL

- [ ] **Step 3: 实现文章标签 handler**

`handlers.go` 末尾新增（`handleNote` 之后区域）：

```go
// handleArticleTagsPage 渲染单文标签管理页（全量保存）。
func (s *Server) handleArticleTagsPage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}
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
	allTags, _ := s.db.ListTags()
	current, _ := s.db.GetArticleTags(id)
	currentIDs := map[int64]bool{}
	for _, t := range current {
		currentIDs[t.ID] = true
	}
	data := map[string]interface{}{
		"ID":         id,
		"Title":      article.Title,
		"URL":        article.URL,
		"AllTags":    allTags,
		"CurrentIDs": currentIDs,
	}
	s.renderTemplate(w, "article-tags.gohtml", data)
}

// handleArticleTagsEditPartial 渲染卡片弹层片段。
func (s *Server) handleArticleTagsEditPartial(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}
	allTags, _ := s.db.ListTags()
	current, _ := s.db.GetArticleTags(id)
	currentIDs := map[int64]bool{}
	for _, t := range current {
		currentIDs[t.ID] = true
	}
	data := map[string]interface{}{
		"ID":         id,
		"AllTags":    allTags,
		"Current":    current,
		"CurrentIDs": currentIDs,
	}
	s.renderTemplate(w, "article-tags-edit.gohtml", data)
}

// handleArticleTagAdd 增量加标签（name，不存在自动建）。
func (s *Server) handleArticleTagAdd(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" || len(name) > 50 {
		http.Error(w, "Invalid tag name", http.StatusBadRequest)
		return
	}
	article, err := s.db.GetArticleByID(id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if article == nil {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}
	tag, err := s.db.CreateTag(name)
	if err != nil {
		log.Printf("Error creating tag '%s': %v", name, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if err := s.db.AddArticleTag(id, tag.ID); err != nil {
		log.Printf("Error adding tag %d to article %d: %v", tag.ID, id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Trigger", "articleListUpdated")
	s.renderArticleTagsEditPartial(w, r, id)
}

// handleArticleTagRemove 增量移除标签。
func (s *Server) handleArticleTagRemove(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}
	tagID, err := strconv.ParseInt(r.PathValue("tagID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}
	if err := s.db.RemoveArticleTag(id, tagID); err != nil {
		log.Printf("Error removing tag %d from article %d: %v", tagID, id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Trigger", "articleListUpdated")
	s.renderArticleTagsEditPartial(w, r, id)
}

// handleArticleTagSave 全量替换文章标签。
func (s *Server) handleArticleTagSave(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	var tagIDs []int64
	for _, raw := range r.Form["tag_ids"] {
		if tid, err := strconv.ParseInt(raw, 10, 64); err == nil {
			tagIDs = append(tagIDs, tid)
		}
	}
	if err := s.db.SetArticleTags(id, tagIDs); err != nil {
		log.Printf("Error saving tags for article %d: %v", id, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("HX-Trigger", "articleListUpdated")
	// 重渲染管理页
	s.handleArticleTagsPage(w, r)
}

// renderArticleTagsEditPartial 抽取：渲染弹层片段（避免重复 PathValue 解析）。
func (s *Server) renderArticleTagsEditPartial(w http.ResponseWriter, r *http.Request, id int64) {
	allTags, _ := s.db.ListTags()
	current, _ := s.db.GetArticleTags(id)
	currentIDs := map[int64]bool{}
	for _, t := range current {
		currentIDs[t.ID] = true
	}
	data := map[string]interface{}{
		"ID":         id,
		"AllTags":    allTags,
		"Current":    current,
		"CurrentIDs": currentIDs,
	}
	s.renderTemplate(w, "article-tags-edit.gohtml", data)
}
```

> 重构：`handleArticleTagsEditPartial` 调用 `renderArticleTagsEditPartial`，二者等价，保留 handler 作路由入口。

- [ ] **Step 4: 注册文章标签路由**

`routes.go` 末尾追加：

```go
	// Article tags
	s.mux.HandleFunc("GET /articles/{id}/tags", s.handleArticleTagsPage)
	s.mux.HandleFunc("GET /articles/{id}/tags/edit", s.handleArticleTagsEditPartial)
	s.mux.HandleFunc("POST /articles/{id}/tags", s.handleArticleTagAdd)
	s.mux.HandleFunc("DELETE /articles/{id}/tags/{tagID}", s.handleArticleTagRemove)
	s.mux.HandleFunc("POST /articles/{id}/tags/save", s.handleArticleTagSave)
```

- [ ] **Step 5: 运行测试验证通过**

Run: `go test ./internal/server/ -run "TestHandleArticleTagAdd|TestHandleArticleTagRemove|TestHandleArticleTagSave|TestHandleArticleTags_NonExistentArticle" -v`
Expected: PASS（需 Task 11 模板就绪后渲染才不报错——见 Task 9 Step 5 同样约束）

- [ ] **Step 6: 运行全量 server 测试**

Run: `go test ./internal/server/ -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/server/handlers.go internal/server/routes.go internal/server/handlers_test.go
git commit -m "feat(tags): server 文章标签 handler 与路由"
```

---

## Task 11: UI — 模板与样式

**Files:**
- Create: `assets/templates/pages/tags.gohtml`
- Create: `assets/templates/pages/article-tags.gohtml`
- Create: `assets/templates/partials/article-tags-edit.gohtml`
- Create: `assets/templates/partials/tags-list.gohtml`
- Modify: `assets/templates/partials/sidebar.gohtml`
- Modify: `assets/templates/partials/article-items.gohtml`
- Modify: `assets/templates/partials/article-list.gohtml`
- Modify: `assets/static/styles.css`

> 前置：确认模板加载机制。`renderTemplate(w, name, data)` 如何解析模板名→文件（grep `renderTemplate` 与 `template.ParseFS`/`New` 的注册）。新模板文件须加入模板注册集，否则渲染时报"未定义"。实现时先 grep `tmpl.Execute` / `template.New` / `ParseFS` 确认注册方式（很可能用 `embed` + `template.ParseFS(fs, "templates/*.gohtml", "templates/**/*.gohtml")` 一次性加载，则新建文件自动纳入）。

- [ ] **Step 1: 新建全局标签管理页 tags.gohtml**

`assets/templates/pages/tags.gohtml`（仿 `note.gohtml` 的独立页面骨架）：

```html
{{define "tags.gohtml"}}
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>标签管理 - BlogWatcher</title>
    <link rel="stylesheet" href="/static/styles.css">
</head>
<body>
    <div class="tags-page">
        <header class="tags-header">
            <h1>标签管理</h1>
        </header>
        <main class="tags-content">
            {{range .Tags}}
            <div class="tag-row" id="tag-{{.ID}}">
                <form class="tag-rename-form"
                      hx-put="/tags/{{.ID}}"
                      hx-target="#tag-{{.ID}}"
                      hx-swap="outerHTML">
                    <input type="text" name="name" value="{{.Name}}" maxlength="50" class="tag-name-input">
                    <button type="submit" class="action-btn">重命名</button>
                </form>
                <span class="tag-count">{{.ArticleCount}} 篇</span>
                <button class="action-btn action-btn-danger"
                        hx-delete="/tags/{{.ID}}"
                        hx-target="#tag-{{.ID}}"
                        hx-swap="outerHTML"
                        hx-confirm="确认删除标签「{{.Name}}」？将解除所有关联。"
                        onclick="event.stopPropagation();">
                    删除
                </button>
            </div>
            {{else}}
            <p class="tags-empty">暂无标签</p>
            {{end}}
            <form class="tag-create-form" hx-post="/tags" hx-target=".tags-content" hx-swap="innerHTML">
                <input type="text" name="name" placeholder="新标签名" maxlength="50" class="tag-name-input">
                <button type="submit" class="action-btn">新建</button>
            </form>
        </main>
    </div>
</body>
</html>
{{end}}
```

- [ ] **Step 2: 新建单文标签管理页 article-tags.gohtml**

`assets/templates/pages/article-tags.gohtml`：

```html
{{define "article-tags.gohtml"}}
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - 标签 - BlogWatcher</title>
    <link rel="stylesheet" href="/static/styles.css">
</head>
<body>
    <div class="article-tags-page">
        <header class="article-tags-header">
            <h1>{{.Title}}</h1>
            <a href="{{.URL}}" target="_blank" rel="noopener noreferrer">查看原文 ↗</a>
        </header>
        <main class="article-tags-content">
            <form hx-post="/articles/{{.ID}}/tags/save"
                  hx-target=".article-tags-content"
                  hx-swap="innerHTML">
                <div class="tag-chips-selectable">
                    {{range .AllTags}}
                    <label class="tag-chip-selectable{{if index $.CurrentIDs .ID}} active{{end}}">
                        <input type="checkbox" name="tag_ids" value="{{.ID}}"
                               {{if index $.CurrentIDs .ID}}checked{{end}}>
                        <span>{{.Name}}</span>
                    </label>
                    {{end}}
                </div>
                <button type="submit" class="action-btn">保存</button>
            </form>
        </main>
    </div>
</body>
</html>
{{end}}
```

- [ ] **Step 3: 新建卡片弹层片段 article-tags-edit.gohtml**

`assets/templates/partials/article-tags-edit.gohtml`：

```html
{{define "article-tags-edit.gohtml"}}
<div class="tag-popover" id="tag-popover-{{.ID}}">
    <div class="tag-popover-current">
        {{range .Current}}
        <span class="tag-chip">
            {{.Name}}
            <button class="tag-chip-remove"
                    hx-delete="/articles/{{$.ID}}/tags/{{.ID}}"
                    hx-target="#tag-popover-{{$.ID}}"
                    hx-swap="outerHTML"
                    onclick="event.stopPropagation();">×</button>
        </span>
        {{else}}
        <span class="tag-popover-empty">暂无标签</span>
        {{end}}
    </div>
    <form class="tag-popover-form"
          hx-post="/articles/{{.ID}}/tags"
          hx-target="#tag-popover-{{.ID}}"
          hx-swap="outerHTML">
        <input type="text" name="name" list="tag-options-{{.ID}}" placeholder="选择或输入标签" autocomplete="off">
        <datalist id="tag-options-{{.ID}}">
            {{range .AllTags}}{{if not (index $.CurrentIDs .ID)}}
            <option value="{{.Name}}">{{.Name}}</option>
            {{end}}{{end}}
        </datalist>
        <button type="submit" class="action-btn">添加</button>
    </form>
</div>
{{end}}
```

- [ ] **Step 4: 新建标签列表片段 tags-list.gohtml（侧边栏/工具栏共用）**

`assets/templates/partials/tags-list.gohtml`：

```html
{{define "tags-list.gohtml"}}
{{range .Tags}}
<a href="/articles?filter=tag&tag={{.Name}}"
   hx-get="/articles?filter=tag&tag={{.Name}}"
   hx-target="#main-content"
   hx-push-url="true"
   class="nav-link tag-nav{{if eq .CurrentTag .Name}} active{{end}}">
    <span>#{{.Name}}</span>
    <span class="tag-nav-count">{{.ArticleCount}}</span>
</a>
{{else}}
<div class="tags-empty-hint">暂无标签</div>
{{end}}
{{end}}
```

> 注：`CurrentTag` 需在侧边栏根模板数据中透传。检查 `handleArticleList`/`handleIndex` 的模板数据是否含 `CurrentTag`，若无则补：`"CurrentTag": r.URL.Query().Get("tag")`。同理侧边栏 `sidebar.gohtml` 渲染时需能访问 `CurrentTag`——若 sidebar 由 base 模板统一渲染，则需在所有页面 handler 的数据里带 `CurrentTag`（默认 `""`）。

- [ ] **Step 5: 修改 sidebar.gohtml 加入 Tags 分区**

`assets/templates/partials/sidebar.gohtml`，在 Subscriptions 分区之前（`:73` 之前）插入：

```html
    {{/* Tags section */}}
    <div class="tags-section">
        <div class="nav-section-title">Tags</div>
        <div id="tag-list"
             hx-get="/tags/list"
             hx-trigger="load, articleListUpdated from:body"
             hx-swap="innerHTML">
        </div>
    </div>
```

- [ ] **Step 6: 修改 article-items.gohtml 加标签按钮与 chips**

`assets/templates/partials/article-items.gohtml`：

在 `.article-meta` 的 `<span class="article-id"...>` 之后（`:52` 之后）追加 chips 渲染：

```html
            {{range .Tags}}
            <span class="tag-chip tag-chip-sm">#{{.Name}}</span>
            {{end}}
```

在 `.article-actions` 内、favorite 按钮之前（`:55` 之前）加标签按钮：

```html
        <button class="action-btn"
                hx-get="/articles/{{.ID}}/tags/edit"
                hx-target="#article-{{.ID}}"
                hx-swap="beforeend"
                title="标签"
                onclick="event.stopPropagation();">
            <span class="action-btn-label">Tag</span>
        </button>
```

> 注：`hx-swap="beforeend"` 把弹层片段追加到卡片内；片段根容器 `.tag-popover` 用绝对定位浮于卡片上方。如需更精确的定位，改为 `hx-target="#tag-popover-anchor-{{.ID}}"` 并在卡片内预留锚点——以实现时调试为准。

- [ ] **Step 7: 修改 article-list.gohtml 工具栏加标签下拉**

定位 `assets/templates/partials/article-list.gohtml` 的工具栏区域，加：

```html
<select name="tag" hx-get="/articles" hx-include hx-push-url="true" hx-target="#main-content" hx-swap="innerHTML">
    <option value="">所有标签</option>
    {{/* 选项由 /tags/list 注入或页面初始数据渲染 */}}
</select>
```

> 实现前读 `article-list.gohtml` 确认工具栏现有结构与 `hx-include` 包含的表单字段，使标签下拉能与其他筛选共存。选项若需动态填充，可在 `handleArticleList` 数据中带 `AllTags` 并在模板 `{{range}}`。

- [ ] **Step 8: 新增样式 styles.css**

`assets/static/styles.css` 末尾追加：

```css
/* Tag chips */
.tag-chip {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 8px;
    border-radius: 12px;
    font-size: 0.8em;
    background: var(--chip-bg, #e5e5e5);
    color: var(--chip-fg, #333);
}
.tag-chip-sm { font-size: 0.7em; padding: 1px 6px; }
.tag-chip-remove {
    border: none;
    background: transparent;
    cursor: pointer;
    color: inherit;
    opacity: 0.6;
    padding: 0;
}
.tag-chip-remove:hover { opacity: 1; }

/* Tag popover (card inline) */
.tag-popover {
    position: absolute;
    z-index: 50;
    min-width: 220px;
    padding: 8px;
    border-radius: 8px;
    background: var(--popover-bg, #fff);
    box-shadow: 0 4px 16px rgba(0,0,0,0.15);
    border: 1px solid var(--border, #ddd);
}
.tag-popover-current { display: flex; flex-wrap: wrap; gap: 4px; margin-bottom: 6px; }
.tag-popover-empty { font-size: 0.8em; opacity: 0.6; }
.tag-popover-form { display: flex; gap: 4px; }

/* Tag management page */
.tag-row { display: flex; align-items: center; gap: 8px; padding: 4px 0; }
.tag-name-input { flex: 0 0 auto; }
.tag-count { font-size: 0.8em; opacity: 0.7; }

/* Sidebar tags */
.tags-section { margin: 8px 0; }
.tag-nav { display: flex; justify-content: space-between; }
.tag-nav-count { opacity: 0.6; font-size: 0.8em; }

/* Article tags page selectable chips */
.tag-chips-selectable { display: flex; flex-wrap: wrap; gap: 6px; }
.tag-chip-selectable { cursor: pointer; }
.tag-chip-selectable input { display: none; }
.tag-chip-selectable.active { background: var(--accent, #3b82f6); color: #fff; }

/* Dark mode overrides */
.dark .tag-chip { background: #333; color: #ddd; }
.dark .tag-popover { background: #1e1e1e; border-color: #444; }
```

- [ ] **Step 9: 编译验证模板注册**

Run: `go build ./...`
Expected: 编译通过

Run: `go test ./internal/server/ -v`
Expected: PASS（含 Task 9/10 渲染测试）

- [ ] **Step 10: Commit**

```bash
git add assets/templates assets/static/styles.css
git commit -m "feat(tags): UI 模板与样式（管理页/弹层/侧边栏/卡片chips）"
```

---

## Task 12: 部署与验证

**Files:** 无（执行部署流程）

> 遵循 `CLAUDE.md` 开发流程。

- [ ] **Step 1: 全量测试**

Run: `go test ./...`
Expected: 全部 PASS

- [ ] **Step 2: 编译并安装 CLI**

Run: `go install ./cmd/blogwatcher`
Expected: 成功，`blogwatcher.exe` 安装到 `C:\Users\Melodyi\go\bin\`

- [ ] **Step 3: CLI 烟雾测试**

```bash
blogwatcher tag list
blogwatcher article tag 1 Go
blogwatcher article list --tag Go
blogwatcher tag list
blogwatcher article untag 1 Go
blogwatcher tag rename Go Golang
blogwatcher tag delete Golang
```
Expected: 各命令正常执行，输出符合预期

- [ ] **Step 4: 构建并部署 Docker**

Run: `docker compose build blogwatcher-ui --no-cache && docker compose up -d blogwatcher-ui`
Expected: 构建成功、容器启动

- [ ] **Step 5: Web 端验证（手动）**

打开 Web UI 验证：侧边栏 Tags 分区、卡片 Tag 按钮 + chips、单文管理页 `/articles/{id}/tags`、全局管理页 `/tags`、按标签筛选。用 webapp-testing skill 或浏览器手动验证。

- [ ] **Step 6: 最终提交（如有遗留）**

```bash
git status
# 若有未提交的修复
git add -A && git commit -m "feat(tags): 完成文章标签功能"
```

---

## 实现顺序与注意事项

1. **Task 1→4（storage 层）可独立完成并测试**，是后续所有层的基础。
2. **Task 9/10/11 需协调**：handler 渲染依赖模板存在，模板语义依赖 handler 数据结构。建议实现时把 Task 11（模板骨架）与 Task 9/10（handler）放在相邻提交，或先建空模板骨架再实现 handler。
3. **CLI 测试范式**：Task 6/7 实现前务必先读 `internal/cli/commands/article_test.go` 确认既有断言写法（stdout 捕获方式），据此调整测试骨架——这是本计划唯一显式标注"以既有范式为准"的点。
4. **模板注册机制**：Task 11 Step 前置 grep `renderTemplate`/`template.ParseFS` 确认新建 gohtml 是否自动纳入加载；若需显式注册则补。
5. **不擅自建分支**：全程在当前 `tag` worktree 的 `tag` 分支提交。
