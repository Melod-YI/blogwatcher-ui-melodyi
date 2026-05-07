---
phase: 12-cli-foundation
plan: 03
subsystem: CLI
tags:
  - article-command
  - list-filter
  - mark-read
  - output-format
requires:
  - CLI-02
  - CLI-03
  - CLI-04
provides:
  - article list 命令（多筛选参数、多格式输出）
  - article mark-read 命令（单篇/批量）
  - article mark-unread 命令
  - ListArticlesWithFilters storage 方法
  - 输出格式化器（table/json/simple）
affects:
  - internal/storage/database.go（新增方法）
  - internal/cli/output/（新包）
  - internal/cli/commands/article.go（新命令）
  - internal/cli/commands/root.go（添加子命令）
tech-stack:
  added:
    - text/tabwriter（表格输出）
    - encoding/json（JSON 输出）
  patterns:
    - Cobra 命令框架
    - 筛选参数组合（--blog + --unread + --after）
    - 多格式输出（table/json/simple）
key-files:
  created:
    - internal/cli/output/table.go
    - internal/cli/output/json.go
    - internal/cli/output/simple.go
    - internal/cli/commands/article.go
  modified:
    - internal/storage/database.go
    - internal/cli/commands/root.go
    - .gitignore
decisions:
  - 使用 ListFilterOptions 结构体替代 SearchOptions（支持博客名称筛选）
  - 输出格式化器作为独立包（internal/cli/output）
  - 使用 Cobra MarkFlagsMutuallyExclusive 处理 --unread/--read 互斥
  - 博客名称筛选通过子查询实现（无需预先获取 BlogID）
metrics:
  duration: 10分钟
  tasks_completed: 6
  files_created: 4
  files_modified: 3
  commits: 6
completed_date: 2026-05-07
---

# Phase 12 Plan 03: Article 子命令实现 Summary

## 一句话概述

实现 article 子命令：list（多筛选、多格式）、mark-read（单篇/批量）、mark-unread，以及输出格式化器。

## 任务完成情况

### Task 1: 添加 storage 方法支持日期筛选

**文件：**
- `internal/storage/database.go`：新增 ListFilterOptions 结构体和 ListArticlesWithFilters 方法

**实现：**
- 定义 ListFilterOptions：BlogName、IsRead、AfterDate、Limit
- 实现 ListArticlesWithFilters：构建动态 SQL 查询，支持组合筛选
- 博客名称通过子查询 `(SELECT id FROM blogs WHERE name = ?)` 实现

**Commit:** `307deec`

### Task 2: 创建输出格式化器

**文件：**
- `internal/cli/output/table.go`：表格格式输出
- `internal/cli/output/json.go`：JSON 格式输出
- `internal/cli/output/simple.go`：简洁格式输出

**实现：**
- FormatTable：列宽固定，包含 ID、Title、Blog、Status、Published 列
- FormatJSON：简化结构，包含 id、title、url、blog、read、published 字段
- FormatSimple：每行一个文章，格式 "[状态] Title (Blog) - Date"

**Commit:** `5727cdf`

### Task 3-5: 实现 article 子命令框架和执行逻辑

**文件：**
- `internal/cli/commands/article.go`：article 命令组和子命令
- `internal/cli/commands/root.go`：添加 article 命令到 rootCmd

**实现：**
- NewArticleCmd()：article 命令组，包含 list、mark-read、mark-unread
- NewListCmd()：list 命令，支持 --blog、--unread/--read、--after、--format
- NewMarkReadCmd()：mark-read 命令，支持 [id] 参数和 --all flag
- NewMarkUnreadCmd()：mark-unread 命令，必须提供 <id> 参数
- runList()：解析筛选参数，调用 storage 方法，格式化输出
- runMarkRead()：支持单篇和批量标记
- runMarkUnread()：支持单篇标记未读

**Commit:** `b1ece68`

### Task 6: 验证 article 命令功能

**验证结果：**
- `./blogwatcher.exe article --help`：显示 article 命令组帮助 ✓
- `./blogwatcher.exe article list --help`：显示所有筛选参数和格式选项 ✓
- `./blogwatcher.exe article mark-read --help`：显示 --all flag ✓
- `./blogwatcher.exe article mark-unread --help`：显示命令帮助 ✓
- `./blogwatcher.exe article list`：成功列出文章（当前数据库无文章） ✓
- `./blogwatcher.exe article list --format json`：输出 JSON 格式 ✓
- `./blogwatcher.exe article list --format simple`：输出简洁格式 ✓

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking Issue] Worktree 缺少 Wave 1 基础设施**
- **Found during:** Task 3 准备阶段
- **Issue:** Worktree 从旧提交创建，缺少 Wave 1 创建的 CLI 基础设施（flags、commands、serve）
- **Fix:** Cherry-pick Wave 1 提交（1016d74、15f2b00、dfadc0d）到 worktree
- **Files modified:** internal/cli/commands/root.go、serve.go、flags.go、cmd/blogwatcher/main.go
- **Commit:** `7e69457`、`96252d9`、`1e6071c`

**2. [Rule 1 - Bug] 初次提交误提交到主仓库 main 分支**
- **Found during:** Task 1 提交阶段
- **Issue:** 使用 `cd` 进入主仓库路径后提交，导致 Task 1 提交到 main 分支
- **Fix:** 在 worktree 目录重新编辑文件并提交
- **Files modified:** internal/storage/database.go（worktree 版本）
- **Commit:** `307deec`（worktree 分支）

## Auth Gates

无。

## Known Stubs

无。

## Threat Flags

无。计划中的威胁模型已正确处理：
- T-12-07（mark-read id）：strconv.ParseInt 错误处理已实现
- T-12-08（--after date）：time.Parse 错误处理已实现
- T-12-09（article list output）：输出文章标题和博客名，无敏感信息
- T-12-10（mark-read --all）：本地操作，用户主动执行

## Verification Results

### 构建测试
```
go build -o blogwatcher.exe ./cmd/blogwatcher
```
**结果：** 成功

### CLI 功能验证
```
./blogwatcher.exe article --help      → 显示命令组帮助
./blogwatcher.exe article list --help → 显示筛选参数和格式选项
./blogwatcher.exe article mark-read --help → 显示 --all flag
./blogwatcher.exe article mark-unread --help → 显示命令帮助
./blogwatcher.exe article list --format json → 输出 []
./blogwatcher.exe article list --format simple → 输出 "没有找到文章"
```
**结果：** 全部通过

## Self-Check

| Item | Status |
|------|--------|
| internal/storage/database.go 方法存在 | ✓ FOUND |
| internal/cli/output/table.go 存在 | ✓ FOUND |
| internal/cli/output/json.go 存在 | ✓ FOUND |
| internal/cli/output/simple.go 存在 | ✓ FOUND |
| internal/cli/commands/article.go 存在 | ✓ FOUND |
| Commit 307deec 存在 | ✓ FOUND |
| Commit 5727cdf 存在 | ✓ FOUND |
| Commit b1ece68 存在 | ✓ FOUND |
| Commit 90cfb84 存在 | ✓ FOUND |

**Self-Check: PASSED**

## Next Steps

后续计划（12-04 及之后）将基于此框架添加：
- 其他 CLI 功能扩展
- 完整集成测试（需要测试数据库）