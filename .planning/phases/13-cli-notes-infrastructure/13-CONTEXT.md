# Phase 13: CLI Notes Infrastructure - Context

**Gathered:** 2026-05-07
**Status:** Ready for planning

## Phase Boundary

实现 CLI 备注基础设施：
- 顶层 `note` 命令（写入备注）
- `note delete` 子命令（删除备注）
- 备注文件存储于 ~/.blogwatcher/notes/{article_id}.md
- 数据库 schema 更新：articles 表新增 has_note 字段
- 写入/删除备注时同步更新 has_note 状态

本阶段不包含：
- UI 备注显示（Phase 15）
- CLI 备注筛选功能（Phase 14）

## Implementation Decisions

### 命令结构
- **D-01:** `note` 为顶层命令，而非 `article` 的子命令
- **D-02:** 命令形式：`blogwatcher note --article-id <id> --file <path>`
- **D-03:** 子命令形式：`blogwatcher note delete --article-id <id>`
- **D-04:** 与现有 CLI 命令风格一致（使用 Cobra，全局 --db flag）

### 文章验证
- **D-05:** 写入备注前必须验证文章 ID 存在于数据库
- **D-06:** 若文章不存在，输出错误消息并退出
- **D-07:** 验证需要数据库查询开销，但确保数据一致性

### 备注文件处理
- **D-08:** 使用文件复制（字节级操作），而非读写处理
- **D-09:** Go 实现：使用 `io.Copy` 或 `os.ReadFile` + `os.WriteFile`
- **D-10:** 不涉及编码转换，直接复制字节

### 覆盖行为
- **D-11:** 已有备注时静默覆盖，无需用户确认
- **D-12:** 符合备注可能需要更新的使用场景

### 目录管理
- **D-13:** ~/.blogwatcher/notes/ 目录不存在时自动创建
- **D-14:** 使用 `os.MkdirAll` 确保目录存在

### 错误处理
- **D-15:** 缺少必填参数（--article-id 或 --file）时简单报错退出
- **D-16:** 错误消息风格与现有 CLI 命令一致（简洁的单行消息）
- **D-17:** 删除不存在的备注时报错提示，告知用户操作无效

### Claude's Discretion
无 — 用户对所有决策提供了明确选择。

## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### CLI 模式
- `internal/cli/commands/root.go` — Cobra 根命令定义，子命令注册模式
- `internal/cli/commands/article.go` — article 命令结构，筛选参数处理模式
- `internal/cli/commands/blog.go` — blog 命令结构，子命令组织模式
- `internal/cli/flags/flags.go` — 全局 --db flag 定义和使用模式

### 存储层
- `internal/storage/database.go` — 数据库操作方法，schema 迁移模式
- `internal/model/model.go` — Article 数据模型

### 项目配置
- `.planning/ROADMAP.md` — Phase 13 需求和成功标准
- `.planning/REQUIREMENTS.md` — NOTE-01/02/03/06/07/08 需求定义
- `.planning/PROJECT.md` — 技术栈约束（Go + Cobra）

## Existing Code Insights

### Reusable Assets
- `internal/cli/flags.DBPath()` — 获取全局数据库路径，新命令可直接使用
- `storage.OpenDatabase()` — 打开数据库连接，幂等初始化
- `storage.ensureMigrations()` — schema 迁移模式，可添加 has_note 字段
- `storage.GetBlogByID()` — 现有 GetByID 模式，可参考实现 GetArticleByID

### Established Patterns
- Cobra 命令注册：在 `init()` 中调用 `rootCmd.AddCommand()`
- 命令结构：命令组包含子命令，子命令定义 flags 和 run 函数
- 错误处理：`fmt.Fprintf(os.Stderr, "错误消息: %v\n", err)` + `os.Exit(1)`
- 输出格式：成功时简单打印确认消息
- 幂等迁移：使用 `columnExists()` 检查字段是否存在

### Integration Points
- `internal/cli/commands/root.go:init()` — 注册新 note 命令
- `internal/storage/database.go:ensureMigrations()` — 添加 has_note 字段迁移
- `internal/storage/database.go` — 添加备注相关数据库方法（GetArticleByID, UpdateArticleHasNote）
- `~/.blogwatcher/notes/` — 新的备注存储目录

## Specific Ideas

无特殊要求 — 按标准 CLI 命令风格实现。

## Deferred Ideas

None — 讨论始终保持在阶段范围内。

---

*Phase: 13-CLI Notes Infrastructure*
*Context gathered: 2026-05-07*