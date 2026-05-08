# BlogWatcher UI

## What This Is

A web-based reader UI for the existing blogwatcher CLI tool with note-taking capabilities. It provides an Omnivore-style dark/light interface to browse, read, manage blog articles, and write Markdown notes. Single-user, self-hosted, accessed via browser on desktop or mobile.

## Core Value

Read and manage blog articles through a clean, responsive web interface without touching the CLI.

## Current Milestone: v1.5 Blog Management Enhancement

**Goal:** 增强博客管理体验，提供更灵活的配置、分类组织和预览能力。

**Target features:**
- Blog 设置页面可查看和修改 Blog URL 和 Feed URL
- 分类系统：blog 分类管理 + UI 分层展示 + CLI 分类过滤
- 添加 Blog 预览：新增前可预览 feed 解析结果
- 文章去重优化：改进去重机制，优雅跳过已存在文章

**Status:** Planning started on 2026-05-08.

## Requirements

### Validated

Shipped in v1.0-v1.3 (see previous milestone details in MILESTONES.md)

Shipped in v1.0:

- ✓ **INFRA-01**: Go HTTP server serving web UI — v1.0
- ✓ **INFRA-02**: Connect to existing blogwatcher SQLite database — v1.0
- ✓ **INFRA-03**: HTMX for dynamic updates without full page reloads — v1.0
- ✓ **UI-01**: Responsive layout with collapsible sidebar on mobile — v1.0
- ✓ **UI-02**: Filter views in sidebar (Inbox/unread, Archived/read) — v1.0
- ✓ **UI-03**: Subscriptions list in sidebar showing tracked blogs — v1.0
- ✓ **UI-04**: Dark/light theme toggle — v1.0
- ✓ **DISP-01**: Article cards show thumbnail or site favicon — v1.0
- ✓ **DISP-02**: Article cards show time ago ("7 hours ago") — v1.0
- ✓ **DISP-03**: Article cards show title and source blog name — v1.0
- ✓ **DISP-04**: Clicking article opens original URL in new tab — v1.0
- ✓ **MGMT-01**: Button to mark individual article as read — v1.0
- ✓ **MGMT-02**: Button to mark article as unread — v1.0
- ✓ **MGMT-03**: "Mark all read" button for bulk action — v1.0
- ✓ **MGMT-04**: Manual sync button to scan blogs for new articles — v1.0

Shipped in v1.1:

- ✓ **POLISH-01**: Entire article card clickable (opens URL in new tab) — v1.1
- ✓ **POLISH-02**: Masonry grid layout as alternative to list view — v1.1
- ✓ **POLISH-03**: View toggle to switch between list and grid layouts — v1.1
- ✓ **POLISH-04**: View preference persists across sessions — v1.1
- ✓ **THUMB-01**: Extract thumbnail URL from RSS media/enclosures during sync — v1.1
- ✓ **THUMB-02**: Extract thumbnail from Open Graph meta tags as fallback — v1.1
- ✓ **THUMB-03**: Fall back to favicon when no thumbnail available — v1.1
- ✓ **THUMB-04**: Display thumbnail in article card (both list and grid views) — v1.1
- ✓ **SRCH-01**: Search articles by title text — v1.1
- ✓ **SRCH-02**: Search input with 300ms debounce (HTMX active search) — v1.1
- ✓ **SRCH-03**: Date filter: Last Week shortcut — v1.1
- ✓ **SRCH-04**: Date filter: Last Month shortcut — v1.1
- ✓ **SRCH-05**: Date filter: Custom date range picker — v1.1
- ✓ **SRCH-06**: Combined filters (blog + status + search + date together) — v1.1
- ✓ **SRCH-07**: Display results count showing how many articles match — v1.1

Shipped in v1.2:

- ✓ **SETT-01**: Settings page displays blog list — v1.2
- ✓ **SETT-02**: Blog list shows article counts — v1.2
- ✓ **SETT-03**: Settings page accessible from sidebar — v1.2
- ✓ **ADD-01**: Add blog by URL with auto feed discovery — v1.2
- ✓ **ADD-02**: Auto-sync new blog after adding — v1.2
- ✓ **ADD-03**: New blog appears in sidebar immediately — v1.2
- ✓ **ADD-04**: FAB for quick blog addition — v1.2
- ✓ **ADD-05**: Add blog form validates URL — v1.2
- ✓ **ADD-06**: Error feedback for invalid URLs — v1.2
- ✓ **EDIT-01**: Edit blog name inline — v1.2
- ✓ **REM-01**: Remove blog with confirmation dialog — v1.2
- ✓ **REM-02**: Confirmation shows article count — v1.2
- ✓ **REM-03**: Cascade delete blog + articles — v1.2

Shipped in v1.3:

- ✓ **CLI-01**: blog scan 命令扫描所有博客 — v1.3
- ✓ **CLI-02**: article list 命令支持多筛选 — v1.3
- ✓ **CLI-03**: article mark-read 命令 — v1.3
- ✓ **CLI-04**: article mark-unread 命令 — v1.3
- ✓ **CLI-05**: 全局 --db flag — v1.3
- ✓ **CLI-06**: go install 安装 — v1.3
- ✓ **CLI-07**: serve 子命令启动 UI — v1.3

Shipped in v1.4:

