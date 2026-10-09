// ABOUTME: HTML 抓取模块单元测试
// ABOUTME: 覆盖三种选择器模式：直接匹配 <a>、匹配包含 <a> 的容器、匹配 <a> 的后代元素（如卡片内标题标签）
package scraper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// servePage 启动一个返回固定 HTML 的测试服务器
func servePage(t *testing.T, html string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(html))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestScrapeBlog_SelectorMatchesAnchor(t *testing.T) {
	// 模式 1：选择器直接匹配 <a> 元素，标题取 <a> 自身文本
	srv := servePage(t, `
		<html><body>
			<a class="article-link" href="/posts/1">First Post</a>
			<a class="article-link" href="/posts/2">Second Post</a>
			<a class="other" href="/posts/3">Ignored</a>
		</body></html>`)

	articles, err := ScrapeBlog(context.Background(), srv.URL, "a.article-link")
	if err != nil {
		t.Fatalf("ScrapeBlog() error = %v", err)
	}
	if len(articles) != 2 {
		t.Fatalf("ScrapeBlog() returned %d articles, want 2", len(articles))
	}
	if articles[0].Title != "First Post" || articles[0].URL != srv.URL+"/posts/1" {
		t.Errorf("article[0] = {%s %s}, want {First Post %s/posts/1}",
			articles[0].Title, articles[0].URL, srv.URL)
	}
}

func TestScrapeBlog_SelectorMatchesContainer(t *testing.T) {
	// 模式 2：选择器匹配包含 <a> 的容器，向下找第一个 <a>，标题取 <a> 文本
	srv := servePage(t, `
		<html><body>
			<div class="card"><a href="/posts/1">First Post</a></div>
			<div class="card"><a href="/posts/2">Second Post</a></div>
		</body></html>`)

	articles, err := ScrapeBlog(context.Background(), srv.URL, "div.card")
	if err != nil {
		t.Fatalf("ScrapeBlog() error = %v", err)
	}
	if len(articles) != 2 {
		t.Fatalf("ScrapeBlog() returned %d articles, want 2", len(articles))
	}
	if articles[1].Title != "Second Post" || articles[1].URL != srv.URL+"/posts/2" {
		t.Errorf("article[1] = {%s %s}, want {Second Post %s/posts/2}",
			articles[1].Title, articles[1].URL, srv.URL)
	}
}

func TestScrapeBlog_SelectorMatchesDescendantOfAnchor(t *testing.T) {
	// 模式 3：选择器匹配 <a> 的后代元素（如整卡链接内的标题标签）。
	// 场景：claude.com/resources/articles 的卡片是整个 <a> 包住卡片（内含
	// 类型/日期/标题/摘要），选择器配 "a[href^='/resources/articles/'] h3"
	// 以取干净的标题文本，链接从最近的祖先 <a> 取。
	// RSSHub claude/blog 路由 2026-10 因该页面改版（旧 .blog_cms_list 结构
	// 消失）解析为空，scraper 此模式作为兜底数据源。
	srv := servePage(t, `
		<html><body>
			<a href="/resources/articles/claude-code-mods" class="ResourceCard__card">
				<div class="header"><span class="type">Article</span><span class="meta">Oct 1, 2026</span></div>
				<div class="content"><div class="body">
					<h3 class="title">Customize Claude Code with mods</h3>
					<p class="excerpt">Change how Claude Code behaves.</p>
				</div></div>
			</a>
			<a href="/resources/articles/another-post" class="ResourceCard__card">
				<h3>Another Post</h3>
			</a>
		</body></html>`)

	articles, err := ScrapeBlog(context.Background(), srv.URL, "a[href^='/resources/articles/'] h3")
	if err != nil {
		t.Fatalf("ScrapeBlog() error = %v", err)
	}
	if len(articles) != 2 {
		t.Fatalf("ScrapeBlog() returned %d articles, want 2", len(articles))
	}
	first := articles[0]
	if first.Title != "Customize Claude Code with mods" {
		t.Errorf("article[0].Title = %q, want %q（应只取 h3 文本，不含日期/摘要）",
			first.Title, "Customize Claude Code with mods")
	}
	if first.URL != srv.URL+"/resources/articles/claude-code-mods" {
		t.Errorf("article[0].URL = %s, want %s/resources/articles/claude-code-mods（应从祖先 a 取 href）",
			first.URL, srv.URL)
	}
	if articles[1].Title != "Another Post" || articles[1].URL != srv.URL+"/resources/articles/another-post" {
		t.Errorf("article[1] = {%s %s}, want {Another Post %s/resources/articles/another-post}",
			articles[1].Title, articles[1].URL, srv.URL)
	}
}

