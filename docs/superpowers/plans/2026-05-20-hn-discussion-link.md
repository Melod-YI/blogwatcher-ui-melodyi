# HN 讨论链接检测功能实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为博客文章自动检测并关联 Hacker News 上的讨论链接

**Architecture:** 在现有 sync 流程中新增 HN 搜索阶段，新建独立 hn 模块封装 Algolia API 调用和 URL 匹配逻辑，数据库新增 hn_url 和 hn_status 列记录搜索结果

**Tech Stack:** Go、SQLite、Algolia HN API、Cobra CLI、Go HTML Template

---

## 文件结构

| 文件 | 负责内容 |
|------|----------|
| `internal/model/model.go` | HNStatus 枚举定义、Article/ArticleWithBlog 结构体新增字段 |
| `internal/storage/database.go` | Migration 新增列、UpdateArticleHNStatus 方法、scanArticle/scanArticleWithBlog 函数更新 |
| `internal/hn/client.go` | 新建，封装 Algolia API 调用、URL 匹配逻辑、网络错误重试 |
| `internal/scanner/scanner.go` | 新增 Phase 6 HN 搜索、ScanResult 新增统计字段 |
| `internal/cli/commands/hn.go` | 新建，hn 命令组和 sync 子命令 |
| `internal/cli/commands/root.go` | 注册 hn 命令 |
| `internal/cli/output/json.go` | ArticleJSONOutput 新增 hn_url/hn_status 字段 |
| `assets/templates/partials/article-items.gohtml` | 文章卡片新增 HN 按钮 |

---

## Task 1: Model 更新 - HNStatus 枚举和结构体字段

**Files:**
- Modify: `internal/model/model.go`

- [ ] **Step 1: 在 model.go 末尾添加 HNStatus 枚举定义**

在文件末尾（`DefaultPageSize` 常量之后）添加：

```go
// HNStatus 定义 HN 链接搜索状态枚举
type HNStatus string

const (
	HNStatusNotSearch HNStatus = "not_searched" // 未搜索
	HNStatusExact     HNStatus = "found_exact"  // 精确匹配
	HNStatusFuzzy     HNStatus = "found_fuzzy"  // 模糊匹配
	HNStatusNotFound  HNStatus = "not_found"    // 搜索无结果
	HNStatusFailed    HNStatus = "failed"       // 搜索失败
)
```

- [ ] **Step 2: 在 Article 结构体新增字段**

在 `Article` 结构体中，`HasNote` 字段后添加：

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
	HasNote        bool // 文章备注状态
	HNURL          string   // Hacker News 讨论链接
	HNStatus       HNStatus // HN 链接搜索状态
}
```

- [ ] **Step 3: 在 ArticleWithBlog 结构体新增字段**

在 `ArticleWithBlog` 结构体中，`HasNote` 字段后添加：

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
	HasNote        bool // 文章备注状态
	HNURL          string   // Hacker News 讨论链接
	HNStatus       HNStatus // HN 链接搜索状态
}
```

- [ ] **Step 4: 运行测试确认编译通过**

运行: `go build ./...`
预期: 编译成功，无错误

- [ ] **Step 5: Commit**

```bash
git add internal/model/model.go
git commit -m "feat(model): 添加 HNStatus 枚举和 Article 结构体字段"
```

---

## Task 2: Database Migration - 新增 hn_url 和 hn_status 列

**Files:**
- Modify: `internal/storage/database.go`

- [ ] **Step 1: 在 ensureMigrations 函数末尾添加 HN 列 migration**

在 `ensureMigrations()` 函数的 `return nil` 之前添加：

```go
// Add hn_url column if it doesn't exist
if !db.columnExists("articles", "hn_url") {
	if _, err := db.conn.Exec(`ALTER TABLE articles ADD COLUMN hn_url TEXT`); err != nil {
		return fmt.Errorf("failed to add hn_url column: %w", err)
	}
}

// Add hn_status column if it doesn't exist
if !db.columnExists("articles", "hn_status") {
	if _, err := db.conn.Exec(`ALTER TABLE articles ADD COLUMN hn_status TEXT DEFAULT 'not_searched'`); err != nil {
		return fmt.Errorf("failed to add hn_status column: %w", err)
	}
}
```

