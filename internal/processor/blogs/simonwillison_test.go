// ABOUTME: simonwillison 处理器单元测试
// ABOUTME: 测试 simonwillison.net 的 URL 清洗逻辑
package blogs

import (
	"testing"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/processor"
)

func TestSimonwillisonProcessor_NormalizeArticleURL(t *testing.T) {
	p := SimonwillisonProcessor{}

	tests := []struct {
		name       string
		articleURL string
		want       string
	}{
		{
			name:       "strip atom-everything suffix",
			articleURL: "https://simonwillison.net/2026/May/18/sighting-362781627/#atom-everything",
			want:       "https://simonwillison.net/2026/May/18/sighting-362781627",
		},
		{
			name:       "strip atom-notes suffix",
			articleURL: "https://simonwillison.net/2026/Jun/8/wwdc/#atom-notes",
			want:       "https://simonwillison.net/2026/Jun/8/wwdc",
		},
		{
			name:       "strip atom-blogmarks suffix",
			articleURL: "https://simonwillison.net/2026/Jun/11/anthropic-walks-back-policy/#atom-blogmarks",
			want:       "https://simonwillison.net/2026/Jun/11/anthropic-walks-back-policy",
		},
		{
			name:       "strip hyphenated atom suffix",
			articleURL: "https://simonwillison.net/2026/Jun/1/post/#atom-my-tag",
			want:       "https://simonwillison.net/2026/Jun/1/post",
		},
		{
			name:       "no suffix - unchanged",
			articleURL: "https://simonwillison.net/2026/May/18/another-post",
			want:       "https://simonwillison.net/2026/May/18/another-post",
		},
		{
			// 新 feed 格式：链接已无 #atom 片段，但以尾斜杠结尾。
			// 必须与旧格式（#atom 片段被剥离后无尾斜杠）归一为同一 URL，
			// 否则 feed 格式变更前后同一篇文章会以带/不带尾斜杠两种 URL 重复入库。
			name:       "no suffix but trailing slash - stripped to canonical",
			articleURL: "https://simonwillison.net/2026/Aug/8/john-gruber/",
			want:       "https://simonwillison.net/2026/Aug/8/john-gruber",
		},
		{
			name:       "empty URL",
			articleURL: "",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.NormalizeArticleURL(tt.articleURL)
			if got != tt.want {
				t.Errorf("SimonwillisonProcessor.NormalizeArticleURL(%s) = %s, want %s",
					tt.articleURL, got, tt.want)
			}
		})
	}
}

func TestSimonwillisonProcessor_OtherMethods_UseBase(t *testing.T) {
	p := SimonwillisonProcessor{}

	// NormalizeSearchURL 和 NormalizeThumbnailURL 使用 BaseProcessor 默认实现（原样返回）
	searchURL := "https://simonwillison.net/2026/May/18/post"
	if got := p.NormalizeSearchURL(searchURL); got != searchURL {
		t.Errorf("SimonwillisonProcessor.NormalizeSearchURL(%s) = %s, want %s", searchURL, got, searchURL)
	}

	thumbURL := "https://simonwillison.net/static/image.jpg"
	if got := p.NormalizeThumbnailURL(thumbURL); got != thumbURL {
		t.Errorf("SimonwillisonProcessor.NormalizeThumbnailURL(%s) = %s, want %s", thumbURL, got, thumbURL)
	}
}

func TestSimonwillisonProcessor_ShouldSkipArticle(t *testing.T) {
	p := SimonwillisonProcessor{}

	tests := []struct {
		name  string
		title string
		want  bool
	}{
		{"datasette version release", "datasette 1.0a33", true},
		{"datasette-agent release", "datasette-agent 0.3a0", true},
		{"datasette-apps release", "datasette-apps 0.1a3", true},
		{"sqlite-utils release", "sqlite-utils 4.0rc1", true},
		{"sqlite-migrate release", "sqlite-migrate 0.1a0", true},
		{"luau-wasm release", "luau-wasm 0.1a0", true},
		{"micropython-wasm release", "micropython-wasm 0.1a0", true},
		{"llm release", "llm 0.25a0", true},
		{"asyncinject release", "asyncinject 0.7a0", true},
		{"inaturalist-clumper release", "inaturalist-clumper 0.1a0", true},
		{"asgi-gzip release", "asgi-gzip 0.1a0", true},
		{"Datasette capitalized - not skipped", "Datasette Apps: Host custom HTML applications inside Datasette", false},
		{"SQLite capitalized - not skipped", "SQLite: the database is the application", false},
		{"unrelated title", "Weeknotes: More releases, more museums", false},
		{"empty title", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.ShouldSkipArticle(tt.title)
			if got != tt.want {
				t.Errorf("SimonwillisonProcessor.ShouldSkipArticle(%q) = %v, want %v", tt.title, got, tt.want)
			}
		})
	}
}

func TestSimonwillisonProcessor_RegisteredInDefaultRegistry(t *testing.T) {
	// blogs 包的 init() 会将 simonwillison 处理器注册到 DefaultRegistry（前缀匹配）
	tests := []struct {
		name    string
		feedURL string
	}{
		{"atom-everything", "https://simonwillison.net/atom/everything/"},
		{"atom-notes", "https://simonwillison.net/atom/notes/"},
		{"atom-links", "https://simonwillison.net/atom/links/"},
		{"atom without trailing slash", "https://simonwillison.net/atom/everything"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := processor.DefaultRegistry.Get(tt.feedURL)
			got := p.NormalizeArticleURL("https://simonwillison.net/2026/May/18/post/#atom-everything")
			want := "https://simonwillison.net/2026/May/18/post"
			if got != want {
				t.Errorf("DefaultRegistry.Get(%s).NormalizeArticleURL() = %s, want %s", tt.feedURL, got, want)
			}
		})
	}
}
