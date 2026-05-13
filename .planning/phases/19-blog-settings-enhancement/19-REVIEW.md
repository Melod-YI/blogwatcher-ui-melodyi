---
phase: 19-blog-settings-enhancement
reviewed: 2026-05-09T00:00:00Z
depth: standard
files_reviewed: 6
files_reviewed_list:
  - assets/static/styles.css
  - assets/templates/partials/blog-display-row.gohtml
  - assets/templates/partials/blog-edit-form.gohtml
  - assets/templates/partials/category-list.gohtml
  - internal/server/handlers.go
  - internal/storage/database.go
findings:
  critical: 2
  warning: 6
  info: 2
  total: 10
status: fixed
---

# Phase 19: Code Review Report

**Reviewed:** 2026-05-09
**Depth:** standard
**Files Reviewed:** 6
**Status:** issues_found

## Summary

审查了 Phase 19 的 Blog Settings Enhancement 相关代码，包括 Go handlers、database 层、模板文件和 CSS。发现了 2 个关键安全问题（SQL 注入风险和 URL 验证不足），以及多个代码质量和错误处理问题。

模板代码整体安全，Go 模板自动对输出进行 HTML 转义。CSS 代码结构合理，命名规范一致。但 Go 代码中存在一些需要修复的安全和质量问题。

## Critical Issues

### CR-01: SQL 注入风险 - PRAGMA 查询使用字符串拼接

**File:** `internal/storage/database.go:196, 219`
**Issue:** `columnExists` 和 `columnIsNotNull` 方法使用字符串拼接构建 PRAGMA 查询，而非参数化查询。虽然这些方法只接收内部硬编码的表名（"blogs", "articles"），但这是一个不安全的 SQL 查询模式。

```go
// 行 196
rows, err := db.conn.Query("PRAGMA table_info(" + table + ")")

// 行 219
rows, err := db.conn.Query("PRAGMA table_info(" + table + ")")
```

当前代码中 `table` 参数来自硬编码调用（如 `columnExists("blogs", "category_id")`），所以目前不存在实际漏洞。但这种模式容易被后续修改引入漏洞，应该改为更安全的实现。

**Fix:** 使用白名单验证表名，或改用参数化查询模式：
```go
func (db *Database) columnExists(table, column string) bool {
    // 白名单验证表名
    validTables := map[string]bool{"blogs": true, "articles": true, "categories": true}
    if !validTables[table] {
        return false
    }
    query := fmt.Sprintf("PRAGMA table_info(%s)", table)
    rows, err := db.conn.Query(query)
    ...
}
```

### CR-02: URL 验证不够严格，存在注入风险

**File:** `internal/server/handlers.go:1023-1030`
**Issue:** `validateURL` 函数只检查 URL 是否以 `http://` 或 `https://` 开头，但没有验证 URL 是否是有效的 URL 格式。这允许以下类型的输入：
- `http://javascript:alert(1)` - 虽然以 http:// 开头但包含注入
- `https://@evil.com/path?query=<script>` - 包含特殊字符
- `http://` - 空 URL 部分

当前验证：
```go
func validateURL(url string) error {
    if url == "" {
        return nil
    }
    if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
        return fmt.Errorf("URL 必须以 http:// 或 https:// 开头")
    }
    return nil
}
```

**Fix:** 使用 Go 的 `net/url` 包进行完整的 URL 验证：
```go
import "net/url"

func validateURL(urlStr string) error {
    if urlStr == "" {
        return nil // 允许空值（nullable 字段）
    }
    u, err := url.Parse(urlStr)
    if err != nil {
        return fmt.Errorf("URL 格式无效: %v", err)
    }
    if u.Scheme != "http" && u.Scheme != "https" {
        return fmt.Errorf("URL 必须使用 http 或 https 协议")
    }
    if u.Host == "" {
        return fmt.Errorf("URL 必须包含主机名")
    }
    return nil
}
```

## Warnings

### WR-01: FTS5 搜索查询未验证特殊字符

**File:** `internal/storage/database.go:541-544`
**Issue:** 用户输入的 `SearchQuery` 直接传递给 FTS5 MATCH 查询。FTS5 有特殊的查询语法（如 `AND`, `OR`, `NOT`, `*`, `^` 等），用户输入特殊字符可能导致查询失败或意外结果。

```go
if opts.SearchQuery != "" {
    query.WriteString(` JOIN articles_fts ON a.id = articles_fts.rowid`)
    conditions = append(conditions, "articles_fts MATCH ?")
    args = append(args, opts.SearchQuery)
}
```

虽然使用了参数化查询，但 FTS5 的特殊语法可能导致用户体验问题（查询失败返回错误）。

**Fix:** 在传递给 FTS5 前，对搜索查询进行转义或清理：
```go
// 转义 FTS5 特殊字符
func escapeFTS5Query(query string) string {
    // FTS5 特殊字符: *, ^, -, +, :, (, ), {, }, [, ], ", '
    specialChars := []string{"*", "^", "-", "+", ":", "(", ")", "{", "}", "[", "]", "\"", "'"}
    result := query
    for _, char := range specialChars {
        result = strings.ReplaceAll(result, char, "")
    }
    return result
}
```

或者在搜索失败时提供友好的错误消息，而不是返回 500 错误。

### WR-02: 错误处理不完整 - 非关键数据获取失败时缺少处理

**File:** `internal/server/handlers.go:145-150`
**Issue:** `handleArticleList` 在获取 blogs 失败时只记录日志，但数据可能仍然被用于渲染。如果 `data["Blogs"]` 未设置，模板可能渲染不完整。

