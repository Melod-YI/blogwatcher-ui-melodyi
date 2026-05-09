---
phase: 19-blog-settings-enhancement
plan: 02
subsystem: settings
tags: [ui, form, validation, url-input, css]
requires: [19-01]
provides: [SETT-02, SETT-03, SETT-04]
affects: [blog-edit-form.gohtml, styles.css]
tech_stack:
  added: []
  patterns: [HTML5 pattern validation, CSS :invalid pseudo-class, inline error messages]
key_files:
  created: []
  modified:
    - assets/templates/partials/blog-edit-form.gohtml
    - assets/static/styles.css
decisions: [D-02, D-03]
metrics:
  duration: 3min
  tasks: 2
  files: 2
  completed_date: "2026-05-09"
---

# Phase 19 Plan 02: Blog Edit Form URL Inputs Summary

## One-liner

在 blog-edit-form.gohtml 中添加 Blog URL 和 Feed URL 输入框，实现 HTML5 pattern 验证和 inline 错误提示（per D-02, D-03）。

## Changes

### Task 01: 扩展 blog-edit-form.gohtml 添加 URL 输入框

在编辑表单中添加两个 URL 输入框，位于 name 输入框和 category dropdown 之间：
- Blog URL 输入框（`<input type="url" name="url">`）
- Feed URL 输入框（`<input type="url" name="feed_url">`）
- HTML5 pattern 验证（`pattern="^https?://.*"`）
- 每个输入框后带错误提示 span（`url-error-message`）
- 错误消息："URL 必须以 http:// 或 https:// 开头"

**Commit:** 07494bf

### Task 02: 添加 CSS 样式支持 inline 错误提示

在 styles.css 中添加 URL 输入框样式和验证反馈样式：
- `.blog-edit-url`, `.blog-edit-feed-url`: 容器样式
- `.edit-label`: URL 标签样式
- URL 输入框样式（focus 状态带 accent 边框）
- `.url-error-message`: 默认隐藏，invalid 时显示
- `input:invalid`: 红色边框（#dc2626）
- CSS :invalid 伪类触发错误提示显示

**Commit:** 01ceeb7

## Deviations from Plan

### CSS 文件路径偏差

**Found during:** Task 02
**Issue:** 计划中指定 `assets/css/styles.css`，但实际项目使用 `assets/static/styles.css`
**Fix:** 使用正确的路径 `assets/static/styles.css`
**Files modified:** assets/static/styles.css
**Impact:** 无 — CSS 样式正确应用

### 错误消息类名偏差

**Found during:** Task 01/02
**Issue:** 计划建议使用 `.error-message`，但 styles.css 已有 `.error-message` 用于全局表单错误提示（不同样式）
**Fix:** 使用 `.url-error-message` 作为 URL 验证的专用类名，避免样式冲突
**Files modified:** blog-edit-form.gohtml, styles.css
**Impact:** 无 — 验证逻辑正确，样式独立

## Verification

### Acceptance Criteria Verified

**Template Criteria:**

- [x] blog-edit-form.gohtml 包含 `<input type="url" name="url"` 输入框
- [x] blog-edit-form.gohtml 包含 `<input type="url" name="feed_url"` 输入框
- [x] Blog URL 输入框包含 `pattern="^https?://.*"`
- [x] Feed URL 输入框包含 `pattern="^https?://.*"`
- [x] Blog URL 输入框后有 `<span class="url-error-message">` 元素
- [x] Feed URL 输入框后有 `<span class="url-error-message">` 元素
- [x] 错误消息文本为 "URL 必须以 http:// 或 https:// 开头"
- [x] URL 输入框位于 name 和 category 之间（字段顺序正确）

**CSS Criteria:**

- [x] styles.css 包含 `.blog-edit-url` 样式规则
- [x] styles.css 包含 `.blog-edit-feed-url` 样式规则
- [x] styles.css 包含 `.edit-label` 样式规则
- [x] styles.css 包含 `.url-error-message` 样式规则
- [x] `.url-error-message` 默认 `display: none`
- [x] `.url-error-message` 包含 `color: #dc2626`（危险红色）
- [x] CSS 包含 `:invalid` 状态样式（边框和错误提示显示）

## Known Stubs

None.

## Threat Flags

None.

## Self-Check: PASSED

- [x] assets/templates/partials/blog-edit-form.gohtml exists
- [x] assets/static/styles.css exists
- [x] Commit 07494bf exists in git log
- [x] Commit 01ceeb7 exists in git log

---

*Completed: 2026-05-09*
*Duration: ~3 minutes*