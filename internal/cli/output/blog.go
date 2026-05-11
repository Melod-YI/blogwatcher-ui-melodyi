// ABOUTME: Blog output formatters for CLI blog list command
// ABOUTME: Provides table and JSON formatting for blog list command
package output

import (
	"encoding/json"
	"fmt"
	"strings"

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

// FormatBlogTable 将博客列表格式化为表格输出
func FormatBlogTable(blogs []storage.BlogWithCount) string {
	if len(blogs) == 0 {
		return "没有博客"
	}

	// 定义列宽
	idWidth := 8
	nameWidth := 25
	categoryWidth := 15
	countWidth := 10
	timeWidth := 18

	// 构建表头
	header := fmt.Sprintf("| %-*s | %-*s | %-*s | %-*s | %-*s |",
		idWidth, "ID",
		nameWidth, "Name",
		categoryWidth, "Category",
		countWidth, "Articles",
		timeWidth, "Last Scanned")

	// 构建分隔线
	separator := fmt.Sprintf("|-%s-|-%s-|-%s-|-%s-|-%s-|",
		strings.Repeat("-", idWidth),
		strings.Repeat("-", nameWidth),
		strings.Repeat("-", categoryWidth),
		strings.Repeat("-", countWidth),
		strings.Repeat("-", timeWidth))

	// 构建各行
	var rows []string
	rows = append(rows, header)
	rows = append(rows, separator)

	for _, blog := range blogs {
		name := truncate(blog.Name, nameWidth)
		category := truncate(blog.CategoryName, categoryWidth)
		if category == "" {
			category = "-"
		}
		lastScanned := "-"
		if blog.LastScanned != nil {
			lastScanned = formatTime(blog.LastScanned, nil)
		}

		row := fmt.Sprintf("| %-*d | %-*s | %-*s | %-*d | %-*s |",
			idWidth, blog.ID,
			nameWidth, name,
			categoryWidth, category,
			countWidth, blog.ArticleCount,
			timeWidth, lastScanned)
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
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