- ✓ **NOTE-01**: CLI note --article-id --file 命令写入备注 — v1.4
- ✓ **NOTE-02**: CLI note delete --article-id 命令删除备注 — v1.4
- ✓ **NOTE-03**: 缺少必填参数时报错退出 — v1.4
- ✓ **NOTE-04**: CLI article list --not-noted 筛选无备注文章 — v1.4
- ✓ **NOTE-05**: --not-noted 可与 --unread 组合使用 — v1.4
- ✓ **NOTE-06**: 备注文件存储于 ~/.blogwatcher/notes/{id}.md — v1.4
- ✓ **NOTE-07**: articles 表新增 has_note 字段 — v1.4
- ✓ **NOTE-08**: 写入/删除备注时同步更新 has_note — v1.4
- ✓ **NOTE-09**: 有备注的文章卡片显示备注按钮 — v1.4
- ✓ **NOTE-10**: 点击备注按钮新标签页打开 Markdown 渲染页面 — v1.4
- ✓ **NOTE-11**: Markdown 渲染支持 GFM 格式 — v1.4
- ✓ **NOTE-12**: 备注页面显示文章标题和原文链接 — v1.4

### Active

(None — will be defined during requirements phase)

### Out of Scope

- User authentication/multi-user — single user, local access
- Labels/tags — not needed yet
- In-app reader view — just link to originals
- Auto-sync/background refresh — manual only
- Full-text search — would require fetching/storing article content
- Read time estimates — not in current database
- OPML import/export — deferred to future milestone
- Scrape-based blogs — RSS/Atom only for v1.2

## Context

**Current state (v1.4 shipped 2026-05-08):**
- Full CRUD for blog articles via UI and CLI
- Note-taking system: CLI write/delete, UI Markdown display
- Tech stack: Go server + HTMX + SQLite + Cobra CLI + Goldmark
- Database: ~/.blogwatcher/blogwatcher.db + notes/{id}.md files
- Total milestones: 4 (v1.0-v1.4), 62 requirements validated

**Reference codebase:** `.reference/blogwatcher/` contains the Go CLI tool that:
- Tracks blogs via RSS/Atom feeds or HTML scraping
- Stores data in SQLite at `~/.blogwatcher/blogwatcher.db`
- Has `blogs` table (id, name, url, feed_url, scrape_selector, last_scanned)
- Has `articles` table (id, blog_id, title, url, published_date, discovered_date, is_read, has_note)
- Provides scanning, read/unread management, note management via CLI

**Database location:** `~/.blogwatcher/blogwatcher.db` (shared with CLI)

**Notes storage:** `~/.blogwatcher/notes/{article_id}.md` (Markdown files)

**Existing patterns:** The reference code uses modernc.org/sqlite, clean Go patterns, tested storage layer.

**Architecture:** Go server with HTMX, CSS custom properties for theming, Cobra CLI framework, scanner packages (RSS, scraper), Goldmark Markdown renderer.

## Constraints

- **Tech stack:** Go server with templates + HTMX — server-rendered, minimal JS
- **Database:** Must use existing SQLite database and schema (may add thumbnail_url column if needed)
- **Deployment:** Single binary that serves the web UI
- **Compatibility:** Share database with CLI tool — both can coexist

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go templates + HTMX | Go-native, minimal JS, matches reference codebase style | ✓ Good |
| Share CLI database | Single source of truth, CLI and UI coexist | ✓ Good |
| No in-app reader | Simpler, just link to originals, avoids content fetching complexity | ✓ Good |
| Manual sync only | Keeps it simple, user controls when to refresh | ✓ Good |
| Three-way theme toggle | Light/Dark/System with CSS :has() and localStorage | ✓ Good |

| Shell exec for CLI integration | Leverage existing blogwatcher CLI for feed discovery, keep logic centralized | ✓ Good (v1.2) |

| Cobra framework for CLI | Standard Go CLI framework, persistent flags, subcommand nesting | ✓ Good (v1.3) |
| Unified entry point (cmd/blogwatcher) | Single binary for UI and CLI, go install compatible | ✓ Good (v1.3) |
| Output formatters as package | Separate internal/cli/output for table/json/simple formats | ✓ Good (v1.3) |

| 顶层命令 note 结构 | ROADMAP 指定 'note --article-id' 风格，简单直接 | ✓ Good (v1.4) |
| 必须验证文章存在 | 确保备注与实际文章关联，避免无效 ID | ✓ Good (v1.4) |
| 静默覆盖已有备注 | 符合备注可能需要更新的使用场景，简单直接 | ✓ Good (v1.4) |
| 自动创建 notes 目录 | 用户首次写入时自动创建，无需手动准备 | ✓ Good (v1.4) |
| 报错提示删除不存在的备注 | 一致的用户体验，告知操作无效 | ✓ Good (v1.4) |
| Goldmark + GFM 扩展 | 标准 Go Markdown 解析器，支持表格、删除线、任务列表 | ✓ Good (v1.4) |
| 服务器端 Markdown 渲染 | 无需 JavaScript 增强，Go-native 方案 | ✓ Good (v1.4) |
| CSS 变量系统适配主题 | 自动 Light/Dark 主题切换，一致性 | ✓ Good (v1.4) |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-05-08 after v1.5 milestone started*
