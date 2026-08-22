// ABOUTME: Tag output formatters
// ABOUTME: Provides table and JSON formatting for tag list command
package output

import (
	"encoding/json"
	"fmt"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
)

// TagJSONOutput 用于 JSON 输出的简化标签结构
type TagJSONOutput struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	ArticleCount int64  `json:"article_count"`
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
