// ABOUTME: 博客处理器接口和注册表
// ABOUTME: 提供按博客自定义 URL 处理的扩展机制，覆盖 RSS 解析、HN 搜索、缩略图提取等环节
package processor

import "strings"

// BlogProcessor 定义博客级的自定义处理钩子
// 各环节的默认实现不做任何转换，博客按需重写
type BlogProcessor interface {
	// NormalizeArticleURL 清洗 RSS 解析出的文章 URL
	NormalizeArticleURL(articleURL string) string

	// NormalizeSearchURL 清洗 HN 搜索前使用的 URL
	NormalizeSearchURL(searchURL string) string

	// NormalizeThumbnailURL 清洗缩略图 URL
	NormalizeThumbnailURL(thumbnailURL string) string

	// ShouldSkipArticle 根据标题判断是否跳过该条目（如过滤掉版本发布类文章）
	ShouldSkipArticle(title string) bool
}

// BaseProcessor 默认实现，所有方法原样返回输入值
// 博客处理器可嵌入此结构体，仅重写需要的方法
type BaseProcessor struct{}

func (BaseProcessor) NormalizeArticleURL(articleURL string) string      { return articleURL }
func (BaseProcessor) NormalizeSearchURL(searchURL string) string        { return searchURL }
func (BaseProcessor) NormalizeThumbnailURL(thumbnailURL string) string  { return thumbnailURL }
func (BaseProcessor) ShouldSkipArticle(title string) bool               { return false }

// Registry 按 feedURL 注册和查找处理器
// 支持精确匹配和前缀匹配，查找时优先精确匹配
type Registry struct {
	processors map[string]BlogProcessor
	prefixes   map[string]BlogProcessor
}

// normalizeFeedURL 规范化 feed URL，去除协议前缀和尾部斜杠，用于匹配比较
func normalizeFeedURL(feedURL string) string {
	s := feedURL
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimRight(s, "/")
	return s
}

// NewRegistry 创建空的处理器注册表
func NewRegistry() *Registry {
	return &Registry{
		processors: make(map[string]BlogProcessor),
		prefixes:   make(map[string]BlogProcessor),
	}
}

// Register 为指定 feedURL 注册处理器（精确匹配）
func (r *Registry) Register(feedURL string, p BlogProcessor) {
	r.processors[normalizeFeedURL(feedURL)] = p
}

// RegisterPrefix 为指定 feedURL 前缀注册处理器（前缀匹配）
// 匹配时对 feedURL 做规范化（去除协议前缀和尾部斜杠）
func (r *Registry) RegisterPrefix(feedURLPrefix string, p BlogProcessor) {
	r.prefixes[normalizeFeedURL(feedURLPrefix)] = p
}

// Get 获取指定 feedURL 的处理器
// 查找优先级：精确匹配 > 最长前缀匹配 > BaseProcessor 默认实现
func (r *Registry) Get(feedURL string) BlogProcessor {
	normalized := normalizeFeedURL(feedURL)

	// 优先精确匹配
	if p, ok := r.processors[normalized]; ok {
		return p
	}

	// 前缀匹配，多个匹配时取最长前缀
	var bestPrefix string
	var bestProcessor BlogProcessor
	for prefix, p := range r.prefixes {
		if strings.HasPrefix(normalized, prefix) && len(prefix) > len(bestPrefix) {
			bestPrefix = prefix
			bestProcessor = p
		}
	}
	if bestProcessor != nil {
		return bestProcessor
	}

	return BaseProcessor{}
}

// DefaultRegistry 全局默认注册表，由 setup 包的 init() 填充
var DefaultRegistry = NewRegistry()
