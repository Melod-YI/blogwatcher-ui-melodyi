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

// SearchHNDiscussion 搜索指定 URL 的 HN 讨论
// 返回匹配结果和可能的错误
func SearchHNDiscussion(ctx context.Context, articleURL string) (MatchResult, error) {
	log.Printf("[HN] 开始搜索文章 URL: %s", articleURL)

	// 构建请求 URL
	queryURL := fmt.Sprintf("%s?query=%s", algoliaAPIURL, url.QueryEscape(articleURL))

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

	// 寻找精确匹配
	for _, hit := range algoliaResp.Hits {
		if hit.URL == articleURL {
			hnURL := fmt.Sprintf("https://news.ycombinator.com/item?id=%s", hit.ObjectID)
			log.Printf("[HN] 找到精确匹配，HN ID: %s", hit.ObjectID)
			return MatchResult{HNURL: hnURL, Status: model.HNStatusExact}, nil
		}
	}

	// 无精确匹配，选择最佳模糊匹配
	bestHit := selectBestFuzzyMatch(algoliaResp.Hits, articleURL)
	hnURL := fmt.Sprintf("https://news.ycombinator.com/item?id=%s", bestHit.ObjectID)
	log.Printf("[HN] 使用模糊匹配，HN ID: %s, 原URL: %s", bestHit.ObjectID, bestHit.URL)
	return MatchResult{HNURL: hnURL, Status: model.HNStatusFuzzy}, nil
}

// selectBestFuzzyMatch 选择最佳模糊匹配结果
// 优先选择 URL 相似度最高的，其次按点赞数排序
func selectBestFuzzyMatch(hits []SearchResult, articleURL string) SearchResult {
	// 提取域名用于相似度判断
	articleDomain := extractDomain(articleURL)

	// 计算每个结果的相似度分数
	bestScore := -1
	bestHit := hits[0] // 默认第一个

	for _, hit := range hits {
		score := calculateSimilarityScore(hit.URL, articleURL, articleDomain, hit.Points)
		if score > bestScore {
			bestScore = score
			bestHit = hit
		}
	}

	return bestHit
}

// extractDomain 从 URL 提取域名
func extractDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		// 尝试添加协议
		parsed, err = url.Parse("https://" + rawURL)
		if err != nil {
			return ""
		}
	}
	return parsed.Host
}

// calculateSimilarityScore 计算 URL 相似度分数
// 分数越高表示匹配度越好
func calculateSimilarityScore(hitURL, articleURL, articleDomain string, points int) int {
	score := 0

	hitDomain := extractDomain(hitURL)

	// 域名匹配加分（最高优先级）
	if hitDomain == articleDomain {
		score += 100
	}

	// URL 前缀匹配加分
	if strings.HasPrefix(articleURL, hitURL) || strings.HasPrefix(hitURL, articleURL) {
		score += 50
	}

	// URL 包含关系加分
	if strings.Contains(hitURL, articleURL) || strings.Contains(articleURL, hitURL) {
		score += 30
	}

	// 点赞数作为次要排序（归一化到 0-20）
	if points > 0 {
		pointScore := points / 10
		if pointScore > 20 {
			pointScore = 20
		}
		score += pointScore
	}

	return score
}