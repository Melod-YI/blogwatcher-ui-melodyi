# Phase 18: Category Display & CLI - Pattern Map

**Mapped:** 2026-05-09
**Files analyzed:** 6 new/modified files
**Analogs found:** 6 / 6

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `assets/templates/partials/sidebar.gohtml` | template | request-response | `assets/templates/partials/sidebar.gohtml` (existing) | exact (modification) |
| `assets/templates/partials/category-group.gohtml` | template | request-response | `assets/templates/partials/blog-list.gohtml` | exact |
| `internal/cli/commands/article.go` | controller (CLI) | request-response | `internal/cli/commands/article.go` (existing) | exact (modification) |
| `internal/storage/database.go` | model | CRUD | `internal/storage/database.go` (existing) | exact (modification) |
| `assets/static/styles.css` | config | static | `assets/static/styles.css` (existing) | exact (modification) |
| `assets/js/sidebar.js` | utility | client-side | `assets/templates/base.gohtml` (localStorage pattern) | role-match |

---

## Pattern Assignments

### `assets/templates/partials/sidebar.gohtml` (template, request-response)

**Analog:** `assets/templates/partials/sidebar.gohtml` (现有文件改造)

**改造要点:** 替换 Subscriptions section 的 blog-list 渲染逻辑为分类分组

**现有结构** (lines 62-71):
```html
{{/* Subscriptions / Blog list */}}
<div class="subscriptions-section">
    <div class="nav-section-title">Subscriptions</div>
    <div id="blog-list"
         hx-get="/blogs"
         hx-trigger="blogListUpdated from:body"
         hx-swap="innerHTML">
        {{template "blog-list.gohtml" .}}
    </div>
</div>
```

**改造模式** (参考 D-01/D-02/D-03):
```html
{{/* Subscriptions / Blog list - Category grouped */}}
<div class="subscriptions-section">
    <div class="nav-section-title">Subscriptions</div>
    <div id="blog-list"
         hx-get="/blogs/grouped"
         hx-trigger="blogListUpdated from:body"
         hx-swap="innerHTML">
        {{template "category-group.gohtml" .}}
    </div>
</div>
```

**HTMX 触发模式** (保持现有模式, lines 66-68):
- `hx-get="/blogs"` → 改为 `/blogs/grouped`
- `hx-trigger="blogListUpdated from:body"` — 保持不变
- `hx-swap="innerHTML"` — 保持不变

---

### `assets/templates/partials/category-group.gohtml` (template, request-response)

**Analog:** `assets/templates/partials/blog-list.gohtml`

**现有 blog-list 模式** (lines 1-16):
```html
{{define "blog-list.gohtml"}}
{{range .Blogs}}
<a href="/articles?blog={{.ID}}"
   hx-get="/articles?blog={{.ID}}"
   hx-target="#main-content"
   hx-push-url="true"
   hx-on:click="document.querySelectorAll('.sidebar-nav .nav-link, .blog-item').forEach(el => el.classList.remove('active')); this.classList.add('active'); document.getElementById('sidebar-toggle').checked = false;"
   class="blog-item{{if eq $.CurrentBlogID .ID}} active{{end}}">
    {{.Name}}
</a>
{{else}}
<p class="empty-state">No blogs tracked yet. Use the blogwatcher CLI to add blogs.</p>
{{end}}
{{end}}
```

