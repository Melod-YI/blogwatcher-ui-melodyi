---
name: article-tags
description: 文章标签功能设计文档
---

# 文章标签功能设计

## 概述

在收藏等已有文章维度之外，新增"标签"维度用于对文章标注与分类。标签是独立维度：任何文章（不论是否收藏）都能打标签。一个文章可拥有多个标签，标签种类可自定义。

## 核心决策

| 方面 | 决定 |
|------|------|
| 作用范围 | 标签为独立维度，不限收藏文章；任何文章均可打标签 |
| 创建方式 | 即用即建（on-the-fly）+ 独立标签管理页（重命名/删除）；UI 组合框可选可输 |
| 生命周期 | 重命名 + 删除（级联解除关联）；不做"合并"（YAGNI） |
| UI 打标签 | 卡片内联弹层（增量增删）+ 单文标签管理页（全量保存），两者都要 |
| 标签筛选 | 侧边栏 Tags 分区 + 文章列表工具栏下拉，两者都要；多标签筛选仅支持单标签（一次按一个标签过滤） |
| 视觉 | 纯文字 chips，无颜色（YAGNI） |
| 数据模型 | `tags` 表 + `article_tags` 关联表（规范多对多） |
| 命名 | Tag，表 `tags` / `article_tags`，列 `name`；CLI 变量 `--tag` |

---

## 第一部分：数据库设计

### 新增表

在 `database.go` 的 `ensureMigrations()` 中 `CREATE TABLE IF NOT EXISTS`：

```sql
CREATE TABLE IF NOT EXISTS tags (
    id          INTEGER PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS article_tags (
    tag_id      INTEGER NOT NULL,
    article_id  INTEGER NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tag_id, article_id)
);
```

外键约束声明仅作文档意图，不依赖其级联：modernc.org/sqlite 默认未开 `PRAGMA foreign_keys=ON`，现有 migration 也不依赖 FK。删除标签/全量替换在 storage 层用事务显式处理关联行，与现有代码风格一致。

### Model 更新

`model.go`：

```go
type Tag struct {
    ID           int64
    Name         string
    CreatedAt    time.Time
    ArticleCount int64 // 列表展示用，ListTags 时填充
}

// ArticleWithBlog 新增字段
Tags []Tag // 卡片渲染标签 chips
```

### SearchOptions / ListFilterOptions 更新

```go
type SearchOptions struct {
    // 现有字段...
    TagName string // 空字符串 = 不按标签筛选；非空 = 该标签的文章
}
// ListFilterOptions 同样新增 TagName string
```

---

## 第二部分：Storage 层

新增方法：

```go
// CreateTag 创建标签，重名幂等：已存在则返回现有标签
func (db *Database) CreateTag(name string) (Tag, error)

func (db *Database) GetTagByName(name string) (*Tag, error)
func (db *Database) GetTagByID(id int64) (*Tag, error)

// ListTags 列出所有标签，带关联文章计数（LEFT JOIN COUNT）
func (db *Database) ListTags() ([]Tag, error)

func (db *Database) RenameTag(id int64, newName string) error

// DeleteTag 事务内：先 DELETE FROM article_tags WHERE tag_id=?，
// 再 DELETE FROM tags WHERE id=?
func (db *Database) DeleteTag(id int64) (affected int64, err error)

// AddArticleTag 幂等：INSERT OR IGNORE
func (db *Database) AddArticleTag(articleID, tagID int64) error

// RemoveArticleTag 移除关联，无影响行算成功
func (db *Database) RemoveArticleTag(articleID, tagID int64) error

// SetArticleTags 全量替换：事务内 DELETE 旧关联，INSERT 新关联
func (db *Database) SetArticleTags(articleID int64, tagIDs []int64) error

// GetArticleTags 取单文的标签列表
func (db *Database) GetArticleTags(articleID int64) ([]Tag, error)

// GetTagsForArticles 批量取多文的标签（按 article_id 聚合），避免列表渲染 N+1
func (db *Database) GetTagsForArticles(articleIDs []int64) (map[int64][]Tag, error)
```

### 现有查询更新

