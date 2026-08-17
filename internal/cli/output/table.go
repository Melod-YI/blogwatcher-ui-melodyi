// ABOUTME: 表格格式输出器
// ABOUTME: 将文章列表格式化为表格形式，用于 CLI article list 命令默认输出
package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
)

// FormatTable 将文章列表格式化为表格输出
// 表头：ID、Title、Blog、Status、Fav、Tags、Published
// 标题截断到 50 字符，状态使用中文显示，标签逗号拼接后截断
// 最后显示分页信息
func FormatTable(articles []model.ArticleWithBlog, meta PaginationMeta) string {
	if len(articles) == 0 {
		return "没有找到文章"
	}

	// 定义列宽
	idWidth := 8
	titleWidth := 50
	blogWidth := 20
	statusWidth := 8
	favWidth := 5
	tagWidth := 18
	timeWidth := 20

	// 构建表头
	header := fmt.Sprintf("| %-*s | %-*s | %-*s | %-*s | %-*s | %-*s | %-*s |",
		idWidth, "ID",
		titleWidth, "Title",
		blogWidth, "Blog",
		statusWidth, "Status",
		favWidth, "Fav",
		tagWidth, "Tags",
		timeWidth, "Published")

	// 构建分隔线
	separator := fmt.Sprintf("|-%s-|-%s-|-%s-|-%s-|-%s-|-%s-|-%s-|",
		strings.Repeat("-", idWidth),
		strings.Repeat("-", titleWidth),
		strings.Repeat("-", blogWidth),
		strings.Repeat("-", statusWidth),
		strings.Repeat("-", favWidth),
		strings.Repeat("-", tagWidth),
		strings.Repeat("-", timeWidth))

	// 构建各行
	var rows []string
	rows = append(rows, header)
	rows = append(rows, separator)

	for _, article := range articles {
		// 截断标题
		title := truncate(article.Title, titleWidth)

		// 截断博客名称
		blogName := truncate(article.BlogName, blogWidth)

		// 状态（中文）
		status := "未读"
		if article.IsRead {
			status = "已读"
		}

		// 收藏标记
		fav := ""
		if article.IsFavorited {
			fav = "★"
		}

		// 标签（逗号拼接后截断）
		tags := ""
		if len(article.Tags) > 0 {
			names := make([]string, len(article.Tags))
			for i, tag := range article.Tags {
				names[i] = tag.Name
			}
			tags = truncate(strings.Join(names, ","), tagWidth)
		}

		// 时间（相对时间或日期）
		published := formatTime(article.PublishedDate, article.DiscoveredDate)

		row := fmt.Sprintf("| %-*d | %-*s | %-*s | %-*s | %-*s | %-*s | %-*s |",
			idWidth, article.ID,
			titleWidth, title,
			blogWidth, blogName,
			statusWidth, status,
			favWidth, fav,
			tagWidth, tags,
			timeWidth, published)

		rows = append(rows, row)
	}

	// 添加分页信息
	rows = append(rows, "")
	rows = append(rows, formatPaginationFooter(meta))

	return strings.Join(rows, "\n")
}

// truncate 截断字符串到指定长度
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// formatTime 格式化时间为相对时间或日期字符串
func formatTime(publishedDate, discoveredDate *time.Time) string {
	var t time.Time
	if publishedDate != nil {
		t = *publishedDate
	} else if discoveredDate != nil {
		t = *discoveredDate
	} else {
		return "未知"
	}

	// 计算相对时间
	now := time.Now()
	diff := now.Sub(t)

	// 小于 1 分钟
	if diff < time.Minute {
		return "刚刚"
	}

	// 小于 1 小时
	if diff < time.Hour {
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%d 分钟前", minutes)
	}

	// 小于 24 小时
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%d 小时前", hours)
	}

	// 小于 7 天
	if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d 天前", days)
	}

	// 超过 7 天，显示日期
	return t.Format("2006-01-02")
}
