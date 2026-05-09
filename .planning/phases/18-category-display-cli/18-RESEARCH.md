# Phase 18: Category Display & CLI - Research

**Researched:** 2026-05-09
**Domain:** Go templates + HTMX for UI category grouping; Cobra CLI for category filtering
**Confidence:** HIGH

## Summary

Phase 18 实现订阅列表按分类分层展示和 CLI `article list --category` 过滤功能。UI 部分需要改造 sidebar.gohtml 渲染逻辑，引入 localStorage 持久化折叠状态；CLI 部分需要扩展 article.go 的 --category flag 并修改数据库查询逻辑。

现有基础设施完备：
- ListBlogsWithCounts() 已返回 category_id 字段
- ListCategoriesWithBlogCount() 已实现（Phase 16）
- ListFilterOptions 结构体存在，可直接扩展
- localStorage 已用于主题和视图偏好，模式成熟
- HTMX 导航模式成熟（hx-get + hx-target + hx-push-url）

**Primary recommendation:** 方案C 设计（分类在前，未分类在后）+ localStorage 'sidebar-category-expand-state' 键 + CLI --category 验证复用 --blog 模式。

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** 分类在前，未分类在后（类似原型方案C）
  - sidebar.gohtml 渲染顺序：categories → separator → uncategorized blogs
- **D-02:** 未分类 blog 显示在底部，有分隔线 + "未分类" 标签
  - `<div class="separator">` + `<div class="uncategorized-header">未分类</div>`
- **D-03:** 空分类不显示
  - 模板渲染时过滤 blog_count > 0 的分类
- **D-04:** 分类标题可展开/折叠，状态持久化到 localStorage
  - `hx-on:click` toggle + localStorage 读写
- **D-05:** 默认全部展开
  - 初始渲染无 collapsed class
- **D-06:** 按分类名称筛选（如 `--category tech`）
  - article.go 新增 `cmd.Flags().String("category", "", "分类名称筛选")`
- **D-07:** 统一组合过滤（加入 ListFilterOptions.CategoryName）
  - database.go 扩展 ListFilterOptions struct，添加 CategoryName string 字段
- **D-08:** 验证分类存在性（查询分类表，不存在则报错退出）
  - runList 函数调用 `db.GetCategoryByName(categoryName)` 验证

### Claude's Discretion
- 分类标题具体样式细节（字体大小、hover 效果、chevron 样式）
- localStorage 键名设计（如 `sidebar-category-expand-state-v1`）
- CLI 错误消息格式（与现有风格一致即可）

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CATG-08 | Subscriptions 按分类分层展示 | sidebar.gohtml 改造 + blog-list.gohtml 替换为 category-grouped 渲染 |
| CATG-09 | 未分类 blog 在 Subscriptions 顶层显示 | 方案C 设计：分类在前，分隔线 + "未分类" 标签在底部 |
| CATG-10 | CLI article list --category 过滤 | article.go 新增 --category flag + ListFilterOptions 扩展 + GetCategoryByName 方法 |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Blog 按分类分组渲染 | Frontend Server (Go template) | — | Go templates 负责 HTML 渲染，HTMX 处理交互 |
| 折叠状态持久化 | Browser (localStorage) | — | 状态仅存客户端，无需服务器参与 |
| CLI 分类过滤参数 | CLI (Cobra) | — | Cobra flag 解析，调用 storage 方法 |
| 分类验证查询 | API / Backend (Database) | — | GetCategoryByName 方法需新增 |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| HTMX | 2.0.10 [VERIFIED] | Dynamic UI updates without full reload | 项目核心架构，已嵌入 assets/static/htmx.min.js |
| Go templates | Go 1.22+ [CITED: go.dev] | Server-side HTML rendering | Go-native 方案，匹配项目风格 |
| Cobra | Latest [VERIFIED: go.mod] | CLI framework | Phase 17 已用于分类 CRUD CLI |
| SQLite (modernc.org/sqlite) | Latest [VERIFIED] | Database | 项目标准数据库层 |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| localStorage API | Native browser | 客户端状态持久化 | 已用于主题/视图偏好，模式成熟 |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| localStorage | Cookie / SessionStorage | localStorage 持久化跨 session，更适合折叠状态 |