- `ListArticlesWithFilters` / `SearchArticles`：
  - WHERE 子句新增标签过滤：当 `opts.TagName != ""` 时追加
    `EXISTS(SELECT 1 FROM article_tags at JOIN tags t ON t.id=at.tag_id WHERE at.article_id=a.id AND t.name=?)`。
  - 列表查询返回后，用本页文章 ID 调 `GetTagsForArticles` 装配 `ArticleWithBlog.Tags`（一次查询，避免 N+1）。
- `scanArticleWithBlog` 等扫描函数本身不变（标签由独立查询装配，非同行扫描）。

---

## 第三部分：API 路由与 Handler

### 路由

| 方法 | 路由 | Handler | 说明 |
|------|------|---------|------|
| GET | `/tags` | `handleTagsPage` | 全局标签管理页：列出标签+计数，可重命名/删除 |
| GET | `/tags/list` | `handleTagsListPartial` | 标签列表片段（侧边栏 Tags 分区 + 工具栏下拉用，HTMX） |
| POST | `/tags` | `handleTagCreate` | 创建标签 |
| PUT | `/tags/{id}` | `handleTagRename` | 重命名 |
| DELETE | `/tags/{id}` | `handleTagDelete` | 删除（级联解除关联） |
| GET | `/articles/{id}/tags` | `handleArticleTagsPage` | 单文标签管理页（全量保存） |
| GET | `/articles/{id}/tags/edit` | `handleArticleTagsEditPartial` | 卡片弹层片段：组合框 + 当前 chips + 可选标签 |
| POST | `/articles/{id}/tags` | `handleArticleTagAdd` | 增量加标签（body: `name`，不存在自动建） |
| DELETE | `/articles/{id}/tags/{tagID}` | `handleArticleTagRemove` | 增量移除标签 |
| POST | `/articles/{id}/tags/save` | `handleArticleTagSave` | 全量替换（body: `tag_ids`） |

### Handler 行为约定

- 仿现有 `handleFavorite`/`handleMarkRead` 模式：解析 ID → 调 storage → 重渲染对应片段 → `HX-Trigger: articleListUpdated`。
- 增删 handler 成功后重渲染弹层片段（`handleArticleTagsEditPartial` 内容）；管理页保存后重渲染 chips 区。均触发卡片刷新以同步 chips 与按钮态。
- `parseSearchOptions` 新增 `filter=tag&tag=<name>`：`case "tag": opts.TagName = ...`，`filter` 字符串为 `"tag"`。模板 active 判断扩展为"filter=tag 且 tag 名匹配"（需在列表页模板数据中带 `CurrentTag`）。
- 错误响应：
  - 非法 ID → 400
  - 文章/标签不存在 → 404
  - 重命名撞名（DB UNIQUE 冲突）→ 409
  - 标签名为空或长度 > 50（对齐 `MaxInputLength`）→ 400
  - 事务失败 → 500 + 日志

---

## 第四部分：CLI 命令设计

新增 `tag` 命令组 + article 子命令扩展，沿用 `article.go` / `category.go` 模式：

```
# 标签管理（tag 命令组）
blogwatcher tag list                          # 列出所有标签 + 文章计数（table/json）
blogwatcher tag rename <oldName> <newName>    # 重命名
blogwatcher tag delete <name>                 # 删除（级联解除关联，带确认提示）

# 文章-标签关联（article 命令组下新增子命令）
blogwatcher article tag <id> <name>            # 给文章打标签（不存在自动创建）
blogwatcher article untag <id> <name>          # 移除文章标签（不影响标签本身）

# 文章列表筛选扩展
blogwatcher article list --tag <name>          # 按标签筛选文章
```

实现要点：
- `runArticleTag`：解析 id+name → `GetArticleByID` 验证 → `CreateTag(name)`（幂等）→ `AddArticleTag` → 输出"已为文章 N 添加标签: xxx"。
- `runArticleUntag`：解析 → 验证文章 → `GetTagByName(name)`（不存在报错）→ `RemoveArticleTag` → 输出。
- `runTagList`：`ListTags()` → output 包新增 `FormatTagTable`/`FormatTagJSON`（仿 `FormatCategoryTable`）。
- `runTagRename`：按名查标签 → 不存在报错 → `RenameTag`；撞名友好化提示。
- `runTagDelete`：按名查标签 → 不存在报错 → `DeleteTag` → 输出"已删除标签 xxx（解除 N 篇关联）"。
- `article list --tag`：`runList` 读取 `--tag` 标志置入 `opts.TagName`，与 `--blog`/`--favorited` 等并列。
- 日志：每命令入口/结果按全局约束输出（命令名、目标 id/name、影响行数）。

