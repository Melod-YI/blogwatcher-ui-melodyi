---
phase: 20-blog-preview
plan: 01
subsystem: ui
tags: [htmx, form, preview-button, go-templates]

requires:
  - phase: 09-settings-page-foundation
    provides: Add Blog form template structure
provides:
  - Preview button in Add Blog form for feed preview flow
affects: [20-02-preview-handler, 20-03-preview-save]

tech-stack:
  added: []
  patterns: [htmx-multiple-actions, inline-buttons]

key-files:
  created: []
  modified:
    - assets/templates/partials/add-blog-form.gohtml

key-decisions:
  - "Preview button uses type=button with hx-include=closest form to capture form data without form submission"
  - "Add Blog button styled with btn-primary class for visual distinction"

patterns-established:
  - "Multiple action buttons: use hx-post with hx-include=closest form for non-submit buttons"

requirements-completed: [PREV-01]

duration: 2min
completed: 2026-05-11
---
# Phase 20 Plan 01: Preview Button Summary

Add Blog 表单中添加 Preview 按钮，使用 HTMX 发送表单数据到 /blogs/preview 路由，实现 Feed 预览流程的入口。

## Performance

- **Duration:** 2 min
- **Started:** 2026-05-11T02:58:08Z
- **Completed:** 2026-05-11T03:00:10Z
- **Tasks:** 1
- **Files modified:** 1

## Accomplishments
- Preview 按钮添加到 Add Blog 表单，与 Add Blog 按钮并列
- Preview 按钮使用 hx-post="/blogs/preview" 和 hx-include="closest form" 发送表单数据
- Add Blog 按钮添加 btn-primary 类，实现蓝色背景视觉区分

## Task Commits

每个任务原子性提交：

1. **Task 1: 添加 Preview 按钮** - `f41e4ba` (feat)

## Files Created/Modified
- `assets/templates/partials/add-blog-form.gohtml` - 添加 Preview 按钮，修改 Add Blog 按钮样式

## Decisions Made
- Preview 按钮使用 `type="button"` 避免 form submit 行为
- 使用 `hx-include="closest form"` 获取表单数据，符合 HTMX 多按钮操作模式
- Add Blog 按钮添加 `btn-primary` 类，与 Preview 按钮灰色背景区分

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Preview 按钮 UI 已完成，等待 20-02 实现 /blogs/preview handler
- /blogs/preview 路由需要在 routes.go 注册

## Self-Check: PASSED
- [x] File modified exists: assets/templates/partials/add-blog-form.gohtml
- [x] Commit exists: f41e4ba feat(20-01): add Preview button to Add Blog form

---
*Phase: 20-blog-preview*
*Completed: 2026-05-11*