**分类分组结构** (方案C, 参考 `.superpowers/mockups/category-sidebar-grouping.html` lines 373-469):
```html
{{define "category-group.gohtml"}}
{{/* Categories first (only show categories with blog_count > 0) */}}
{{range .Categories}}
{{if gt .BlogCount 0}}
<div class="category-group" data-category="{{.Name}}">
    <div class="category-header" onclick="toggleCategory(this)">
        <span class="expand-icon">
            <svg class="icon-chevron" viewBox="0 0 24 24">
                <polyline points="6 9 12 15 18 9"></polyline>
            </svg>
        </span>
        {{.Name}}
        <span class="category-count">{{.BlogCount}}</span>
    </div>
    <div class="category-blogs">
        {{range .Blogs}}
        <a href="/articles?blog={{.ID}}"
           hx-get="/articles?blog={{.ID}}"
           hx-target="#main-content"
           hx-push-url="true"
           hx-on:click="document.querySelectorAll('.sidebar-nav .nav-link, .blog-item').forEach(el => el.classList.remove('active')); this.classList.add('active'); document.getElementById('sidebar-toggle').checked = false;"
           class="blog-item{{if eq $.CurrentBlogID .ID}} active{{end}}">
            {{.Name}}
        </a>
        {{end}}
    </div>
</div>
{{end}}
{{end}}

{{/* Uncategorized blogs at the END with separator */}}
{{if .Uncategorized}}
<div class="uncategorized-section">
    <div class="uncategorized-header">未分类</div>
    {{range .Uncategorized}}
    <a href="/articles?blog={{.ID}}"
       hx-get="/articles?blog={{.ID}}"
       hx-target="#main-content"
       hx-push-url="true"
       hx-on:click="..."
       class="blog-item{{if eq $.CurrentBlogID .ID}} active{{end}}">
        {{.Name}}
    </a>
    {{end}}
</div>
{{end}}
{{end}}
```

**blog-item HTMX 导航模式** (复制自 blog-list.gohtml lines 5-10):
- `hx-get="/articles?blog={{.ID}}"` — HTMX 导航
- `hx-target="#main-content"` — 内容替换目标
- `hx-push-url="true"` — URL 更新
- `hx-on:click` — active 状态切换 + 关闭 mobile sidebar

---

### `internal/cli/commands/article.go` (CLI command, request-response)

**Analog:** `internal/cli/commands/article.go` (现有 --blog flag 模式)

**Flag 定义模式** (lines 69-78):
```go
// 添加筛选 flags
cmd.Flags().String("blog", "", "博客名称筛选")
cmd.Flags().Bool("unread", false, "仅未读文章")
cmd.Flags().Bool("read", false, "仅已读文章")
cmd.Flags().Bool("not-noted", false, "仅无备注文章")
cmd.Flags().String("after", "", "日期筛选（格式 YYYY-MM-DD）")
cmd.Flags().String("format", "table", "输出格式（table|json|simple）")

// 标记 --unread 和 --read 为互斥
cmd.MarkFlagsMutuallyExclusive("unread", "read")
```

**新增 --category flag** (参考 D-06):
```go
// 添加 --category flag
cmd.Flags().String("category", "", "分类名称筛选")
```

**Flag 解析模式** (lines 143-148):
```go
// 解析筛选参数
blogName, _ := cmd.Flags().GetString("blog")
unread, _ := cmd.Flags().GetBool("unread")
read, _ := cmd.Flags().GetBool("read")
notNoted, _ := cmd.Flags().GetBool("not-noted")
afterStr, _ := cmd.Flags().GetString("after")
format, _ := cmd.Flags().GetString("format")
```

**新增 category 解析** (参考 D-07):
```go
// 解析 --category flag
categoryName, _ := cmd.Flags().GetString("category")
```

**验证模式** (--blog 验证, lines 188-200):
```go
// 检查博客名称是否存在（如果指定了）
if blogName != "" && len(articles) == 0 {
    // 验证博客是否存在
    blog, err := db.GetBlogByName(blogName)
    if err != nil {
        fmt.Fprintf(os.Stderr, "查询博客失败: %v\n", err)
        os.Exit(1)
    }
    if blog == nil {
        fmt.Fprintf(os.Stderr, "博客 '%s' 不存在\n", blogName)
        os.Exit(1)
    }
}
```

**新增 category 验证** (参考 D-08, 复用 --blog 验证模式):
```go
// 检查分类名称是否存在（如果指定了）
if categoryName != "" {
    category, err := db.GetCategoryByName(categoryName)
    if err != nil {
        fmt.Fprintf(os.Stderr, "查询分类失败: %v\n", err)
        os.Exit(1)
    }
    if category == nil {
        fmt.Fprintf(os.Stderr, "分类 '%s' 不存在\n", categoryName)
        os.Exit(1)
    }
}
```

