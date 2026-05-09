---
phase: 19-blog-settings-enhancement
plan: 03
subsystem: settings
tags: [backend, handler, validation, url, database]
requires: [19-01, 19-02]
provides: [SETT-05]
affects: [handlers.go, database.go]
tech_stack:
  added: []
  patterns: [URL validation, nullable field handling, sql.NullInt64]
key_files:
  created: []
  modified:
    - internal/server/handlers.go
    - internal/storage/database.go
decisions: [D-03a, D-04, SETT-05]
metrics:
  duration: 5min
  tasks: 2
  files: 2
  completed_date: "2026-05-09"
---

# Phase 19 Plan 03: handleUpdateBlogName URL Parameter Processing Summary

## One-liner

扩展 handleUpdateBlogName handler 解析 URL 参数并验证格式，使用 UpdateBlog 一次性更新所有字段（实现 SETT-05）。

## Changes

### Task 01: 扩展 handleUpdateBlogName 解析 URL 参数并验证

在 handleUpdateBlogName handler 中添加 URL 参数解析和验证逻辑：
- 解析 `url` 和 `feed_url` 表单参数
- 添加 `validateURL` 函数验证 HTTP/HTTPS 前缀
- 空值允许（per D-03a，nullable 字段）
- 无效 URL 返回 400 Bad Request

### Task 02: 扩展 handleUpdateBlogName 使用 UpdateBlog 更新所有字段

替换 UpdateBlogName + UpdateBlogCategory 调用为单一 UpdateBlog 方法：
- 获取当前 Blog 数据保留其他字段
- 更新 name, url, feed_url, category_id 字段
- 使用 UpdateBlog 一次性提交所有更改
- 日志记录所有更新字段

**Commit:** cc13a56

## Deviations from Plan

### Rule 2 - Auto-fix missing critical functionality

**Found during:** Task 02
**Issue:** UpdateBlog 方法未包含 category_id 字段，计划期望一次性更新所有字段
**Fix:** 在 database.go 的 UpdateBlog SQL 查询中添加 category_id 列，使用 sql.NullInt64 处理 nullable 值
**Files modified:** internal/storage/database.go
**Commit:** cc13a56

## Verification

### Acceptance Criteria Verified

**Task 01 Criteria:**

- [x] handlers.go 包含 `url := strings.TrimSpace(r.FormValue("url"))`
- [x] handlers.go 包含 `feedURL := strings.TrimSpace(r.FormValue("feed_url"))`
- [x] handlers.go 包含 validateURL 函数定义
- [x] validateURL 函数包含空值处理（`if url == "" { return nil }`）
- [x] validateURL 函数验证 HTTP/HTTPS 前缀
- [x] handlers.go 调用 validateURL(url) 和 validateURL(feedURL)
- [x] URL 验证失败时返回 400 错误

**Task 02 Criteria:**

- [x] handlers.go 包含 `blog.Name = name`
- [x] handlers.go 包含 `blog.URL = url`
- [x] handlers.go 包含 `blog.FeedURL = feedURL`
- [x] handlers.go 包含 `blog.CategoryID = categoryID`
- [x] handlers.go 调用 `s.db.UpdateBlog(*blog)`
- [x] handlers.go 包含日志记录所有更新字段
- [x] 不再调用 UpdateBlogName 方法（仅保留 handler 函数名）
- [x] 不再调用 UpdateBlogCategory 方法

**Database Fix:**

- [x] UpdateBlog SQL 包含 category_id 列
- [x] UpdateBlog 使用 sql.NullInt64 处理 nullable category_id
- [x] 存储层测试全部通过

## Known Stubs

None.

## Threat Flags

None.

## Self-Check: PASSED

- [x] internal/server/handlers.go exists in worktree
- [x] internal/storage/database.go exists in worktree
- [x] Commit cc13a56 exists in git log
- [x] Go build passes
- [x] Go tests pass

---

*Completed: 2026-05-09*
*Duration: ~5 minutes*