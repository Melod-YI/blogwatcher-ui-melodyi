# Roadmap: BlogWatcher UI

**Milestone:** v1.5 Blog Management Enhancement
**Created:** 2026-02-02
**Depth:** standard
**Total Phases:** 20

## Milestones

- **v1.0 MVP** - Phases 1-5 (shipped 2026-02-03)
- **v1.1 UI Polish & Search** - Phases 6-8 (shipped 2026-02-03)
- **v1.2 Blog Management** - Phases 9-11 (shipped 2026-02-09)
- **v1.3 CLI System** - Phase 12 (shipped 2026-05-07)
- **v1.4 Article Notes** - Phases 13-15 (shipped 2026-05-08)
- **v1.5 Blog Management Enhancement** - Phases 16-20 (planning)

## Phases

<details>
<summary>v1.0 MVP (Phases 1-5) - SHIPPED 2026-02-03</summary>

### Phase 1: Infrastructure Setup

**Goal:** Foundation server and database connection are functional and ready to serve the UI.

**Requirements:**
- INFRA-01: Go HTTP server serving web UI
- INFRA-02: Connect to existing blogwatcher SQLite database
- INFRA-03: HTMX for dynamic updates without full page reloads

**Success Criteria:**
1. User can navigate to localhost URL and see a basic page served by Go server
2. Server successfully reads articles and blogs from existing SQLite database at ~/.blogwatcher/blogwatcher.db
3. HTMX requests can fetch data from server endpoints and update page sections without full reload

**Depends on:** None

**Plans:** 3 plans

Plans:
- [x] 01-01-PLAN.md — Project setup, database layer, HTMX static file
- [x] 01-02-PLAN.md — HTTP server with NewServer pattern and templates
- [x] 01-03-PLAN.md — Wire handlers to database, full integration verification

---

### Phase 2: UI Layout & Navigation

**Goal:** User can navigate between different views of their articles using a responsive sidebar.

**Requirements:**
- UI-01: Responsive layout with collapsible sidebar on mobile
- UI-02: Filter views in sidebar (Inbox/unread, Archived/read)
- UI-03: Subscriptions list in sidebar showing tracked blogs

**Success Criteria:**
1. User sees a sidebar with "Inbox" and "Archived" filter options
2. User sees list of subscribed blogs in sidebar matching their blogwatcher database
3. User can collapse/expand sidebar on mobile screen sizes
4. Clicking a filter or blog in sidebar changes the main content area (even if just showing placeholder)

**Depends on:** Phase 1 (Infrastructure Setup)

**Plans:** 2 plans

Plans:
- [x] 02-01-PLAN.md — CSS foundation with dark theme, responsive grid layout, collapsible sidebar structure
- [x] 02-02-PLAN.md — HTMX navigation wiring, filter query params, active state highlighting

---

### Phase 3: Article Display

**Goal:** User can see their articles with rich metadata and open them to read.

**Requirements:**
- DISP-01: Article cards show thumbnail or site favicon
- DISP-02: Article cards show time ago ("7 hours ago")
- DISP-03: Article cards show title and source blog name
- DISP-04: Clicking article opens original URL in new tab

**Success Criteria:**
1. User sees article cards displaying title, source blog name, and relative time ("2 hours ago")
2. Each article card shows either a thumbnail image or favicon for the source blog
3. User can click an article card and original blog post opens in new browser tab
4. Articles from database appear in correct filtered view (unread in Inbox, read in Archived)
5. Clicking a blog in sidebar filters articles to only show that blog's content

**Depends on:** Phase 2 (UI Layout & Navigation)

**Plans:** 2 plans

Plans:
- [x] 03-01-PLAN.md — Template functions (timeAgo, faviconURL), ArticleWithBlog model, database JOIN query
- [x] 03-02-PLAN.md — Article card template and CSS styling with rich metadata display

---

### Phase 4: Article Management

**Goal:** User can mark articles as read/unread and trigger blog syncing from the UI.

