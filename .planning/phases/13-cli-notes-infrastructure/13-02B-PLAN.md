---
phase: 13-cli-notes-infrastructure
plan: 02B
type: execute
wave: 2
depends_on:
  - 13-02A
files_modified:
  - internal/storage/database.go
autonomous: true
requirements:
  - NOTE-07
user_setup: []

must_haves:
  truths:
    - "scanArticle 扫描 has_note 字段"
    - "scanArticleWithBlog 扫描 has_note 字段"
    - "scanArticleWithBlogAndCount 扫描 has_note 字段"
    - "所有 Article 查询 SELECT 语句包含 has_note 字段"
  artifacts:
    - path: "internal/storage/database.go"
      provides: "scanArticle and scanArticleWithBlog with has_note scanning"
      contains:
        - "hasNote bool"
        - "&hasNote"
        - "SELECT.*has_note.*FROM articles"
      min_lines: 1200
  key_links:
    - from: "scanArticle"
      to: "model.Article"
      via: "Scan(&hasNote)"
      pattern: "scanner.Scan.*hasNote"
    - from: "SELECT queries"
      to: "articles.has_note"
      via: "column in SELECT list"
      pattern: "SELECT.*has_note.*FROM articles"
---

<objective>
更新所有 scan 函数和 SELECT 查询，添加 has_note 字段支持。

Purpose: 确保数据库查询和扫描函数支持 has_note 字段（per D-07）。
Output: database.go 更新 scanArticle、scanArticleWithBlog、scanArticleWithBlogAndCount 和所有 SELECT 查询。
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
@.planning/phases/13-cli-notes-infrastructure/13-02A-SUMMARY.md

<interfaces>
<!-- 从 database.go 和 model.go 提取的关键接口 -->

Plan 13-02A 已完成的模型更新：
- Article.HasNote bool 字段已添加
- ArticleWithBlog.HasNote bool 字段已添加

现有 scanArticle 函数 (database.go:854-892):
```go
func scanArticle(scanner interface{ Scan(dest ...any) error }) (*model.Article, error) {
    var (
        id            int64
        blogID        int64
        title         string
        url           string
        thumbnailURL  sql.NullString
        publishedDate sql.NullString
        discovered    sql.NullString
        isRead        bool
    )
    if err := scanner.Scan(&id, &blogID, &title, &url, &thumbnailURL, &publishedDate, &discovered, &isRead); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil
        }
        return nil, err
    }
    // ... 字段赋值 ...
}
```

需要更新的 SELECT 查询：
- ListArticles (约第 381 行)
- ListArticlesByReadStatus (约第 415 行)
- ListArticlesWithBlog (约第 447 行)
- SearchArticles (约第 484 行)
- ListArticlesWithFilters (约第 1069 行)
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: 更新 scanArticle 函数</name>
  <files>internal/storage/database.go</files>
  <read_first>
    - internal/storage/database.go (查看 scanArticle 函数定义，约第 854 行)
    - internal/model/model.go (确认 Article.HasNote 字段已由 13-02A 添加)
  </read_first>
  <behavior>
    - Test 1: scanArticle 包含 `hasNote bool` 变量声明
    - Test 2: scanner.Scan 参数包含 `&hasNote`（9个参数）
    - Test 3: 返回的 article.HasNote 值正确
    - Test 4: go build ./internal/storage/ 成功
  </behavior>
  <action>
更新 database.go 中的 scanArticle 函数，添加 has_note 字段扫描。

**具体修改（database.go scanArticle 函数）：**

1. **变量声明（isRead 之后添加）：**
```go
var (
    id            int64
    blogID        int64
    title         string
    url           string
    thumbnailURL  sql.NullString
    publishedDate sql.NullString
    discovered    sql.NullString
    isRead        bool
    hasNote       bool  // 新增
)
```

2. **Scan 参数（&isRead 之后添加）：**
```go
if err := scanner.Scan(&id, &blogID, &title, &url, &thumbnailURL, &publishedDate, &discovered, &isRead, &hasNote); err != nil {
```