---

## 第五部分：UI 设计

### 模板变更/新增

| 文件 | 变更 | 内容 |
|------|------|------|
| `sidebar.gohtml` | 修改 | 新增 `Tags` 分区：`<div hx-get="/tags/list" hx-trigger="load, articleListUpdated from:body">`，拉取标签列表；每标签链接 `?filter=tag&tag=xxx`。active 判断扩展为 filter=tag 且 tag 名匹配 |
| `article-items.gohtml` | 修改 | ① 操作区加"标签"按钮（`hx-get="/articles/{id}/tags/edit"`，弹层容器 swap）；② 标题 meta 行渲染 `{{range .Tags}}<span class="tag-chip">{{.Name}}</span>{{end}}` |
| `article-list.gohtml` | 修改 | 工具栏加标签下拉筛选（选项来自 `/tags/list`），切换即改 URL `?filter=tag&tag=xxx` |
| `tags.gohtml` | 新增 | 全局标签管理页：标签名内联编辑（`PUT /tags/{id}`）、计数、删除按钮（`DELETE`，二次确认） |
| `article-tags.gohtml` | 新增 | 单文标签管理页：标题 + 全部标签可勾选 chips + 组合框新建 → 保存 `POST /articles/{id}/tags/save` |
| `article-tags-edit.gohtml` | 新增 | 卡片弹层片段：当前 chips（带 × 删除）+ 组合框（datalist 可选可输）+ 提交 |

### 组合框实现

原生 `<input list="tag-options">` + `<datalist>`（选项 = 全部标签 − 当前已有）。零 JS，可选可输，符合现有 gohtml 风格。

### 卡片 chips 与列表渲染

`handleArticleList` 取本页文章后，调 `GetTagsForArticles(ids)` 一次查询装配 `ArticleWithBlog.Tags`，避免 N+1。

### 样式（styles.css）

- `.tag-chip`：圆角小标签，中性底色，内边距 `2px 8px`，小字号；卡片内用更小尺寸。
- `.tag-popover`：`position: absolute`（相对卡片），窄宽，含 chips 区 + 输入框。
- `.tag-btn`：复用 `.action-btn` 基类。
- 重命名/删除按钮复用现有按钮样式。

---

## 第六部分：数据流

### 卡片弹层打标签（增量）

1. 点卡片"标签"按钮 → `GET /articles/{id}/tags/edit` → `GetArticleTags(id)` + `ListTags()` 渲染弹层（当前 chips + 组合框，选项 = 全部 − 已有）。
2. 提交 → `POST /articles/{id}/tags`（`name`）→ `CreateTag`（幂等）→ `AddArticleTag` → 重渲染弹层 + `HX-Trigger: articleListUpdated`。
3. 点 chips × → `DELETE /articles/{id}/tags/{tagID}` → `RemoveArticleTag` → 重渲染弹层 + 触发卡片刷新。
4. 关闭弹层 → 纯前端，无后端状态。

### 单文管理页（全量）

1. `GET /articles/{id}/tags` → 渲染：标题 + 全部标签可勾选 chips（已关联高亮）+ 组合框可新建。
2. 保存 → `POST /articles/{id}/tags/save`（`tag_ids`）→ `SetArticleTags`（事务删旧插新）→ 重渲染页面。

### 侧边栏/工具栏筛选

1. `GET /tags/list` → `ListTags()` 渲染侧边栏 Tags 分区（名称 + 计数），链接 `?filter=tag&tag=xxx`。
2. 点击 → `GET /articles?filter=tag&tag=xxx` → `parseSearchOptions` 置 `opts.TagName` → `SearchArticles` 用 `EXISTS` 过滤 → 渲染列表。
3. 工具栏下拉切换同理。

---

## 第七部分：测试计划

