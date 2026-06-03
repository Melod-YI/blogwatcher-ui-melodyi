// ABOUTME: 博客处理器接口和注册表单元测试
// ABOUTME: 测试 BaseProcessor 默认行为和 Registry 注册查找逻辑
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

// mockProcessor 用于测试的自定义处理器
type mockProcessor struct {
	articleURL string
}

func (m *mockProcessor) NormalizeArticleURL(url string) string    { return m.articleURL }
func (m *mockProcessor) NormalizeSearchURL(url string) string     { return url }
func (m *mockProcessor) NormalizeThumbnailURL(url string) string  { return url }
