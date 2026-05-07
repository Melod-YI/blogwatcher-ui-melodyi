// ABOUTME: JSON 格式输出器
// ABOUTME: 将文章列表格式化为 JSON 数组，用于 CLI article list --format json 输出
package output

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
)

// ArticleJSONOutput 用于 JSON 输出的简化文章结构
type ArticleJSONOutput struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Blog      string `json:"blog"`
	Read      bool   `json:"read"`
	Published string `json:"published,omitempty"`
}

// FormatJSON 将文章列表格式化为 JSON 输出
// 使用简化结构，输出格式化的 JSON 字符串
func FormatJSON(articles []model.ArticleWithBlog) string {
	if len(articles) == 0 {
		return "[]"
	}

	// 转换为 JSON 输出结构
	output := make([]ArticleJSONOutput, len(articles))
	for i, article := range articles {
		output[i] = ArticleJSONOutput{
			ID:    article.ID,
			Title: article.Title,
			URL:   article.URL,
			Blog:  article.BlogName,
			Read:  article.IsRead,
		}

		// 格式化时间
		var t time.Time
		if article.PublishedDate != nil {
			t = *article.PublishedDate
		} else if article.DiscoveredDate != nil {
			t = *article.DiscoveredDate
		}

		if !t.IsZero() {
			output[i].Published = t.Format("2006-01-02")
		}
	}

	// 序列化为 JSON
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf("JSON 序列化错误: %v", err)
	}

	return string(data)
}