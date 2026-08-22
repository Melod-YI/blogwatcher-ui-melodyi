// ABOUTME: Blog output formatters for CLI blog list command
// ABOUTME: Provides table and JSON formatting for blog list command
package output

import (
	"encoding/json"
	"fmt"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
)

// BlogJSONOutput 用于 JSON 输出的简化博客结构
type BlogJSONOutput struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	FeedURL      string `json:"feed_url,omitempty"`
	Category     string `json:"category,omitempty"`
	ArticleCount int    `json:"article_count"`
	LastScanned  string `json:"last_scanned,omitempty"`
}

// FormatBlogJSON 将博客列表格式化为 JSON 输出
func FormatBlogJSON(blogs []storage.BlogWithCount) string {
	if len(blogs) == 0 {
		return "[]"
	}

	output := make([]BlogJSONOutput, len(blogs))
	for i, blog := range blogs {
		output[i] = BlogJSONOutput{
			ID:           blog.ID,
			Name:         blog.Name,
			URL:          blog.URL,
			FeedURL:      blog.FeedURL,
			Category:     blog.CategoryName,
			ArticleCount: blog.ArticleCount,
		}
		if blog.LastScanned != nil {
			output[i].LastScanned = blog.LastScanned.Format("2006-01-02 15:04")
		}
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf("JSON 序列化错误: %v", err)
	}

	return string(data)
}