**Installation:**
无需新增安装 — 所有依赖已在项目中。

**Version verification:**
```bash
# HTMX 已嵌入
ls -lh assets/static/htmx.min.js
# 输出确认存在（文件大小约 50KB）

# Go 版本
go version
# 输出: go version go1.22.x

# SQLite driver
grep modernc.org/sqlite go.mod
```

## Architecture Patterns

### System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                      Browser / Client                        │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Sidebar Component (sidebar.gohtml)                   │  │
│  │  ├─ Category Group Header (click → toggle collapse)  │  │
│  │  ├─ Blog List (hx-get → navigate to articles)        │  │
│  │  └─ Uncategorized Section (separator + blogs)        │  │
│  └──────────────────────────────────────────────────────┘  │
│         │ localStorage 'sidebar-category-expand-state'      │
│         │ (save/load collapsed state per category)          │
└─────────────────────────────────────────────────────────────┘
         │ ↑ HTMX hx-get /hx-swap          │ ↓ JavaScript
         │                                  │
┌─────────────────────────────────────────────────────────────┐
│                   Frontend Server (Go)                       │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ HTTP Handlers                                         │  │
│  │  ├─ handleBlogList → ListBlogsWithCounts()           │  │
│  │  ├─ handleCategoryGrouped (NEW) → grouped data       │  │
│  └──────────────────────────────────────────────────────┘  │
│         │ Call storage methods                               │
└─────────────────────────────────────────────────────────────┘
         │
         │ SQL queries (LEFT JOIN + GROUP BY)
         │
┌─────────────────────────────────────────────────────────────┐
│              Database / Storage Layer                        │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ database.go                                           │  │
│  │  ├─ ListBlogsWithCounts() → category_id field        │  │
│  │  ├─ ListCategoriesWithBlogCount() → blog_count       │  │
│  │  ├─ GetCategoryByName(name) (NEW) → validate         │  │
│  │  ├─ ListFilterOptions.CategoryName (NEW field)       │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                      CLI (article.go)                        │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ NewListCmd                                            │  │
│  │  ├─ --category flag (NEW)                             │  │
│  │  ├─ runList() → parse flags                           │  │
│  │  ├─ GetCategoryByName() → validate                    │  │
│  │  ├─ ListArticlesWithFilters(opts) → query             │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Recommended Project Structure
```
assets/templates/
├── partials/
│   ├── sidebar.gohtml          # 改造：替换 blog-list 渲染逻辑
│   ├── blog-list.gohtml        # 保留（可能用于其他场景）
│   ├── category-group.gohtml   # 新增：分类分组渲染片段
│   └── category-header.gohtml  # 新增：可折叠分类标题

internal/
├── cli/commands/
│   └── article.go              # 改造：新增 --category flag
├── storage/
│   └── database.go             # 改造：GetCategoryByName + ListFilterOptions 扩展
├── server/
│   └── handlers.go             # 改造：handleBlogList 返回分类分组数据
```

### Pattern 1: HTMX Blog Navigation (Existing)
**What:** 点击 blog 链接触发 HTMX 请求，替换内容区域并更新 URL
**When to use:** 所有 blog 导航交互
**Example:**
```html
<!-- Source: assets/templates/partials/blog-list.gohtml -->
<a href="/articles?blog={{.ID}}"
   hx-get="/articles?blog={{.ID}}"
   hx-target="#main-content"
   hx-push-url="true"
   hx-on:click="document.querySelectorAll('.blog-item').forEach(el => el.classList.remove('active')); this.classList.add('active');"
   class="blog-item{{if eq $.CurrentBlogID .ID}} active{{end}}">
    {{.Name}}
</a>
```

