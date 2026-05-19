---
name: hn-discussion-link
description: Hacker News 讨论链接检测功能设计文档
---

# Hacker News 讨论链接检测功能设计

## 概述

为博客文章自动检测并关联 Hacker News 上的讨论链接，让用户能快速查看相关讨论。

## 核心需求

| 方面 | 决定 |
|------|------|
| 多结果处理 | URL 匹配度排序，精确匹配优先 |
| 状态记录 | 使用状态枚举记录完整搜索状态 |
| 同步时机 | 在 sync 阶段同步调用 |
| UI 位置 | 操作按钮区域，类似备注按钮 |
| CLI 输出 | 只在 JSON 格式中添加字段 |
| CLI 命令 | 提供单独命令手动触发 |
| API 处理 | 网络错误重试，限流记录跳过 |

---

## 第一部分：数据库设计

### 状态枚举

`hn_status` 字段使用字符串枚举：

| 状态值 | 说明 |
|--------|------|
| `not_searched` | 未搜索（默认状态，新文章初始值） |
| `found_exact` | 找到精确匹配的 HN 链接 |
| `found_fuzzy` | 找到模糊匹配的 HN 链接 |
| `not_found` | 搜索过但未找到讨论 |
| `failed` | 搜索失败（网络错误、限流等） |

### 新增列

在 `articles` 表新增两列：

| 列名 | 类型 | 说明 |
|------|------|------|
| `hn_url` | TEXT | Hacker News 讨论链接（仅当 `found_exact` 或 `found_fuzzy` 时有值） |
| `hn_status` | TEXT | 搜索状态枚举，默认 `not_searched` |

### Migration 实现

在 `database.go` 的 `ensureMigrations()` 中添加（参考现有 `columnExists` 方法签名）：

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

### Model 更新

在 `model.go` 新增枚举类型和结构体字段：

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

// Article 结构体新增字段
HNURL    string    // Hacker News 讨论链接
HNStatus HNStatus  // 搜索状态

// ArticleWithBlog 结构体同样新增这两个字段
```

---

## 第二部分：HN 模块设计

### 新建 `internal/hn/client.go`

```go
package hn

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
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

// SearchHNDiscussion 搜索指定 URL 的 HN 讨论
func SearchHNDiscussion(ctx context.Context, articleURL string) (MatchResult, error) {
    // 1. 调用 Algolia API
    // 2. 解析响应
    // 3. URL 匹配逻辑
    // 4. 返回最佳匹配
}
```

### URL 匹配算法

1. 调用 `https://hn.algolia.com/api/v1/search?query=<url>`
2. 解析返回的 `hits` 数组
3. 匹配逻辑：
   - 首选精确匹配：`hit.URL == articleURL` → 返回 `found_exact`
   - 若无精确匹配，按相似度排序：
     - URL 前缀匹配（相同域名优先）
     - URL 包含关系（hit.URL 包含 articleURL 或反之）
     - 按点赞数 `points` 作为次要排序
   - 选择最佳匹配，标记 `found_fuzzy`
   - 无结果 → 返回 `not_found`

### 网络错误处理

- 最多重试 3 次，每次间隔 500ms
- HTTP 429（限流）→ 直接返回限流错误，不重试
- 重试仍失败 → 返回 error

---

## 第三部分：Scanner 集成

### ScanBlog 流程变更

在 `scanner.go` 新增 **Phase 6**：HN 链接搜索

```go
// Phase 6: Search HN discussion for new articles
for _, article := range newArticles {
    match, err := hn.SearchHNDiscussion(ctx, article.URL)
    if err != nil {
        log.Printf("HN search failed for article %d: %v", article.ID, err)
        db.UpdateArticleHNStatus(article.ID, "", model.HNStatusFailed)
        continue
    }
    db.UpdateArticleHNStatus(article.ID, match.HNURL, match.Status)
}
```

**Why:** 新文章保存后立即搜索，用户能同步完成时看到结果。
**How to apply:** 在 Phase 5 批量保存后，逐篇调用 HN 搜索。

### ScanResult 新增统计字段

```go
type ScanResult struct {
    // 现有字段...
    HNSearched  int  // 尝试搜索 HN 的文章数
    HNFound     int  // 找到 HN 链接的文章数
    HNFailed    int  // HN 搜索失败数
}
```

---

## 第四部分：CLI 命令设计

### 新增 `hn` 命令组

在 `internal/cli/commands/` 新建 `hn.go`：

```go
func NewHNCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "hn",
        Short: "Hacker News 讨论链接管理",
    }
    cmd.AddCommand(NewHNSyncCmd())
    return cmd
}
```

### `hn sync` 子命令

```bash
blogwatcher hn sync                # 为所有未搜索的文章同步
blogwatcher hn sync --all          # 强制重新搜索所有文章
blogwatcher hn sync --blog "Name"  # 仅搜索指定博客
blogwatcher hn sync --failed       # 仅重新搜索失败的文章
blogwatcher hn sync --limit 50     # 限制搜索数量
```

**输出示例**：

```
HN 链接同步完成
搜索: 100 篇文章
找到: 45 篇（精确: 30, 模糊: 15）
未找到: 50 篇
失败: 5 篇
```

### JSON 输出更新

在 `internal/cli/output/json.go` 添加字段：

```json
{
  "hn_url": "https://news.ycombinator.com/item?id=12345",
  "hn_status": "found_exact"
}
```

---

## 第五部分：UI 设计

### 文章卡片按钮

在 `article-items.gohtml` 的操作按钮区域新增：

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

**视觉区分**：
- `found_exact`: 普通按钮样式
- `found_fuzzy`: 使用 `action-btn-muted` 类，视觉上更低调（如灰色边框）

---

## 文件变更清单

| 文件 | 变更类型 |
|------|----------|
| `internal/model/model.go` | 新增 HNStatus 枚举、Article/ArticleWithBlog 字段 |
| `internal/storage/database.go` | Migration、UpdateArticleHNStatus 方法 |
| `internal/hn/client.go` | 新建，HN 搜索模块 |
| `internal/scanner/scanner.go` | 新增 Phase 6、ScanResult 字段 |
| `internal/cli/commands/hn.go` | 新建，hn 命令组 |
| `internal/cli/commands/root.go` | 注册 hn 命令 |
| `internal/cli/output/json.go` | 添加 hn_url/hn_status 字段 |
| `assets/templates/partials/article-items.gohtml` | 添加 HN 按钮 |