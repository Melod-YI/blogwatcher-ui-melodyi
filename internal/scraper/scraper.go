// ABOUTME: Provides HTML scraping functionality as fallback for blogs without RSS feeds.
// ABOUTME: Used by scanner when no RSS feed is available but a scrape selector is configured.
package scraper

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ScrapedArticle struct {
	Title         string
	URL           string
	PublishedDate *time.Time
}

type ScrapeError struct {
	Message string
}

func (e ScrapeError) Error() string {
	return e.Message
}

func ScrapeBlog(ctx context.Context, blogURL string, selector string) ([]ScrapedArticle, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, blogURL, nil)
	if err != nil {
		return nil, ScrapeError{Message: fmt.Sprintf("failed to build request: %v", err)}
	}
	client := &http.Client{Transport: http.DefaultTransport}
	response, err := client.Do(req)
	if err != nil {
		return nil, ScrapeError{Message: fmt.Sprintf("failed to fetch page: %v", err)}
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, ScrapeError{Message: fmt.Sprintf("failed to fetch page: status %d", response.StatusCode)}
	}

	base, err := url.Parse(blogURL)
	if err != nil {
		return nil, ScrapeError{Message: "invalid blog url"}
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, ScrapeError{Message: fmt.Sprintf("failed to parse page: %v", err)}
	}

	seen := make(map[string]struct{})
	var articles []ScrapedArticle

	doc.Find(selector).Each(func(_ int, selection *goquery.Selection) {
		link := selection
		descendant := false // 选择器匹配的是 <a> 的后代元素（如整卡链接内的标题标签）
		if goquery.NodeName(selection) != "a" {
			link = selection.Find("a").First()
			if link.Length() == 0 {
				// 向下找不到 <a> 时向上取最近的祖先 <a>（场景：整卡即 <a>，
				// 选择器匹配卡片内的标题元素，如 "a[href^='...'] h3"）
				link = selection.Closest("a")
				descendant = true
			}
		}
		if link.Length() == 0 {
			return
		}
		href, exists := link.Attr("href")
		if !exists {
			return
		}
		resolved := resolveURL(base, href)
		if resolved == "" {
			return
		}
		if _, ok := seen[resolved]; ok {
			return
		}
		seen[resolved] = struct{}{}

		var title string
		if descendant {
			// 匹配元素即标题载体，直接取其文本，避免整卡文本（日期/摘要）混入
			title = strings.TrimSpace(selection.Text())
			if title == "" {
				title = extractTitle(link, selection)
			}
		} else {
			title = extractTitle(link, selection)
		}
		if title == "" {
			return
		}
		articles = append(articles, ScrapedArticle{
			Title:         title,
			URL:           resolved,
			PublishedDate: extractPublishedDate(link),
		})
	})

	return articles, nil
}

// cardDateLayouts 卡片内日期串的常见格式（完整日期，避免误匹配年份等碎片）
var cardDateLayouts = []string{
	time.RFC3339,
	"January 2, 2006", // October 6, 2026
	"Jan 2, 2006",     // Oct 1, 2026
	"2006-01-02",      // 2026-10-06
}

// cardDateRe 匹配卡片文本中的完整日期串：月名+日+年（全名/缩写）或 ISO 日期
var cardDateRe = regexp.MustCompile(`(?:January|February|March|April|May|June|July|August|September|October|November|December|Jan|Feb|Mar|Apr|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+\d{1,2},\s+\d{4}|\d{4}-\d{2}-\d{2}`)

// extractPublishedDate 从卡片（链接元素）内提取发布日期：
// 优先 HTML 标准的 <time datetime="...">，回退到卡片文本中的第一个完整日期串。
// 提取不到返回 nil，由调用方决定兜底（如以发现时间代替）。
func extractPublishedDate(link *goquery.Selection) *time.Time {
	if dt, ok := link.Find("time").First().Attr("datetime"); ok && dt != "" {
		if t := parseCardDate(dt); t != nil {
			return t
		}
	}
	if m := cardDateRe.FindString(link.Text()); m != "" {
		return parseCardDate(m)
	}
	return nil
}

// parseCardDate 按已知 layout 解析日期串，全部失败返回 nil
func parseCardDate(s string) *time.Time {
	s = strings.TrimSpace(strings.Replace(s, "Sept ", "Sep ", 1)) // 非标准缩写归一
	for _, layout := range cardDateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}

func extractTitle(link *goquery.Selection, parent *goquery.Selection) string {
	text := strings.TrimSpace(link.Text())
	if text != "" {
		return text
	}
	if title, exists := link.Attr("title"); exists {
		title = strings.TrimSpace(title)
		if title != "" {
			return title
		}
	}
	if parent != nil && parent != link {
		text = strings.TrimSpace(parent.Text())
		if text != "" {
			return text
		}
	}
	return ""
}

func resolveURL(base *url.URL, href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}
	parsed, err := url.Parse(href)
	if err != nil {
		return ""
	}
	return base.ResolveReference(parsed).String()
}

func IsScrapeError(err error) bool {
	var scrapeErr ScrapeError
	return errors.As(err, &scrapeErr)
}
