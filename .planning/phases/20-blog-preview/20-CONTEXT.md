# Phase 20: Blog Preview - Context

**Gathered:** 2026-05-11
**Status:** Ready for planning

<domain>
## Phase Boundary

在添加 Blog 之前，用户可以预览 Feed 解析结果，确认文章列表正确后再保存：
- PREV-01: 添加 blog 表单有预览按钮
- PREV-02: 点击预览触发临时 feed 解析
- PREV-03: 预览页面显示解析的文章列表（最多 20 条）
- PREV-04: 预览失败显示错误信息
- PREV-05: 预览页面有保存按钮（保存为正式 blog）
- PREV-06: 预览页面有返回修改按钮（返回添加表单）

此阶段专注于预览流程，不改变现有 Add Blog 直接保存的行为。

**In scope:**
- Preview 按钮（与 Add Blog 并排）
- Preview 页面（文章卡片列表）
- Preview Handler（调用 rss.ParseFeed）
- Preview Save Handler（调用 BlogService.AddBlog）
- 错误处理（Inline 显示 + Save Anyway 选项）

**Out of scope:**
- Edit Blog 预览（Phase 19 已完成编辑，预览仅用于新增场景）
- Feed URL 手动输入（现有 Auto-discovery 保持不变）
- 预览文章持久化（仅临时展示，不写入数据库）

</domain>

<decisions>
## Implementation Decisions

### 预览按钮位置（通过 Visual Companion 原型讨论）
- **D-01:** Preview 和 Add Blog 按钮并列
  - **Why:** 操作直观，用户可自主选择先预览或直接添加，与现有表单风格一致
  - **How to apply:** `add-blog-form.gohtml` 在 `.form-actions` 添加两个并列按钮

### 预览页面布局（通过 Visual Companion 原型讨论）
- **D-02:** 卡片列表展示文章（与主页风格一致）
  - **Why:** 视觉层次清晰，用户熟悉现有卡片样式，信息密度适中
  - **How to apply:** 创建 `preview-page.gohtml`，使用与 `article-card.gohtml` 相似的卡片样式，最多显示 20 条

### 错误显示方式（通过 Visual Companion 原型讨论）
- **D-03:** Inline 错误提示 + Save Anyway 选项
  - **Why:** 用户可自主选择下一步，错误信息详细，允许保存无效 URL 稍后修复
  - **How to apply:** 创建 `preview-error.gohtml` fragment，显示红色错误框 + Back to Edit / Save Anyway 按钮

### 保存后跳转（通过 Visual Companion 原型讨论）
- **D-04:** 跳转到 Settings 页面
  - **Why:** 流程闭环（Settings → Add → Preview → Save → Settings），可立即验证博客已添加
  - **How to apply:** `handleBlogPreviewSave` 保存成功后返回 Settings + 成功消息

### Claude's Discretion
- 卡片样式细节（间距、字体大小）— 按现有 article-card.gohtml 风格设计
- 错误消息具体措辞 — 与现有风格一致即可
- Preview 按钮样式（颜色、图标）— 与 Add Blog 按钮区分即可

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### UI 设计文档
- `docs/superpowers/specs/2026-05-11-blog-preview-design.md` — UI 设计决策详情，包含布局图、数据流、错误处理、测试场景

### Requirements & Roadmap
- `.planning/REQUIREMENTS.md` §PREV — PREV-01~06 需求定义
- `.planning/ROADMAP.md` §Phase 20 — Phase 目标、Success Criteria

### 项目架构
- `.planning/PROJECT.md` — 技术栈（Go templates + HTMX）、Key Decisions

### 现有代码模式
- `internal/rss/rss.go` — ParseFeed() 和 DiscoverFeedURL() 方法，用于解析 Feed
- `internal/service/blog_service.go` — AddBlog() 方法，用于保存博客
- `assets/templates/partials/add-blog-form.gohtml` — 现有添加表单结构
- `assets/templates/partials/article-card.gohtml` — 文章卡片样式参考
- `internal/server/handlers.go` — handleAddBlog 处理逻辑参考

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/rss/rss.go:ParseFeed()` — 解析 Feed URL，返回 []FeedArticle（标题、URL、时间、缩略图）
- `internal/rss/rss.go:DiscoverFeedURL()` — 自动发现 Feed URL（现有 Add Blog 使用）
- `internal/service/blog_service.go:AddBlog()` — 保存博客的 Service 方法
- `assets/templates/partials/article-card.gohtml` — 文章卡片样式，可复用于预览卡片
- `assets/templates/partials/add-blog-form.gohtml` — 现有表单，需扩展添加 Preview 按钮

### Established Patterns
- **HTMX 表单提交:** hx-post + hx-target + hx-swap="innerHTML"，返回 partial HTML
- **成功跳转:** Add Blog 成功后显示成功消息 + 操作按钮（View Articles / Back to Settings）
- **错误显示:** Inline 红色错误框 + 操作按钮（返回 / 重试）

### Integration Points
- `assets/templates/partials/add-blog-form.gohtml` — 添加 Preview 按钮（与 Add Blog 并排）
- `internal/server/routes.go` — 新增 POST `/blogs/preview` 和 POST `/blogs/preview/save` 路由
- `internal/server/handlers.go` — 新增 handleBlogPreview 和 handleBlogPreviewSave
- `assets/templates/partials/preview-page.gohtml` — 新建预览页面模板
- `assets/templates/partials/preview-article-card.gohtml` — 新建预览文章卡片模板

</code_context>

<specifics>
## Specific Ideas

- Preview 按钮样式：灰色背景（与 Add Blog 蓝色区分），文字 "Preview"
- 预览卡片：标题 + 时间（relative format）+ 外链图标（target="_blank"）
- 错误框：红色背景 (#fef2f2)，警告图标 "⚠"，具体错误消息
- Save Anyway 按钮：灰色背景，允许保存无效 URL
- 最多 20 条文章：超过时显示 "Showing 20 of N articles"

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 20-Blog Preview*
*Context gathered: 2026-05-11*