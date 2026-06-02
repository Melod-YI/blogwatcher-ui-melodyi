---
name: article-favorites
description: 文章收藏功能设计文档
---

# 文章收藏功能设计

## 概述

为博客文章添加收藏功能，用户可以收藏/取消收藏文章，并通过 UI 和 CLI 查看收藏的文章列表。

## 核心决策

| 方面 | 决定 |
|------|------|
| 命名 | 收藏（Favorite），变量名 `is_favorited` |
| 存储方案 | articles 表添加 `is_favorited BOOLEAN` 列，与 `is_read`/`has_note` 模式一致 |
| 侧边栏 | 添加"收藏"导航入口，与 Inbox/Archived 并列 |
| CLI 命令 | 分离模式：`article favorite <id>` 和 `article unfavorite <id>` |
| 筛选 | `article list` 添加 `--favorited` 标志，Web UI 支持 `?filter=favorites` |

---

## 第一部分：数据库设计

### 新增列

在 `articles` 表新增一列：

| 列名 | 类型 | 说明 |
|------|------|------|
| `is_favorited` | BOOLEAN | 收藏状态，默认 `FALSE` |

### Migration 实现

在 `database.go` 的 `ensureMigrations()` 中添加：

```go
// Add is_favorited column if it doesn't exist
if !db.columnExists("articles", "is_favorited") {
    if _, err := db.conn.Exec(`ALTER TABLE articles ADD COLUMN is_favorited BOOLEAN DEFAULT FALSE`); err != nil {
        return fmt.Errorf("failed to add is_favorited column: %w", err)
    }
}
```

### Model 更新

在 `model.go` 中为 `Article` 和 `ArticleWithBlog` 结构体新增字段：

```go
// Article 结构体新增字段
IsFavorited bool // 文章收藏状态

// ArticleWithBlog 结构体同样新增该字段
IsFavorited bool // 文章收藏状态
```

### SearchOptions 更新

```go
type SearchOptions struct {
    // 现有字段...
    IsFavorited *bool      // nil = 所有, true = 仅收藏, false = 仅非收藏
}
```

---

## 第二部分：Storage 层

### 新增方法

```go
// FavoriteArticle 收藏文章
func (db *Database) FavoriteArticle(id int64) error {
    result, err := db.conn.Exec(`UPDATE articles SET is_favorited = TRUE WHERE id = ?`, id)
    if err != nil {
        return fmt.Errorf("failed to favorite article: %w", err)
    }
    rows, _ := result.RowsAffected()
    if rows == 0 {
        return fmt.Errorf("article not found: %d", id)
    }
    return nil
}

// UnfavoriteArticle 取消收藏文章
func (db *Database) UnfavoriteArticle(id int64) error {
    result, err := db.conn.Exec(`UPDATE articles SET is_favorited = FALSE WHERE id = ?`, id)
    if err != nil {
        return fmt.Errorf("failed to unfavorite article: %w", err)
    }
    rows, _ := result.RowsAffected()
    if rows == 0 {
        return fmt.Errorf("article not found: %d", id)
    }
    return nil
}
```

### 现有查询更新

- 所有 `scanArticle`/`scanArticleWithBlog` 函数添加 `is_favorited` 字段扫描
- `SearchArticles` 方法的 WHERE 子句添加 `is_favorited` 过滤条件
- `GetArticleByID` 的 SELECT 列表添加 `is_favorited`

---

## 第三部分：API 路由与 Handler

### 新增路由

| 方法 | 路由 | Handler | 说明 |
|------|------|---------|------|
| POST | `/articles/{id}/favorite` | `handleFavorite` | 收藏文章 |
| POST | `/articles/{id}/unfavorite` | `handleUnfavorite` | 取消收藏 |

### Handler 实现

遵循 `handleMarkRead`/`handleMarkUnread` 模式：

```go
func (s *Server) handleFavorite(w http.ResponseWriter, r *http.Request) {
    // 1. 解析文章 ID
    // 2. 调用 db.FavoriteArticle(id)
    // 3. HTMX 请求：重新渲染文章卡片 partial
    // 4. 设置 HX-Trigger: articleListUpdated
}

func (s *Server) handleUnfavorite(w http.ResponseWriter, r *http.Request) {
    // 1. 解析文章 ID
    // 2. 调用 db.UnfavoriteArticle(id)
    // 3. HTMX 请求：重新渲染文章卡片 partial
    // 4. 设置 HX-Trigger: articleListUpdated
}
```

### 筛选支持

在 `parseSearchOptions()` 中添加 `filter=favorites` 支持：

```go
case "favorites":
    isFav := true
    opts.IsFavorited = &isFav
```

---

## 第四部分：CLI 命令设计

### 新增子命令

在 `article` 命令组下新增两个子命令：

```bash
blogwatcher article favorite <id>      # 收藏文章
blogwatcher article unfavorite <id>    # 取消收藏
```

### 命令实现

遵循 `mark-read`/`mark-unread` 模式：

