# Roadmap: BlogWatcher UI

**Created:** 2026-02-02
**Depth:** standard
**Total Phases:** 12

## Milestones

- **v1.0 MVP** - Phases 1-5 (shipped 2026-02-03)
- **v1.1 UI Polish & Search** - Phases 6-8 (shipped 2026-02-03)
- **v1.2 Blog Management** - Phases 9-11 (shipped 2026-02-09)
- **v1.3 CLI System** - Phase 12 (planning)

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
4. User can click "Last Month" filter and see only articles from past 30 days
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

## v1.3 CLI System (Planning)

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

**Plans:** 3 plans

Plans:
- [ ] 12-01-PLAN.md — CLI framework setup (cobra root, global flags, serve command)
- [ ] 12-02-PLAN.md — Blog commands (scan all, scan specific)
- [ ] 12-03-PLAN.md — Article commands (list, mark-read, mark-unread)

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
| CLI-01 | 12 | Planned |
| CLI-02 | 12 | Planned |
| CLI-03 | 12 | Planned |
| CLI-04 | 12 | Planned |
| CLI-05 | 12 | Planned |
| CLI-06 | 12 | Planned |
| CLI-07 | 12 | Planned |

**v1.3 Coverage:** 7/7 requirements mapped (100%)

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

### v1.3 (Planning)

| Phase | Status | Progress |
|-------|--------|----------|
| 12 - CLI Foundation | Planned | 0% |

**v1.0 Progress:** 5/5 phases complete (100%)
**v1.1 Progress:** 3/3 phases complete (100%)
**v1.2 Progress:** 3/3 phases complete (100%)
**v1.3 Progress:** 0/1 phases complete (0%)
**Overall Progress:** 11/12 phases complete (92%)

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
*Last updated: 2026-05-07*