**ListFilterOptions 构建** (lines 151-179):
```go
// 构建筛选选项
opts := storage.ListFilterOptions{
    BlogName: blogName,
}

// 设置 IsRead 状态筛选
if unread {
    isRead := false
    opts.IsRead = &isRead
} else if read {
    isRead := true
    opts.IsRead = &isRead
}

// 设置 HasNote 状态筛选
if notNoted {
    hasNote := false
    opts.HasNote = &hasNote
}

// 解析日期筛选
if afterStr != "" {
    afterDate, err := time.Parse("2006-01-02", afterStr)
    if err != nil {
        fmt.Fprintf(os.Stderr, "日期格式错误: %v（格式应为 YYYY-MM-DD）\n", err)
        os.Exit(1)
    }
    opts.AfterDate = &afterDate
}
```

**新增 CategoryName 到 opts** (参考 D-07):
```go
opts := storage.ListFilterOptions{
    BlogName:     blogName,
    CategoryName: categoryName, // 新增字段
}
```

---

### `internal/storage/database.go` (model, CRUD)

**Analog:** `internal/storage/database.go` (现有 GetBlogByName, ListFilterOptions 模式)

**GetBlogByName 模式** (lines 659-663):
```go
// GetBlogByName returns a blog by its name, or nil if not found.
func (db *Database) GetBlogByName(name string) (*model.Blog, error) {
    row := db.conn.QueryRow(`SELECT id, name, url, feed_url, scrape_selector, last_scanned, category_id FROM blogs WHERE name = ?`, name)
    return scanBlog(row)
}
```

**新增 GetCategoryByName** (复用 scanCategory 模式, 参考 D-08):
```go
// GetCategoryByName returns a category by its name, or nil if not found.
func (db *Database) GetCategoryByName(name string) (*model.Category, error) {
    row := db.conn.QueryRow(`SELECT id, name, created_at FROM categories WHERE name = ?`, name)
    return scanCategory(row)
}

// scanCategory scans a category row from the database.
func scanCategory(scanner interface{ Scan(dest ...any) error }) (*model.Category, error) {
    var (
        id        int64
        name      string
        createdAt sql.NullString
    )
    if err := scanner.Scan(&id, &name, &createdAt); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil
        }
        return nil, err
    }
    
    category := &model.Category{
        ID:   id,
        Name: name,
    }
    if createdAt.Valid {
        if parsed, err := parseTime(createdAt.String); err == nil {
            category.CreatedAt = parsed
        }
    }
    return category, nil
}
```

**ListFilterOptions 结构** (lines 1132-1140):
```go
// ListFilterOptions 筛选文章的选项参数
// 用于 CLI article list 命令，支持按博客名称、已读状态、备注状态、日期筛选
type ListFilterOptions struct {
    BlogName  string     // 博客名称筛选（空表示所有博客）
    IsRead    *bool      // 已读状态筛选（nil 表示所有状态）
    HasNote   *bool      // 备注状态筛选（nil 表示所有状态，false 表示无备注）
    AfterDate *time.Time // 日期筛选（nil 表示无限制）
    Limit     int        // 结果数量限制（0 表示无限制）
}
```

**新增 CategoryName 字段** (参考 D-07):
```go
type ListFilterOptions struct {
    BlogName     string     // 博客名称筛选（空表示所有博客）
    CategoryName string     // 分类名称筛选（空表示所有分类）
    IsRead       *bool      // 已读状态筛选（nil 表示所有状态）
    HasNote      *bool      // 备注状态筛选（nil 表示所有状态，false 表示无备注）
    AfterDate    *time.Time // 日期筛选（nil 表示无限制）
    Limit        int        // 结果数量限制（0 表示无限制）
}
```

**ListArticlesWithFilters 方法** (lines 1144-1211):
```go
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
    
    // ... 其他筛选条件
    
    // 添加 WHERE 条件
    if len(conditions) > 0 {
        query += " AND " + strings.Join(conditions, " AND ")
    }
    
    // ... 排序和执行
}
```

