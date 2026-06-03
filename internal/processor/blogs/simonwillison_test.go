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
			name:       "no suffix - unchanged",
			articleURL: "https://simonwillison.net/2026/May/18/another-post",
			want:       "https://simonwillison.net/2026/May/18/another-post",
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

func TestSimonwillisonProcessor_RegisteredInDefaultRegistry(t *testing.T) {
	// blogs 包的 init() 会将 simonwillison 处理器注册到 DefaultRegistry
	p := processor.DefaultRegistry.Get(simonwillisonFeedURL)
	got := p.NormalizeArticleURL("https://simonwillison.net/2026/May/18/post/#atom-everything")
	want := "https://simonwillison.net/2026/May/18/post"
	if got != want {
		t.Errorf("DefaultRegistry.Get(simonwillisonFeedURL).NormalizeArticleURL() = %s, want %s", got, want)
	}
}