func TestScrapeBlog_DeduplicatesByURL(t *testing.T) {
	// 同一 URL 的多个匹配只保留第一条
	srv := servePage(t, `
		<html><body>
			<a href="/posts/1">First</a>
			<div><a href="/posts/1">First Duplicate</a></div>
		</body></html>`)

	articles, err := ScrapeBlog(context.Background(), srv.URL, "a, div")
	if err != nil {
		t.Fatalf("ScrapeBlog() error = %v", err)
	}
	if len(articles) != 1 {
		t.Fatalf("ScrapeBlog() returned %d articles, want 1 (deduplicated)", len(articles))
	}
}

func TestScrapeBlog_ExtractsPublishedDate(t *testing.T) {
	// 日期提取：优先 <time datetime="...">（HTML 标准），回退到卡片文本中的完整日期串。
	// 场景：claude.com 卡片的日期是 span 文本 "Oct 1, 2026"。
	tests := []struct {
		name     string
		html     string
		selector string
		wantDate string // 空串表示应为 nil
	}{
		{
			name: "time 元素 datetime 属性（RFC3339）",
			html: `<a href="/p/1">Post One<time datetime="2026-09-30T10:00:00Z"></time></a>`,
			wantDate: "2026-09-30",
		},
		{
			name: "time 元素 datetime 属性（纯日期）",
			html: `<a href="/p/1">Post One<time datetime="2026-09-30"></time></a>`,
			wantDate: "2026-09-30",
		},
		{
			name: "卡片文本日期（缩写月名）",
			html: `<a href="/p/1"><span>Article</span><span>Oct 1, 2026</span><h3>Post One</h3></a>`,
			wantDate: "2026-10-01",
		},
		{
			name: "卡片文本日期（全名月名）",
			html: `<a href="/p/1"><span>October 6, 2026</span><h3>Post One</h3></a>`,
			wantDate: "2026-10-06",
		},
		{
			name: "无日期信息",
			html: `<a href="/p/1"><h3>Post One</h3><p>Just a teaser</p></a>`,
			wantDate: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := servePage(t, `<html><body>`+tt.html+`</body></html>`)
			articles, err := ScrapeBlog(context.Background(), srv.URL, "a[href^='/p/']")
			if err != nil {
				t.Fatalf("ScrapeBlog() error = %v", err)
			}
			if len(articles) != 1 {
				t.Fatalf("ScrapeBlog() returned %d articles, want 1", len(articles))
			}
			a := articles[0]
			if tt.wantDate == "" {
				if a.PublishedDate != nil {
					t.Errorf("PublishedDate = %v, want nil", a.PublishedDate)
				}
				return
			}
			if a.PublishedDate == nil {
				t.Fatalf("PublishedDate = nil, want %s", tt.wantDate)
			}
			if got := a.PublishedDate.Format("2006-01-02"); got != tt.wantDate {
				t.Errorf("PublishedDate = %s, want %s", got, tt.wantDate)
			}
		})
	}
}

func TestScrapeBlog_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	if _, err := ScrapeBlog(context.Background(), srv.URL, "a"); err == nil {
		t.Fatal("expected error for non-OK status")
	}
}
