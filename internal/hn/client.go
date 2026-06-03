// ABOUTME: Hacker News 讨论链接搜索模块
// ABOUTME: 使用 Algolia API 搜索文章对应的 HN 讨论，支持 URL 匹配和网络错误重试
package hn

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
)

// SearchResult 表示 Algolia API 返回的单个结果
type SearchResult struct {
	ObjectID    string `json:"objectID"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Points      int    `json:"points"`
	NumComments int    `json:"num_comments"`
}

// AlgoliaResponse 表示 Algolia API 响应
type AlgoliaResponse struct {
	Hits []SearchResult `json:"hits"`
}

// MatchResult 表示 URL 匹配结果
type MatchResult struct {
	HNURL  string        // HN 讨论链接
	Status model.HNStatus // 搜索状态
}

const (
	algoliaAPIURL = "https://hn.algolia.com/api/v1/search"
	maxRetries    = 3
	retryDelay    = 500 * time.Millisecond
	httpTimeout   = 10 * time.Second
)

// isRateLimitError 检查是否为限流错误
func isRateLimitError(statusCode int) bool {
	return statusCode == 429
}

// normalizeURL 归一化 URL，消除常见的格式差异以便比较
// 处理：尾部斜杠、协议(http/https)、www 前缀、域名大小写
func normalizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// 获取主机名（去除端口）
	host := u.Hostname()
	if host == "" {
		host = u.Host
	}

	// 去除 www 前缀
	host = strings.TrimPrefix(strings.ToLower(host), "www.")

	// 去除路径尾部斜杠（保留根路径 "/"）
	path := strings.TrimRight(u.Path, "/")
	if path == "" {
		path = "/"
	}

	// 重建 URL（不含协议），用于比较
	result := host + path
	if u.RawQuery != "" {
		result += "?" + u.RawQuery
	}
	if u.Fragment != "" {
		result += "#" + u.Fragment
	}

	return result
}

// SearchHNDiscussion 搜索指定 URL 的 HN 讨论
// 返回匹配结果和可能的错误
func SearchHNDiscussion(ctx context.Context, articleURL string) (MatchResult, error) {
	log.Printf("[HN] 开始搜索文章 URL: %s", articleURL)

	// 构建请求 URL：限定只搜索 story 且只匹配 URL 字段
	queryURL := fmt.Sprintf("%s?query=%s&tags=story&restrictSearchableAttributes=url", algoliaAPIURL, url.QueryEscape(articleURL))

	var resp *http.Response
	var err error

	// 重试逻辑
	for retry := 0; retry < maxRetries; retry++ {
		if retry > 0 {
			log.Printf("[HN] 重试第 %d 次，等待 %v", retry, retryDelay)
			time.Sleep(retryDelay)
		}

		// 创建请求
		req, reqErr := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
		if reqErr != nil {
			return MatchResult{}, fmt.Errorf("创建请求失败: %w", reqErr)
		}

		// 发送请求
		client := &http.Client{Timeout: httpTimeout}
		resp, err = client.Do(req)

		if err != nil {
			log.Printf("[HN] 请求失败: %v", err)
			continue // 网络错误，重试
		}

		// 检查限流
		if isRateLimitError(resp.StatusCode) {
			resp.Body.Close()
			return MatchResult{}, fmt.Errorf("HN API 限流 (429)")
		}

		// 成功响应
		if resp.StatusCode == 200 {
			break
		}

		// 其他错误状态码
		resp.Body.Close()
		err = fmt.Errorf("HN API 返回状态码: %d", resp.StatusCode)
		log.Printf("[HN] API 错误: %v", err)
	}

	if err != nil {
		return MatchResult{}, fmt.Errorf("HN 搜索失败: %w", err)
	}
	if resp == nil {
		return MatchResult{}, fmt.Errorf("HN 搜索失败: 无响应")
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return MatchResult{}, fmt.Errorf("读取响应失败: %w", err)
	}

	var algoliaResp AlgoliaResponse
	if err := json.Unmarshal(body, &algoliaResp); err != nil {
		return MatchResult{}, fmt.Errorf("解析响应失败: %w", err)
	}

	log.Printf("[HN] 收到 %d 个搜索结果", len(algoliaResp.Hits))

	// 匹配逻辑
	if len(algoliaResp.Hits) == 0 {
		log.Printf("[HN] 无搜索结果，状态: not_found")
		return MatchResult{HNURL: "", Status: model.HNStatusNotFound}, nil
	}

	// 寻找精确匹配（使用 URL 归一化比较）
	normalizedArticleURL := normalizeURL(articleURL)
	for _, hit := range algoliaResp.Hits {
		if normalizeURL(hit.URL) == normalizedArticleURL {
			hnURL := fmt.Sprintf("https://news.ycombinator.com/item?id=%s", hit.ObjectID)
			log.Printf("[HN] 找到精确匹配，HN ID: %s", hit.ObjectID)
			return MatchResult{HNURL: hnURL, Status: model.HNStatusExact}, nil
		}
	}

	// 无精确匹配，记录归一化后的 URL 以便调试
	log.Printf("[HN] 无精确匹配，状态: not_found (归一化后: %s)", normalizedArticleURL)
	for _, hit := range algoliaResp.Hits {
		log.Printf("[HN]   候选 URL: %s (归一化后: %s)", hit.URL, normalizeURL(hit.URL))
	}
	return MatchResult{HNURL: "", Status: model.HNStatusNotFound}, nil
}