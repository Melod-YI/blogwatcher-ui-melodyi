// ABOUTME: TSV 格式输出器（默认输出格式）
// ABOUTME: 表头(schema) + 数据行 + key=value 补充信息，面向 token 高效消费
package output

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
)

// cstZone 东八区，时间戳按 +08:00 输出，避免依赖系统时区
var cstZone = time.FixedZone("CST", 8*60*60)

// formatTSV 构建一个 TSV 块：首行为 schema 表头，随后每行一条数据，最后追加 key=value 补充行。
// rows 为空时返回空字符串（由调用方决定空态提示文案）。
func formatTSV(headers []string, rows [][]string, footer []string) string {
	if len(rows) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(strings.Join(headers, "\t"))
	b.WriteByte('\n')
	for _, r := range rows {
		b.WriteString(strings.Join(r, "\t"))
		b.WriteByte('\n')
	}
	for _, f := range footer {
		b.WriteString(f)
		b.WriteByte('\n')
	}
	return strings.TrimSuffix(b.String(), "\n")
}

// formatTSVFooter 由分页元信息生成 key=value 补充行。
// count 必有；has_more 必有；total 仅当总数大于当前返回条数时给出（否则与 count 重复）；
// offset 仅在 >0 时给出。
func formatTSVFooter(meta PaginationMeta) []string {
	footer := []string{fmt.Sprintf("count=%d", meta.Count)}
	if meta.Total > int64(meta.Count) {
		footer = append(footer, fmt.Sprintf("total=%d", meta.Total))
	}
	footer = append(footer, fmt.Sprintf("has_more=%t", meta.HasMore))
	if meta.Offset > 0 {
		footer = append(footer, fmt.Sprintf("offset=%d", meta.Offset))
	}
	return footer
}

// formatTSVTime 将文章时间格式化为 +08:00 的 ISO8601 字符串。
// 优先 PublishedDate，回退 DiscoveredDate；无可用时间返回空串。
func formatTSVTime(publishedDate, discoveredDate *time.Time) string {
	var t time.Time
	if publishedDate != nil {
		t = *publishedDate
	} else if discoveredDate != nil {
		t = *discoveredDate
	}
	if t.IsZero() {
		return ""
	}
	return t.In(cstZone).Format("2006-01-02T15:04:05+08:00")
}

// FormatTSV 将文章列表格式化为 TSV 输出。
// schema: id title blog status fav tags published
// 状态/收藏标记保留中文，标签逗号拼接，时间 +08:00 ISO8601。
func FormatTSV(articles []model.ArticleWithBlog, meta PaginationMeta) string {
	if len(articles) == 0 {
		return "没有找到文章"
	}

	headers := []string{"id", "title", "blog", "status", "fav", "tags", "published"}
	rows := make([][]string, 0, len(articles))
	for _, a := range articles {
		status := "未读"
		if a.IsRead {
			status = "已读"
		}

		fav := ""
		if a.IsFavorited {
			fav = "★"
		}

		tags := ""
		if len(a.Tags) > 0 {
			names := make([]string, len(a.Tags))
			for i, tag := range a.Tags {
				names[i] = tag.Name
			}
			tags = strings.Join(names, ",")
		}

		rows = append(rows, []string{
			strconv.FormatInt(a.ID, 10),
			a.Title,
			a.BlogName,
			status,
			fav,
			tags,
			formatTSVTime(a.PublishedDate, a.DiscoveredDate),
		})
	}

	return formatTSV(headers, rows, formatTSVFooter(meta))
}

// FormatBlogTSV 将博客列表格式化为 TSV 输出。
// schema: id name category articles last_scanned
func FormatBlogTSV(blogs []storage.BlogWithCount) string {
	if len(blogs) == 0 {
		return "没有博客"
	}

	headers := []string{"id", "name", "category", "articles", "last_scanned"}
	rows := make([][]string, 0, len(blogs))
	for _, b := range blogs {
		category := b.CategoryName
		if category == "" {
			category = "-"
		}

		lastScanned := ""
		if b.LastScanned != nil {
			lastScanned = b.LastScanned.In(cstZone).Format("2006-01-02T15:04:05+08:00")
		}

		rows = append(rows, []string{
			strconv.FormatInt(b.ID, 10),
			b.Name,
			category,
			strconv.Itoa(b.ArticleCount),
			lastScanned,
		})
	}

	return formatTSV(headers, rows, nil)
}

// FormatCategoryTSV 将分类列表格式化为 TSV 输出。
// schema: id name blogs
func FormatCategoryTSV(categories []storage.CategoryWithBlogCount) string {
	if len(categories) == 0 {
		return "没有分类"
	}

	headers := []string{"id", "name", "blogs"}
	rows := make([][]string, 0, len(categories))
	for _, c := range categories {
		rows = append(rows, []string{
			strconv.FormatInt(c.ID, 10),
			c.Name,
			strconv.Itoa(c.BlogCount),
		})
	}

	return formatTSV(headers, rows, nil)
}

// FormatTagTSV 将标签列表格式化为 TSV 输出。
// schema: id name articles
func FormatTagTSV(tags []model.Tag) string {
	if len(tags) == 0 {
		return "没有标签"
	}

	headers := []string{"id", "name", "articles"}
	rows := make([][]string, 0, len(tags))
	for _, t := range tags {
		rows = append(rows, []string{
			strconv.FormatInt(t.ID, 10),
			t.Name,
			strconv.FormatInt(t.ArticleCount, 10),
		})
	}

	return formatTSV(headers, rows, nil)
}