**Requirements:**
- MGMT-01: Button to mark individual article as read
- MGMT-02: Button to mark article as unread
- MGMT-03: "Mark all read" button for bulk action
- MGMT-04: Manual sync button to scan blogs for new articles

**Success Criteria:**
1. User can mark individual article as read and see it move from Inbox to Archived
2. User can mark individual article as unread and see it move from Archived to Inbox
3. User can click "Mark all read" and see all visible articles move to Archived
4. User can click sync button and see new articles appear from blogs after scan completes
5. Read/unread state persists to database and matches CLI tool's view

**Depends on:** Phase 3 (Article Display)

**Plans:** 2 plans

Plans:
- [x] 04-01-PLAN.md — Scanner package setup (copy RSS, scraper, scanner from reference + database extensions)
- [x] 04-02-PLAN.md — Article management handlers, routes, templates with action buttons and toolbar

---

### Phase 5: Theme Toggle

**Goal:** User can switch between dark and light themes with preference persisted.

**Requirements:**
- UI-04: Dark/light theme toggle

**Success Criteria:**
1. User sees theme toggle control in UI
2. User can click toggle and interface switches between dark and light themes
3. Theme preference persists across browser sessions

**Depends on:** Phase 2 (UI Layout & Navigation)

**Plans:** 1 plan

Plans:
- [x] 05-01-PLAN.md — CSS light theme variables, toggle component, FOUC prevention, localStorage persistence

</details>

---

<details>
<summary>v1.1 UI Polish & Search (Phases 6-8) - SHIPPED 2026-02-03</summary>

**Milestone Goal:** Improve visual presentation with masonry layout and thumbnails, add search and filtering capabilities.

### Phase 6: Enhanced Card Interaction

**Goal:** User can click entire article card to open article, and cards display rich thumbnails with fallback chain.

**Requirements:**
- POLISH-01: Entire article card is clickable (opens URL in new tab)
- THUMB-01: Extract thumbnail URL from RSS media/enclosures during sync
- THUMB-02: Extract thumbnail from Open Graph meta tags as fallback
- THUMB-03: Fall back to favicon when no thumbnail available
- THUMB-04: Display thumbnail in article card (both list and grid views)

**Success Criteria:**
1. User can click anywhere on article card and original article opens in new tab
2. Article cards display thumbnails extracted from RSS media/enclosures when available
3. When RSS has no thumbnail, article cards display Open Graph image from article URL
4. When neither RSS nor Open Graph provide thumbnail, article cards display blog favicon
5. Thumbnail images render with proper aspect ratio and no cumulative layout shift

**Depends on:** Phase 5 (Theme Toggle)

**Plans:** 2 plans

Plans:
- [x] 06-01-PLAN.md — Schema migration, models, thumbnail extraction package
- [x] 06-02-PLAN.md — Sync pipeline integration, template updates, CSS for clickable cards

---

### Phase 7: Search & Date Filtering

**Goal:** User can find articles by title search and filter by date ranges.

**Requirements:**
- SRCH-01: Search articles by title text
- SRCH-02: Search input with 300ms debounce (HTMX active search)
- SRCH-03: Date filter: Last Week shortcut
- SRCH-04: Date filter: Last Month shortcut
- SRCH-05: Date filter: Custom date range picker
- SRCH-06: Combined filters (blog + status + search + date together)
- SRCH-07: Display results count showing how many articles match

**Success Criteria:**
1. User can type in search box and see results filter to articles matching title text
2. Search input debounces at 300ms and does not trigger on every keystroke
3. User can click "Last Week" filter and see only articles from past 7 days
4. User can click "Last Month" filter and see articles from past 30 days
5. User can select custom date range and see articles within that range
6. User can combine multiple filters (blog + status + search + date) and see articles matching all conditions
7. Results count displays "Showing X articles" or "No articles found" based on active filters

**Depends on:** Phase 5 (Theme Toggle)

**Plans:** 2 plans