3. **字段赋值（IsRead 之后添加）：**
```go
article := &model.Article{
    ID:           id,
    BlogID:       blogID,
    Title:        title,
    URL:          url,
    ThumbnailURL: thumbnailURL.String,
    IsRead:       isRead,
    HasNote:      hasNote,  // 新增
}
```

**重要说明：**
- scanArticle 现在扫描 9 个字段（包含 has_note）
- Plan 13-02A 已在 model.go 添加 HasNote 字段，可直接使用
- Scan 参数顺序必须与后续 SELECT 查询字段顺序一致
</action>
  <verify>
    <automated>go build ./internal/storage/</automated>
  </verify>
  <acceptance_criteria>
    - scanArticle 函数包含 `hasNote bool` 变量声明
    - scanner.Scan 参数包含 `&hasNote`（位于 &isRead 之后，共 9 个参数）
    - article 结构体赋值包含 `HasNote: hasNote`
    - go build ./internal/storage/ 成功
  </acceptance_criteria>
  <done>
scanArticle 函数更新完成，支持 has_note 字段扫描，编译通过。
</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: 更新 scanArticleWithBlog 函数</name>
  <files>internal/storage/database.go</files>
  <read_first>
    - internal/storage/database.go (查看 scanArticleWithBlog 函数定义，约第 894 行)
    - internal/model/model.go (确认 ArticleWithBlog.HasNote 字段已由 13-02A 添加)
  </read_first>
  <behavior>
    - Test 1: scanArticleWithBlog 包含 `hasNote bool` 变量声明
    - Test 2: scanner.Scan 参数包含 `&hasNote`（11个参数）
    - Test 3: 返回的 article.HasNote 值正确
    - Test 4: go build ./internal/storage/ 成功
  </behavior>
  <action>
更新 database.go 中的 scanArticleWithBlog 函数，添加 has_note 字段扫描。

**具体修改（database.go scanArticleWithBlog 函数）：**

1. **变量声明（blogURL 之后添加）：**
```go
var (
    id            int64
    blogID        int64
    title         string
    url           string
    thumbnailURL  sql.NullString
    publishedDate sql.NullString
    discovered    sql.NullString
    isRead        bool
    blogName      string
    blogURL       string
    hasNote       bool  // 新增
)
```

2. **Scan 参数（&blogURL 之后添加）：**
```go
if err := scanner.Scan(&id, &blogID, &title, &url, &thumbnailURL, &publishedDate, &discovered, &isRead, &blogName, &blogURL, &hasNote); err != nil {
```

3. **字段赋值（BlogURL 之后添加）：**
```go
article := &model.ArticleWithBlog{
    ID:           id,
    BlogID:       blogID,
    Title:        title,
    URL:          url,
    ThumbnailURL: thumbnailURL.String,
    IsRead:       isRead,
    BlogName:     blogName,
    BlogURL:      blogURL,
    HasNote:      hasNote,  // 新增
}
```
</action>
  <verify>
    <automated>go build ./internal/storage/</automated>
  </verify>
  <acceptance_criteria>
    - scanArticleWithBlog 函数包含 `hasNote bool` 变量声明
    - scanner.Scan 参数包含 `&hasNote`（位于 &blogURL 之后，共 11 个参数）
    - article 结构体赋值包含 `HasNote: hasNote`
    - go build ./internal/storage/ 成功
  </acceptance_criteria>
  <done>
scanArticleWithBlog 函数更新完成，编译通过。
</done>
</task>

<task type="auto" tdd="true">
  <name>Task 3: 更新所有 Article SELECT 查询语句</name>
  <files>internal/storage/database.go</files>
  <read_first>
    - internal/storage/database.go (查看所有 Article SELECT 查询)
  </read_first>
  <behavior>
    - Test 1: ListArticles SELECT 包含 has_note
    - Test 2: ListArticlesByReadStatus SELECT 包含 has_note
    - Test 3: ListArticlesWithBlog SELECT 包含 has_note
    - Test 4: SearchArticles SELECT 包含 has_note
    - Test 5: ListArticlesWithFilters SELECT 包含 has_note
    - Test 6: go build ./internal/storage/ 成功
  </behavior>
  <action>