**新增分类筛选条件** (参考 D-07):
```go
// 分类名称筛选（通过子查询获取 category_id）
if opts.CategoryName != "" {
    conditions = append(conditions, "b.category_id = (SELECT id FROM categories WHERE name = ?)")
    args = append(args, opts.CategoryName)
}
```

**ListBlogsWithCounts 模式** (lines 356-414) — 可参考用于分组数据结构:
```go
// ListBlogsWithCounts returns all blogs with their article counts.
// Uses LEFT JOIN to include blogs with zero articles.
func (db *Database) ListBlogsWithCounts() ([]BlogWithCount, error) {
    query := `SELECT
        b.id, b.name, b.url, b.feed_url, b.scrape_selector, b.last_scanned, b.category_id,
        COUNT(a.id) as article_count
    FROM blogs b
    LEFT JOIN articles a ON b.id = a.blog_id
    GROUP BY b.id
    ORDER BY b.name`
    // ... scan results
}
```

**新增分组数据结构** (handler 使用):
```go
// CategoryWithBlogs extends CategoryWithBlogCount with blog list for sidebar grouping.
type CategoryWithBlogs struct {
    Category model.Category
    Blogs    []BlogWithCount
}

// GroupedSidebarData holds data for category-grouped sidebar rendering.
type GroupedSidebarData struct {
    Categories    []CategoryWithBlogs  // blog_count > 0
    Uncategorized []BlogWithCount      // category_id = NULL
}
```

**新增分组查询方法**:
```go
// ListBlogsGroupedByCategory returns blogs grouped by category for sidebar.
// Categories with blog_count > 0 first, uncategorized blogs at the end.
func (db *Database) ListBlogsGroupedByCategory() (GroupedSidebarData, error) {
    // 1. 获取分类 + blog count (已过滤 blog_count > 0)
    categories, err := db.ListCategoriesWithBlogCount()
    if err != nil {
        return GroupedSidebarData{}, err
    }
    
    // 2. 获取所有 blogs + category_id
    blogs, err := db.ListBlogsWithCounts()
    if err != nil {
        return GroupedSidebarData{}, err
    }
    
    // 3. 分组逻辑
    var grouped []CategoryWithBlogs
    var uncategorized []BlogWithCount
    
    for _, cat := range categories {
        if cat.BlogCount > 0 {
            var catBlogs []BlogWithCount
            for _, blog := range blogs {
                if blog.CategoryID != nil && *blog.CategoryID == cat.ID {
                    catBlogs = append(catBlogs, blog)
                }
            }
            grouped = append(grouped, CategoryWithBlogs{
                Category: cat.Category,
                Blogs:    catBlogs,
            })
        }
    }
    
    // 4. 未分类 blogs
    for _, blog := range blogs {
        if blog.CategoryID == nil {
            uncategorized = append(uncategorized, blog)
        }
    }
    
    return GroupedSidebarData{
        Categories:    grouped,
        Uncategorized: uncategorized,
    }, nil
}
```

---

### `assets/static/styles.css` (config, static)

**Analog:** `assets/static/styles.css` (现有 sidebar/nav-link/blog-item 样式)

**nav-section-title 模式** (lines 173-180):
```css
.nav-section-title {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
  margin-bottom: 0.5rem;
  padding: 0 0.75rem;
}
```

**blog-item 模式** (lines 212-232):
```css
.blog-item {
  display: block;
  padding: 0.5rem 0.75rem;
  border-radius: 6px;
  color: var(--text-primary);
  text-decoration: none;
  transition: background-color var(--transition-speed) ease;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.blog-item:hover {
  background-color: var(--bg-elevated);
  text-decoration: none;
}

.blog-item.active {
  background-color: var(--bg-elevated);
  color: var(--accent);
}
```

