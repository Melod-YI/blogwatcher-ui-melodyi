# Phase 18: Category Display & CLI - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-09
**Phase:** 18-Category Display & CLI
**Areas discussed:** UI 分组结构（frontend-design 原型）, CLI --category 参数

---

## UI 分组结构（frontend-design 原型讨论）

通过 `.superpowers/mockups/category-sidebar-grouping.html` 原型交互讨论，展示三个方案：

| Option | Description | Selected |
|--------|-------------|----------|
| 方案 A | 极简主义：无分类标签，未分类 blog 在顶层无标识 | |
| 方案 B | 分组 + 计数徽章：背景块 + 蓝色 badge，"未分类"标签在顶层 | |
| 方案 C | 混合：分类在前，未分类在后（分隔线区分） | ✓ |

**User's choice:** 方案 C + 空分类不显示
**Notes:** "分类在前，未分类在后。空分类不显示。整体比较类似方案C。"

**Additional decisions from prototype interaction:**
- 未分类分隔线样式：底部分隔线 + "未分类" 标签
- 分类标题样式：背景色 + 右侧计数（灰色）+ chevron 旋转动画
- 展开/折叠：默认全部展开，状态持久化到 localStorage

---

## CLI --category 参数

| Option | Description | Selected |
|--------|-------------|----------|
| 按名称筛选 | 用户友好，如 '--category tech'，但需查询分类ID | ✓ |
| 按ID筛选 | 精确匹配，如 '--category 1'，无需额外查询 | |

**User's choice:** 按名称筛选（推荐）
**Notes:** 用户友好，符合 CLI 使用习惯

---

## CLI 组合过滤

| Option | Description | Selected |
|--------|-------------|----------|
| 统一组合 | 现有 ListFilterOptions.IsRead、BlogName、AfterDate、HasNote，--category 加入该结构 | ✓ |
| 单独过滤 | --category 独立过滤，不与其他参数组合 | |

**User's choice:** 统一组合（推荐）
**Notes:** 与现有筛选参数无缝组合，灵活性高

---

## CLI 错误处理

| Option | Description | Selected |
|--------|-------------|----------|
| 验证存在 | 查询分类表验证存在性，不存在则报错退出（类似现有 --blog 验证） | ✓ |
| 不验证 | 直接过滤 blog.category_id，不验证分类存在 | |

**User's choice:** 验证存在（推荐）
**Notes:** 与现有 --blog 验证模式一致，提供友好错误提示

---

## Claude's Discretion

- 分类标题样式细节（字体大小、hover 效果、chevron 样式）
- localStorage 键名设计
- CLI 错误消息格式（与现有风格一致）

---

## Deferred Ideas

None — discussion stayed within phase scope.