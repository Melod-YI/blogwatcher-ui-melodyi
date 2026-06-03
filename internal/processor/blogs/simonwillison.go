// ABOUTME: simonwillison.net 博客处理器
// ABOUTME: 处理 simonwillison.net 的 atom feed URL 中的 /#atom-everything 后缀问题
package blogs

import (
	"strings"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/processor"
)

const (
	simonwillisonFeedURL = "https://simonwillison.net/atom/everything/"
	atomSuffix           = "/#atom-everything"
)

// SimonwillisonProcessor 处理 simonwillison.net 的特殊 URL 问题
// atom/everything feed 中的文章链接包含 /#atom-everything 后缀，需要去除
type SimonwillisonProcessor struct {
	processor.BaseProcessor
}

// NormalizeArticleURL 去除 simonwillison 文章 URL 中的 /#atom-everything 后缀
func (SimonwillisonProcessor) NormalizeArticleURL(articleURL string) string {
	return strings.TrimSuffix(articleURL, atomSuffix)
}

func init() {
	processor.DefaultRegistry.Register(simonwillisonFeedURL, SimonwillisonProcessor{})
}