**新增分类分组样式** (参考 mockup 方案C, `.superpowers/mockups/category-sidebar-grouping.html` lines 373-470):
```css
/* ============================================
   Category Group Styles (Sidebar)
   ============================================ */

.category-group {
  margin-bottom: 0.75rem;
}

.category-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-secondary);
  background-color: var(--bg-elevated);
  border-radius: 6px;
  transition: all var(--transition-speed) ease;
  margin-bottom: 0.25rem;
}

.category-header:hover {
  background-color: var(--bg-surface);
  color: var(--text-primary);
}

.category-header .expand-icon {
  width: 12px;
  height: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 250ms ease;
  color: var(--text-secondary);
}

.category-header.collapsed .expand-icon {
  transform: rotate(-90deg);
}

.category-count {
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-left: auto;
}

.category-blogs {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding-left: 1rem;
  overflow: hidden;
  max-height: 500px;
  transition: max-height 250ms ease, opacity 150ms ease;
}

.category-group.collapsed .category-blogs {
  max-height: 0;
  opacity: 0;
}

/* Uncategorized section - at the END with separator */
.uncategorized-section {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--border);
}

.uncategorized-header {
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
  text-transform: uppercase;
  margin-bottom: 0.25rem;
  padding: 0 0.75rem;
}

/* Chevron icon style */
.icon-chevron {
  width: 12px;
  height: 12px;
  fill: none;
  stroke: currentColor;
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}
```

---

### `assets/js/sidebar.js` (utility, client-side)

**Analog:** `assets/templates/base.gohtml` (localStorage pattern)

**localStorage 主题持久化模式** (base.gohtml lines 5-22):
```javascript
(function() {
  var radios = document.querySelectorAll('input[name="theme"]');
  var stored = localStorage.getItem('theme') || 'system';

  // Set initial checked state
  var initial = document.getElementById('theme-' + stored);
  if (initial) initial.checked = true;

  function updateTheme(value) {
    localStorage.setItem('theme', value);
    var isDark = value === 'dark' ||
      (value === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches);
    document.documentElement.classList.toggle('dark', isDark);
  }

  radios.forEach(function(r) {
    r.addEventListener('change', function(e) {
      updateTheme(e.target.value);
    });
  });
})();
```

**新增 sidebar.js 内容** (参考 mockup lines 864-925, D-04):
```javascript
// sidebar.js - Category expand/collapse persistence

(function() {
  var STORAGE_KEY = 'sidebar-category-expand-state';

  // Toggle category expansion
  function toggleCategory(header) {
    var group = header.parentElement;
    var isCollapsed = group.classList.contains('collapsed');

    if (isCollapsed) {
      group.classList.remove('collapsed');
    } else {
      group.classList.add('collapsed');
    }

    // Persist to localStorage
    saveExpandState();
  }

  // Save expand state to localStorage
  function saveExpandState() {
    var state = {};
    document.querySelectorAll('.category-group').forEach(function(group) {
      var category = group.dataset.category;
      state[category] = !group.classList.contains('collapsed');
    });
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
  }

  // Load expand state from localStorage
  function loadExpandState() {
    var saved = localStorage.getItem(STORAGE_KEY);
    if (saved) {
      var state = JSON.parse(saved);
      document.querySelectorAll('.category-group').forEach(function(group) {
        var category = group.dataset.category;
        if (state[category] === false) {
          group.classList.add('collapsed');
        }
      });
    }
  }

  // Initialize on page load
  document.addEventListener('DOMContentLoaded', loadExpandState);

  // Re-apply after HTMX swaps (sidebar refreshes)
  document.body.addEventListener('htmx:afterSwap', function(evt) {
    if (evt.detail.target.id === 'blog-list') {
      loadExpandState();
    }
  });

  // Expose toggleCategory globally for onclick handlers
  window.toggleCategory = toggleCategory;
})();
```

**集成方式:**
- 在 `assets/templates/base.gohtml` 的 `{{define "scripts"}}` block 中添加 `<script src="/static/sidebar.js"></script>`
- 或直接内嵌到 sidebar.gohtml 中

---

## Shared Patterns

### HTMX 导航模式
**Source:** `assets/templates/partials/blog-list.gohtml`
**Apply to:** 所有 blog-item 链接（分类分组和未分类）

