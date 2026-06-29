// ABOUTME: simonwillison.net 博客处理器
// ABOUTME: 处理 simonwillison.net atom feed URL 中的 /#atom-xxx 后缀问题，以及过滤版本发布类文章
package blogs

import (
	"regexp"
	"strings"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/processor"
)

const (
	atomURLPrefix = "https://simonwillison.net/atom"
)

// atomSuffixRe 匹配文章 URL 末尾的 /#atom-xxx 片段
var atomSuffixRe = regexp.MustCompile(`/#[aA]tom-[\w-]+$`)

// skipTitlePrefixes 是需要过滤的版本发布类文章标题前缀（大小写敏感）
var skipTitlePrefixes = []string{
	"sqlite-utils",
	"datasette",
	"luau-wasm",
	"micropython-wasm",
	"llm",
	"asyncinject",
	"inaturalist-clumper",
	"asgi-gzip",
}

// SimonwillisonProcessor 处理 simonwillison.net 的特殊 URL 问题
// 所有 atom feed（everything、notes、links 等）的文章链接包含 /#atom-xxx 后缀，需要去除
// 同时过滤掉以特定项目名开头的版本发布类文章
type SimonwillisonProcessor struct {
	processor.BaseProcessor
}

// NormalizeArticleURL 去除 simonwillison 文章 URL 中的 /#atom-xxx 后缀
func (SimonwillisonProcessor) NormalizeArticleURL(articleURL string) string {
	return atomSuffixRe.ReplaceAllString(articleURL, "")
}

// ShouldSkipArticle 跳过标题以特定项目名开头的版本发布类文章（大小写敏感）
// 例如 "datasette 1.0a33"、"sqlite-utils 4.0rc1" 会被跳过
// 但 "Datasette Apps: ..." 不会被跳过（首字母大写）
func (SimonwillisonProcessor) ShouldSkipArticle(title string) bool {
	for _, prefix := range skipTitlePrefixes {
		if strings.HasPrefix(title, prefix) {
			return true
		}
	}
	return false
}

func init() {
	processor.DefaultRegistry.RegisterPrefix(atomURLPrefix, SimonwillisonProcessor{})
}