### Pattern 2: localStorage State Persistence (Existing)
**What:** 使用 localStorage 存储客户端状态（主题、视图偏好）
**When to use:** 折叠状态持久化，无需服务器参与
**Example:**
```javascript
// Source: assets/templates/base.gohtml (theme persistence pattern)
var stored = localStorage.getItem('theme') || 'system';
localStorage.setItem('theme', value);

// Mockup pattern for category collapse state
function saveExpandState() {
    const state = {};
    document.querySelectorAll('.category-group').forEach(group => {
        const category = group.dataset.category;
        state[category] = !group.classList.contains('collapsed');
    });
    localStorage.setItem('sidebar-category-expand-state', JSON.stringify(state));
}

function loadExpandState() {
    const saved = localStorage.getItem('sidebar-category-expand-state');
    if (saved) {
        const state = JSON.parse(saved);
        document.querySelectorAll('.category-group').forEach(group => {
            const category = group.dataset.category;
            if (state[category] === false) {
                group.classList.add('collapsed');
            }
        });
    }
}
```

### Pattern 3: CLI Flag Validation (Existing)
**What:** CLI 参数验证：查询数据库 → 不存在 → 报错退出
**When to use:** --category 验证复用 --blog 验证模式
**Example:**
```go
// Source: internal/cli/commands/article.go:189-200
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

### Anti-Patterns to Avoid
- **Anti-Pattern:** 将折叠状态发送到服务器存储
  - **Why it's bad:** localStorage 已足够，服务器存储增加复杂度和延迟
  - **What to do instead:** 使用 localStorage + JSON.stringify/state 对象

- **Anti-Pattern:** 在模板中直接查询数据库
  - **Why it's bad:** 违反分层原则，handler 应准备数据
  - **What to do instead:** handler.go 中组装分组数据传递给模板

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| 分类分组数据结构 | 手动在模板中循环分类和 blog | handler 组装 GroupedBlogs struct | 数据分组逻辑应在 handler 层，模板只渲染 |
| 折叠状态管理 | 自定义状态管理类 | localStorage + hx-on:click | 简单需求，浏览器原生 API 足够 |
| CLI 参数解析 | 手动解析命令行参数 | Cobra Flags() | 项目标准 CLI 框架，已验证 |

**Key insight:** 项目已建立成熟模式（Go templates + HTMX + localStorage + Cobra），不要引入新复杂性。

## Runtime State Inventory

> 此阶段无 rename/refactor/migration，跳过此节。

## Common Pitfalls

### Pitfall 1: Blog With Same Name in Multiple Categories
**What goes wrong:** 如果按 blog.name 分组，多个分类可能有同名 blog（实际不可能，blog 属于单一分类）
**Why it happens:** 理论误解分类关系（blog.category_id 是单值）
**How to avoid:** 理解 blog → category 是 N:1 关系，分组逻辑按 category_id
**Warning signs:** 模板中出现嵌套循环遍历 blogs 两次

### Pitfall 2: localStorage Key Collision
**What goes wrong:** 使用简单键名如 'sidebar-state' 可能与未来功能冲突
**Why it happens:** localStorage 是全局命名空间
**How to avoid:** 使用明确键名如 'sidebar-category-expand-state-v1' 或 'bw-sidebar-collapse'
**Warning signs:** localStorage 键名过短或通用

### Pitfall 3: Empty Category Display Logic
**What goes wrong:** 空分类显示会增加视觉噪音，违背 D-03 决策
**Why it happens:** ListCategoriesWithBlogCount() 返回 blog_count=0 的分类
**How to avoid:** handler 或模板过滤 blog_count > 0
**Warning signs:** 模板渲染空分类容器

### Pitfall 4: HTMX Swap Target Incorrect
**What goes wrong:** 点击分类标题触发折叠，但 HTMX swap 替换了错误元素
**Why it happens:** hx-target 指向错误选择器
**How to avoid:** 折叠用 hx-on:click（纯 JS），不用 HTMX swap
**Warning signs:** 分类标题有 hx-get/hx-swap 属性

### Pitfall 5: CLI Validation Error Message Format
**What goes wrong:** 错误消息与现有风格不一致（如 "Category tech not found" vs "博客 'Tech' 不存在"）
**Why it happens:** 未参考现有 --blog 验证消息格式
**How to avoid:** 复用现有格式：`分类 '{name}' 不存在`
**Warning signs:** 错误消息中英文混用或格式不一致

## Code Examples

Verified patterns from official sources:

### Blog List Rendering (Current Pattern)
```html
<!-- Source: assets/templates/partials/blog-list.gohtml -->
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