| 层 | 用例 | 说明 |
|----|------|------|
| Storage | `TestCreateTag` / `TestCreateTag_Duplicate` | 创建；重名幂等返回现有 |
| Storage | `TestListTags_WithCount` | 计数正确 |
| Storage | `TestRenameTag` / `TestRenameTag_Conflict` | 重命名；撞名报错 |
| Storage | `TestDeleteTag_CascadesAssociations` | 删标签后关联清空、文章仍在 |
| Storage | `TestAddArticleTag_Idempotent` / `TestRemoveArticleTag` | 重复加幂等；移除 |
| Storage | `TestSetArticleTags_FullReplace` | 全量替换正确 |
| Storage | `TestSearchArticles_TagFilter` | `TagName` 筛选命中 |
| Storage | `TestGetTagsForArticles_Batch` | 批量装配，避免 N+1 |
| Handler | `TestHandleArticleTagAdd` / `TestHandleArticleTagRemove` | 增删并触发 `articleListUpdated` |
| Handler | `TestHandleArticleTagSave` | 全量替换 |
| Handler | `TestHandleTagCreate/Rename/Delete` | CRUD + 404/409 |
| Handler | `TestHandleArticleTags_NonExistentArticle` → 404 | |
| Handler | `TestParseSearchOptions_TagFilter` | `filter=tag&tag=xxx` 解析正确 |
| CLI | `TestArticleTagCmd` / `TestArticleUntagCmd` | 自动建标签 / 移除 |
| CLI | `TestTagListCmd` / `TestTagRenameCmd` / `TestTagDeleteCmd` | list/rename/delete |
| CLI | `TestArticleListCmd_TagFilter` | `--tag` 筛选 |

---

## 文件变更清单

| 文件 | 变更类型 | 说明 |
|------|----------|------|
| `internal/model/model.go` | 修改 | 新增 `Tag` 结构；`ArticleWithBlog` 加 `Tags`；`SearchOptions`/`ListFilterOptions` 加 `TagName` |
| `internal/storage/database.go` | 修改 | Migration 两张表；Tag/ArticleTag 全套 storage 方法；列表查询加 TagName 过滤 + 批量标签装配 |
| `internal/server/routes.go` | 修改 | 注册 10 条新路由 |
| `internal/server/handlers.go` | 修改 | 新增对应 handler；`parseSearchOptions` 加 `tag` 分支；列表 handler 装配标签 |
| `internal/cli/commands/article.go` | 修改 | 新增 `tag`/`untag` 子命令；`list` 加 `--tag` 标志 |
| `internal/cli/commands/tag.go` | 新增 | `tag` 命令组：list/rename/delete |
| `internal/cli/commands/root.go` | 修改 | 注册 `tag` 命令组 |
| `internal/cli/output/tag.go` | 新增 | `FormatTagTable`/`FormatTagJSON`（仿 `output/category.go`） |
| `assets/templates/partials/sidebar.gohtml` | 修改 | 新增 Tags 分区 |
| `assets/templates/partials/article-items.gohtml` | 修改 | 加标签按钮 + chips 渲染 |
| `assets/templates/partials/article-list.gohtml` | 修改 | 工具栏标签下拉 |
| `assets/templates/pages/tags.gohtml` | 新增 | 全局标签管理页 |
| `assets/templates/pages/article-tags.gohtml` | 新增 | 单文标签管理页 |
| `assets/templates/partials/article-tags-edit.gohtml` | 新增 | 卡片弹层片段 |
| `assets/static/styles.css` | 修改 | tag-chip / popover / tag-btn 样式 |
| `internal/storage/database_test.go` | 修改 | Storage 测试 |
| `internal/server/handlers_test.go` | 修改 | Handler 测试 |
| `internal/cli/commands/*_test.go` | 修改/新增 | CLI 测试 |

---

## 开发与部署流程

遵循项目 `CLAUDE.md`：

1. `go install ./cmd/blogwatcher` — 编译并安装 CLI 到全局。
2. `docker compose build blogwatcher-ui --no-cache && docker compose up -d blogwatcher-ui` — 构建并部署 Docker 服务。
3. `go test ./...` — 所有测试通过后再部署。

## 实现顺序建议

1. DB migration + Model + Storage 方法 + Storage 测试（可独立验证）。
2. Server handler + 路由 + `parseSearchOptions` + Handler 测试。
3. CLI 命令 + output 格式化 + CLI 测试。
4. UI 模板 + 样式。
5. `go install` + `docker compose build` + `go test ./...` 部署验证。