Plans:
- [x] 07-01-PLAN.md — FTS5 infrastructure and SearchArticles method
- [x] 07-02-PLAN.md — Search UI, date filters, results count display

---

### Phase 8: Masonry Layout

**Goal:** User can toggle between list and masonry grid layouts with preference persisted.

**Requirements:**
- POLISH-02: Masonry grid layout as alternative to list view
- POLISH-03: View toggle to switch between list and grid layouts
- POLISH-04: View preference persists across sessions

**Success Criteria:**
1. User sees view toggle button to switch between list and grid layouts
2. User can click grid view and see articles arranged in masonry layout with varied card heights
3. Masonry layout responds to viewport width (1 col mobile, 2 col tablet, 3-4 col desktop)
4. User can switch back to list view and see traditional vertical layout
5. View preference persists across browser sessions (remembered on next visit)

**Depends on:** Phase 6 (Enhanced Card Interaction)

**Plans:** 1 plan

Plans:
- [x] 08-01-PLAN.md — CSS Grid layout, view toggle component, localStorage persistence

</details>

---

## v1.2 Blog Management (SHIPPED 2026-02-09)

**Milestone Goal:** Manage blogs entirely from the web UI — add, edit, remove — without touching the CLI.
**Started:** 2026-02-08
**Completed:** 2026-02-09

### Phase 9: Settings Page Foundation

**Goal:** User can access settings page and view all tracked blogs with their metadata.

**Requirements:**
- SETT-01: User can access settings page from sidebar gear icon
- SETT-02: Settings page displays list of all tracked blogs
- SETT-03: Each blog entry shows name, URL, and article count

**Success Criteria:**
1. User sees gear icon in sidebar navigation
2. User can click gear icon and settings page loads via HTMX
3. Settings page displays complete list of all blogs from database
4. Each blog entry shows display name, homepage URL, and count of articles
5. Blog list updates when blogs are added/removed (via HTMX swap)

**Depends on:** Phase 8 (Masonry Layout)

**Plans:** 1 plan

Plans:
- [x] 09-01-PLAN.md — Database layer (ListBlogsWithCounts), settings handler/route/templates, gear icon in sidebar

---

### Phase 10: Add Blog Flow

**Goal:** User can add new blogs by URL with auto-discovery and automatic first sync.

**Requirements:**
- ADD-01: User can enter blog URL in add form
- ADD-02: System auto-discovers RSS/Atom feed via blogwatcher CLI
- ADD-03: User sees success/error feedback after add attempt
- ADD-04: System displays discovered feed URL to user
- ADD-05: System auto-syncs newly added blog to fetch articles
- ADD-06: User can access quick add via floating action button (FAB)

**Success Criteria:**
1. User sees "Add Blog" form on settings page with URL input field
2. User can enter a blog URL and submit the form
3. System executes `blogwatcher add <url>` command via os/exec to discover RSS/Atom feed
4. User sees success message showing discovered feed URL after successful add
5. User sees error message when feed discovery fails (invalid URL, no feed found)
6. System automatically triggers sync for newly added blog (fetches articles with thumbnails)
7. User sees floating action button (FAB) on main article list page
8. User can click FAB to open quick add dialog without navigating to settings

**Depends on:** Phase 9 (Settings Page Foundation)

**Plans:** 2 plans

Plans:
- [x] 10-01-PLAN.md — Add blog handler with CLI integration, auto-sync, form template, and FAB component
- [x] 10-02-PLAN.md — Verify complete add blog flow (form, CLI, error handling, auto-sync, FAB)

---

### Phase 11: Edit and Remove Blogs

**Goal:** User can modify blog names and remove blogs with confirmation dialog.

**Requirements:**
- EDIT-01: User can edit blog display name
- REM-01: User sees confirmation dialog before deletion
- REM-02: User can choose to delete blog only or blog + articles (simplified to single cascade delete)
- REM-03: Confirmation dialog shows article count that would be deleted

