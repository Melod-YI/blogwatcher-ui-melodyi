# Requirements: BlogWatcher UI

**Defined:** 2026-05-08
**Core Value:** Read and manage blog articles through a clean, responsive web interface without touching the CLI.

## v1.5 Requirements

Requirements for Blog Management Enhancement milestone. Each maps to roadmap phases.

### SETT - Blog Settings Enhancement

- [ ] **SETT-01**: 设置页面显示 Blog URL 和 Feed URL
- [ ] **SETT-02**: 设置页面可编辑 Blog URL（inline 编辑）
- [ ] **SETT-03**: 设置页面可编辑 Feed URL（inline 编辑）
- [ ] **SETT-04**: 编辑时验证 URL 格式（HTTP/HTTPS）
- [ ] **SETT-05**: 保存后立即更新数据库

### CATG - Category System

- [ ] **CATG-01**: 数据库创建 categories 表（id, name, created_at）
- [ ] **CATG-02**: 数据库添加 blog.category_id 字段（nullable, foreign key）
- [ ] **CATG-03**: 设置页面添加分类管理区
- [ ] **CATG-04**: 用户可创建新分类（输入名称）
- [ ] **CATG-05**: 用户可编辑分类名称（inline 编辑）
- [ ] **CATG-06**: 用户可删除分类（删除时 blog.category_id 置空）
- [ ] **CATG-07**: Blog 编辑时可选择分类（下拉选择）
- [ ] **CATG-08**: Subscriptions 按分类分层展示
- [ ] **CATG-09**: 未分类 blog 在 Subscriptions 顶层显示
- [ ] **CATG-10**: CLI article list --category 过滤

### PREV - Blog Preview

- [ ] **PREV-01**: 添加 blog 表单有预览按钮
- [ ] **PREV-02**: 点击预览触发临时 feed 解析
- [ ] **PREV-03**: 预览页面显示解析的文章列表（最多 20 条）
- [ ] **PREV-04**: 预览失败显示错误信息
- [ ] **PREV-05**: 预览页面有保存按钮（保存为正式 blog）
- [ ] **PREV-06**: 预览页面有返回修改按钮（返回添加表单）

### DEDUP - Article Deduplication

- [ ] **DEDUP-01**: AddArticlesBulk 使用 INSERT OR IGNORE
- [ ] **DEDUP-02**: 扫描时遇到重复 URL 静默跳过
- [ ] **DEDUP-03**: 扫描结果统计跳过的文章数量

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Full-text Search

- **FTS-01**: Full-text search across article content (would require storing content)

### OPML Import/Export

- **OPML-01**: Import subscriptions from OPML file
- **OPML-02**: Export subscriptions to OPML file

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Edit blog preview | PREV requirements only cover new blog scenario, edit preview deferred |
| Multi-category blogs | Single category sufficient for v1.5, complexity deferred |
| Category ordering | Categories displayed alphabetically, custom order deferred |
| Blog URL change cascade | Changing Blog URL does not auto-update articles, manual rescan needed |
| Preview for edit scenario | Add blog preview sufficient, edit preview adds complexity without clear value |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| (Empty — will be populated during roadmap creation) |

**Coverage:**
- v1.5 requirements: 24 total
- Mapped to phases: 0
- Unmapped: 24 ⚠️

---
*Requirements defined: 2026-05-08*
*Last updated: 2026-05-08 after initial definition*