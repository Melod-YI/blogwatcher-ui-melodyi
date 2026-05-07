# Features Research: Article Notes 功能

**研究日期:** 2026-05-07
**里程碑:** v1.4 Article Notes

## 功能类别

### CLI 备注命令

**Table stakes（必备功能）:**
- note --article-id <id> --file <path> 写入备注
- note delete --article-id <id> 删除备注
- 必填参数校验（缺少参数报错）
- 文件内容完整复制（不依赖源文件）

**Differentiators（差异化）:**
- 输出友好提示（成功/失败消息）

**Anti-features（不做）:**
- CLI 内编辑备注（用户使用外部编辑器）
- 备注版本历史

### CLI 筛选增强

**Table stakes:**
- article list --not-noted 筛选未读且无备注文章
- 与现有筛选参数组合使用

**Differentiators:**
- --has-notes 筛选有备注文章（反向筛选）

### UI 备注展示

**Table stakes:**
- 有备注的文章卡片显示备注按钮
- 点击按钮新标签页打开渲染页面
- Markdown 渲染（GFM 格式）

**Differentiators:**
- 备注创建时间显示

**Anti-features:**
- UI 内编辑备注（仅 CLI 编辑）
- 备注附件支持

## 研究发现

### Markdown 渲染最佳实践

1. GFM 扩展支持表格、删除线、任务列表、自动链接
2. HTML 输出需配置安全选项防止 XSS
3. 使用 CSS 配置 Markdown 渲染样式

---
*研究完成: 2026-05-07*
