# Phase 20: Blog Preview - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-11
**Phase:** 20-Blog Preview
**Areas discussed:** 预览按钮位置, 预览页面布局, 错误显示方式, 保存后跳转

---

## 预览按钮位置

| Option | Description | Selected |
|--------|-------------|----------|
| Inline 并列 | Preview 和 Add Blog 按钮并排在表单底部 | ✓ |
| 下拉菜单 | 主按钮 Add Blog，下拉可选 Preview First | |
| 链接样式 | Add Blog 是按钮，Preview 是链接 | |

**User's choice:** Inline 并列
**Notes:** 用户选择方案 A，认为操作直观，与现有表单风格一致

---

## 预览页面布局

| Option | Description | Selected |
|--------|-------------|----------|
| 卡片列表 | 每篇文章一个卡片，显示标题、时间、链接 | ✓ |
| 紧凑表格 | 表格形式展示，信息密度高 | |
| 简单列表 | 极简设计，标题 + 时间 | |

**User's choice:** 卡片列表
**Notes:** 用户选择方案 A，认为与主页风格一致，视觉层次清晰

---

## 错误显示方式

| Option | Description | Selected |
|--------|-------------|----------|
| Inline 错误提示 | 页面内显示红色错误框，可返回修改或强制保存 | ✓ |
| 模态弹窗 | 强制返回修改，不允许保存无效 URL | |
| Toast 提示 | 简短提示，禁用 Save 按钮 | |

**User's choice:** Inline 错误提示
**Notes:** 用户选择方案 A，认为可自主选择下一步，错误信息详细

---

## 保存后跳转

| Option | Description | Selected |
|--------|-------------|----------|
| Settings 页面 | 返回 Settings，流程闭环 | ✓ |
| Articles 主页 | 直接跳转到主页，后台同步 | |
| 停留在预览页 | 显示成功消息和同步进度 | |

**User's choice:** Settings 页面
**Notes:** 用户选择方案 A，认为流程闭环，可验证博客已添加

---

## Claude's Discretion

- 卡片样式细节（间距、字体）— 按现有 article-card.gohtml 风格
- 错误消息具体措辞 — 与现有风格一致
- Preview 按钮样式（颜色、图标）— 与 Add Blog 按钮区分

## Deferred Ideas

None — discussion stayed within phase scope.