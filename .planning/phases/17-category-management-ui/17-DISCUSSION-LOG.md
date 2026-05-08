# Phase 17: Category Management UI - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-08
**Phase:** 17-Category Management UI
**Areas discussed:** 分类管理区位置, 分类管理交互风格, Blog分类选择方式, 删除分类行为

---

## 分类管理区位置

| Option | Description | Selected |
|--------|-------------|----------|
| A - Tracked Blogs 之前 | 分类管理在最上方，用户先看到分类，再看到 blog 列表。逻辑上：先有分类，再分配 blog。 | ✓ |
| B - Tracked Blogs 之后 | 保持现有 Tracked Blogs 区不变，分类管理在其下方添加。不改变用户已熟悉的布局顺序。 | |
| C - 合并在同一区域 | 分类和 blog 列表合并在一个大的 "管理区"，用分隔线区分。更紧凑，但可能不够清晰。 | |

**User's choice:** A
**Notes:** 推荐选项，逻辑顺序清晰

---

## 分类管理交互风格

| Option | Description | Selected |
|--------|-------------|----------|
| A - Inline 操作 | 创建/编辑 inline，删除用 dialog。与现有 blog name 编辑风格一致（click-to-edit），流畅直观。 | ✓ |
| B - Modal 弹窗 | 点击按钮弹出对话框操作。操作更明确，适合新手用户。 | |

**User's choice:** A
**Notes:** 与现有 UI 交互风格保持一致

---

## Blog 分类选择方式

| Option | Description | Selected |
|--------|-------------|----------|
| A - 编辑表单内下拉选择 | 点击 Edit 进入编辑模式后，在表单中显示分类下拉框。与现有编辑流程一致。 | ✓ |
| B - Display 行内下拉选择 | 在 display 行直接显示分类下拉框，点击下拉框即可更改分类。无需进入编辑模式。 | |

**User's choice:** A
**Notes:** 与现有 blog 编辑流程统一，避免功能重叠

---

## 删除分类行为

| Option | Description | Selected |
|--------|-------------|----------|
| A - 简单删除 + 自动置空 | 删除分类时，直接将关联 blog 的 category_id 置空，变为 "未分类"。符合 CATG-06 要求。 | ✓ |
| B - 强制选择新分类 | 删除分类时，如果该分类下有 blog，要求用户选择将这些 blog 移到哪个新分类。 | |

**User's choice:** A
**Notes:** 符合 CATG-06 要求，操作简单，与删除 blog 的 cascade delete 风格一致

---

## Claude's Discretion

- HTMX 路由设计、模板命名、CSS 样式细节
- 空分类删除是否需要确认 dialog（可简化处理）

## Deferred Ideas

None — discussion stayed within phase scope.

---

*Discussion completed: 2026-05-08*