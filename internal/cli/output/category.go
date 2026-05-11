// ABOUTME: Category output formatters
// ABOUTME: Provides table and JSON formatting for category list command
package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
)

// CategoryJSONOutput 用于 JSON 输出的简化分类结构
type CategoryJSONOutput struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	BlogCount int    `json:"blog_count"`
}

// FormatCategoryTable 将分类列表格式化为表格输出
func FormatCategoryTable(categories []storage.CategoryWithBlogCount) string {
	if len(categories) == 0 {
		return "没有分类"
	}

	// 定义列宽
	idWidth := 8
	nameWidth := 30
	countWidth := 12

	// 构建表头
	header := fmt.Sprintf("| %-*s | %-*s | %-*s |",
		idWidth, "ID",
		nameWidth, "Name",
		countWidth, "Blog Count")

	// 构建分隔线
	separator := fmt.Sprintf("|-%s-|-%s-|-%s-|",
		strings.Repeat("-", idWidth),
		strings.Repeat("-", nameWidth),
		strings.Repeat("-", countWidth))

	// 构建各行
	var rows []string
	rows = append(rows, header)
	rows = append(rows, separator)

	for _, cat := range categories {
		name := truncate(cat.Name, nameWidth)
		row := fmt.Sprintf("| %-*d | %-*s | %-*d |",
			idWidth, cat.ID,
			nameWidth, name,
			countWidth, cat.BlogCount)
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

// FormatCategoryJSON 将分类列表格式化为 JSON 输出
func FormatCategoryJSON(categories []storage.CategoryWithBlogCount) string {
	if len(categories) == 0 {
		return "[]"
	}

	output := make([]CategoryJSONOutput, len(categories))
	for i, cat := range categories {
		output[i] = CategoryJSONOutput{
			ID:        cat.ID,
			Name:      cat.Name,
			BlogCount: cat.BlogCount,
		}
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf("JSON 序列化错误: %v", err)
	}

	return string(data)
}