---
phase: 12-cli-foundation
plan: 01
subsystem: CLI
tags:
  - infrastructure
  - cobra
  - entry-point
requires:
  - CLI-05
  - CLI-06
  - CLI-07
provides:
  - CLI 统一入口点
  - 全局 flags 解析
  - serve 子命令
affects:
  - go.mod（添加 cobra 依赖）
  - cmd/blogwatcher（新入口点）
tech-stack:
  added:
    - github.com/spf13/cobra@v1.9.1
    - github.com/spf13/pflag@v1.0.6
  patterns:
    - Cobra 命令框架
    - PersistentFlags 全局配置
    - signal.NotifyContext graceful shutdown
key-files:
  created:
    - cmd/blogwatcher/main.go
    - internal/cli/flags/flags.go
    - internal/cli/commands/root.go
    - internal/cli/commands/serve.go
  modified:
    - go.mod
    - go.sum
decisions:
  - 选择 Cobra 框架作为 CLI 基础（标准 Go CLI 框架）
  - 使用 PersistentFlags 实现全局 --db flag
  - 在 init() 函数中添加子命令（cobra 推荐模式）
  - 使用 SetVersionTemplate 自定义版本输出格式
metrics:
  duration: 15分钟
  tasks_completed: 3
  files_created: 4
  commits: 3
completed_date: 2026-05-07
---

# Phase 12 Plan 01: CLI 框架基础设施 Summary

## 一句话概述

使用 Cobra 框架创建 CLI 统一入口点，支持 --db 全局 flag 和 serve 子命令，替代原 cmd/server 入口。

## 任务完成情况

### Task 1: 创建 CLI 包结构和全局 flags

**文件：**
- `internal/cli/flags/flags.go`：定义 dbPath 变量和 SetGlobalFlags 函数
- `internal/cli/commands/root.go`：定义 rootCmd 和 Execute 函数

**实现：**
- 添加 cobra v1.9.1 依赖到 go.mod
- 创建 flags 包，导出 DBPath() 和 SetGlobalFlags()
- 创建 commands 包，定义根命令（Use="blogwatcher", Short="博客文章管理工具"）
- 使用 init() 函数设置版本模板

**Commit:** `1016d74`

### Task 2: 创建 serve 子命令

**文件：**
- `internal/cli/commands/serve.go`

**实现：**
- 定义 NewServeCmd() 返回 serve 子命令
- 实现 runServe() 启动 HTTP 服务器
- 复用 cmd/server/main.go 的服务器启动逻辑
- 支持 graceful shutdown（SIGINT/SIGTERM）
- 端口从 PORT 环境变量或默认 8080

**Commit:** `15f2b00`

### Task 3: 创建统一入口并验证 CLI 安装

**文件：**
- `cmd/blogwatcher/main.go`

**验证：**
- `./blogwatcher.exe --help`：显示可用命令列表（包含 serve）
- `./blogwatcher.exe --version`：显示版本信息（从 go install 读取）
- `./blogwatcher.exe serve --help`：显示 serve 命令帮助

**Commit:** `dfadc0d`

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking Issue] 网络连接问题导致 go mod download 失败**
- **Found during:** Task 1
- **Issue:** proxy.golang.org 连接超时，无法下载 cobra 依赖
- **Fix:** 使用 GOPROXY=https://goproxy.cn,direct 通过国内镜像下载
- **Files modified:** 无（仅构建配置）
- **Commit:** N/A

**2. [Rule 1 - Bug] VersionTemplate 字段不存在**
- **Found during:** Task 1
- **Issue:** cobra.Command 的 VersionTemplate 是私有字段（versionTemplate），无法直接设置
- **Fix:** 使用 SetVersionTemplate() 方法在 init() 函数中设置版本模板
- **Files modified:** internal/cli/commands/root.go
- **Commit:** `15f2b00`

**3. [Rule 3 - Blocking Issue] flags 包目录结构错误**
- **Found during:** Task 1
- **Issue:** flags.go 包名是 flags，但文件在 internal/cli/ 目录下，导致 Go 无法识别包
- **Fix:** 创建 internal/cli/flags/ 目录，将 flags.go 移动到该目录
- **Files modified:** internal/cli/flags/flags.go
- **Commit:** `1016d74`

**4. [Rule 3 - Blocking Issue] 远程同名模块冲突**
- **Found during:** Task 1
- **Issue:** 远程存在 github.com/esttorhe/blogwatcher-ui v1.2.0，Go 工具链尝试从远程下载 internal/cli/flags 包
- **Fix:** 使用 GOPROXY=https://goproxy.cn,direct 和 GOSUMDB=sum.golang.org 绕过远程模块检查
- **Files modified:** 无（仅构建配置）
- **Commit:** N/A

## Auth Gates

无。

## Known Stubs

无。

## Threat Flags

无。计划中的威胁模型已正确处理：
- T-12-01（flags.dbPath）：storage.OpenDatabase 已验证路径有效性
- T-12-02（serve 命令）：本地服务，端口可配置，无敏感信息暴露
- T-12-03（go install）：标准工具链，用户主动安装

## Verification Results

### 构建测试
```
go build ./cmd/blogwatcher && ./blogwatcher --help && ./blogwatcher --version
```
**结果：** 成功

### CLI 功能验证
```
./blogwatcher.exe --help      → 显示可用命令（包含 serve）
./blogwatcher.exe --version   → 显示版本信息（v2.3.1-...）
./blogwatcher.exe serve --help → 显示 serve 命令帮助
```
**结果：** 全部通过

### 数据库路径验证
- 默认路径：~/.blogwatcher/blogwatcher.db（通过 storage.DefaultDBPath()）
- 自定义路径：`./blogwatcher serve --db /custom/path.db`（通过 --db flag）

## Self-Check

| Item | Status |
|------|--------|
| cmd/blogwatcher/main.go 存在 | ✓ FOUND |
| internal/cli/flags/flags.go 存在 | ✓ FOUND |
| internal/cli/commands/root.go 存在 | ✓ FOUND |
| internal/cli/commands/serve.go 存在 | ✓ FOUND |
| Commit 1016d74 存在 | ✓ FOUND |
| Commit 15f2b00 存在 | ✓ FOUND |
| Commit dfadc0d 存在 | ✓ FOUND |

**Self-Check: PASSED**

## Next Steps

后续计划（12-02 及之后）将基于此框架添加：
- blog 子命令（扫描博客）
- article 子命令（列出/标记文章）
- 输出格式支持（table/json/simple）