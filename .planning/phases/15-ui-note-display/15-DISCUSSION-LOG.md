# Phase 15: UI Note Display - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-08
**Phase:** 15-UI Note Display
**Areas discussed:** 备注按钮位置与样式、备注页面路由与布局、Markdown 渲染方案、Markdown 样式与主题

---

## 备注按钮位置与样式

| Option | Description | Selected |
|--------|-------------|----------|
| 放在 Summarize 之后 | 保持 article-actions 视觉一致性，所有按钮在同一行 | ✓ |
| 放在 Read/Unread 之后 | 可能与阅读状态按钮冲突，视觉优先级降低 | |
| 独立区域 | 需要修改卡片布局结构，增加复杂度 | |

**User's choice:** 自动决策
**Notes:** 用户选择让 Claude 自动决策，推荐放在 Summarize 之后，保持现有 action-actions 区域结构

---

## 备注页面路由与布局

| Option | Description | Selected |
|--------|-------------|----------|
| `/note/{id}` | 简洁明了，与现有路由风格一致 | ✓ |
| `/articles/{id}/note` | RESTful 风格，但路径较长 | |
| `/notes/{id}` | 复数形式，与其他路由风格不一致 | |

**User's choice:** 自动决策
**Notes:** 推荐 `/note/{id}`，页面布局为独立专注页面（无 sidebar），顶部显示标题和原文链接

---

## Markdown 渲染方案

| Option | Description | Selected |
|--------|-------------|----------|
| goldmark | Go 生态成熟库，内置 GFM 支持，服务器端渲染 | ✓ |
| blackfriday | 简单快速，但 GFM 支持较弱 | |
| JavaScript 库 | 需要客户端渲染，不符合项目架构 | |

**User's choice:** 自动决策
**Notes:** 推荐 goldmark + GFM 扩展，符合服务器渲染架构，支持表格、删除线、任务列表

---

## Markdown 样式与主题

| Option | Description | Selected |
|--------|-------------|----------|
| 使用现有 CSS 变量 | 自动适配 Light/Dark 主题，保持风格一致 | ✓ |
| 独立 Markdown 样式 | 需要单独定义主题切换逻辑 | |
| 外部 CSS 框架 | 增加 CSS 文件大小，可能与应用风格冲突 | |

**User's choice:** 自动决策
**Notes:** 推荐 .markdown-body 样式类，复用现有 CSS 变量系统，定义代码块、表格、链接等样式

---

## Claude's Discretion

用户选择自动决策，Claude 有以下灵活性：
- 备注按钮的具体图标选择（推荐 Heroicons document-text）
- Markdown 样式的具体细节（颜色、间距等）
- 备注页面顶部布局的具体样式

---

## Deferred Ideas

None — 讨论始终保持在阶段范围内。