```go
// article favorite <id>
func runArticleFavorite(cmd *cobra.Command, args []string) {
    // 1. 解析文章 ID
    // 2. 打开数据库
    // 3. 验证文章存在
    // 4. 调用 db.FavoriteArticle(id)
    // 5. 输出: "文章已收藏: <title>"
}

// article unfavorite <id>
func runArticleUnfavorite(cmd *cobra.Command, args []string) {
    // 1. 解析文章 ID
    // 2. 打开数据库
    // 3. 验证文章存在
    // 4. 调用 db.UnfavoriteArticle(id)
    // 5. 输出: "已取消收藏: <title>"
}
```

### article list 筛选

`article list` 命令添加 `--favorited` 标志：

```bash
blogwatcher article list --favorited   # 仅显示收藏文章
```

---

## 第五部分：UI 设计

### 文章卡片按钮

在 `article-items.gohtml` 的操作按钮区域新增收藏按钮：

```html
{{if .IsFavorited}}
<button hx-post="/articles/{{.ID}}/unfavorite"
        hx-target="#article-{{.ID}}"
        hx-swap="outerHTML swap:300ms"
        class="action-btn action-btn-favorite active"
        title="取消收藏"
        onclick="event.stopPropagation();">
    ★
</button>
{{else}}
<button hx-post="/articles/{{.ID}}/favorite"
        hx-target="#article-{{.ID}}"
        hx-swap="outerHTML swap:300ms"
        class="action-btn action-btn-favorite"
        title="收藏"
        onclick="event.stopPropagation();">
    ☆
</button>
{{end}}
```

**视觉样式**：
- 已收藏：实心星标 ★，高亮颜色（如金色/warning色）
- 未收藏：空心星标 ☆，默认颜色

### 侧边栏导航

在 `sidebar.gohtml` 中 Inbox 和 Archived 之间添加"收藏"入口：

```
📥 Inbox (未读)
⭐ 收藏 (Favorites)    ← 新增
📦 Archived (已读)
```

链接指向 `/articles?filter=favorites`。

### 文章列表筛选

在文章列表头部的筛选标签中添加"收藏"选项。

---

## 第六部分：测试计划

### Storage 层测试

| 测试用例 | 说明 |
|----------|------|
| `TestFavoriteArticle` | 验证收藏操作正确更新 `is_favorited` 为 TRUE |
| `TestUnfavoriteArticle` | 验证取消收藏操作正确更新 `is_favorited` 为 FALSE |
| `TestFavoriteArticle_NotFound` | 验证不存在的文章返回错误 |
| `TestSearchArticles_FavoriteFilter` | 验证 `IsFavorited` 筛选条件正确过滤 |
| `TestArticleDefaultNotFavorited` | 验证新文章默认 `is_favorited` 为 FALSE |

### Handler 层测试

| 测试用例 | 说明 |
|----------|------|
| `TestHandleFavorite` | 验证 POST 端点正确收藏并返回成功 |
| `TestHandleUnfavorite` | 验证 POST 端点正确取消收藏并返回成功 |
| `TestHandleFavorite_InvalidID` | 验证非法 ID 返回 400 |
| `TestHandleFavorite_ArticleNotFound` | 验证不存在的文章返回 404 |

### CLI 测试

| 测试用例 | 说明 |
|----------|------|
| `TestArticleFavoriteCmd` | 验证 favorite 命令正确执行 |
| `TestArticleUnfavoriteCmd` | 验证 unfavorite 命令正确执行 |

### 日志

- Handler 入口：记录请求的文章 ID 和操作类型
- Storage 操作结果：记录成功/失败状态
- CLI 命令执行：记录操作结果

---

## 文件变更清单

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `internal/model/model.go` | 修改 | Article/ArticleWithBlog 添加 IsFavorited 字段，SearchOptions 添加 IsFavorited |
| `internal/storage/database.go` | 修改 | Migration、FavoriteArticle/UnfavoriteArticle 方法、查询更新 |
| `internal/server/routes.go` | 修改 | 注册 favorite/unfavorite 路由 |
| `internal/server/handlers.go` | 修改 | 新增 handleFavorite/handleUnfavorite |
| `internal/cli/commands/article.go` | 修改 | 新增 favorite/unfavorite 子命令，list 添加 --favorited 标志 |
| `internal/cli/commands/root.go` | 修改 | 注册新命令（如需要） |
| `assets/templates/partials/article-items.gohtml` | 修改 | 添加收藏按钮 |
| `assets/templates/partials/sidebar.gohtml` | 修改 | 添加"收藏"导航入口 |
| `assets/templates/partials/article-list.gohtml` | 修改 | 添加收藏筛选标签 |
| `assets/static/styles.css` | 修改 | 收藏按钮样式 |
| `internal/storage/database_test.go` | 修改 | 添加收藏相关测试 |
| `internal/server/handlers_test.go` | 修改 | 添加 Handler 测试 |
