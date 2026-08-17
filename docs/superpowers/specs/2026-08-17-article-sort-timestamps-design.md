---
name: article-sort-timestamps
description: 文章收藏/已读时间戳与按时间排序功能设计文档
---

# 文章收藏/已读时间戳与排序设计

## 概述

为文章的收藏与已读状态记录时间戳（`favorited_at`、`read_at`），并在 Web UI 提供按这两个时间排序的能力。收藏页默认按收藏时间排序，已读页默认按已读时间排序，二者均可切换为按发布时间排序；收件箱仅支持发布时间排序。

## 核心决策

| 方面 | 决定 |
|------|------|
| 新增列 | `articles` 表新增 `favorited_at TIMESTAMP`、`read_at TIMESTAMP`，均 nullable，与 `is_read`/`is_favorited` 模式一致 |
| 时间戳生命周期 | 收藏/标为已读时写入当前时间；取消收藏/标为未读时置 NULL（「取消即清空」），即记录「最近一次进入该状态的时间」 |
| 存量回填 | 不回填，保持 NULL；排序查询用 `COALESCE(favorited_at, discovered_date)` / `COALESCE(read_at, discovered_date)` 回退，避免凭空造数据 |
| 排序入口 | 在 `SearchOptions` 增加 `Sort` 字段，`SearchArticles` 一处 ORDER BY 分支，复用单一查询入口 |
| 默认排序 | `filter=favorites` → `favorited`；`filter=read` → `read`；其余 → `published` |
| 排序控件 | 仅收藏页/已读页渲染排序分段按钮（对应时间 / 发布时间）；收件箱不渲染控件 |
| 卡片时间显示 | 不改动，维持现状（显示 `timeAgo(published_date)`）。排序结果仅以顺序体现，不新增时间字段展示，避免 `renderUpdatedArticleCard` 单卡重渲染需透传 `CurrentFilter` 的连锁改动 |
| CLI | 不在范围内，`article list` 维持发布时间排序 |

---

## 第一部分：数据库设计

### 新增列

在 `articles` 表新增：

| 列名 | 类型 | 说明 |
|------|------|------|
| `favorited_at` | TIMESTAMP | 收藏时间，nullable；取消收藏置 NULL |
| `read_at` | TIMESTAMP | 已读时间，nullable；标为未读置 NULL |

### Migration 实现

在 `database.go` 的 `ensureMigrations()` 末尾追加（幂等，沿用 `is_favorited` 的写法）：

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

存量已收藏/已读行的新列保持 NULL，不回填。

---

## 第二部分：Model 更新

`internal/model/model.go`：

```go
type Article struct {
    // 现有字段...
    IsFavorited  bool
    FavoritedAt  *time.Time // 收藏时间，未收藏为 nil
    ReadAt       *time.Time // 已读时间，未读为 nil
}

type ArticleWithBlog struct {
    // 现有字段...
    IsFavorited  bool
    FavoritedAt  *time.Time
    ReadAt       *time.Time
}

type SearchOptions struct {
    // 现有字段...
    Sort string // "published"(默认) | "favorited" | "read"
}

// Sort 取值常量
const (
    SortPublished = "published"
    SortFavorited = "favorited"
    SortRead      = "read"
)
```

---

## 第三部分：Storage 层

`internal/storage/database.go`：

### 状态写方法更新

```go
// FavoriteArticle 收藏文章并记录时间
func (db *Database) FavoriteArticle(id int64) error {
    now := time.Now().Format(sqliteTimeLayout)
    result, err := db.conn.Exec(
        `UPDATE articles SET is_favorited = 1, favorited_at = ? WHERE id = ?`, now, id)
    // ... rows==0 → not found
}

// UnfavoriteArticle 取消收藏并清空时间
func (db *Database) UnfavoriteArticle(id int64) error {
    result, err := db.conn.Exec(
        `UPDATE articles SET is_favorited = 0, favorited_at = NULL WHERE id = ?`, id)
    // ...
}

// MarkArticleRead 标记已读并记录时间
func (db *Database) MarkArticleRead(id int64) (bool, error) {
    now := time.Now().Format(sqliteTimeLayout)
    result, err := db.conn.Exec(
        `UPDATE articles SET is_read = 1, read_at = ? WHERE id = ?`, now, id)
    // ...
}

// MarkArticleUnread 标记未读并清空时间
func (db *Database) MarkArticleUnread(id int64) (bool, error) {
    result, err := db.conn.Exec(
        `UPDATE articles SET is_read = 0, read_at = NULL WHERE id = ?`, id)
    // ...
}

// MarkAllUnreadArticlesRead 批量标记已读并记录时间
func (db *Database) MarkAllUnreadArticlesRead(blogID *int64) error {
    now := time.Now().Format(sqliteTimeLayout)
    query := `UPDATE articles SET is_read = 1, read_at = ? WHERE is_read = 0`
    // blog 过滤同现状
}
```

### SearchArticles ORDER BY 分支

`SearchArticles` 在拼装完 WHERE 后，按 `opts.Sort` 选择 ORDER BY：

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