```html
<a href="/articles?blog={{.ID}}"
   hx-get="/articles?blog={{.ID}}"
   hx-target="#main-content"
   hx-push-url="true"
   hx-on:click="document.querySelectorAll('.sidebar-nav .nav-link, .blog-item').forEach(el => el.classList.remove('active')); this.classList.add('active'); document.getElementById('sidebar-toggle').checked = false;"
   class="blog-item{{if eq $.CurrentBlogID .ID}} active{{end}}">
    {{.Name}}
</a>
```

### CLI Flag 验证模式
**Source:** `internal/cli/commands/article.go` lines 188-200
**Apply to:** --category flag 验证

```go
// 查询数据库验证存在性
if name != "" {
    entity, err := db.GetByName(name)
    if err != nil {
        fmt.Fprintf(os.Stderr, "查询失败: %v\n", err)
        os.Exit(1)
    }
    if entity == nil {
        fmt.Fprintf(os.Stderr, "'%s' 不存在\n", name)
        os.Exit(1)
    }
}
```

### localStorage 状态持久化模式
**Source:** `assets/templates/base.gohtml` lines 5-22
**Apply to:** sidebar.js 分类折叠状态

```javascript
var stored = localStorage.getItem('key') || 'default';
localStorage.setItem('key', value);

// HTMX swap 后重新应用状态
document.body.addEventListener('htmx:afterSwap', function(evt) {
  // re-apply saved state
});
```

### CSS 变量主题适配模式
**Source:** `assets/static/styles.css` lines 4-61
**Apply to:** 所有新增 CSS 类

```css
/* Light theme (default) */
:root {
  --bg-primary: #FAF8F5;
  --bg-surface: #FFFFFF;
  --bg-elevated: #F5F3F0;
  --text-primary: #37352F;
  --text-secondary: #6B6B6B;
  --border: #E5E3E0;
  --accent: #2563EB;
}

/* Dark theme */
html.dark {
  --bg-primary: #121212;
  --bg-surface: #1e1e1e;
  --bg-elevated: #2d2d2d;
  --text-primary: #e0e0e0;
  --text-secondary: #a0a0a0;
  --border: #333333;
  --accent: #64b5f6;
}
```

---

## No Analog Found

所有文件都有现有 analog:

| File | Role | Data Flow | Analog Found |
|------|------|-----------|--------------|
| sidebar.gohtml | template | request-response | ✓ 现有 sidebar.gohtml |
| category-group.gohtml | template | request-response | ✓ blog-list.gohtml |
| article.go | CLI command | request-response | ✓ 现有 article.go |
| database.go | model | CRUD | ✓ 现有 database.go |
| styles.css | config | static | ✓ 现有 styles.css |
| sidebar.js | utility | client-side | ✓ base.gohtml localStorage |

---

## Metadata

**Analog search scope:**
- `assets/templates/**/*.gohtml`
- `assets/static/*.css`
- `internal/cli/commands/*.go`
- `internal/storage/*.go`
- `internal/server/*.go`
- `.superpowers/mockups/*.html`

**Files scanned:** 19 template files, 1 CSS file, 3 Go files, 1 mockup

**Pattern extraction date:** 2026-05-09

---

## Key Insights

1. **HTMX 导航模式成熟:** blog-list.gohtml 提供完整的 HTMX 导航模式，可直接复制到分类分组中的 blog-item。

2. **CLI Flag 验证模式可复用:** --blog 验证逻辑可直接复制用于 --category，只需替换 GetBlogByName → GetCategoryByName。

3. **localStorage 模式已验证:** base.gohtml 的主题持久化证明了 localStorage + HTMX afterSwap 重新应用的模式有效。

4. **CSS 主题适配完善:** 所有新增样式只需使用现有 CSS 变量，无需手动处理 light/dark 主题切换。

5. **GetCategoryByName 需新增:** database.go 中不存在此方法，需按 GetBlogByName 模式新增。

6. **分组数据结构需定义:** handler 需组装 GroupedSidebarData 传递给模板，包含 Categories + Uncategorized。