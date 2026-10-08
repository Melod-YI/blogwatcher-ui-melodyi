// ABOUTME: RSS 解析模块单元测试
// ABOUTME: 测试 ParseFeed 和 extractHNURLFromComments 函数
package rss

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/processor"
)

func TestExtractHNURLFromComments(t *testing.T) {
	tests := []struct {
		name     string
		comments string
		want     string
	}{
		{
			name:     "HN comments URL",
			comments: "https://news.ycombinator.com/item?id=12345",
			want:     "https://news.ycombinator.com/item?id=12345",
		},
		{
			name:     "HN comments URL with whitespace",
			comments: "  https://news.ycombinator.com/item?id=67890  ",
			want:     "https://news.ycombinator.com/item?id=67890",
		},
		{
			name:     "non-HN URL",
			comments: "https://example.com/comments",
			want:     "",
		},
		{
			name:     "empty comments",
			comments: "",
			want:     "",
		},
		{
			name:     "HN URL without ID",
			comments: "https://news.ycombinator.com/",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractHNURLFromComments(tt.comments)
			if got != tt.want {
				t.Errorf("extractHNURLFromComments(%s) = %s, want %s",
					tt.comments, got, tt.want)
			}
		})
	}
}

func TestExtractHNURLFromDescription(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        string
	}{
		{
			name:        "RSSub comments anchor",
			description: `<a href="https://news.ycombinator.com/item?id=48808482">Comments on Hacker News</a> | <a href="https://openwrt.org/toh/openwrt/one">Source</a>`,
			want:        "https://news.ycombinator.com/item?id=48808482",
		},
		{
			name:        "RSSub anchor with surrounding whitespace",
			description: `  <a href="https://news.ycombinator.com/item?id=48823557">Comments on Hacker News</a>  `,
			want:        "https://news.ycombinator.com/item?id=48823557",
		},
		{
			name:        "HN link in body text - not extracted",
			description: `<p>See this HN thread: <a href="https://news.ycombinator.com/item?id=12345">a random discussion</a></p>`,
			want:        "",
		},
		{
			name:        "non-HN comments anchor",
			description: `<a href="https://example.com/comments">Comments on Hacker News</a>`,
			want:        "",
		},
		{
			name:        "empty description",
			description: ``,
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractHNURLFromDescription(tt.description)
			if got != tt.want {
				t.Errorf("extractHNURLFromDescription(%s) = %s, want %s",
					tt.description, got, tt.want)
			}
		})
	}
}

// TestParseFeed_TrimsTrailingSlashFromArticleURL 回归测试：跨 feed 重复收录 bug。
// HN 官方 RSS / RSSHub 路由等聚合 feed 给出的文章 URL 带尾斜杠，而原博客 feed
// 已按无斜杠形式收录。文章查重按 URL 精确匹配，若聚合 feed 的 URL 不做归一化，
// 同一篇文章会重复入库（如 simonwillison.net 文章在 Simon 博客与 Hackernews News
// 两条博客下各存一条）。
func TestParseFeed_TrimsTrailingSlashFromArticleURL(t *testing.T) {
	rssXML := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel>
<title>Aggregate Feed</title>
<item><title>Post with slash</title><link>https://example.com/post/</link></item>
<item><title>Post without slash</title><link>https://example.com/post</link></item>
</channel></rss>`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(rssXML))
	}))
	defer srv.Close()

	articles, err := ParseFeed(context.Background(), srv.URL, processor.BaseProcessor{})
	if err != nil {
		t.Fatalf("ParseFeed() error = %v", err)
	}
	if len(articles) != 2 {
		t.Fatalf("ParseFeed() returned %d articles, want 2", len(articles))
	}
	for _, a := range articles {
		if a.URL != "https://example.com/post" {
			t.Errorf("article %q URL = %s, want https://example.com/post", a.Title, a.URL)
		}
	}
}