> `ORDER BY` 引用的列无需出现在 SELECT 列表，但为保持 scan 列顺序统一，SELECT 仍会带上 `favorited_at, read_at`。

### scan 函数与查询列同步

`scanArticle`、`scanArticleWithBlog`、`scanArticleWithBlogAndCount` 的 SELECT 列表与 Scan 目标新增 `favorited_at, read_at`（`sql.NullString` → 解析为 `*time.Time`）。下列方法的 SELECT 需同步追加这两列，保持顺序一致：

- `ListArticles`
- `ListArticlesByReadStatus`
- `ListArticlesWithBlog`
- `SearchArticles`
- `GetArticleByID`
- `ListArticlesWithFilters`

CLI 侧（`ListArticlesWithFilters`）排序仍固定为发布时间，不在本次范围。

---

## 第四部分：Handler 层

`internal/server/handlers.go`：

### parseSearchOptions 解析 sort

```go
// 解析显式 sort 参数
sortParam := r.URL.Query().Get("sort")
switch sortParam {
case model.SortFavorited, model.SortRead, model.SortPublished:
    opts.Sort = sortParam
default:
    opts.Sort = "" // 未指定，按 filter 决定默认
}

// 按 filter 设默认（仅当 sort 未显式指定时）
if opts.Sort == "" {
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

返回值新增 `sort`（或直接把 `opts.Sort` 放进模板数据）。`handleIndex`、`handleArticleList`、`handleMarkAllRead`、`handleSync` 的模板 data 注入 `"CurrentSort": opts.Sort`。

---

## 第五部分：模板

`assets/templates/partials/article-list.gohtml`：

### 排序控件

在 `.filter-bar` 内、`#filter-hidden` 附近新增排序分段按钮，**仅当 `CurrentFilter` 为 `favorites` 或 `read` 时渲染**：

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
  <input type="radio" name="sort" id="sort-read" value="read" ...>
  <label for="sort-read">已读时间</label>
  {{end}}
  <input type="radio" name="sort" id="sort-published" value="published"
         {{if eq .CurrentSort "published"}}checked{{end}} ...>
  <label for="sort-published">发布时间</label>
</div>
{{end}}
<input type="hidden" name="sort" id="sort-hidden" value="{{.CurrentSort}}">
```

`#sort-hidden` 用于 search/date 输入的 `hx-include` 携带当前 sort，避免切换搜索/日期时排序丢失。

### load-more 链接

`#load-more-trigger` 的 `hx-get` URL 追加 `&sort={{.CurrentSort}}`。

> 卡片 meta 时间维持现状（`timeAgo(published_date)`），不在本次改动卡片渲染。

---

## 第六部分：测试计划

`internal/storage/database_test.go`：

| 测试用例 | 说明 |
|----------|------|
| `TestFavoriteArticle_SetsFavoritedAt` | 收藏后 `favorited_at` 非空 |
| `TestUnfavoriteArticle_ClearsFavoritedAt` | 取消收藏后 `favorited_at` 为 NULL |
| `TestMarkArticleRead_SetsReadAt` | 标记已读后 `read_at` 非空 |
| `TestMarkArticleUnread_ClearsReadAt` | 标记未读后 `read_at` 为 NULL |
| `TestMarkAllUnreadArticlesRead_SetsReadAt` | 批量标记后所有受影响行 `read_at` 非空 |
| `TestSearchArticles_SortByFavorited` | `Sort=favorited` 按 `COALESCE(favorited_at, discovered_date)` 倒序 |
| `TestSearchArticles_SortByRead` | `Sort=read` 按 `COALESCE(read_at, discovered_date)` 倒序 |
| `TestSearchArticles_SortNullFallback` | 存量 NULL 行回退 `discovered_date` 参与排序 |
| migration 幂等 | 已有 `favorited_at`/`read_at` 列时不再 ALTER |

### 日志

- `FavoriteArticle`/`UnfavoriteArticle`/`MarkArticleRead`/`MarkArticleUnread`/`MarkAllUnreadArticlesRead` 入口与结果补日志（带 id 与时间），符合 CLAUDE.md 核心业务日志要求。
- Handler 已有入口日志，`parseSearchOptions` 无需新增。

---

## 文件变更清单

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `internal/model/model.go` | 修改 | Article/ArticleWithBlog 加 FavoritedAt/ReadAt；SearchOptions 加 Sort 及常量 |
| `internal/storage/database.go` | 修改 | migration 两列；状态写方法更新时间戳；SearchArticles ORDER BY 分支；scan 与查询列同步 |
| `internal/server/handlers.go` | 修改 | parseSearchOptions 解析 sort + 默认；各 handler 注入 CurrentSort |
| `assets/templates/partials/article-list.gohtml` | 修改 | 排序分段控件、load-more 带 sort、search/date 输入 include `#sort-hidden` |
| `assets/static/styles.css` | 修改 | `.sort-toggle` 样式（沿用 `.view-toggle` 风格） |
| `internal/storage/database_test.go` | 修改 | 新增时间戳与排序测试 |

## 不在范围内

- CLI `article list` 排序标志。
- 收件箱排序控件（仅发布时间）。
- `ListArticlesWithFilters`（CLI 用）排序变更。