**Success Criteria:**
1. User sees "Edit" action next to each blog in settings list
2. User can click edit and modify blog display name inline or in modal
3. Changed blog name persists to database and updates in sidebar subscriptions list
4. User sees "Remove" action next to each blog in settings list
5. User clicking remove triggers confirmation dialog (does not delete immediately)
6. Confirmation dialog displays count of articles that belong to the blog
7. User can choose "Delete blog only" (keeps articles orphaned) or "Delete blog + articles"
8. Selected deletion action executes and blog list updates via HTMX

**Depends on:** Phase 9 (Settings Page Foundation)

**Plans:** 3 plans

Plans:
- [x] 11-01-PLAN.md — Inline blog name editing with HTMX click-to-edit pattern
- [x] 11-02-PLAN.md — Schema migration to make articles.blog_id nullable
- [x] 11-03-PLAN.md — Delete blog with confirmation dialog (simplified to cascade delete)

---

<details>
<summary>v1.3 CLI System (Phase 12) - SHIPPED 2026-05-07</summary>

**Milestone Goal:** 提供独立的 CLI 工具，可通过 `go install` 安装，完成扫描、列出文章、标记已读/未读等核心操作，共享 UI 项目数据库。

### Phase 12: CLI Foundation

**Goal:** 用户可以通过子命令风格的 CLI 完成博客扫描、文章列表查看、标记操作。

**Requirements:**
- CLI-01: 扫描博客命令
- CLI-02: 列出文章命令（支持筛选参数和多格式输出）
- CLI-03: 标记已读命令（单篇和批量）
- CLI-04: 标记未读命令
- CLI-05: 共享数据库路径
- CLI-06: 可通过 go install 安装

**Success Criteria:**
1. 用户可以通过 `go install` 安装 CLI
2. `blogwatcher serve` 启动 UI 服务器（替代原 server 入口）
3. `blogwatcher blog scan` 成功扫描所有博客并获取新文章
4. `blogwatcher blog scan <name>` 成功扫描指定博客
5. `blogwatcher article list` 显示所有文章（默认表格格式）
6. `blogwatcher article list --blog <name>` 按博客筛选
7. `blogwatcher article list --unread/--read` 按状态筛选
8. `blogwatcher article list --after <date>` 按日期筛选
9. `blogwatcher article list --format json` 输出 JSON 格式
10. `blogwatcher article mark-read <id>` 标记单篇已读
11. `blogwatcher article mark-read --all` 标记全部已读
12. `blogwatcher article mark-unread <id>` 标记单篇未读
13. CLI 与 UI 共享同一数据库（默认 ~/.blogwatcher/blogwatcher.db）

**Depends on:** Phase 4 (Article Management - scanner package)

**Plans:** 3/3 plans complete

Plans:
- [x] 12-01-PLAN.md — CLI framework setup (cobra root, global flags, serve command)
- [x] 12-02-PLAN.md — Blog commands (scan all, scan specific)
- [x] 12-03-PLAN.md — Article commands (list, mark-read, mark-unread)

</details>

---

<details>
<summary>v1.4 Article Notes (Phases 13-15) - SHIPPED 2026-05-08</summary>

**Milestone Goal:** 允许用户为文章编写备注，备注为 Markdown 文档，可在 UI 查看。

### Phase 13: CLI Notes Infrastructure

**Goal:** 用户可以通过 CLI 写入和删除文章备注，备注以 Markdown 文件存储。

**Requirements:**
- NOTE-01: CLI `note --article-id <id> --file <path>` 写入备注
- NOTE-02: CLI `note delete --article-id <id>` 删除备注
- NOTE-03: 缺少必填参数时报错退出
- NOTE-06: 备注存储于 ~/.blogwatcher/notes/{article_id}.md
- NOTE-07: articles 表新增 has_note 字段
- NOTE-08: 写入/删除备注时同步更新 has_note 字段

