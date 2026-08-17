// ABOUTME: Tag output formatters
// ABOUTME: Provides table and JSON formatting for tag list command
package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
)

// TagJSONOutput 用于 JSON 输出的简化标签结构
type TagJSONOutput struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	ArticleCount int64  `json:"article_count"`
}

// FormatTagTable 将标签列表格式化为表格输出
func FormatTagTable(tags []model.Tag) string {
	if len(tags) == 0 {
		return "没有标签"
	}

	// 定义列宽
	idWidth := 8
	nameWidth := 30
	countWidth := 14

	// 构建表头
	header := fmt.Sprintf("| %-*s | %-*s | %-*s |",
		idWidth, "ID",
		nameWidth, "Name",
		countWidth, "Article Count")

	// 构建分隔线
	separator := fmt.Sprintf("|-%s-|-%s-|-%s-|",
		strings.Repeat("-", idWidth),
		strings.Repeat("-", nameWidth),
		strings.Repeat("-", countWidth))

	// 构建各行
	var rows []string
	rows = append(rows, header)
	rows = append(rows, separator)

	for _, t := range tags {
		name := truncate(t.Name, nameWidth)
		row := fmt.Sprintf("| %-*d | %-*s | %-*d |",
			idWidth, t.ID,
			nameWidth, name,
			countWidth, t.ArticleCount)
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

// FormatTagJSON 将标签列表格式化为 JSON 输出
func FormatTagJSON(tags []model.Tag) string {
	if len(tags) == 0 {
		return "[]"
	}

	output := make([]TagJSONOutput, len(tags))
	for i, t := range tags {
		output[i] = TagJSONOutput{
			ID:           t.ID,
			Name:         t.Name,
			ArticleCount: t.ArticleCount,
		}
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf("JSON 序列化错误: %v", err)
	}

	return string(data)
}
