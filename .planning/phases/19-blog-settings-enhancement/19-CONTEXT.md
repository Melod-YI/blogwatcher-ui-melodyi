# Phase 19: Blog Settings Enhancement - Context

**Gathered:** 2026-05-09
**Status:** Ready for planning

<domain>
## Phase Boundary

设置页面显示和编辑 Blog URL 和 Feed URL：
- SETT-01: 设置页面显示 Blog URL 和 Feed URL
- SETT-02: 设置页面可编辑 Blog URL（inline 编辑）
- SETT-03: 设置页面可编辑 Feed URL（inline 编辑）
- SETT-04: 编辑时验证 URL 格式（HTTP/HTTPS）
- SETT-05: 保存后立即更新数据库

此阶段专注于扩展现有博客编辑表单，添加 URL 显示和编辑功能。

**In scope:**
- `blog-display-row.gohtml` 添加两列 URL 显示
- `blog-edit-form.gohtml` 扩展添加 URL/Feed URL 输入框
- `handleBlogUpdate` 验证 + 更新 URL 字段
- 数据库 `UpdateBlog` 方法扩展

**Out of scope:**
- Phase 20 功能（PREV-01~06: 添加 Blog 预览）
- Blog URL 变化后自动更新文章 URL
- Feed URL 变化后自动重新扫描

</domain>

<decisions>
## Implementation Decisions

### URL 显示位置（通过 frontend-design 原型讨论）
- **D-01:** 两列并排显示（Blog URL 左，Feed URL 右）
  - **Why:** 信息一目了然，两 URL 并列对比清晰，与现有卡片布局兼容
  - **How to apply:** `blog-display-row.gohtml` 内 flex 布局两列显示
- **D-01a:** URL 作为可点击链接（target="_blank"）
  - **Why:** 用户可直接打开查看目标博客或 feed
  - **How to apply:** `<a href="{{.Blog.URL}}" target="_blank">`

### 编辑交互模式（通过 frontend-design 原型讨论）
- **D-02:** 统一编辑表单（name + Blog URL + Feed URL + category）
  - **Why:** 与现有 Edit 流程一致，用户习惯已建立，一次编辑所有字段
  - **How to apply:** 扩展 `blog-edit-form.gohtml`，在 name 和 category 之间添加两个 URL 输入框

### 验证反馈方式（通过 frontend-design 原型讨论）
- **D-03:** Inline 错误提示（输入框下方显示红色错误）
  - **Why:** 错误位置精确，阻止提交无效数据，与表单验证通用风格一致
  - **How to apply:** HTML5 `pattern="^https?://.*"` + CSS `input:invalid` 样式
- **D-03a:** 空值允许（nullable 字段）
  - **Why:** Blog URL/Feed URL 可以为空，不强制填写
  - **How to apply:** 后端验证 `url == ""` 时跳过验证

### Claude's Discretion
- URL 输入框宽度、错误提示字体大小等样式细节 — 按现有表单风格自由设计
- 后端验证错误消息具体措辞 — 与现有风格一致即可

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### UI 设计文档
- `docs/superpowers/specs/2026-05-09-blog-settings-enhancement-design.md` — UI 设计决策详情，包含布局图、验证规则、数据流

### Requirements & Roadmap
- `.planning/REQUIREMENTS.md` — SETT-01~05 需求定义
- `.planning/ROADMAP.md` §Phase 19 — Phase 目标、Success Criteria

### 项目架构
- `.planning/PROJECT.md` — 技术栈（Go templates + HTMX）、Key Decisions

### 现有代码模式
- `assets/templates/partials/blog-display-row.gohtml` — 现有博客卡片显示模式
- `assets/templates/partials/blog-edit-form.gohtml` — 现有编辑表单结构（name + category）
- `internal/server/handlers.go` — handleBlogUpdate 处理逻辑

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `assets/templates/partials/blog-edit-form.gohtml` — 现有编辑表单模板，可直接扩展添加 URL 输入框
- `internal/server/handlers.go:handleBlogUpdate()` — 现有更新处理逻辑，可扩展解析 URL 参数
- `internal/model/model.go:Blog` struct — 已有 URL 和 FeedURL 字段

### Established Patterns
- **Inline 编辑风格:** blog-edit-form.gohtml + blog-display-row.gohtml，点击 Edit 进入编辑模式，Save/Cancel 按钮
- **HTMX 交互:** hx-put 提交 + hx-target + hx-swap="outerHTML" 刷新卡片
- **验证模式:** 后端验证返回错误，前端显示（如 Phase 17 分类名称验证）

### Integration Points
- `assets/templates/partials/blog-display-row.gohtml` — 添加两列 URL 显示区域
- `assets/templates/partials/blog-edit-form.gohtml` — 扩展表单添加 URL/Feed URL 输入框
- `internal/server/handlers.go:handleBlogUpdate()` — 解析 name + url + feed_url + category_id 参数
- `internal/storage/database.go` — 扩展 UpdateBlog 方法或新增 UpdateBlogURLs 方法

</code_context>

<specifics>
## Specific Ideas

- 两列并排布局：`display: flex; gap: 24px;` 左侧 Blog URL，右侧 Feed URL
- 统一编辑表单字段顺序：名称 → Blog URL → Feed URL → 分类 → Save/Cancel
- HTML5 pattern 验证：`pattern="^https?://.*"` 阻止非 HTTP/HTTPS URL 提交
- 错误提示样式：输入框下方红色文字 `color: var(--danger); font-size: 13px; margin-top: 4px;`
- 空值允许：后端验证 `url == ""` 时返回 nil（不强制填写）

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 19-Blog Settings Enhancement*
*Context gathered: 2026-05-09*