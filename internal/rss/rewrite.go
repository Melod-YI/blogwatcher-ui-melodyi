// ABOUTME: feed URL 主机重写
// ABOUTME: 在 HTTP 抓取前按 BLOGWATCHER_FEED_HOSTMAP 把内部主机名（如 rsshub:1200）改写为当前运行时可访问的地址
package rss

import (
	"net/url"
	"os"
	"strings"
)

// envFeedHostmap 是存放 feed URL 主机重写映射的环境变量名。
// 格式：srchost[:port]=dsthost[:port]，多条以逗号分隔。
// 示例：BLOGWATCHER_FEED_HOSTMAP=rsshub:1200=localhost:19998
//
// 用途：Docker 容器内可直接通过服务名 rsshub:1200 访问 RSSHub；
// 但在宿主机上跑 CLI 时该主机名不可达，需改写为 localhost:19998。
// 容器侧不设置此变量即保持原样。
const envFeedHostmap = "BLOGWATCHER_FEED_HOSTMAP"

// parseHostmap 解析 "srchost[:port]=dsthost[:port],..." 形式的映射。
// key 归一为小写，value 原样保留。空或全无效时返回 nil。
func parseHostmap(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	m := make(map[string]string)
	for _, pair := range strings.Split(raw, ",") {
		k, v, found := strings.Cut(pair, "=")
		if !found {
			continue
		}
		k = strings.ToLower(strings.TrimSpace(k))
		v = strings.TrimSpace(v)
		if k == "" || v == "" {
			continue
		}
		m[k] = v
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

// loadHostmap 从环境变量读取并解析主机映射。
func loadHostmap() map[string]string {
	return parseHostmap(os.Getenv(envFeedHostmap))
}

// rewriteFeedURL 按 hostmap 重写 feedURL 的主机部分。
// 匹配规则：先尝试 "host:port" 精确匹配，再尝试 "host"（任意端口）匹配。
// value 含端口则使用指定端口，否则保留原端口。无匹配或解析失败时原样返回。
//
// 注意：仅在 HTTP 抓取前调用，不要用它改写传给 processor 注册表用于选择的 URL，
// 否则会破坏按原始域名（如 simonwillison.net）匹配的处理器逻辑。
func rewriteFeedURL(feedURL string, hostmap map[string]string) string {
	if len(hostmap) == 0 {
		return feedURL
	}
	u, err := url.Parse(feedURL)
	if err != nil || u.Host == "" {
		return feedURL
	}

	host := strings.ToLower(u.Hostname())
	port := u.Port()

	// 先按 host:port 精确匹配，再按 host-only 匹配（任意端口）
	var dst string
	if port != "" {
		if d, ok := hostmap[host+":"+port]; ok {
			dst = d
		}
	}
	if dst == "" {
		if d, ok := hostmap[host]; ok {
			dst = d
		}
	}
	if dst == "" {
		return feedURL
	}

	dh, dp, hasPort := strings.Cut(dst, ":")
	dh = strings.TrimSpace(dh)
	if dh == "" {
		return feedURL
	}
	if hasPort {
		u.Host = dh + ":" + strings.TrimSpace(dp)
	} else if port != "" {
		u.Host = dh + ":" + port // 保留原端口
	} else {
		u.Host = dh
	}
	return u.String()
}

// RewriteFeedURL 按环境变量 BLOGWATCHER_FEED_HOSTMAP 重写 feedURL 的主机部分。
// 在 ParseFeed 抓取前调用，使 web 与 CLI 共用同一份逻辑地址的 feed URL。
func RewriteFeedURL(feedURL string) string {
	return rewriteFeedURL(feedURL, loadHostmap())
}