- [ ] **Step 2: 更新 validTableNames 添加 articles 白名单确认**

确认 `validTableNames` 已包含 `"articles": true`（已存在，无需修改）

- [ ] **Step 3: 运行数据库测试确认 migration 正常**

运行: `go test ./internal/storage/... -v`
预期: 测试通过，migration 在新建数据库时执行

- [ ] **Step 4: Commit**

```bash
git add internal/storage/database.go
git commit -m "feat(storage): 添加 hn_url 和 hn_status 列 migration"
```

---

## Task 3: Database 方法 - scanArticle 函数更新

**Files:**
- Modify: `internal/storage/database.go`

- [ ] **Step 1: 更新 scanArticle 函数扫描 hn_url 和 hn_status**

找到 `scanArticle` 函数（约第 1075 行），修改为：

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
		hasNote       bool
		hnURL         sql.NullString
		hnStatus      sql.NullString
	)
	if err := scanner.Scan(&id, &blogID, &title, &url, &thumbnailURL, &publishedDate, &discovered, &isRead, &hasNote, &hnURL, &hnStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	article := &model.Article{
		ID:           id,
		BlogID:       blogID,
		Title:        title,
		URL:          url,
		ThumbnailURL: thumbnailURL.String,
		IsRead:       isRead,
		HasNote:      hasNote,
		HNURL:        hnURL.String,
		HNStatus:     model.HNStatus(hnStatus.String),
	}
	if publishedDate.Valid {
		if parsed, err := parseTime(publishedDate.String); err == nil {
			article.PublishedDate = &parsed
		}
	}
	if discovered.Valid {
		if parsed, err := parseTime(discovered.String); err == nil {
			article.DiscoveredDate = &parsed
		}
	}

	return article, nil
}
```

- [ ] **Step 2: 更新 scanArticleWithBlog 函数**

找到 `scanArticleWithBlog` 函数（约第 1143 行），修改为：

```go
func scanArticleWithBlog(scanner interface{ Scan(dest ...any) error }) (*model.ArticleWithBlog, error) {
	var (
		id            int64
		blogID        int64
		title         string
		url           string
		thumbnailURL  sql.NullString
		publishedDate sql.NullString
		discovered    sql.NullString
		isRead        bool
		hasNote       bool
		blogName      string
		blogURL       string
		hnURL         sql.NullString
		hnStatus      sql.NullString
	)
	if err := scanner.Scan(&id, &blogID, &title, &url, &thumbnailURL, &publishedDate, &discovered, &isRead, &hasNote, &blogName, &blogURL, &hnURL, &hnStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	article := &model.ArticleWithBlog{
		ID:           id,
		BlogID:       blogID,
		Title:        title,
		URL:          url,
		ThumbnailURL: thumbnailURL.String,
		IsRead:       isRead,
		BlogName:     blogName,
		BlogURL:      blogURL,
		HasNote:      hasNote,
		HNURL:        hnURL.String,
		HNStatus:     model.HNStatus(hnStatus.String),
	}
	if publishedDate.Valid {
		if parsed, err := parseTime(publishedDate.String); err == nil {
			article.PublishedDate = &parsed
		}
	}
	if discovered.Valid {
		if parsed, err := parseTime(discovered.String); err == nil {
			article.DiscoveredDate = &parsed
		}
	}

	return article, nil
}
```

- [ ] **Step 3: 更新 scanArticleWithBlogAndCount 函数**

找到 `scanArticleWithBlogAndCount` 函数（约第 1189 行），修改为：

```go
func scanArticleWithBlogAndCount(scanner interface{ Scan(dest ...any) error }) (*model.ArticleWithBlog, int, error) {
	var (
		id            int64
		blogID        int64
		title         string
		url           string
		thumbnailURL  sql.NullString
		publishedDate sql.NullString
		discovered    sql.NullString
		isRead        bool
		hasNote       bool
		blogName      string
		blogURL       string
		totalCount    int
		hnURL         sql.NullString
		hnStatus      sql.NullString
	)
	if err := scanner.Scan(&id, &blogID, &title, &url, &thumbnailURL, &publishedDate, &discovered, &isRead, &hasNote, &blogName, &blogURL, &totalCount, &hnURL, &hnStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, nil
		}
		return nil, 0, err
	}

	article := &model.ArticleWithBlog{
		ID:           id,
		BlogID:       blogID,
		Title:        title,
		URL:          url,
		ThumbnailURL: thumbnailURL.String,
		IsRead:       isRead,
		HasNote:      hasNote,
		BlogName:     blogName,
		BlogURL:      blogURL,
		HNURL:        hnURL.String,
		HNStatus:     model.HNStatus(hnStatus.String),
	}
	if publishedDate.Valid {
		if parsed, err := parseTime(publishedDate.String); err == nil {
			article.PublishedDate = &parsed
		}
	}
	if discovered.Valid {
		if parsed, err := parseTime(discovered.String); err == nil {
			article.DiscoveredDate = &parsed
		}
	}

	return article, totalCount, nil
}
```

- [ ] **Step 4: 更新所有查询语句添加 hn_url 和 hn_status 列**

需要更新的查询语句位置：
- `ListArticles` (约第 546 行): `SELECT id, blog_id, title, url, thumbnail_url, published_date, discovered_date, is_read, has_note, hn_url, hn_status`
- `ListArticlesByReadStatus` (约第 581 行): 同上
- `ListArticlesWithBlog` (约第 612 行): `SELECT a.id, a.blog_id, a.title, a.url, a.thumbnail_url, a.published_date, a.discovered_date, a.is_read, a.has_note, a.hn_url, a.hn_status, b.name, b.url`
- `SearchArticles` (约第 650 行): 同上
- `GetArticleByID` (约第 817 行): `SELECT id, blog_id, title, url, thumbnail_url, published_date, discovered_date, is_read, has_note, hn_url, hn_status`
- `ListArticlesWithFilters` (约第 1328 行): `SELECT a.id, a.blog_id, a.title, a.url, a.thumbnail_url, a.published_date, a.discovered_date, a.is_read, a.has_note, a.hn_url, a.hn_status, b.name, b.url`

- [ ] **Step 5: 运行测试确认编译通过**

运行: `go build ./...`
预期: 编译成功

- [ ] **Step 6: Commit**

```bash
git add internal/storage/database.go
git commit -m "feat(storage): 更新 scanArticle 函数扫描 hn_url 和 hn_status"
```

---

## Task 4: Database 方法 - UpdateArticleHNStatus 和查询方法

**Files:**
- Modify: `internal/storage/database.go`

- [ ] **Step 1: 添加 UpdateArticleHNStatus 方法**

在文件末尾（`UpdateBlogCategory` 函数之后）添加：

```go
// UpdateArticleHNStatus 更新文章的 HN 链接和搜索状态
func (db *Database) UpdateArticleHNStatus(id int64, hnURL string, status model.HNStatus) error {
	_, err := db.conn.Exec(
		`UPDATE articles SET hn_url = ?, hn_status = ? WHERE id = ?`,
		nullIfEmpty(hnURL),
		string(status),
		id,
	)
	return err
}
```

- [ ] **Step 2: 添加 GetArticlesForHNSync 方法**

在 `UpdateArticleHNStatus` 后添加：

```go
// ArticleForHNSync 用于 HN 同步的文章数据
type ArticleForHNSync struct {
	ID  int64
	URL string
}

