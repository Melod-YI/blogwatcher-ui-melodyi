---
phase: 17-category-management-ui
plan: 02
subsystem: ui
tags: [htmx, go-templates, category-crud]

# Dependency graph
requires:
  - phase: 17-01
    provides: Category database methods (CreateCategory, ListCategoriesWithBlogCount, UpdateCategoryName, DeleteCategory)
provides:
  - 7 category HTTP handlers (create, read, update, delete)
  - 6 category Go templates (section, item, forms, dialog)
  - HTMX-powered inline create/edit/delete workflow
affects:
  - phase-17 (plan 03: blog category assignment UI)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - HTMX inline editing pattern (click-to-edit, outerHTML swap)
    - HTMX dialog confirmation pattern (showModal, hx-delete)

key-files:
  created:
    - assets/templates/partials/category-section.gohtml
    - assets/templates/partials/category-item.gohtml
    - assets/templates/partials/category-add-form.gohtml
    - assets/templates/partials/category-edit-form.gohtml
    - assets/templates/partials/category-list.gohtml
    - assets/templates/partials/delete-category-dialog.gohtml
  modified:
    - internal/server/handlers.go
    - internal/server/routes.go

key-decisions:
  - "Inline editing pattern matches existing blog name editing (click-to-edit)"
  - "Delete confirmation dialog embedded in category-item to avoid extra requests"
  - "Cancel buttons restore original display via hx-get to category/{id}"

patterns-established:
  - "HTMX click-to-edit: span with hx-get triggers inline form swap"
  - "HTMX dialog: onclick showModal() + hx-delete on confirm button"
  - "List refresh: hx-trigger='categoryListUpdated from:body' pattern"

requirements-completed: [CATG-03, CATG-04, CATG-05, CATG-06]

# Metrics
duration: 8min
completed: 2026-05-09
---

# Plan 17-02: Category Management UI Handlers and Templates

**HTMX-powered category CRUD: create, inline edit, and delete with confirmation dialog**

## Performance

- **Duration:** 8 min
- **Started:** 2026-05-09T01:24:00Z
- **Completed:** 2026-05-09T01:32:24Z
- **Tasks:** 2
- **Files modified:** 8 (2 modified, 6 created)

## Accomplishments

- 7 category HTTP handlers implemented (handleCategoriesNew, handleCategoriesList, handleGetCategory, handleCategoriesCreate, handleCategoryEdit, handleCategoryUpdate, handleCategoryDelete)
- 7 RESTful routes registered for category management
- 6 Go templates created for category UI (section container, display item, add form, edit form, list partial, delete dialog)
- HTMX inline create/edit pattern: click category name to edit, inline Save/Cancel workflow
- Delete confirmation dialog with blog count warning

## Task Commits

Each task was committed atomically:

1. **Task 1: Category handlers and routes** - `a3ab932` (feat)
2. **Task 2: Category templates** - `a3ab932` (feat)

**Plan metadata:** `a3ab932` (feat: complete plan 17-02)

## Files Created/Modified

- `internal/server/handlers.go` - Added 7 category handlers with HTMX response support
- `internal/server/routes.go` - Registered 7 category management routes
- `assets/templates/partials/category-section.gohtml` - Section container with list and add button
- `assets/templates/partials/category-item.gohtml` - Display row with click-to-edit and embedded delete dialog
- `assets/templates/partials/category-add-form.gohtml` - Inline input form for new category
- `assets/templates/partials/category-edit-form.gohtml` - Inline input form for editing category name
- `assets/templates/partials/category-list.gohtml` - Partial for refreshing category list
- `assets/templates/partials/delete-category-dialog.gohtml` - Standalone delete confirmation dialog

## Decisions Made

- Used inline editing pattern matching existing blog name editing (click span to edit)
- Embedded delete dialog in category-item to minimize requests
- Cancel buttons restore display via hx-get to category/{id} endpoint
- Created handleCategoriesList for list refresh after various operations

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

**1. Missing storage import**
- **Issue:** `storage.CategoryWithBlogCount` referenced but storage package not imported
- **Fix:** Added `"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"` to handlers.go imports
- **Verification:** `go build ./internal/server/...` passes

**2. Missing route: GET /categories/list**
- **Issue:** Templates referenced `/categories/list` for list refresh but route not implemented
- **Fix:** Added handleCategoriesList handler and route
- **Verification:** Route registered, handler returns category-list.gohtml

**3. Missing route: GET /categories/{id}**
- **Issue:** Cancel button in edit form needed endpoint to restore display state
- **Fix:** Added handleGetCategory handler and route
- **Verification:** Route registered, handler returns category-item.gohtml

## Next Phase Readiness

- Category CRUD handlers and templates complete
- Ready for Plan 17-03: Blog category assignment (integrating category selection into blog edit form)

---
*Plan: 17-02*
*Completed: 2026-05-09*
