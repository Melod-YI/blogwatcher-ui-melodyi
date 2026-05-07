---
phase: 13-cli-notes-infrastructure
plan: 02A
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/model/model.go
autonomous: true
requirements:
  - NOTE-07
user_setup: []

must_haves:
  truths:
    - "Article 结构体包含 HasNote bool 字段"
    - "ArticleWithBlog 结构体包含 HasNote bool 字段"
  artifacts:
    - path: "internal/model/model.go"
      provides: "Article and ArticleWithBlog models with HasNote field"
      contains:
        - "HasNote bool"
      min_lines: 50
  key_links:
    - from: "model.Article"
      to: "articles.has_note"
      via: "HasNote field"
      pattern: "HasNote bool"
---

<objective>
更新 Article 和 ArticleWithBlog 模型，添加 HasNote 字段。

Purpose: 确保数据模型支持 has_note 字段，为后续 scan 函数和查询更新提供基础。
Output: model.go 添加 HasNote 字段。
</objective>

<execution_context>
@C:/workspace/blogwatcher-ui-melodyi/blogwatcher-ui/.claude/get-shit-done/workflows/execute-plan.md
@C:/workspace/blogwatcher-ui-melodyi/blogwatcher-ui/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/phases/13-cli-notes-infrastructure/13-CONTEXT.md
@.planning/phases/13-cli-notes-infrastructure/13-PATTERNS.md

<interfaces>
<!-- 从 model.go 提取的关键接口 -->

现有 Article 结构体:
```go
type Article struct {
    ID             int64
    BlogID         int64
    Title          string
    URL            string
    ThumbnailURL   string
    PublishedDate  *time.Time
    DiscoveredDate *time.Time
    IsRead         bool
}
```

现有 ArticleWithBlog 结构体:
```go
type ArticleWithBlog struct {
    ID             int64
    BlogID         int64
    Title          string
    URL            string
    ThumbnailURL   string
    PublishedDate  *time.Time
    DiscoveredDate *time.Time
    IsRead         bool
    BlogName       string
    BlogURL        string
}
```
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: 更新 Article 和 ArticleWithBlog 结构体</name>
  <files>internal/model/model.go</files>
  <read_first>
    - internal/model/model.go (查看 Article 和 ArticleWithBlog 结构体定义)
  </read_first>
  <behavior>
    - Test 1: Article 结构体包含 `HasNote bool` 字段
    - Test 2: ArticleWithBlog 结构体包含 `HasNote bool` 字段
    - Test 3: 字段位于 IsRead 之后（Article）和 BlogURL 之后（ArticleWithBlog）
    - Test 4: go build ./internal/model/ 成功
  </behavior>
  <action>
在 model.go 中为 Article 和 ArticleWithBlog 结构体添加 HasNote bool 字段。

**具体修改：**

1. **Article 结构体（IsRead 之后添加）：**
```go
type Article struct {
    ID             int64
    BlogID         int64
    Title          string
    URL            string
    ThumbnailURL   string
    PublishedDate  *time.Time
    DiscoveredDate *time.Time
    IsRead         bool
    HasNote        bool  // 新增：文章备注状态
}
```

2. **ArticleWithBlog 结构体（BlogURL 之后添加）：**
```go
type ArticleWithBlog struct {
    ID             int64
    BlogID         int64
    Title          string
    URL            string
    ThumbnailURL   string
    PublishedDate  *time.Time
    DiscoveredDate *time.Time
    IsRead         bool
    BlogName       string
    BlogURL        string
    HasNote        bool  // 新增：文章备注状态
}
```

**字段位置说明：**
- Article 结构体：HasNote 位于 IsRead 之后
- ArticleWithBlog 结构体：HasNote 位于 BlogURL 之后
- 字段顺序与后续 SELECT 查询字段顺序保持一致
</action>
  <verify>
    <automated>go build ./internal/model/</automated>
  </verify>
  <acceptance_criteria>
    - model.go Article 结构体包含 `HasNote bool` 字段
    - model.go ArticleWithBlog 结构体包含 `HasNote bool` 字段
    - 字段位于 IsRead 之后（Article）和 BlogURL 之后（ArticleWithBlog）
    - go build ./internal/model/ 成功
  </acceptance_criteria>
  <done>
Article 和 ArticleWithBlog 结构体添加 HasNote 字段完成，编译通过。为 Plan 13-02B 的 scan 函数更新提供基础。
</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Model → Database | 结构体字段映射数据库列，字段顺序必须与查询一致 |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-13-04 | Tampering | Article.HasNote | accept | 字段添加不改变现有数据，迁移添加默认值 FALSE |
</threat_model>

<verification>
## 整体阶段验证

1. **编译验证：**
   ```bash
   go build ./internal/model/
   ```

2. **字段存在验证：**
   ```bash
   grep -c "HasNote bool" internal/model/model.go
   ```
   应返回 2（Article 和 ArticleWithBlog 各一个）
</verification>

<success_criteria>
1. Article 结构体包含 HasNote 字段
2. ArticleWithBlog 结构体包含 HasNote 字段
3. 所有代码编译通过
</success_criteria>

<output>
完成后创建 `.planning/phases/13-cli-notes-infrastructure/13-02A-SUMMARY.md`
</output>