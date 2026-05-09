---
phase: 18
plan: 01
subsystem: CLI
tags: [category, filtering, CLI, CATG-10]
dependency_graph:
  requires: [Phase 17 - Category Management UI]
  provides: [CLI --category filtering]
  affects: [article list command, ListFilterOptions struct]
tech_stack:
  added: [GetCategoryByName method, scanCategory function, CategoryName field]
  patterns: [CLI flag pattern, GetByName pattern, SQL subquery filtering]
key_files:
  created: []
  modified:
    - internal/storage/database.go
    - internal/cli/commands/article.go
decisions:
  - D-06: 按分类名称筛选（如 --category tech）
  - D-07: 统一组合过滤（加入 ListFilterOptions.CategoryName）
  - D-08: 验证分类存在性（查询分类表，不存在则报错退出）
metrics:
  duration: ~5 minutes
  tasks_completed: 3
  files_modified: 2
  commits: 2
  completed_date: "2026-05-09"
---

# Phase 18 Plan 01: CLI --category Filtering Summary

**One-liner:** CLI `article list --category` filtering implemented with validation and combined filter support (CATG-10).

## Tasks Completed

| Task | Name | Commit | Files Modified |
|------|------|--------|----------------|
| 1 | 扩展 ListFilterOptions 结构添加 CategoryName 字段 | 94e1587 | internal/storage/database.go |
| 2 | 新增 GetCategoryByName 数据库方法 | 94e1587 | internal/storage/database.go |
| 3 | CLI article.go 新增 --category flag 和验证逻辑 | db02644 | internal/cli/commands/article.go |

## Changes Summary

### Task 1: ListFilterOptions Extension

Added `CategoryName string` field to `ListFilterOptions` struct:
- Enables combined filtering with existing options (--blog, --unread, --not-noted, --after)
- Added category filtering condition to `ListArticlesWithFilters`:
  ```go
  if opts.CategoryName != "" {
      conditions = append(conditions, "b.category_id = (SELECT id FROM categories WHERE name = ?)")
      args = append(args, opts.CategoryName)
  }
  ```

### Task 2: GetCategoryByName Method

Added database methods for category lookup:
- `GetCategoryByName(name string) (*model.Category, error)` - returns category or nil if not found
- `scanCategory(scanner) (*model.Category, error)` - scans category rows from database

Pattern matches existing `GetBlogByName` implementation.

### Task 3: CLI --category Flag

Added to `internal/cli/commands/article.go`:
- Flag definition: `cmd.Flags().String("category", "", "分类名称筛选")`
- Help text updated with category parameter and example
- Parameter parsing: `categoryName, _ := cmd.Flags().GetString("category")`
- Options construction: `opts.CategoryName = categoryName`
- Validation logic: calls `GetCategoryByName`, errors if category doesn't exist

## Verification Results

All verification tests passed:

| Test | Result |
|------|--------|
| `./blogwatcher article list --help \| grep "category"` | PASS |
| `./blogwatcher article list --category nonexistent` | Error: "分类 'nonexistent' 不存在" |
| Go build | PASS |

## Deviations from Plan

None - plan executed exactly as written.

## Key Decisions

- **D-06**: Category filtering by name (user-friendly, matches --blog pattern)
- **D-07**: Combined filtering through ListFilterOptions extension
- **D-08**: Existence validation before filtering (consistent error messages)

## Files Modified

1. **internal/storage/database.go**
   - ListFilterOptions struct: added CategoryName field
   - ListArticlesWithFilters: added category filtering condition
   - GetCategoryByName method: new category lookup
   - scanCategory function: new row scanner

2. **internal/cli/commands/article.go**
   - NewListCmd: added --category flag
   - Long description: updated with category examples
   - runList: added categoryName parsing, opts construction, validation

## Requirements Satisfied

- **CATG-10**: CLI `article list --category <name>` filtering implemented

## Next Steps

- Plan 18-02: UI 分类分组展示
- Plan 18-03: 样式和 localStorage 持久化

---

*Plan executed: 2026-05-09*
*Commits: 94e1587, db02644*