// GetArticlesForHNSync 返回需要 HN 搜索的文章列表
// mode: "not_searched" - 仅未搜索的文章
// mode: "failed" - 仅失败的文章
// mode: "all" - 所有文章（重新搜索）
// blogName: 可选，筛选指定博客的文章
func (db *Database) GetArticlesForHNSync(mode string, blogName string, limit int) ([]ArticleForHNSync, error) {
	var query string
	var args []interface{}

	// 基础查询
	query = `SELECT a.id, a.url FROM articles a`

	// 博客筛选
	if blogName != "" {
		query += ` INNER JOIN blogs b ON a.blog_id = b.id WHERE b.name = ?`
		args = append(args, blogName)
		query += ` AND`
	} else {
		query += ` WHERE`
	}

	// 状态筛选
	switch mode {
	case "not_searched":
		query += ` a.hn_status = 'not_searched'`
	case "failed":
		query += ` a.hn_status = 'failed'`
	case "all":
		query += ` 1=1` // 所有文章
	default:
		query += ` a.hn_status = 'not_searched'` // 默认未搜索
	}

	query += ` ORDER BY a.id DESC`

	// 数量限制
	if limit > 0 {
		query += fmt.Sprintf(` LIMIT %d`, limit)
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []ArticleForHNSync
	for rows.Next() {
		var a ArticleForHNSync
		if err := rows.Scan(&a.ID, &a.URL); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, rows.Err()
}
```

- [ ] **Step 3: 运行测试确认编译通过**

运行: `go build ./...`
预期: 编译成功

- [ ] **Step 4: Commit**

```bash
git add internal/storage/database.go
git commit -m "feat(storage): 添加 UpdateArticleHNStatus 和 GetArticlesForHNSync 方法"
```

---

## Task 5: HN 模块 - 新建 internal/hn/client.go

**Files:**
- Create: `internal/hn/client.go`

- [ ] **Step 1: 创建 hn 目录**

```bash
mkdir -p internal/hn
```

- [ ] **Step 2: 创建 client.go 文件**

写入完整代码：

```go
// ABOUTME: Hacker News 讨论链接搜索模块
// ABOUTME: 使用 Algolia API 搜索文章对应的 HN 讨论，支持 URL 匹配和网络错误重试
package hn

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
)

// SearchResult 表示 Algolia API 返回的单个结果
type SearchResult struct {
	ObjectID    string `json:"objectID"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Points      int    `json:"points"`
	NumComments int    `json:"num_comments"`
}

// AlgoliaResponse 表示 Algolia API 响应
type AlgoliaResponse struct {
	Hits []SearchResult `json:"hits"`
}

// MatchResult 表示 URL 匹配结果
type MatchResult struct {
	HNURL  string        // HN 讨论链接
	Status model.HNStatus // 搜索状态
}

const (
	algoliaAPIURL = "https://hn.algolia.com/api/v1/search"
	maxRetries    = 3
	retryDelay    = 500 * time.Millisecond
	httpTimeout   = 10 * time.Second
)

// isRateLimitError 检查是否为限流错误
func isRateLimitError(statusCode int) bool {
	return statusCode == 429
}

// SearchHNDiscussion 搜索指定 URL 的 HN 讨论
// 返回匹配结果和可能的错误
func SearchHNDiscussion(ctx context.Context, articleURL string) (MatchResult, error) {
	log.Printf("[HN] 开始搜索文章 URL: %s", articleURL)

	// 构建请求 URL
	queryURL := fmt.Sprintf("%s?query=%s", algoliaAPIURL, url.QueryEscape(articleURL))

	var resp *http.Response
	var err error

	// 重试逻辑
	for retry := 0; retry < maxRetries; retry++ {
		if retry > 0 {
			log.Printf("[HN] 重试第 %d 次，等待 %v", retry, retryDelay)
			time.Sleep(retryDelay)
		}

		// 创建请求
		req, reqErr := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
		if reqErr != nil {
			return MatchResult{}, fmt.Errorf("创建请求失败: %w", reqErr)
		}

		// 发送请求
		client := &http.Client{Timeout: httpTimeout}
		resp, err = client.Do(req)

		if err != nil {
			log.Printf("[HN] 请求失败: %v", err)
			continue // 网络错误，重试
		}

		// 检查限流
		if isRateLimitError(resp.StatusCode) {
			resp.Body.Close()
			return MatchResult{}, fmt.Errorf("HN API 限流 (429)")
		}

		// 成功响应
		if resp.StatusCode == 200 {
			break
		}

		// 其他错误状态码
		resp.Body.Close()
		err = fmt.Errorf("HN API 返回状态码: %d", resp.StatusCode)
		log.Printf("[HN] API 错误: %v", err)
	}

	if err != nil {
		return MatchResult{}, fmt.Errorf("HN 搜索失败: %w", err)
	}
	if resp == nil {
		return MatchResult{}, fmt.Errorf("HN 搜索失败: 无响应")
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return MatchResult{}, fmt.Errorf("读取响应失败: %w", err)
	}

	var algoliaResp AlgoliaResponse
	if err := json.Unmarshal(body, &algoliaResp); err != nil {
		return MatchResult{}, fmt.Errorf("解析响应失败: %w", err)
	}

	log.Printf("[HN] 收到 %d 个搜索结果", len(algoliaResp.Hits))

	// 匹配逻辑
	if len(algoliaResp.Hits) == 0 {
		log.Printf("[HN] 无搜索结果，状态: not_found")
		return MatchResult{HNURL: "", Status: model.HNStatusNotFound}, nil
	}

	// 寻找精确匹配
	for _, hit := range algoliaResp.Hits {
		if hit.URL == articleURL {
			hnURL := fmt.Sprintf("https://news.ycombinator.com/item?id=%s", hit.ObjectID)
			log.Printf("[HN] 找到精确匹配，HN ID: %s", hit.ObjectID)
			return MatchResult{HNURL: hnURL, Status: model.HNStatusExact}, nil
		}
	}

	// 无精确匹配，选择最佳模糊匹配
	bestHit := selectBestFuzzyMatch(algoliaResp.Hits, articleURL)
	hnURL := fmt.Sprintf("https://news.ycombinator.com/item?id=%s", bestHit.ObjectID)
	log.Printf("[HN] 使用模糊匹配，HN ID: %s, 原URL: %s", bestHit.ObjectID, bestHit.URL)
	return MatchResult{HNURL: hnURL, Status: model.HNStatusFuzzy}, nil
}

// selectBestFuzzyMatch 选择最佳模糊匹配结果
// 优先选择 URL 相似度最高的，其次按点赞数排序
func selectBestFuzzyMatch(hits []SearchResult, articleURL string) SearchResult {
	// 提取域名用于相似度判断
	articleDomain := extractDomain(articleURL)

	// 计算每个结果的相似度分数
	bestScore := -1
	bestHit := hits[0] // 默认第一个

	for _, hit := range hits {
		score := calculateSimilarityScore(hit.URL, articleURL, articleDomain, hit.Points)
		if score > bestScore {
			bestScore = score
			bestHit = hit
		}
	}

	return bestHit
}

// extractDomain 从 URL 提取域名
func extractDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		// 尝试添加协议
		parsed, err = url.Parse("https://" + rawURL)
		if err != nil {
			return ""
		}
	}
	return parsed.Host
}

// calculateSimilarityScore 计算 URL 相似度分数
// 分数越高表示匹配度越好
func calculateSimilarityScore(hitURL, articleURL, articleDomain, points int) int {
	score := 0

	hitDomain := extractDomain(hitURL)

	// 域名匹配加分（最高优先级）
	if hitDomain == articleDomain {
		score += 100
	}

	// URL 前缀匹配加分
	if strings.HasPrefix(articleURL, hitURL) || strings.HasPrefix(hitURL, articleURL) {
		score += 50
	}

	// URL 包含关系加分
	if strings.Contains(hitURL, articleURL) || strings.Contains(articleURL, hitURL) {
		score += 30
	}

	// 点赞数作为次要排序（归一化到 0-20）
	if points > 0 {
		pointScore := points / 10
		if pointScore > 20 {
			pointScore = 20
		}
		score += pointScore
	}

	return score
}
```

- [ ] **Step 3: 运行测试确认编译通过**

运行: `go build ./...`
预期: 编译成功

- [ ] **Step 4: Commit**

```bash
git add internal/hn/client.go
git commit -m "feat(hn): 新建 HN 讨论链接搜索模块"
```

---

## Task 6: Scanner 集成 - 新增 Phase 6

**Files:**
- Modify: `internal/scanner/scanner.go`

- [ ] **Step 1: 添加 hn 包导入**

在 import 块中添加：

```go
import (
	// 现有导入...
	"log"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/hn"
)
```

- [ ] **Step 2: 更新 ScanResult 结构体添加 HN 统计字段**

找到 `ScanResult` 结构体（约第 18 行），修改为：

```go
type ScanResult struct {
	BlogName     string
	NewArticles  int
	SkippedCount int // 因重复 URL 跳过的文章数量
	TotalFound   int
	Source       string
	Error        string
	HNSearched   int // 尝试搜索 HN 的文章数
	HNFound      int // 找到 HN 链接的文章数（精确+模糊）
	HNFailed     int // HN 搜索失败数
}
```

- [ ] **Step 3: 在 ScanBlog 函数末尾添加 Phase 6**

在 `ScanBlog` 函数中，Phase 5 之后、`UpdateBlogLastScanned` 之前添加：

```go
// Phase 6: Search HN discussion for new articles
hnSearched := 0
hnFound := 0
hnFailed := 0

if len(newArticles) > 0 {
	log.Printf("[Scanner] 博客 '%s': 开始搜索 %d 篇新文章的 HN 讨论", blog.Name, len(newArticles))
	for _, article := range newArticles {
		hnSearched++
		match, err := hn.SearchHNDiscussion(ctx, article.URL)
		if err != nil {
			log.Printf("[Scanner] HN 搜索失败，文章 ID %d: %v", article.ID, err)
			hnFailed++
			_ = db.UpdateArticleHNStatus(article.ID, "", model.HNStatusFailed)
			continue
		}

		if match.Status == model.HNStatusExact || match.Status == model.HNStatusFuzzy {
			hnFound++
		}

		_ = db.UpdateArticleHNStatus(article.ID, match.HNURL, match.Status)
	}
	log.Printf("[Scanner] 博客 '%s': HN 搜索完成，找到 %d/%d，失败 %d", blog.Name, hnFound, hnSearched, hnFailed)
}
```

- [ ] **Step 4: 更新 ScanResult 返回值包含 HN 统计**

找到 `return ScanResult{...}` 语句，添加 HN 字段：

```go
return ScanResult{
	BlogName:     blog.Name,
	NewArticles:  newCount,
	SkippedCount: skippedCount,
	TotalFound:   len(seenURLs),
	Source:       source,
	Error:        errText,
	HNSearched:   hnSearched,
	HNFound:      hnFound,
	HNFailed:     hnFailed,
}
```

- [ ] **Step 5: 运行测试确认编译通过**

运行: `go build ./...`
预期: 编译成功

- [ ] **Step 6: Commit**

```bash
git add internal/scanner/scanner.go
git commit -m "feat(scanner): 新增 Phase 6 HN 讨论链接搜索"
```

---

## Task 7: CLI 命令 - 新建 hn.go

**Files:**
- Create: `internal/cli/commands/hn.go`

- [ ] **Step 1: 创建 hn.go 文件**

写入完整代码：

```go
// ABOUTME: hn 子命令定义
// ABOUTME: 提供 HN 讨论链接同步 CLI 命令
package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/cli/flags"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/hn"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
	"github.com/spf13/cobra"
)

// NewHNCmd 创建 hn 命令组
func NewHNCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hn",
		Short: "Hacker News 讨论链接管理",
		Long: `Hacker News 讨论链接管理命令。

提供同步命令为文章搜索对应的 HN 讨论。`,
	}
	cmd.AddCommand(NewHNSyncCmd())
	return cmd
}

// NewHNSyncCmd 创建 sync 子命令
func NewHNSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "同步 HN 讨论链接",
		Long: `为文章搜索 Hacker News 上的讨论链接。

默认只搜索未搜索的文章（hn_status='not_searched'）。

参数：
  --all          强制重新搜索所有文章（覆盖已有状态）
  --failed       仅重新搜索失败的文章
  --blog <name>  仅搜索指定博客的文章
  --limit <n>    限制搜索数量（默认无限制）

示例：
  blogwatcher hn sync
  blogwatcher hn sync --all
  blogwatcher hn sync --failed
  blogwatcher hn sync --blog "Tech Blog"
  blogwatcher hn sync --limit 50`,
		Run: runHNSync,
	}

	cmd.Flags().Bool("all", false, "重新搜索所有文章")
	cmd.Flags().Bool("failed", false, "仅重新搜索失败的文章")
	cmd.Flags().String("blog", "", "指定博客名称")
	cmd.Flags().Int("limit", 0, "搜索数量限制（0 表示无限制）")

	cmd.MarkFlagsMutuallyExclusive("all", "failed")

	return cmd
}

// runHNSync 执行 HN 同步命令
func runHNSync(cmd *cobra.Command, args []string) {
	// 获取数据库路径
	dbPath := flags.DBPath()
	if dbPath == "" {
		var err error
		dbPath, err = storage.DefaultDBPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取默认数据库路径失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 打开数据库
	db, err := storage.OpenDatabase(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// 解析参数
	all, _ := cmd.Flags().GetBool("all")
	failed, _ := cmd.Flags().GetBool("failed")
	blogName, _ := cmd.Flags().GetString("blog")
	limit, _ := cmd.Flags().GetInt("limit")

	// 确定搜索模式
	mode := "not_searched"
	if all {
		mode = "all"
	} else if failed {
		mode = "failed"
	}

	// 获取需要搜索的文章
	articles, err := db.GetArticlesForHNSync(mode, blogName, limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查询文章失败: %v\n", err)
		os.Exit(1)
	}

	if len(articles) == 0 {
		fmt.Println("没有需要搜索的文章")
		return
	}

	fmt.Printf("开始搜索 %d 篇文章的 HN 讨论...\n", len(articles))

	// 统计结果
	var searched, foundExact, foundFuzzy, notFound, failedCount int

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// 逐篇搜索
	for i, article := range articles {
		searched++
		fmt.Printf("[%d/%d] 搜索文章 ID %d...\n", i+1, len(articles), article.ID)

		match, err := hn.SearchHNDiscussion(ctx, article.URL)
		if err != nil {
			failedCount++
			fmt.Printf("  失败: %v\n", err)
			_ = db.UpdateArticleHNStatus(article.ID, "", model.HNStatusFailed)
			continue
		}

		switch match.Status {
		case model.HNStatusExact:
			foundExact++
			fmt.Printf("  精确匹配: %s\n", match.HNURL)
		case model.HNStatusFuzzy:
			foundFuzzy++
			fmt.Printf("  模糊匹配: %s\n", match.HNURL)
		case model.HNStatusNotFound:
			notFound++
			fmt.Printf("  未找到讨论\n")
		}

		_ = db.UpdateArticleHNStatus(article.ID, match.HNURL, match.Status)
	}

	// 输出统计
	fmt.Println("\nHN 链接同步完成")
	fmt.Printf("搜索: %d 篇文章\n", searched)
	fmt.Printf("找到: %d 篇（精确: %d, 模糊: %d）\n", foundExact+foundFuzzy, foundExact, foundFuzzy)
	fmt.Printf("未找到: %d 篇\n", notFound)
	fmt.Printf("失败: %d 篇\n", failedCount)
}
```

- [ ] **Step 2: 在 root.go 注册 hn 命令**

在 `init()` 函数中添加：

```go
// 添加 hn 子命令
rootCmd.AddCommand(NewHNCmd())
```

- [ ] **Step 3: 运行测试确认编译通过**

运行: `go build ./cmd/blogwatcher`
预期: 编译成功

- [ ] **Step 4: 测试 CLI 命令帮助**

运行: `./blogwatcher.exe hn --help`
预期: 显示 hn 命令帮助信息

- [ ] **Step 5: Commit**

```bash
git add internal/cli/commands/hn.go internal/cli/commands/root.go
git commit -m "feat(cli): 新增 hn 命令组和 sync 子命令"
```

---

## Task 8: JSON 输出更新 - 添加 hn_url/hn_status 字段

**Files:**
- Modify: `internal/cli/output/json.go`

- [ ] **Step 1: 更新 ArticleJSONOutput 结构体**

找到 `ArticleJSONOutput` 结构体，添加新字段：

```go
// ArticleJSONOutput 用于 JSON 输出的简化文章结构
type ArticleJSONOutput struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Blog      string `json:"blog"`
	Read      bool   `json:"read"`
	HasNote   bool   `json:"has_note"`
	Published string `json:"published,omitempty"`
	HNURL     string `json:"hn_url,omitempty"`
	HNStatus  string `json:"hn_status,omitempty"`
}
```

- [ ] **Step 2: 更新 FormatJSON 函数填充新字段**

找到 `FormatJSON` 函数中的 `data[i] = ArticleJSONOutput{...}` 部分，添加：

```go
data[i] = ArticleJSONOutput{
	ID:        article.ID,
	Title:     article.Title,
	URL:       article.URL,
	Blog:      article.BlogName,
	Read:      article.IsRead,
	HasNote:   article.HasNote,
	HNURL:     article.HNURL,
	HNStatus:  string(article.HNStatus),
}
```

注意：保留现有的时间格式化逻辑不变。

- [ ] **Step 3: 运行测试确认编译通过**

运行: `go build ./...`
预期: 编译成功

- [ ] **Step 4: Commit**

```bash
git add internal/cli/output/json.go
git commit -m "feat(output): JSON 输出添加 hn_url 和 hn_status 字段"
```

---

## Task 9: UI 更新 - 添加 HN 按钮

**Files:**
- Modify: `assets/templates/partials/article-items.gohtml`
- Modify: `assets/static/styles.css`

- [ ] **Step 1: 在 article-items.gohtml 添加 HN 按钮**

找到 `article-actions` div（约第 48-77 行），在备注按钮之后、Read/Unread 按钮之前添加：

```html
{{if .HNURL}}
<a href="{{.HNURL}}"
   target="_blank"
   rel="noopener noreferrer"
   class="action-btn{{if eq .HNStatus "found_fuzzy"}} action-btn-muted{{end}}"
   title="查看 HN 讨论{{if eq .HNStatus "found_fuzzy"}}（模糊匹配）{{end}}"
   onclick="event.stopPropagation();">
    <span class="action-btn-label">HN</span>
</a>
{{end}}
```

- [ ] **Step 2: 在 styles.css 添加 action-btn-muted 样式**

在 action-btn 相关样式之后添加：

```css
/* HN 模糊匹配按钮样式 - 更低调 */
.action-btn-muted {
  opacity: 0.7;
  border-color: #9ca3af;
}
.action-btn-muted:hover {
  opacity: 1;
}
```

- [ ] **Step 3: 运行测试确认编译通过**

运行: `go build ./...`
预期: 编译成功

- [ ] **Step 4: 启动服务器测试 UI**

运行: `./blogwatcher.exe serve`
预期: 服务器启动，访问 localhost:8080 检查文章卡片

- [ ] **Step 5: Commit**

```bash
git add assets/templates/partials/article-items.gohtml assets/static/styles.css
git commit -m "feat(ui): 文章卡片添加 HN 讨论链接按钮"
```

---

## Task 10: 集成测试和最终验收

**Files:**
- 无文件变更，仅测试

- [ ] **Step 1: 运行完整测试套件**

运行: `go test ./... -v`
预期: 所有测试通过

- [ ] **Step 2: 手动测试 CLI hn sync 命令**

运行: `./blogwatcher.exe hn sync --limit 5`
预期: 显示搜索进度和统计结果

- [ ] **Step 3: 手动测试 CLI article list JSON 输出**

运行: `./blogwatcher.exe article list --format json --limit 3`
预期: JSON 输出包含 hn_url 和 hn_status 字段

- [ ] **Step 4: 手动测试 UI HN 按钮**

运行: `./blogwatcher.exe serve`
访问: `localhost:8080`
预期: 有 HN 链接的文章显示 HN 按钮

- [ ] **Step 5: 最终提交（如有遗漏修改）**

```bash
git status
git add -A
git commit -m "feat: HN 讨论链接检测功能完成"
```