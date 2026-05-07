# Architecture Research: Article Notes 功能

**研究日期:** 2026-05-07
**里程碑:** v1.4 Article Notes

## 组件新增

### CLI 层

**新增命令:** internal/cli/commands/note.go

note 命令结构:
- note (命令组)
  - --article-id <id> (必填)
  - --file <path> (写入时必填)
  - delete 子命令
    - --article-id <id>

**集成点:**
- root.go: 添加 NewNoteCmd()
- article.go: 添加 --not-noted flag

### Storage 层

**新增结构:** internal/storage/note.go

方法:
- WriteNote(articleID int64, content []byte) error
- DeleteNote(articleID int64) error
- HasNote(articleID int64) bool
- GetNoteContent(articleID int64) ([]byte, error)
- ListArticlesWithoutNotes(unreadOnly bool) ([]model.Article, error)

**数据库变更:**

ALTER TABLE articles ADD COLUMN has_note BOOLEAN DEFAULT FALSE;

### Server 层

**新增路由:** GET /note/{id} → 渲染 Markdown 备注

**新增处理器:** HandleNoteView

**新增模板:** templates/note.html

### Markdown 渲染

**新增包:** internal/markdown/renderer.go

## 构建顺序建议

| 顺序 | 组件 | 依赖 |
|---|---|---|
| 1 | 数据库迁移（has_note 字段） | 无 |
| 2 | NoteStore 实现 | 数据库迁移 |
| 3 | CLI note 命令 | NoteStore |
| 4 | article list --not-noted | NoteStore |
| 5 | Markdown 渲染器 | 无 |
| 6 | Server 路由和处理器 | Markdown 渲染器 + NoteStore |
| 7 | UI 模板 | Server 路由 |
| 8 | 文章卡片备注按钮 | has_note 字段 |

---
*研究完成: 2026-05-07*