```go
// Return full page for direct navigation
data["Title"] = "BlogWatcher"
data["Version"] = s.version
blogs, err := s.db.ListBlogs()
if err != nil {
    log.Printf("Error fetching blogs: %v", err)
} else {
    data["Blogs"] = blogs
}
```

**Fix:** 明确处理错误情况，确保数据完整性：
```go
blogs, err := s.db.ListBlogs()
if err != nil {
    log.Printf("Error fetching blogs for sidebar: %v", err)
    data["Blogs"] = []model.Blog{} // 设置空数组避免模板错误
} else {
    data["Blogs"] = blogs
}
```

### WR-03: 日志记录不够详细 - 更新操作缺少字段变化记录

**File:** `internal/server/handlers.go:729`
**Issue:** `handleUpdateBlogName` 的日志只记录了新值，没有记录旧值，这使得问题排查困难。

```go
log.Printf("Updated blog %d: name='%s', url='%s', feed_url='%s', category=%v",
    id, name, url, feedURL, categoryID)
```

**Fix:** 记录新旧值对比：
```go
log.Printf("Updated blog %d: name='%s' (was '%s'), url='%s' (was '%s'), feed_url='%s' (was '%s'), category=%v (was %v)",
    id, name, oldBlog.Name, url, oldBlog.URL, feedURL, oldBlog.FeedURL, categoryID, oldBlog.CategoryID)
```

### WR-04: 前端验证依赖不可靠 - HTML5 pattern 验证可被绕过

**File:** `assets/templates/partials/blog-edit-form.gohtml:16-18, 25-27`
**Issue:** URL 输入使用 HTML5 `pattern` 属性进行前端验证，但这种验证可以被绕过：
- 用户可以禁用浏览器验证
- 用户可以直接通过 API/命令行提交
- pattern 验证不阻止恶意数据提交

```html
<input type="url" name="url" value="{{.Blog.URL}}"
       pattern="^https?://.*"
       placeholder="https://example.com">
<span class="url-error-message">URL 必须以 http:// 或 https:// 开头</span>
```

虽然后端有 validateURL，但后端验证不够严格（见 CR-02）。

**Fix:** 前端验证只是用户体验增强，关键验证必须在后端完成。修复后端 validateURL（见 CR-02）。

### WR-05: 变量命名冲突 - error 参数名与内置类型重名

**File:** `internal/storage/database.go:835`
**Issue:** `AddArticlesBulk` 函数的第三个返回参数名为 `error`，与 Go 内置的 `error` 类型重名，这可能导致代码混淆和类型检查问题。

```go
func (db *Database) AddArticlesBulk(articles []model.Article) (inserted int, skipped int, error error) {
```

**Fix:** 使用不同的参数名：
```go
func (db *Database) AddArticlesBulk(articles []model.Article) (inserted int, skipped int, err error) {
```

### WR-06: 错误处理不一致 - 部分错误未包装上下文

**File:** `internal/storage/database.go:132-134`
**Issue:** `ensureMigrations` 方法中，某些错误使用 `fmt.Errorf` 包装上下文，但某些直接返回原始错误，不一致的错误处理使得问题定位困难。

```go
// 行 132-134 - 直接返回错误
if _, err := db.conn.Exec(`ALTER TABLE articles ADD COLUMN thumbnail_url TEXT`); err != nil {
    return err // 没有包装上下文
}

// 其他位置有包装
if _, err := db.conn.Exec(`ALTER TABLE blogs ADD COLUMN category_id INTEGER REFERENCES categories(id)`); err != nil {
    return fmt.Errorf("failed to add category_id column to blogs: %w", err) // 有包装
}
```

**Fix:** 所有错误都应包装上下文：
```go
if _, err := db.conn.Exec(`ALTER TABLE articles ADD COLUMN thumbnail_url TEXT`); err != nil {
    return fmt.Errorf("failed to add thumbnail_url column to articles: %w", err)
}
```

## Info

### IN-01: Magic numbers - 硬编码的时间常量

**File:** `internal/server/handlers.go:306, 365, 557`
**Issue:** 多处使用硬编码的时间常量：
- `3*time.Minute`（行 306, 365）- sync timeout
- `2*time.Minute`（行 557）- auto-sync timeout

```go
ctx, cancel := context.WithTimeout(r.Context(), 3*time.Minute)
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
```

**Fix:** 定义常量并在文档中说明超时时间的理由：
```go
const (
    // SyncTimeout - sync 操作超时时间（考虑慢速外部站点）
    SyncTimeout = 3 * time.Minute
    // AutoSyncTimeout - 新博客自动同步超时时间
    AutoSyncTimeout = 2 * time.Minute
)

func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), SyncTimeout)
    ...
}
```

### IN-02: CSS 变量命名不一致 - text-muted 变量位置

**File:** `assets/static/styles.css:1906-1909`
**Issue:** `--text-muted` CSS 变量定义在文件末尾的 "Category Grouping Styles" 部分（行 1906），而其他文本颜色变量（`--text-primary`, `--text-secondary`）定义在文件开头的 ":root" 部分（行 7-24）。这可能导致变量查找困难和维护问题。

```css
/* 行 1906-1909 - 在文件末尾定义 */
:root {
  --text-muted: #9ca3af;
}
```

**Fix:** 将 `--text-muted` 移到文件开头的 CSS 变量定义区域：
```css
:root {
  color-scheme: light dark;
  --bg-primary: #FAF8F5;
  --text-primary: #37352F;
  --text-secondary: #6B6B6B;
  --text-muted: #9ca3af;  /* 移到这里 */
  ...
}
```

---

_Reviewed: 2026-05-09_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_