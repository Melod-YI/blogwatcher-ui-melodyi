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

### rsshub 本地服务主机映射

部分博客的 feed URL 指向本地部署的 RSSHub，使用 Docker 服务名形式的地址，例如 `http://rsshub:1200/...`。该地址在两个运行时下的可达性不同：

- **Web（Docker 容器内）**：`blogwatcher-ui` 与 `rsshub` 同属一个 compose 网络，可直接通过服务名 `rsshub:1200` 访问，无需额外配置。
- **CLI（宿主机）**：宿主没有 `rsshub` 这条 DNS，只能经端口映射 `localhost:19998` 访问 RSSHub。

为让两端共用同一份逻辑地址（DB 中统一存 `rsshub:1200` 形式），`rss.ParseFeed` 在 HTTP 抓取前会按环境变量 `BLOGWATCHER_FEED_HOSTMAP` 重写主机。格式：`srchost[:port]=dsthost[:port]`，多条逗号分隔，key 大小写不敏感。value 含端口则使用指定端口，否则保留原端口。

示例（宿主机使用 CLI 时设置）：

```bash
# bash
export BLOGWATCHER_FEED_HOSTMAP=rsshub:1200=localhost:19998

# PowerShell
$env:BLOGWATCHER_FEED_HOSTMAP="rsshub:1200=localhost:19998"
```

容器侧不设置该变量即保持原样（直接用 `rsshub:1200`）。重写仅作用于 HTTP 抓取地址，不影响 `processor` 注册表按原始域名选择处理器（如 simonwillison.net）。实际发生重写时会输出日志 `[RSS] feed URL 主机已重写: <原> -> <新>`。

注意：走 rsshub 的博客应直接填好 FeedURL，不要依赖 `DiscoverFeedURL` 自动发现——自动发现返回的地址会被写回 DB，重写会导致 `localhost` 形式落库而破坏一致性。

### simonwillison.net 处理

**URL 清洗**：simonwillison.net 的所有 atom feed（everything、notes、links 等）中的文章链接包含 `/#atom-xxx` 后缀（如 `/#atom-everything`、`/#atom-notes`、`/#atom-blogmarks`），这会导致 HN 搜索功能无法正确匹配。RSS 解析时会自动匹配 feed URL 以 `simonwillison.net/atom` 开头的博客并去除该后缀，无需手动处理。

**标题过滤**：自动跳过标题以小写 `sqlite-utils`、`datasette`、`luau-wasm`、`micropython-wasm`、`llm`、`asyncinject`、`inaturalist-clumper`、`asgi-gzip` 开头的版本发布类文章（大小写敏感）。例如 `datasette 1.0a33` 会被跳过，但 `Datasette Apps: Host custom HTML applications inside Datasette` 不会。