更新 database.go 中所有 Article SELECT 查询，添加 has_note 字段。

**需要更新的查询（共 5 处）：**

1. **ListArticles (约第 381 行)：**
```go
query := `SELECT id, blog_id, title, url, thumbnail_url, published_date, discovered_date, is_read, has_note FROM articles WHERE 1=1`
```

2. **ListArticlesByReadStatus (约第 415 行)：**
```go
query := `SELECT id, blog_id, title, url, thumbnail_url, published_date, discovered_date, is_read, has_note FROM articles WHERE is_read = ?`
```

3. **ListArticlesWithBlog (约第 447 行)：**
```go
query := `SELECT a.id, a.blog_id, a.title, a.url, a.thumbnail_url, a.published_date, a.discovered_date, a.is_read, a.has_note, b.name, b.url
    FROM articles a
    INNER JOIN blogs b ON a.blog_id = b.id
    WHERE a.is_read = ?`
```

4. **SearchArticles (约第 484 行)：**
```go
query.WriteString(`SELECT a.id, a.blog_id, a.title, a.url, a.thumbnail_url, a.published_date, a.discovered_date, a.is_read, a.has_note, b.name, b.url, COUNT(*) OVER() as total_count
    FROM articles a`)
```

5. **ListArticlesWithFilters (约第 1069 行)：**
```go
query := `SELECT a.id, a.blog_id, a.title, a.url, a.thumbnail_url, a.published_date, a.discovered_date, a.is_read, a.has_note, b.name, b.url
    FROM articles a
    INNER JOIN blogs b ON a.blog_id = b.id
    WHERE 1=1`
```

**字段顺序说明：**
- Article 单表查询：is_read 之后添加 has_note
- ArticleWithBlog 查询：is_read 之后添加 has_note，blogName、blogURL 保持原位
- SearchArticles：is_read 之后添加 has_note，totalCount 保持末尾
</action>
  <verify>
    <automated>go build ./internal/storage/</automated>
  </verify>
  <acceptance_criteria>
    - ListArticles SELECT 包含 `is_read, has_note`（9个字段）
    - ListArticlesByReadStatus SELECT 包含 `is_read, has_note`
    - ListArticlesWithBlog SELECT 包含 `a.is_read, a.has_note`
    - SearchArticles SELECT 包含 `a.is_read, a.has_note`
    - ListArticlesWithFilters SELECT 包含 `a.is_read, a.has_note`
    - 所有查询字段顺序与 scanArticle/scanArticleWithBlog Scan 参数顺序一致
    - go build ./internal/storage/ 成功
  </acceptance_criteria>
  <done>
所有 Article SELECT 查询更新完成，包含 has_note 字段，编译通过。
</done>
</task>

<task type="auto" tdd="true">
  <name>Task 4: 更新 scanArticleWithBlogAndCount 函数</name>
  <files>internal/storage/database.go</files>
  <read_first>
    - internal/storage/database.go (查看 scanArticleWithBlogAndCount 函数，约第 938 行)
    - internal/model/model.go (确认 ArticleWithBlog.HasNote 字段已由 13-02A 添加)
  </read_first>
  <behavior>
    - Test 1: scanArticleWithBlogAndCount 包含 `hasNote bool` 变量
    - Test 2: scanner.Scan 参数包含 `&hasNote`（12个参数）
    - Test 3: 返回的 article.HasNote 值正确
    - Test 4: go build ./internal/storage/ 成功
  </behavior>
  <action>
更新 database.go 中的 scanArticleWithBlogAndCount 函数，添加 has_note 字段扫描。

**具体修改（database.go scanArticleWithBlogAndCount 函数）：**

