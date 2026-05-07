// ABOUTME: 简洁格式输出器
// ABOUTME: 将文章列表格式化为简洁的单行输出，用于 CLI article list --format simple 输出
package output

import (
	"fmt"
	"time"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
)

// FormatSimple 将文章列表格式化为简洁输出
// 格式：[状态] Title (BlogName) - PublishedDate
// 每行一个文章
func FormatSimple(articles []model.ArticleWithBlog) string {
	if len(articles) == 0 {
		return "没有找到文章"
	}

	var lines []string
	for _, article := range articles {
		// 状态（中文）
		status := "[未读]"
		if article.IsRead {
			status = "[已读]"
		}

		// 时间
		published := formatSimpleTime(article.PublishedDate, article.DiscoveredDate)

		// 构建行
		line := fmt.Sprintf("%s %s (%s) - %s",
			status,
			article.Title,
			article.BlogName,
			published)

		lines = append(lines, line)
	}

	return fmt.Sprintf("%s\n", lines)
}

// formatSimpleTime 格式化时间为日期字符串
func formatSimpleTime(publishedDate, discoveredDate *time.Time) string {
	var t time.Time
	if publishedDate != nil {
		t = *publishedDate
	} else if discoveredDate != nil {
		t = *discoveredDate
	}

	if t.IsZero() {
		return "未知"
	}

	return t.Format("2006-01-02")
}