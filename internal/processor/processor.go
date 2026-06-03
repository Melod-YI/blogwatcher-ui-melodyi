// ABOUTME: 博客处理器接口和注册表
// ABOUTME: 提供按博客自定义 URL 处理的扩展机制，覆盖 RSS 解析、HN 搜索、缩略图提取等环节
package processor

// BlogProcessor 定义博客级的自定义处理钩子
// 各环节的默认实现不做任何转换，博客按需重写
type BlogProcessor interface {
	// NormalizeArticleURL 清洗 RSS 解析出的文章 URL
	NormalizeArticleURL(articleURL string) string

	// NormalizeSearchURL 清洗 HN 搜索前使用的 URL
	NormalizeSearchURL(searchURL string) string

	// NormalizeThumbnailURL 清洗缩略图 URL
	NormalizeThumbnailURL(thumbnailURL string) string
}

// BaseProcessor 默认实现，所有方法原样返回输入值
// 博客处理器可嵌入此结构体，仅重写需要的方法
type BaseProcessor struct{}

func (BaseProcessor) NormalizeArticleURL(articleURL string) string    { return articleURL }
func (BaseProcessor) NormalizeSearchURL(searchURL string) string      { return searchURL }
func (BaseProcessor) NormalizeThumbnailURL(thumbnailURL string) string { return thumbnailURL }

// Registry 按 feedURL 注册和查找处理器
type Registry struct {
	processors map[string]BlogProcessor
}

// NewRegistry 创建空的处理器注册表
func NewRegistry() *Registry {
	return &Registry{processors: make(map[string]BlogProcessor)}
}

// Register 为指定 feedURL 注册处理器
func (r *Registry) Register(feedURL string, p BlogProcessor) {
	r.processors[feedURL] = p
}

// Get 获取指定 feedURL 的处理器，未注册时返回 BaseProcessor
func (r *Registry) Get(feedURL string) BlogProcessor {
	if p, ok := r.processors[feedURL]; ok {
		return p
	}
	return BaseProcessor{}
}

// DefaultRegistry 全局默认注册表，由 setup 包的 init() 填充
var DefaultRegistry = NewRegistry()
