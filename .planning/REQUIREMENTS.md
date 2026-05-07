# Requirements: BlogWatcher UI

**Defined:** 2026-05-07
**Core Value:** Read and manage blog articles through a clean, responsive web interface without touching the CLI.

## v1.4 Requirements

Requirements for Article Notes milestone. Each maps to roadmap phases.

### CLI Notes

- [ ] **NOTE-01**: CLI `note --article-id <id> --file <path>` 写入备注（完整复制文件内容）
- [ ] **NOTE-02**: CLI `note delete --article-id <id>` 删除备注
- [ ] **NOTE-03**: 缺少必填参数时报错退出

### CLI Filtering

- [ ] **NOTE-04**: CLI `article list --not-noted` 筛选无备注文章（仅过滤备注状态）
- [ ] **NOTE-05**: --not-noted 可与 --unread 组合使用，筛选未读且无备注文章

### Storage

- [ ] **NOTE-06**: 备注存储于 ~/.blogwatcher/notes/{article_id}.md
- [ ] **NOTE-07**: articles 表新增 has_note BOOLEAN 字段
- [ ] **NOTE-08**: 写入/删除备注时同步更新 has_note 字段

### UI Display

- [ ] **NOTE-09**: 有备注的文章卡片显示备注按钮
- [ ] **NOTE-10**: 点击备注按钮新标签页打开 Markdown 渲染页面
- [ ] **NOTE-11**: Markdown 渲染支持 GFM 格式（表格、删除线、任务列表）
- [ ] **NOTE-12**: 备注 页面显示文章标题和原文链接

## v2 Requirements

Deferred to future release.

### Notes Enhancement

- **NOTE-V2-01**: CLI `note show --article-id <id>` 查看备注内容
- **NOTE-V2-02**: article list --has-notes 筛选有备注文章
- **NOTE-V2-03**: 备注创建时间显示

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| UI 内编辑备注 | CLI 专注，用户使用外部编辑器 |
| 备注版本历史 | 增加复杂度，v1 仅支持覆盖写入 |
| 备注附件支持 | 保持简单，纯 Markdown 文本 |
| 备注搜索 | 需要额外索引机制，推迟 |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| NOTE-01 | Phase 13 | Pending |
| NOTE-02 | Phase 13 | Pending |
| NOTE-03 | Phase 13 | Pending |
| NOTE-04 | Phase 14 | Pending |
| NOTE-05 | Phase 14 | Pending |
| NOTE-06 | Phase 13 | Pending |
| NOTE-07 | Phase 13 | Pending |
| NOTE-08 | Phase 13 | Pending |
| NOTE-09 | Phase 15 | Pending |
| NOTE-10 | Phase 15 | Pending |
| NOTE-11 | Phase 15 | Pending |
| NOTE-12 | Phase 15 | Pending |

**Coverage:**
- v1.4 requirements: 12 total
- Mapped to phases: 12
- Unmapped: 0 ✓

---
*Requirements defined: 2026-05-07*
*Last updated: 2026-05-07 after initial definition*
