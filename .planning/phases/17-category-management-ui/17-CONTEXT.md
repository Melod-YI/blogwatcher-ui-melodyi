# Phase 17: Category Management UI - Context

**Gathered:** 2026-05-08
**Status:** Ready for planning

<domain>
## Phase Boundary

为 BlogWatcher UI 设置页面添加分类管理功能，包括：
- 分类管理区（创建、编辑、删除分类）
- Blog 编辑时选择分类（下拉选择）

此阶段专注于 UI 交互，Phase 16 已完成数据库基础设施（categories 表 + blogs.category_id 字段）。

**In scope:**
- CATG-03: 设置页面添加分类管理区
- CATG-04: 用户可创建新分类（inline 输入）
- CATG-05: 用户可编辑分类名称（inline 编辑）
- CATG-06: 用户可删除分类（删除时 blog.category_id 置空）
- CATG-07: Blog 编辑时可选择分类（下拉选择）

**Out of scope:**
- Phase 18 功能（CATG-08~10: Subscriptions 分层展示、CLI 过滤）
- 分类排序/颜色/图标
- 多分类、分类嵌套

</domain>

<decisions>
## Implementation Decisions

### 分类管理区位置
- **D-01:** 分类管理区位于 Tracked Blogs 之前
  - **Why:** 逻辑上先有分类再分配 blog，新增 blog 时可以先看到分类选项
  - **How to apply:** settings-page.gohtml 中 Categories Section 渲染在 Add Blog Form 和 Tracked Blogs 之间

### 分类管理交互风格
- **D-02:** 分类管理使用 Inline 操作风格
  - **Why:** 与现有 blog name 编辑风格一致（click-to-edit），流畅直观
  - **How to apply:**
    - 创建：点击 "+ New Category" → inline 输入框 → Save/Cancel
    - 编辑：点击分类名称 → inline 编辑 → Enter/Save
    - 删除：点击 Delete → 确认 dialog → Delete 确认

### Blog 分类选择方式
- **D-03:** Blog 编辑表单内下拉选择分类
  - **Why:** 与现有编辑流程一致，点击 Edit 进入编辑模式后统一编辑 name + category
  - **How to apply:** blog-edit-form.gohtml 扩展，添加 `<select name="category_id">` 下拉框

### 删除分类行为
- **D-04:** 删除分类时，关联 blog 自动置空（变为 "未分类"）
  - **Why:** 符合 CATG-06 要求，操作简单，与删除 blog 的 cascade delete 风格一致
  - **How to apply:** 确认 dialog 显示关联 blog 数量，确认后执行删除 + blog.category_id 置空

### Claude's Discretion
- HTMX 路由设计、模板命名、CSS 样式细节 — 按现有模式自由设计
- 空分类删除是否需要确认 dialog — 可简化处理

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### 设计文档
- `docs/superpowers/specs/2026-05-08-category-management-ui-design.md` — UI 设计文档，包含交互流程、布局结构、技术实现要点

### Requirements & Roadmap
- `.planning/REQUIREMENTS.md` — CATG-03~07 需求定义
- `.planning/ROADMAP.md` §Phase 17 — Phase 目标、Success Criteria

### 项目架构
- `.planning/PROJECT.md` — 技术栈（Go templates + HTMX）、Key Decisions

### UI 原型参考
- `.superpowers/mockups/category-position.html` — 分类管理区位置选择原型
- `.superpowers/mockups/category-interaction.html` — 分类管理交互风格原型
- `.superpowers/mockups/blog-category-select.html` — Blog 分类选择方式原型
- `.superpowers/mockups/delete-category.html` — 删除分类行为原型

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/storage/database.go`: ListCategoriesWithBlogCount() — Phase 16 已实现，获取分类 + blog 数量
- `internal/storage/database.go`: CreateCategory(), UpdateCategoryName(), DeleteCategory() — Phase 16 已实现
- `internal/storage/database.go`: UpdateBlogCategory() — Phase 16 已实现，更新 blog.category_id
- `internal/model/model.go`: Category struct, BlogWithCount struct

### Established Patterns
- **Inline 编辑风格:** blog-edit-form.gohtml + blog-display-row.gohtml，点击 Edit 进入编辑模式，Save/Cancel 按钮
- **Click-to-edit:** blog name 可直接点击编辑，无需弹窗
- **Dialog 确认:** delete-blog-dialog.gohtml 使用 `<dialog>` 元素 + HTMX 确认删除
- **HTMX 交互:** hx-get/hx-post/hx-put/hx-delete + hx-target + hx-swap="outerHTML"

### Integration Points
- `internal/server/handlers.go`: handleSettings() — 需扩展，传递 Categories 数据
- `assets/templates/partials/settings-page.gohtml` — 需添加 Categories Section
- `assets/templates/partials/blog-edit-form.gohtml` — 需扩展，添加 category 下拉框
- 新增 Handler: handleCategoriesNew, handleCategoriesCreate, handleCategoryEdit, handleCategoryUpdate, handleCategoryDelete

</code_context>

<specifics>
## Specific Ideas

- 分类名称编辑与 blog name 编辑风格完全一致（click 名称直接编辑）
- 删除分类确认 dialog 显示关联 blog 列表（最多显示 5 个）
- 分类下拉框默认选项为 "-- 未分类 --"（空值）
- 编辑 blog 时，name 和 category 统一在表单内编辑，一次 Save 提交

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 17-Category Management UI*
*Context gathered: 2026-05-08*