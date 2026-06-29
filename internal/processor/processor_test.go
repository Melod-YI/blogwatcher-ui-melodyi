// ABOUTME: 博客处理器接口和注册表单元测试
// ABOUTME: 测试 BaseProcessor 默认行为、Registry 精确匹配、前缀匹配和 URL 规范化逻辑
package processor

import (
	"testing"
)

func TestBaseProcessor_NormalizeArticleURL(t *testing.T) {
	p := BaseProcessor{}
	url := "https://example.com/article/1"
	if got := p.NormalizeArticleURL(url); got != url {
		t.Errorf("BaseProcessor.NormalizeArticleURL(%s) = %s, want %s", url, got, url)
	}
}

func TestBaseProcessor_NormalizeSearchURL(t *testing.T) {
	p := BaseProcessor{}
	url := "https://example.com/article/1"
	if got := p.NormalizeSearchURL(url); got != url {
		t.Errorf("BaseProcessor.NormalizeSearchURL(%s) = %s, want %s", url, got, url)
	}
}

func TestBaseProcessor_NormalizeThumbnailURL(t *testing.T) {
	p := BaseProcessor{}
	url := "https://example.com/thumb.jpg"
	if got := p.NormalizeThumbnailURL(url); got != url {
		t.Errorf("BaseProcessor.NormalizeThumbnailURL(%s) = %s, want %s", url, got, url)
	}
}

func TestBaseProcessor_ShouldSkipArticle(t *testing.T) {
	p := BaseProcessor{}
	if got := p.ShouldSkipArticle("any title"); got {
		t.Errorf("BaseProcessor.ShouldSkipArticle() = true, want false")
	}
}

func TestRegistry_Get_Unregistered(t *testing.T) {
	r := NewRegistry()
	p := r.Get("https://example.com/feed.xml")
	// 未注册时返回 BaseProcessor，行为与原值返回一致
	url := "https://example.com/article"
	if got := p.NormalizeArticleURL(url); got != url {
		t.Errorf("Registry.Get(unregistered).NormalizeArticleURL(%s) = %s, want %s", url, got, url)
	}
}

func TestRegistry_Register_And_Get(t *testing.T) {
	r := NewRegistry()
	mock := &mockProcessor{articleURL: "normalized-url"}
	r.Register("https://example.com/feed.xml", mock)

	p := r.Get("https://example.com/feed.xml")
	if got := p.NormalizeArticleURL("original-url"); got != "normalized-url" {
		t.Errorf("Registry.Get(registered).NormalizeArticleURL() = %s, want normalized-url", got)
	}
}

func TestRegistry_Register_NormalizesTrailingSlash(t *testing.T) {
	r := NewRegistry()
	mock := &mockProcessor{articleURL: "matched"}
	r.Register("https://example.com/feed.xml/", mock)

	p := r.Get("https://example.com/feed.xml")
	if got := p.NormalizeArticleURL("original"); got != "matched" {
		t.Errorf("Register with trailing slash, Get without: got %s, want matched", got)
	}
}

func TestRegistry_Register_NormalizesScheme(t *testing.T) {
	r := NewRegistry()
	mock := &mockProcessor{articleURL: "matched"}
	r.Register("https://example.com/feed.xml", mock)

	p := r.Get("http://example.com/feed.xml")
	if got := p.NormalizeArticleURL("original"); got != "matched" {
		t.Errorf("Register https, Get http: got %s, want matched", got)
	}
}

func TestRegistry_RegisterPrefix_And_Get(t *testing.T) {
	r := NewRegistry()
	mock := &mockProcessor{articleURL: "prefix-matched"}
	r.RegisterPrefix("https://example.com/atom", mock)

	p := r.Get("https://example.com/atom/everything/")
	if got := p.NormalizeArticleURL("original"); got != "prefix-matched" {
		t.Errorf("RegisterPrefix, Get with sub-path: got %s, want prefix-matched", got)
	}
}

func TestRegistry_RegisterPrefix_NormalizesURL(t *testing.T) {
	r := NewRegistry()
	mock := &mockProcessor{articleURL: "prefix-matched"}
	r.RegisterPrefix("https://example.com/atom/", mock)

	tests := []struct {
		name    string
		feedURL string
	}{
		{"no trailing slash", "https://example.com/atom/notes"},
		{"with trailing slash", "https://example.com/atom/notes/"},
		{"http scheme", "http://example.com/atom/links/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := r.Get(tt.feedURL)
			if got := p.NormalizeArticleURL("original"); got != "prefix-matched" {
				t.Errorf("Get(%s): got %s, want prefix-matched", tt.feedURL, got)
			}
		})
	}
}

func TestRegistry_RegisterPrefix_LongestMatchWins(t *testing.T) {
	r := NewRegistry()
	shortMock := &mockProcessor{articleURL: "short"}
	longMock := &mockProcessor{articleURL: "long"}
	r.RegisterPrefix("https://example.com/atom", shortMock)
	r.RegisterPrefix("https://example.com/atom/special", longMock)

	p := r.Get("https://example.com/atom/special/feed")
	if got := p.NormalizeArticleURL("original"); got != "long" {
		t.Errorf("longest prefix match: got %s, want long", got)
	}

	p = r.Get("https://example.com/atom/other/feed")
	if got := p.NormalizeArticleURL("original"); got != "short" {
		t.Errorf("short prefix match: got %s, want short", got)
	}
}

func TestRegistry_ExactMatch_TakesPrecedenceOverPrefix(t *testing.T) {
	r := NewRegistry()
	exactMock := &mockProcessor{articleURL: "exact"}
	prefixMock := &mockProcessor{articleURL: "prefix"}
	r.Register("https://example.com/atom/exact", exactMock)
	r.RegisterPrefix("https://example.com/atom", prefixMock)

	p := r.Get("https://example.com/atom/exact")
	if got := p.NormalizeArticleURL("original"); got != "exact" {
		t.Errorf("exact match should take precedence: got %s, want exact", got)
	}
}

// mockProcessor 用于测试的自定义处理器
type mockProcessor struct {
	articleURL string
}

func (m *mockProcessor) NormalizeArticleURL(url string) string     { return m.articleURL }
func (m *mockProcessor) NormalizeSearchURL(url string) string      { return url }
func (m *mockProcessor) NormalizeThumbnailURL(url string) string   { return url }
func (m *mockProcessor) ShouldSkipArticle(title string) bool       { return false }
