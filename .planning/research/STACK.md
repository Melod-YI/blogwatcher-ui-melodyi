# Stack Research: Article Notes 功能

**研究日期:** 2026-05-07
**里程碑:** v1.4 Article Notes

## 新增依赖

### Markdown 渲染

**推荐:** Goldmark (github.com/yuin/goldmark)

| 库 | 状态 | 特点 |
|---|---|---|
| Goldmark | ✓ 活跃维护 | CommonMark 兼容、GFM 扩展、Hugo 使用、XSS 安全 |
| Blackfriday v2 | 维护模式 | 成熟稳定，但已迁移至 Goldmark |
| Gomarkdown | Blackfriday fork | 继续开发但不如 Goldmark 流行 |

**安装:**
```bash
go get github.com/yuin/goldmark
go get github.com/yuin/goldmark/extension  # GFM 扩展
go get github.com/yuin/goldmark/renderer/html
```

### 已有依赖（无需新增）

| 依赖 | 用途 |
|---|---|
| github.com/spf13/cobra | CLI 命令框架 |
| modernc.org/sqlite | SQLite 数据库 |

## 技术选型决策

### Markdown 存储

- **文件系统:** ~/.blogwatcher/notes/{article_id}.md
- **优点:** 简单直接、便于外部编辑、不占用数据库空间
- **缺点:** 需要管理文件同步（删除文章时需删除对应备注）

### 数据库变更

- **articles 表新增字段:** has_note BOOLEAN DEFAULT FALSE
- **用途:** UI 查询时筛选有备注的文章

## 与现有架构的集成点

1. **CLI:** 新增 note 命令组（与 article、blog 命令同级）
2. **Storage:** 新增 NoteStore 或扩展现有 Database
3. **Server:** 新增 /note/{id} 路由渲染 Markdown
4. **Templates:** 新增 note.html 模板

---
*研究完成: 2026-05-07*