**Success Criteria:**
1. 用户执行 `blogwatcher note --article-id 42 --file ~/note.md` 成功写入备注
2. 源文件内容被完整复制（不依赖源文件后续变化）
3. 用户执行 `blogwatcher note delete --article-id 42` 成功删除备注
4. 缺少 --article-id 或 --file 时输出错误消息并退出
5. 备注 文件存储于 ~/.blogwatcher/notes/ 目录
6. articles 表新增 has_note BOOLEAN 字段
7. 写入备注后 article.has_note = TRUE
8. 删除备注后 article.has_note = FALSE

**Depends on:** Phase 12 (CLI Foundation)

**Plans:** 4 plans

Plans:
- [x] 13-02A-PLAN.md — Model updates (Article.HasNote, ArticleWithBlog.HasNote 字段)
- [x] 13-02B-PLAN.md — scan functions + SELECT queries
- [x] 13-01-PLAN.md — Database layer (has_note migration, GetArticleByID, UpdateArticleHasNote)
- [x] 13-03-PLAN.md — CLI note command (write, delete, command registration)

---

### Phase 14: CLI Filtering Enhancement

**Goal:** 用户可以通过 --not-noted 参数筛选无备注文章，可与 --unread 组合使用。

**Requirements:**
- NOTE-04: CLI `article list --not-noted` 筛选无备注文章
- NOTE-05: --not-noted 可与 --unread 组合使用

**Success Criteria:**
1. 用户执行 `blogwatcher article list --not-noted` 仅显示无备注文章
2. 用户执行 `blogwatcher article list --not-noted --unread` 仅显示未读且无备注文章
3. --not-noted 与 --blog、--after 等参数可组合使用
4. 输出格式 与 --not-noted 兼容

**Depends on:** Phase 13 (CLI Notes Infrastructure)

**Plans:** 1 plan

Plans:
- [x] 14-01-PLAN.md — Add --not-noted filter parameter

---

### Phase 15: UI Note Display

**Goal:** 用户可以在 UI 上查看文章备注，Markdown 渲染显示。

**Requirements:**
- NOTE-09: 有备注的文章卡片显示备注按钮
- NOTE-10: 点击备注按钮新标签页打开 Markdown 渲染页面
- NOTE-11: Markdown 渲染支持 GFM 格式
- NOTE-12: 备注 页面显示文章标题和原文链接

**Success Criteria:**
1. 文章卡片查询包含 has_note 字段
2. has_note = TRUE 的文章卡片显示备注按钮（图标或标签）
3. 点击备注按钮在新标签页打开 /note/{id} 页面
4. 页面渲染 Markdown 内容（GFM 格式：表格、删除线、任务列表）
5. 页面顶部显示文章标题和原文链接
6. Markdown 样式与应用整体风格一致

**Depends on:** Phase 14 (CLI Filtering Enhancement)

**Plans:** 1 plan

Plans:
- [x] 15-01-PLAN.md — 文章卡片备注按钮 + 备注 页面 + Markdown 渲染（goldmark + GFM）

</details>

---

## v1.5 Blog Management Enhancement (PLANNING)

**Milestone Goal:** 增强博客管理体验，提供更灵活的配置、分类组织和预览能力。
**Started:** 2026-05-08

### Phase 16: Database Schema

**Goal:** Schema 扩展支持分类和改进去重机制

**Requirements:**
- CATG-01: 数据库创建 categories 表（id, name, created_at）
- CATG-02: 数据库添加 blog.category_id 字段（nullable, foreign key）
- DEDUP-01: AddArticlesBulk 使用 INSERT OR IGNORE
- DEDUP-02: 扫描时遇到重复 URL 静默跳过
- DEDUP-03: 扫描结果统计跳过的文章数量

**Success Criteria:**
1. categories 表创建成功，包含 id, name, created_at 字段
2. blogs 表添加 category_id 字段（nullable foreign key）
3. AddArticlesBulk 遇到重复 URL 不报错，静默跳过
4. ScanBlog 返回统计包含 skipped_count
5. 数据库迁移向后兼容，现有数据不受影响

