---
phase: 12-cli-foundation
plan: 02
subsystem: CLI
tags:
  - blog-command
  - scanner
  - scan
requires:
  - CLI-01
provides:
  - blog 子命令（scan all + scan specific）
  - 扫描结果输出格式
affects:
  - internal/cli/commands/blog.go（新建）
  - internal/cli/commands/root.go（添加 blog 子命令）
tech-stack:
  added:
    - 无新依赖（复用现有 scanner 和 storage 包）
  patterns:
    - Cobra 子命令嵌套（blog -> scan）
    - 复用 scanner.ScanAllBlogs / scanner.ScanBlogByName
key-files:
  created:
    - internal/cli/commands/blog.go
  modified:
    - internal/cli/commands/root.go
decisions:
  - blog 命令作为命令组，不直接执行，显示帮助
  - scan 命令使用 MaximumNArgs(1) 支持 0 或 1 个参数
  - 输出格式："[博客名] 新文章: N, 总文章: M, 来源: RSS/Scrape"
  - 错误处理：博客不存在返回错误消息并退出码 1
metrics:
  duration: 151秒
  tasks_completed: 3
  files_created: 1
  commits: 1
completed_date: 2026-05-07
---

# Phase 12 Plan 02: blog 子命令 Summary

## 一句话概述

实现 blog 子命令的 scan 功能，支持扫描所有博客或指定博客，复用现有 scanner 包获取新文章。

## 任务完成情况

### Task 1: 创建 blog 子命令框架

**文件：**
- `internal/cli/commands/blog.go`：新建，定义 NewBlogCmd 和 NewScanCmd
- `internal/cli/commands/root.go`：修改，添加 blog 子命令注册

**实现：**
- NewBlogCmd()：blog 命令组，包含 scan 子命令
- NewScanCmd()：scan 命令，支持 [name] 参数
- 使用 cobra.MaximumNArgs(1) 限制参数数量
- 在 root.go init() 中添加 rootCmd.AddCommand(NewBlogCmd())

**Commit:** `4008182`

### Task 2: 实现 scan 命令执行逻辑

**文件：**
- `internal/cli/commands/blog.go`

**实现：**
- runScan()：获取数据库路径，打开数据库，根据参数数量调用 scanner
- scanAllBlogs()：调用 scanner.ScanAllBlogs，遍历结果输出
- scanSingleBlog()：调用 scanner.ScanBlogByName，处理 nil 结果
- outputScanResult()：格式化输出（成功/失败两种格式）

**Commit:** `4008182`（与 Task 1 合并提交）

### Task 3: 验证 scan 命令功能

**验证：**
- `./blogwatcher.exe blog --help`：显示 blog 命令组帮助，包含 scan 子命令
- `./blogwatcher.exe blog scan --help`：显示 scan 命令帮助，说明 [name] 参数用法
- `./blogwatcher.exe blog scan`：扫描所有博客（当前数据库无博客，输出为空）
- `./blogwatcher.exe blog scan "NonExistentBlog"`：输出 "博客 'NonExistentBlog' 不存在"
- `./blogwatcher.exe --db ~/.blogwatcher/blogwatcher.db blog scan`：全局 --db flag 正确传递

**Commit:** 无新提交（仅验证）

## Deviations from Plan

### Auto-fixed Issues

无。计划执行完全按照设计进行。

## Auth Gates

无。

## Known Stubs

无。

## Threat Flags

无。计划中的威胁模型已正确处理：
- T-12-04（scanner.ScanAllBlogs）：scanner 包已处理解析错误
- T-12-05（network timeout）：通过 context 传递，scanner 包使用 context timeout
- T-12-06（scan output）：输出博客名称和文章数，无敏感信息

## Verification Results

### 构建测试
```
go build -o blogwatcher.exe ./cmd/blogwatcher && echo "Build successful"
```
**结果：** 成功

### 命令帮助验证
```
./blogwatcher.exe blog --help         → 显示 blog 命令组（包含 scan）
./blogwatcher.exe blog scan --help    → 显示 scan 命令帮助（包含 [name] 参数说明）
```
**结果：** 全部通过

### 功能验证
```
./blogwatcher.exe blog scan                      → 扫描所有（空数据库输出为空）
./blogwatcher.exe blog scan "NonExistentBlog"    → 输出 "博客不存在"
./blogwatcher.exe --db ~/.blogwatcher/blogwatcher.db blog scan → 全局 flag 正确
```
**结果：** 全部通过

## Self-Check

| Item | Status |
|------|--------|
| internal/cli/commands/blog.go 存在 | ✓ FOUND |
| internal/cli/commands/root.go 修改 | ✓ FOUND |
| Commit 4008182 存在 | ✓ FOUND |

**Self-Check: PASSED**

## Next Steps

后续计划将基于此框架添加：
- article 子命令（列出/标记文章）
- 输出格式支持（table/json/simple）