### Sidebar Container (HTMX Trigger)
```html
<!-- Source: assets/templates/partials/sidebar.gohtml:63-70 -->
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

### Category Dropdown (Phase 17)
```html
<!-- Source: assets/templates/partials/blog-edit-form.gohtml:13-28 -->
<div class="blog-edit-category">
    <select name="category_id"
            hx-trigger="categoryListUpdated from:body"
            hx-get="/blogs/{{.Blog.ID}}/edit"
            hx-target="#blog-{{.Blog.ID}}"
            hx-swap="outerHTML">
        <option value="">-- 未分类 --</option>
        {{range .Categories}}
        <option value="{{.ID}}"
                {{if eq $.Blog.CategoryID .ID}}selected{{end}}>
            {{.Name}}
        </option>
        {{end}}
    </select>
</div>
```

### ListFilterOptions Structure (Current)
```go
// Source: internal/storage/database.go:1132-1140
type ListFilterOptions struct {
    BlogName  string     // 博客名称筛选（空表示所有博客）
    IsRead    *bool      // 已读状态筛选（nil 表示所有状态）
    HasNote   *bool      // 备注状态筛选（nil 表示所有状态，false 表示无备注）
    AfterDate *time.Time // 日期筛选（nil 表示无限制）
    Limit     int        // 结果数量限制（0 表示无限制）
}
```

### ListArticlesWithFilters Method (Current)
```go
// Source: internal/storage/database.go:1144-1211
func (db *Database) ListArticlesWithFilters(opts ListFilterOptions) ([]model.ArticleWithBlog, error) {
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

    // 排序和限制
    query += " ORDER BY COALESCE(a.published_date, a.discovered_date) DESC"
    if opts.Limit > 0 {
        query += fmt.Sprintf(" LIMIT %d", opts.Limit)
    }

    rows, err := db.conn.Query(query, args...)
    // ... scan results
}
```

### ListCategoriesWithBlogCount Method (Phase 16)
```go
// Source: internal/storage/database.go:1245-1289
func (db *Database) ListCategoriesWithBlogCount() ([]CategoryWithBlogCount, error) {
    query := `SELECT
        c.id,
        c.name,
        c.created_at,
        COUNT(b.id) as blog_count
    FROM categories c
    LEFT JOIN blogs b ON b.category_id = c.id
    GROUP BY c.id
    ORDER BY c.name`

    rows, err := db.conn.Query(query)
    // ... scan results
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| CLI 单独 flag 解析 | Cobra framework | Phase 17 (v1.5) | 统一 CLI 架构，易扩展 |
| 手动 blog 分类分组 | 无（新增需求） | Phase 18 | 使用 Go template + handler 分组 |

**Deprecated/outdated:**
- 无 deprecated 内容 — 所有现有模式有效

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | GetCategoryByName 方法不存在，需新增 | Standard Stack | 中 — 需验证，若已存在则无需新增 |
| A2 | ListBlogsWithCounts() 已返回 category_id | Code Examples | 低 — 已在 database.go:356-414 验证 |
| A3 | localStorage 键名 'sidebar-category-expand-state' 无冲突 | Common Pitfalls | 低 — 键名明确，风险可控 |

**Verification needed:** 检查 GetCategoryByName 是否已实现（Phase 17 可能已添加）。

## Open Questions

1. **GetCategoryByName 方法是否存在？**
   - What we know: Phase 17 实现了分类 CRUD，但 grep 未找到 GetCategoryByName
   - What's unclear: 是否在其他文件中实现
   - Recommendation: 新增 GetCategoryByName 方法，复用 scanCategory 模式

2. **分类分组数据结构如何定义？**
   - What we know: ListCategoriesWithBlogCount() 返回 CategoryWithBlogCount，ListBlogsWithCounts() 返回 BlogWithCount
   - What's unclear: handler 如何组装分组数据传递给模板
   - Recommendation: 定义 GroupedSidebar struct:
     ```go
     type GroupedSidebar struct {
         Categories    []CategoryWithBlogs  // blog_count > 0
         Uncategorized []BlogWithCount      // category_id = NULL
     }
     type CategoryWithBlogs struct {
         Category model.Category
         Blogs    []BlogWithCount
     }
     ```

3. **是否需要新增 category-group.gohtml 模板？**
   - What we know: blog-list.gohtml 是现有渲染片段
   - What's unclear: 是否直接修改 sidebar.gohtml 或创建新 partial
   - Recommendation: 创建 category-group.gohtml partial，保持模块化

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go runtime | Server + CLI | ✓ | go1.22+ | — |
| SQLite database | Storage layer | ✓ | modernc.org/sqlite | — |
| HTMX library | UI navigation | ✓ | 2.0.10 | — |
| localStorage API | Client state persistence | ✓ | Native browser | — |

**Missing dependencies with no fallback:**
无 — 所有依赖可用。

**Missing dependencies with fallback:**
无。

## Validation Architecture

> workflow.nyquist_validation 未在 config.json 显式设置，默认启用。

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing package + testify |
| Config file | 无（Go 标准测试） |
| Quick run command | `go test ./internal/storage -run TestCategory -v` |
| Full suite command | `go test ./... -v` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CATG-08 | Blog 按分类分层展示 | integration | `go test ./internal/server -run TestHandleBlogList -v` | ❌ Wave 0 (需新增) |
| CATG-09 | 未分类 blog 在底部显示 | integration | `go test ./internal/server -run TestSidebarRendering -v` | ❌ Wave 0 (需新增) |
| CATG-10 | CLI --category 过滤 | unit | `go test ./internal/cli -run TestListCategoryFilter -v` | ❌ Wave 0 (需新增) |

### Sampling Rate
- **Per task commit:** `go test ./internal/storage ./internal/cli -v`
- **Per wave merge:** `go test ./... -v`
- **Phase gate:** Full suite green before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `internal/storage/database_test.go` — TestGetCategoryByName (CATG-10 验证)
- [ ] `internal/server/handlers_test.go` — TestHandleBlogListGrouped (CATG-08/09 UI 验证)
- [ ] `internal/cli/commands/article_test.go` — TestListCategoryFlag (CATG-10 CLI 验证)
- [ ] 共享 fixture: database_test.go 已有 testDB setup 模式

*(If no gaps: "None — existing test infrastructure covers all phase requirements")*

## Security Domain

> security_enforcement 未在 config.json 显式禁用，默认启用。

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | 单用户本地访问，无认证需求 |
| V3 Session Management | no | 无 session，localStorage 状态客户端 |
| V4 Access Control | no | 单用户，无权限控制 |
| V5 Input Validation | yes | Go template 自动 XSS 防护 + CLI 参数验证 |
| V6 Cryptography | no | 无敏感数据加密需求 |

### Known Threat Patterns for Go + HTMX

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| XSS in blog names | Tampering | Go template 自动 HTML escape [CITED: go.dev] |
| CLI parameter injection | Tampering | Cobra flag validation + GetCategoryByName 查询 |
| localStorage manipulation | Tampering | 无敏感数据，仅 UI 状态 |

## Sources

### Primary (HIGH confidence)
- [VERIFIED] assets/templates/partials/sidebar.gohtml — HTMX 触发模式、blog-list 渲染
- [VERIFIED] assets/templates/partials/blog-list.gohtml — blog 导航 pattern
- [VERIFIED] internal/storage/database.go — ListBlogsWithCounts、ListCategoriesWithBlogCount、ListFilterOptions
- [VERIFIED] internal/cli/commands/article.go — --blog validation pattern、NewListCmd flag pattern
- [CITED: go.dev] Go html/template — XSS 防护文档

### Secondary (MEDIUM confidence)
- [VERIFIED] .superpowers/mockups/category-sidebar-grouping.html — 方案C 设计、localStorage pattern
- [VERIFIED] assets/templates/base.gohtml — localStorage 主题/视图 persistence pattern
- [VERIFIED] internal/server/handlers.go — handleBlogList、handleSettings pattern

### Tertiary (LOW confidence)
- 无 — 所有核心信息已通过代码验证

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — 所有依赖已在项目中验证
- Architecture: HIGH — 现有模式成熟，可直接复用
- Pitfalls: MEDIUM — 需测试验证折叠状态和空分类过滤逻辑

**Research date:** 2026-05-09
**Valid until:** 30 days — 项目稳定，架构无重大变更计划