# BlogWatcher UI 项目开发指南

## 项目概述

BlogWatcher 是一个博客文章管理工具，提供 Web UI 用于浏览和管理博客文章，同时提供 CLI 命令用于扫描博客和管理文章状态。

## 技术栈

- **后端**: Go 1.25+
- **前端**: React + TypeScript + Tailwind CSS + shadcn/ui (嵌入式)
- **数据库**: SQLite (使用 modernc.org/sqlite 纯 Go 实现)
- **全文搜索**: FTS5

## 开发流程要求

完成开发后，需要执行以下步骤：

1. **编译并安装 CLI 到全局**
   ```bash
   go install ./cmd/blogwatcher
   ```
   这会将 `blogwatcher.exe` 安装到 `C:\Users\Melodyi\go\bin\` 目录，使其可在任意位置直接执行。

2. **重新构建并部署 Docker 服务**
   ```bash
   docker compose build blogwatcher-ui --no-cache && docker compose up -d blogwatcher-ui
   ```
   确保所有代码变更都部署到容器中运行。

3. **运行测试**
   ```bash
   go test ./...
   ```
   确保所有测试通过后再部署。

## 项目结构

```
blogwatcher-ui/
├── cmd/
│   ├── server/main.go       # Web 服务器入口
│   └── blogwatcher/main.go  # CLI 入口
├── internal/
│   ├── cli/commands/        # CLI 子命令
│   ├── hn/client.go         # HN 搜索客户端
│   ├── model/model.go       # 数据模型
│   ├── rss/rss.go           # RSS 解析
│   ├── scanner/scanner.go   # 扫描器核心
│   ├── server/              # Web 服务器
│   ├── service/             # 业务逻辑层
│   └── storage/database.go  # 数据库操作
└── assets/                  # 嵌入静态资源
```

## 常用 CLI 命令

```bash
# 启动 Web 服务
blogwatcher serve

# 扫描博客
blogwatcher blog scan [name]
blogwatcher blog scan --all

# HN 搜索同步
blogwatcher hn sync --all
blogwatcher hn sync --failed

# 文章列表
blogwatcher article list --blog <name>
```

## 特殊处理

### simonwillison.net URL 处理

simonwillison.net 的 atom/everything feed 中的文章链接包含 `/#atom-everything` 后缀，这会导致 HN 搜索功能无法正确匹配。RSS 解析时会自动去除该后缀，无需手动处理。