**Depends on:** Phase 15 (UI Note Display)

**Plans:** 2/2 plans complete

Plans:
- [x] 16-01-PLAN.md — 创建 categories 表和 blogs.category_id 字段
- [x] 16-02-PLAN.md — AddArticlesBulk INSERT OR IGNORE + ScanResult skipped_count
---

### Phase 17: Category Management UI

**Goal:** 分类管理界面和 blog 分类选择

**Requirements:**
- CATG-03: 设置页面添加分类管理区
- CATG-04: 用户可创建新分类（输入名称）
- CATG-05: 用户可编辑分类名称（inline 编辑）
- CATG-06: 用户可删除分类（删除时 blog.category_id 置空）
- CATG-07: Blog 编辑时可选择分类（下拉选择）

**Success Criteria:**
1. 设置页面显示分类管理区（独立区域）
2. 点击"新建分类"按钮，输入名称后立即保存
3. 分类名称可 inline 编辑（点击名称直接编辑）
4. 删除分类时，关联 blog 的 category_id 自动置空
5. Blog 编辑行显示分类下拉选择框

**Depends on:** Phase 16 (Database Schema)

**Plans:** 0 plans

---

### Phase 18: Category Display & CLI

**Goal:** Subscriptions 分类分层展示和 CLI 分类过滤

**Requirements:**
- CATG-08: Subscriptions 按分类分层展示
- CATG-09: 未分类 blog 在 Subscriptions 顶层显示
- CATG-10: CLI article list --category 过滤

**Success Criteria:**
1. Subscriptions 区域按分类分组显示 blog
2. 未分类 blog 显示在顶层（无分类标签）
3. 点击分类名称可展开/折叠该分类下的 blog
4. CLI `article list --category <name>` 返回该分类下的所有文章
5. CLI `article list --category tech --unread` 组合过滤生效

**Depends on:** Phase 17 (Category Management UI)

**Plans:** 0 plans

---

### Phase 19: Blog Settings Enhancement

**Goal:** 设置页面可查看和编辑 Blog URL 和 Feed URL

**Requirements:**
- SETT-01: 设置页面显示 Blog URL 和 Feed URL
- SETT-02: 设置页面可编辑 Blog URL（inline 编辑）
- SETT-03: 设置页面可编辑 Feed URL（inline 编辑）
- SETT-04: 编辑时验证 URL 格式（HTTP/HTTPS）
- SETT-05: 保存后立即更新数据库

**Success Criteria:**
1. 设置页面 blog 行显示 Blog URL 和 Feed URL 列
2. Blog URL 可 inline 编辑（点击 URL 直接编辑）
3. Feed URL 可 inline 编辑（点击 URL 直接编辑）
4. 输入非 HTTP/HTTPS URL 时显示错误提示
5. 编辑保存后，数据库立即更新，无需刷新页面

**Depends on:** Phase 18 (Category Display & CLI)

**Plans:** 0 plans

---

### Phase 20: Blog Preview

**Goal:** 新增 blog 前预览 feed 解析结果

**Requirements:**
- PREV-01: 添加 blog 表单有预览按钮
- PREV-02: 点击预览触发临时 feed 解析
- PREV-03: 预览页面显示解析的文章列表（最多 20 条）
- PREV-04: 预览失败显示错误信息
- PREV-05: 预览页面有保存按钮（保存为正式 blog）
- PREV-06: 预览页面有返回修改按钮（返回添加表单）

**Success Criteria:**
1. 添加 blog 表单显示"预览"按钮（与"保存"并列）
2. 点击预览后，页面跳转到临时预览页面
3. 预览页面显示最多 20 篇解析的文章（标题、时间、链接）
4. Feed URL 无效时，预览页面显示错误信息（如"无法解析 feed"）
5. 预览页面显示"保存为 Blog"按钮，点击后保存并跳转到设置页面
6. 预览页面显示"返回修改"按钮，点击后返回添加表单保留输入

