// ABOUTME: Category output formatters
// ABOUTME: Provides table and JSON formatting for category list command
package output

import (
	"encoding/json"
	"fmt"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
)

// CategoryJSONOutput 用于 JSON 输出的简化分类结构
type CategoryJSONOutput struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	BlogCount int    `json:"blog_count"`
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