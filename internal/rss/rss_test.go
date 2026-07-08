// ABOUTME: RSS 解析模块单元测试
// ABOUTME: 测试 ParseFeed 和 extractHNURLFromComments 函数
package rss

import (
	"testing"
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