**Depends on:** Phase 19 (Blog Settings Enhancement)

**Plans:** 0 plans

---

## Coverage Validation

### v1.0 Requirements

| Requirement | Phase | Covered |
|-------------|-------|---------|
| INFRA-01 | 1 | Done |
| INFRA-02 | 1 | Done |
| INFRA-03 | 1 | Done |
| UI-01 | 2 | Done |
| UI-02 | 2 | Done |
| UI-03 | 2 | Done |
| DISP-01 | 3 | Done |
| DISP-02 | 3 | Done |
| DISP-03 | 3 | Done |
| DISP-04 | 3 | Done |
| MGMT-01 | 4 | Done |
| MGMT-02 | 4 | Done |
| MGMT-03 | 4 | Done |
| MGMT-04 | 4 | Done |
| UI-04 | 5 | Done |

**v1.0 Coverage:** 15/15 requirements mapped (100%)

### v1.1 Requirements

| Requirement | Phase | Covered |
|-------------|-------|---------|
| POLISH-01 | 6 | Done |
| POLISH-02 | 8 | Done |
| POLISH-03 | 8 | Done |
| POLISH-04 | 8 | Done |
| THUMB-01 | 6 | Done |
| THUMB-02 | 6 | Done |
| THUMB-03 | 6 | Done |
| THUMB-04 | 6 | Done |
| SRCH-01 | 7 | Done |
| SRCH-02 | 7 | Done |
| SRCH-03 | 7 | Done |
| SRCH-04 | 7 | Done |
| SRCH-05 | 7 | Done |
| SRCH-06 | 7 | Done |
| SRCH-07 | 7 | Done |

**v1.1 Coverage:** 15/15 requirements mapped (100%)

### v1.2 Requirements

| Requirement | Phase | Covered |
|-------------|-------|---------|
| SETT-01 | 9 | Yes |
| SETT-02 | 9 | Yes |
| SETT-03 | 9 | Yes |
| ADD-01 | 10 | Yes |
| ADD-02 | 10 | Yes |
| ADD-03 | 10 | Yes |
| ADD-04 | 10 | Yes |
| ADD-05 | 10 | Yes |
| ADD-06 | 10 | Yes |
| EDIT-01 | 11 | Yes |
| REM-01 | 11 | Yes |
| REM-02 | 11 | Yes |
| REM-03 | 11 | Yes |

**v1.2 Coverage:** 13/13 requirements mapped (100%)

### v1.3 Requirements

| Requirement | Phase | Covered |
|-------------|-------|---------|
| CLI-01 | 12 | Done |
| CLI-02 | 12 | Done |
| CLI-03 | 12 | Done |
| CLI-04 | 12 | Done |
| CLI-05 | 12 | Done |
| CLI-06 | 12 | Done |
| CLI-07 | 12 | Done |

**v1.3 Coverage:** 7/7 requirements mapped (100%)

### v1.4 Requirements

| Requirement | Phase | Covered |
|-------------|-------|---------|
| NOTE-01 | 13 | Yes |
| NOTE-02 | 13 | Yes |
| NOTE-03 | 13 | Yes |
| NOTE-04 | 14 | Yes |
| NOTE-05 | 14 | Yes |
| NOTE-06 | 13 | Yes |
| NOTE-07 | 13 | Yes |
| NOTE-08 | 13 | Yes |
| NOTE-09 | 15 | Yes |
| NOTE-10 | 15 | Yes |
| NOTE-11 | 15 | Yes |
| NOTE-12 | 15 | Yes |

**v1.4 Coverage:** 12/12 requirements mapped (100%)

### v1.5 Requirements

