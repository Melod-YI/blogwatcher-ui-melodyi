// ABOUTME: 简洁格式输出器
// ABOUTME: 将文章列表格式化为简洁的单行输出，用于 CLI article list --format simple 输出
package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
)

// FormatSimple 将文章列表格式化为简洁输出
// 格式：[状态] Title (BlogName) - PublishedDate
// 每行一个文章，最后显示分页信息
func FormatSimple(articles []model.ArticleWithBlog, meta PaginationMeta) string {
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

		// 收藏标记
		favMark := ""
		if article.IsFavorited {
			favMark = " ★"
		}

		// 标签后缀（#tag1 #tag2）
		tagMark := ""
		if len(article.Tags) > 0 {
			parts := make([]string, len(article.Tags))
			for i, tag := range article.Tags {
				parts[i] = "#" + tag.Name
			}
			tagMark = " " + strings.Join(parts, " ")
		}

		// 构建行
		line := fmt.Sprintf("%s %s%s%s (%s) - %s",
			status,
			article.Title,
			favMark,
			tagMark,
			article.BlogName,
			published)

		lines = append(lines, line)
	}

	// 添加分页信息
	lines = append(lines, "")
	lines = append(lines, formatPaginationFooter(meta))

	return strings.Join(lines, "\n")
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
