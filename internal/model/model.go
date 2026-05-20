// ABOUTME: Defines Blog and Article data models matching the database schema.
// ABOUTME: These structs represent the core entities for tracking blogs and their articles.
package model

import "time"

// HNStatus represents the status of Hacker News discussion detection.
type HNStatus string

const (
	HNStatusNone     HNStatus = ""        // 未检测或无讨论
	HNStatusFound    HNStatus = "found"   // 已找到讨论链接
	HNStatusNotFound HNStatus = "notfound" // 已检测但未找到讨论
	HNStatusError    HNStatus = "error"   // 检测出错
)

// Category represents a blog category for organizing blogs.
type Category struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}

type Blog struct {
	ID             int64
	Name           string
	URL            string
	FeedURL        string
	ScrapeSelector string
	LastScanned    *time.Time
	CategoryID     *int64 // 分类ID（nullable，指向 categories.id）
}

type Article struct {
	ID             int64
	BlogID         int64
	Title          string
	URL            string
	ThumbnailURL   string
	PublishedDate  *time.Time
	DiscoveredDate *time.Time
	IsRead         bool
	HasNote        bool    // 文章备注状态
	HNURL          string  // Hacker News 讨论链接（空表示无讨论）
	HNStatus       HNStatus // HN 讨论检测状态
}

// ArticleWithBlog extends Article with blog metadata for display in article cards.
// Used when rendering article lists where blog name and favicon are needed.
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
	HasNote        bool    // 文章备注状态
	HNURL          string  // Hacker News 讨论链接（空表示无讨论）
	HNStatus       HNStatus // HN 讨论检测状态
}

// SearchOptions contains all filter parameters for article search.
// All fields are optional - nil/empty means no filter for that field.
type SearchOptions struct {
	SearchQuery string     // FTS5 search query (empty = skip FTS5)
	IsRead      *bool      // nil = all, true = read only, false = unread only
	BlogID      *int64     // nil = all blogs
	DateFrom    *time.Time // nil = no lower bound
	DateTo      *time.Time // nil = no upper bound
	Limit       int        // 0 = use default (20)
	Offset      int        // 0 = start from beginning
}

// DefaultPageSize is the default number of articles per page.
const DefaultPageSize = 20