| Requirement | Phase | Covered |
|-------------|-------|---------|
| CATG-01 | 16 | Yes |
| CATG-02 | 16 | Yes |
| DEDUP-01 | 16 | Yes |
| DEDUP-02 | 16 | Yes |
| DEDUP-03 | 16 | Yes |
| CATG-03 | 17 | Yes |
| CATG-04 | 17 | Yes |
| CATG-05 | 17 | Yes |
| CATG-06 | 17 | Yes |
| CATG-07 | 17 | Yes |
| CATG-08 | 18 | Yes |
| CATG-09 | 18 | Yes |
| CATG-10 | 18 | Yes |
| SETT-01 | 19 | Yes |
| SETT-02 | 19 | Yes |
| SETT-03 | 19 | Yes |
| SETT-04 | 19 | Yes |
| SETT-05 | 19 | Yes |
| PREV-01 | 20 | Yes |
| PREV-02 | 20 | Yes |
| PREV-03 | 20 | Yes |
| PREV-04 | 20 | Yes |
| PREV-05 | 20 | Yes |
| PREV-06 | 20 | Yes |

**v1.5 Coverage:** 24/24 requirements mapped (100%)

---

## Phase Progress

### v1.0 (Complete)

| Phase | Status | Progress |
|-------|--------|----------|
| 1 - Infrastructure Setup | Complete | 100% |
| 2 - UI Layout & Navigation | Complete | 100% |
| 3 - Article Display | Complete | 100% |
| 4 - Article Management | Complete | 100% |
| 5 - Theme Toggle | Complete | 100% |

### v1.1 (Complete)

| Phase | Status | Progress |
|-------|--------|----------|
| 6 - Enhanced Card Interaction | Complete | 100% |
| 7 - Search & Date Filtering | Complete | 100% |
| 8 - Masonry Layout | Complete | 100% |

### v1.2 (Complete)

| Phase | Status | Progress |
|-------|--------|----------|
| 9 - Settings Page Foundation | Complete | 100% |
| 10 - Add Blog Flow | Complete | 100% |
| 11 - Edit and Remove Blogs | Complete | 100% |

### v1.3 (Complete)

| Phase | Status | Progress |
|-------|--------|----------|
| 12 - CLI Foundation | Complete | 100% |

### v1.4 (Complete)

| Phase | Status | Progress |
|-------|--------|----------|
| 13 - CLI Notes Infrastructure | Complete | 100% |
| 14 - CLI Filtering Enhancement | Complete | 100% |
| 15 - UI Note Display | Complete | 100% |

**v1.0 Progress:** 5/5 phases complete (100%)
**v1.1 Progress:** 3/3 phases complete (100%)
**v1.2 Progress:** 3/3 phases complete (100%)
**v1.3 Progress:** 1/1 phases complete (100%)
**v1.4 Progress:** 3/3 phases complete (100%) — 6 plans complete
**v1.5 Progress:** 0/5 phases complete (0%)
**Overall Progress:** 15/20 phases complete (75%)

---

*Roadmap created: 2026-02-02*
*v1.1 roadmap added: 2026-02-03*
*Phase 6 planned: 2026-02-03*
*Phase 6 complete: 2026-02-03*
*Phase 7 planned: 2026-02-03*
*Phase 7 complete: 2026-02-03*
*Phase 8 planned: 2026-02-03*
*Phase 8 complete: 2026-02-03*
*v1.1 COMPLETE: 2026-02-03*
*v1.2 roadmap added: 2026-02-08*
*Phase 9 planned: 2026-02-08*
*Phase 9 complete: 2026-02-09*
*Phase 10 planned: 2026-02-09*
*Phase 10 complete: 2026-02-09*
*Phase 11 planned: 2026-02-09*
*Phase 11 complete: 2026-02-09*
*v1.2 COMPLETE: 2026-02-09*
*v1.3 roadmap added: 2026-05-07*
*Phase 12 planned: 2026-05-07*
*v1.4 roadmap added: 2026-05-07*
*Plan 13-03 complete: 2026-05-07*
*Phase 15 planned: 2026-05-08*
*v1.5 roadmap added: 2026-05-08*
*Last updated: 2026-05-08*