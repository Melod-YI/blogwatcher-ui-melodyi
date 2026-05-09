# Phase 18: Category Display & CLI - Context

**Gathered:** 2026-05-09
**Status:** Ready for planning

<domain>
## Phase Boundary

Subscriptions 分类分层展示（UI sidebar）+ CLI `article list --category` 过滤：
- CATG-08: Subscriptions 按分类分层展示
- CATG-09: 未分类 blog 在 Subscriptions 顶层显示（实际设计：分类在前，未分类在后）
- CATG-10: CLI article list --category 过滤

此阶段专注于 UI 分类分组展示和 CLI 分类过滤，Phase 17 已完成分类管理基础设施。

**In scope:**
- sidebar.gohtml 改造：分类分组渲染 + 展开/折叠交互
- 未分类 blog 在底部显示（分隔线 + "未分类" 标签）
- 空分类不显示
- CLI article.go 新增 --category flag
- ListFilterOptions 扩展 CategoryName 字段
- 分类存在性验证（类似 --blog 验证）

**Out of scope:**
- Phase 19 功能（SETT-01~05: Blog URL/Feed URL 编辑）
- 分类排序自定义（alphabetically 已满足）
- 分类颜色/图标

</domain>

<decisions>
## Implementation Decisions

### UI 分组结构（通过 frontend-design 原型讨论）
- **D-01:** 分类在前，未分类在后（类似原型方案C）
  - **Why:** 逻辑上用户先看到有组织的分类内容，未分类作为补充
  - **How to apply:** sidebar.gohtml 渲染顺序：categories → separator → uncategorized blogs
- **D-02:** 未分类 blog 显示在底部，有分隔线 + "未分类" 标签
  - **Why:** 清晰区分已分类和未分类内容，避免混淆
  - **How to apply:** `<div class="separator">` + `<div class="uncategorized-header">未分类</div>`
- **D-03:** 空分类不显示
  - **Why:** 减少视觉噪音，无 blog 的分类对用户无意义
  - **How to apply:** 模板渲染时过滤 blog_count > 0 的分类
- **D-04:** 分类标题可展开/折叠，状态持久化到 localStorage
  - **Why:** 大量 blog 时折叠可节省空间，持久化避免每次展开
  - **How to apply:** `hx-on:click` toggle + localStorage 读写
- **D-05:** 默认全部展开
  - **Why:** 新用户可见所有 blog，无需手动展开
  - **How to apply:** 初始渲染无 collapsed class

### CLI --category 过滤参数
- **D-06:** 按分类名称筛选（如 `--category tech`）
  - **Why:** 用户友好，比 ID 更直观
  - **How to apply:** article.go 新增 `cmd.Flags().String("category", "", "分类名称筛选")`
- **D-07:** 统一组合过滤（加入 ListFilterOptions.CategoryName）
  - **Why:** 与现有 --unread/--blog/--after/--not-noted 组合生效，灵活性高
  - **How to apply:** database.go 扩展 ListFilterOptions struct，添加 CategoryName string 字段
- **D-08:** 验证分类存在性（查询分类表，不存在则报错退出）
  - **Why:** 类似现有 --blog 验证模式，提供友好错误提示
  - **How to apply:** runList 函数调用 `db.GetCategoryByName(categoryName)` 验证

### Claude's Discretion
- 分类标题具体样式细节（字体大小、hover 效果、chevron 样式）
- localStorage 键名设计（如 `sidebar-category-expand-state-v1`）
- CLI 错误消息格式（与现有风格一致即可）

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements & Roadmap
- `.planning/REQUIREMENTS.md` — CATG-08/09/10 需求定义
- `.planning/ROADMAP.md` §Phase 18 — Phase 目标、Success Criteria

### 项目架构
- `.planning/PROJECT.md` — 技术栈（Go templates + HTMX）、Key Decisions

### UI 原型
- `.superpowers/mockups/category-sidebar-grouping.html` — 三个方案原型（A/B/C），用户选择方案C

### 现有代码模式
- `assets/templates/partials/sidebar.gohtml` — 现有 sidebar 结构，blog-list 渲染
- `assets/templates/partials/blog-list.gohtml` — 现有 blog 列表渲染模式
- `internal/cli/commands/article.go` — CLI list 命令结构，筛选参数处理
- `internal/storage/database.go` — ListBlogsWithCounts、ListCategoriesWithBlogCount 方法

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/storage/database.go:ListBlogsWithCounts()` — 已含 category_id 字段，可直接按分类分组
- `internal/storage/database.go:ListCategoriesWithBlogCount()` — Phase 16 已实现，获取分类 + blog 数量
- `internal/cli/commands/article.go:runList()` — CLI 筛选参数解析模式，可直接扩展 --category
- `internal/storage/database.go:ListFilterOptions` — 现有筛选结构，扩展 CategoryName 字段

### Established Patterns
- **Sidebar 渲染:** blog-list.gohtml 循环渲染，HTMX hx-get 导航
- **CLI 筛选:** ListFilterOptions 结构 + database.ListArticlesWithFilters() 方法
- **验证模式:** --blog flag 验证博客存在性（article.go:189-200），相同模式用于 --category
- **localStorage:** 前序阶段未使用，本阶段引入持久化折叠状态

### Integration Points
- `assets/templates/partials/sidebar.gohtml` — 替换 blog-list 渲染逻辑，改为分类分组
- `internal/cli/commands/article.go` — NewListCmd 新增 --category flag，runList 解析参数
- `internal/storage/database.go` — 扩展 ListFilterOptions struct，修改 ListArticlesWithFilters 查询逻辑
- 新增 handler: GetCategoryByName(name string) 用于 CLI 验证
- 新增 template: 可能需要 category-group.gohtml partial（分类分组渲染片段）

</code_context>

<specifics>
## Specific Ideas

- 分类标题样式参考方案C原型：背景色 + 右侧计数（灰色）+ chevron 旋转动画
- 未分类分隔线样式：1px solid var(--border-color)，上下间距 16px
- CLI 错误消息格式：`分类 '{name}' 不存在`（与现有 "博客 '{name}' 不存在" 风格一致）
- localStorage 键名：`sidebar-category-expand-state`（简单直接）

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 18-Category Display & CLI*
*Context gathered: 2026-05-09*