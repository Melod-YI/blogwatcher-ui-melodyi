# Phase 15: UI Note Display - Context

**Gathered:** 2026-05-08
**Status:** Ready for planning

## Phase Boundary

在 UI 上显示文章备注，提供 Markdown 渲染页面：
- 有备注的文章卡片显示备注按钮
- 点击备注按钮新标签页打开 Markdown 渲染页面
- Markdown 渲染支持 GFM 格式（表格、删除线、任务列表）
- 备注页面显示文章标题和原文链接

本阶段不包含：
- 备注编辑功能（CLI 专注，用户使用外部编辑器）
- 备注创建/删除 UI（已通过 CLI 完成）
- 备注版本历史

## Implementation Decisions

### 备注按钮位置与样式
- **D-01:** 备注按钮放在 Summarize 按钮之后，Read/Unread 按钮之前
- **D-02:** 只有 HasNote=true 时显示备注按钮
- **D-03:** 使用相同的 action-btn 类，添加 document-text 图标（Heroicons）
- **D-04:** 点击备注按钮在新标签页打开 /note/{id} 页面

### 备注页面路由与布局
- **D-05:** 路由为 `/note/{id}` - 简洁明了
- **D-06:** 页面布局：
  - 顶部：文章标题（大标题）+ 原文链接（在新标签页打开）
  - 主体：Markdown 渲染内容区域
  - 不需要 sidebar，独立专注页面
- **D-07:** 新增 handler: `handleNote` 获取文章和备注内容
- **D-08:** 新增模板: `note.gohtml` 用于备注页面渲染

### Markdown 渲染方案
- **D-09:** 使用 goldmark 库进行 Markdown 渲染
- **D-10:** 启用 GFM 扩展支持：
  - 表格（Tables）
  - 删除线（Strikethrough）
  - 任务列表（Task lists）
  - 自动链接（Autolinks）
- **D-11:** 服务器端渲染，无需 JavaScript 增强
- **D-12:** 备注文件不存在时显示友好提示（如"备注内容为空"）

### Markdown 样式与主题
- **D-13:** 使用现有 CSS 变量系统，定义 .markdown-body 样式类
- **D-14:** 自动适配 Light/Dark 主题
- **D-15:** 样式包括：
  - 段落、标题、列表的基础排版
  - 代码块样式（使用现有代码块背景色）
  - 表格样式（带边框和交替行背景）
  - 链接样式（使用现有链接颜色）

### Claude's Discretion
用户选择自动决策，Claude 有以下灵活性：
- 备注按钮的具体图标选择（推荐 Heroicons document-text）
- Markdown 样式的具体细节（颜色、间距等）
- 备注页面顶部布局的具体样式

## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### 现有 UI 结构
- `assets/templates/partials/article-items.gohtml` — 文章卡片模板，action-actions 区域结构
- `assets/templates/pages/index.gohtml` — 页面模板结构，sidebar 和 main-content 布局
- `assets/templates/base.gohtml` — 基础脚本和主题切换逻辑

### 现有 Handler 和路由
- `internal/server/handlers.go` — 现有 handler 实现模式
- `internal/server/routes.go` — 路由注册模式
- `internal/server/server.go` — Server 结构和模板加载

### 数据模型和存储
- `internal/model/model.go` — Article 和 ArticleWithBlog 模型（包含 HasNote 字段）
- `internal/storage/database.go` — 数据库操作方法
- `~/.blogwatcher/notes/{article_id}.md` — 备注文件存储位置

### 项目配置
- `.planning/ROADMAP.md` — Phase 15 需求和成功标准
- `.planning/REQUIREMENTS.md` — NOTE-09/10/11/12 需求定义
- `.planning/PROJECT.md` — 技术栈约束（Go templates + HTMX）

### Markdown 渲染参考
- goldmark 库文档 — https://github.com/yuin/goldmark
- GFM 扩展 — https://github.com/yuin/goldmark#extensions

## Existing Code Insights

### Reusable Assets
- `article-items.gohtml` 的 action-btn 类 — 备注按钮可直接使用
- `internal/server/server.go` 的模板加载逻辑 — 新增 note.gohtml 模板
- 现有的 CSS 变量系统（主题色、背景色等）— Markdown 样式可复用
- `parseSearchOptions` 模式 — handler 参数解析参考

### Established Patterns
- Go 1.22+ method routing with http.ServeMux — 新路由 `/note/{id}`
- HTMX 用于动态更新，但备注页面是独立页面（非 HTMX partial）
- 模板命名：pages/*.gohtml 用于完整页面，partials/*.gohtml 用于片段
- 备注按钮触发新标签页打开（target="_blank"）

### Integration Points
- `internal/server/routes.go:registerRoutes()` — 添加 GET /note/{id} 路由
- `internal/server/handlers.go` — 新增 handleNote handler
- `assets/templates/partials/article-items.gohtml` — 在 action-actions 区域添加备注按钮
- `assets/templates/pages/note.gohtml` — 新建备注页面模板
- `internal/server/server.go:loadTemplates()` — 加载新模板文件
- 备注文件读取：使用 os.ReadFile 读取 ~/.blogwatcher/notes/{id}.md

## Specific Ideas

无特殊要求 — 按标准 UI 页面和按钮风格实现。

## Deferred Ideas

None — 讨论始终保持在阶段范围内。

---

*Phase: 15-UI Note Display*
*Context gathered: 2026-05-08*