1. **变量声明（isRead 之后添加）：**
```go
var (
    id            int64
    blogID        int64
    title         string
    url           string
    thumbnailURL  sql.NullString
    publishedDate sql.NullString
    discovered    sql.NullString
    isRead        bool
    hasNote       bool  // 新增（位于 isRead 之后）
    blogName      string
    blogURL       string
    totalCount    int
)
```

2. **Scan 参数（&isRead 之后添加，与 SearchArticles SELECT 字段顺序一致）：**
```go
if err := scanner.Scan(&id, &blogID, &title, &url, &thumbnailURL, &publishedDate, &discovered, &isRead, &hasNote, &blogName, &blogURL, &totalCount); err != nil {
```

**Scan 参数顺序必须匹配 SearchArticles SELECT 字段顺序：**
```
a.id, a.blog_id, a.title, a.url, a.thumbnail_url, a.published_date, a.discovered_date, a.is_read, a.has_note, b.name, b.url, COUNT(*) OVER() as total_count
  1     2        3       4          5                   6                  7               8           9         10      11            12
```

3. **字段赋值（IsRead 之后添加）：**
```go
article := &model.ArticleWithBlog{
    ID:           id,
    BlogID:       blogID,
    Title:        title,
    URL:          url,
    ThumbnailURL: thumbnailURL.String,
    IsRead:       isRead,
    HasNote:      hasNote,  // 新增
    BlogName:     blogName,
    BlogURL:      blogURL,
}
```

**重要：SearchArticles SELECT 字段顺序（Task 3 已更新）：**
```go
a.id, a.blog_id, a.title, a.url, a.thumbnail_url, a.published_date, a.discovered_date, a.is_read, a.has_note, b.name, b.url, COUNT(*) OVER() as total_count
```
对应 Scan 参数：&id, &blogID, &title, &url, &thumbnailURL, &publishedDate, &discovered, &isRead, &hasNote, &blogName, &blogURL, &totalCount
</action>
  <verify>
    <automated>go build ./internal/storage/</automated>
  </verify>
  <acceptance_criteria>
    - scanArticleWithBlogAndCount 函数包含 `hasNote bool` 变量声明
    - scanner.Scan 参数包含 `&hasNote`（位于 &isRead 之后、&blogName 之前，共 12 个参数）
    - article 结构体赋值包含 `HasNote: hasNote`
    - go build ./internal/storage/ 成功
  </acceptance_criteria>
  <done>
scanArticleWithBlogAndCount 函数更新完成，编译通过。
</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Database → Model | 数据库查询结果映射到模型，字段顺序必须一致 |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-13-05 | Tampering | scanArticle | accept | 字段顺序错误会导致运行时错误，编译时无法检测。需通过测试验证。 |
| T-13-06 | Information Disclosure | HasNote field | accept | has_note 状态公开，无敏感信息 |
</threat_model>

<verification>
## 整体阶段验证

1. **编译验证：**
   ```bash
   go build ./internal/storage/
   ```

2. **字段扫描验证：**
   ```bash
   grep -c "hasNote bool" internal/storage/database.go
   ```
   应返回 3（scanArticle, scanArticleWithBlog, scanArticleWithBlogAndCount）

3. **SELECT 查询验证：**
   ```bash
   grep -c "has_note" internal/storage/database.go
   ```
   应返回至少 6（5 个 SELECT 查询 + 1 个迁移语句）

4. **运行验证（可选）：**
   ```bash
   go run ./cmd/blogwatcher article list --format simple
   ```
   验证查询执行无 column count mismatch 错误。
</verification>

<success_criteria>
1. scanArticle 支持 has_note 扫描（9 个字段）
2. scanArticleWithBlog 支持 has_note 扫描（11 个字段）
3. scanArticleWithBlogAndCount 支持 has_note 扫描（12 个字段）
4. 所有 Article SELECT 查询包含 has_note 字段
5. 所有代码编译通过
6. 运行 article list 命令无 column count mismatch 错误
</success_criteria>

<output>
完成后创建 `.planning/phases/13-cli-notes-infrastructure/13-02B-SUMMARY.md`
</output>