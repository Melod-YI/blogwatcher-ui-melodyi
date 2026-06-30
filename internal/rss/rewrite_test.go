// ABOUTME: feed URL host 重写单元测试
// ABOUTME: 测试 RewriteFeedURL 与 parseHostmap，覆盖端口匹配、host-only 匹配、scheme 保留等
package rss

import "testing"

func TestRewriteFeedURL(t *testing.T) {
	// rsshub:1200 -> localhost:19998，含端口精确匹配
	hostmap := map[string]string{
		"rsshub:1200": "localhost:19998",
		"rsshub2":     "localhost", // host-only，保留原端口
	}

	tests := []struct {
		name    string
		feedURL string
		want    string
	}{
		{
			name:    "host:port 精确匹配",
			feedURL: "http://rsshub:1200/simonwillison/atom/everything",
			want:    "http://localhost:19998/simonwillison/atom/everything",
		},
		{
			name:    "https scheme 保留",
			feedURL: "https://rsshub:1200/path",
			want:    "https://localhost:19998/path",
		},
		{
			name:    "host 大小写不敏感",
			feedURL: "http://RSSHUB:1200/path",
			want:    "http://localhost:19998/path",
		},
		{
			name:    "host-only 匹配保留原端口",
			feedURL: "http://rsshub2:1200/path",
			want:    "http://localhost:1200/path",
		},
		{
			name:    "无匹配保持不变",
			feedURL: "https://simonwillison.net/atom/everything",
			want:    "https://simonwillison.net/atom/everything",
		},
		{
			name:    "无端口 URL 不匹配带端口 key",
			feedURL: "http://rsshub/path",
			want:    "http://rsshub/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rewriteFeedURL(tt.feedURL, hostmap)
			if got != tt.want {
				t.Errorf("rewriteFeedURL(%q) = %q, want %q", tt.feedURL, got, tt.want)
			}
		})
	}
}

func TestRewriteFeedURL_EmptyHostmap(t *testing.T) {
	// 空 hostmap 原样返回，且不 panic
	got := rewriteFeedURL("http://rsshub:1200/path", nil)
	if got != "http://rsshub:1200/path" {
		t.Errorf("expected unchanged, got %q", got)
	}
}

func TestRewriteFeedURL_Malformed(t *testing.T) {
	// 畸形 URL 或无 host 不应 panic，原样返回
	hostmap := map[string]string{"rsshub:1200": "localhost:19998"}
	for _, in := range []string{"", "not a url", "://nohost", "ftp://rsshub:1200/path"} {
		got := rewriteFeedURL(in, hostmap) //nolint:errcheck // 仅验证不 panic
		_ = got
	}
}

func TestParseHostmap(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want map[string]string
	}{
		{
			name: "单条",
			raw:  "rsshub:1200=localhost:19998",
			want: map[string]string{"rsshub:1200": "localhost:19998"},
		},
		{
			name: "多条逗号分隔带空格",
			raw:  "rsshub:1200=localhost:19998 , svc=srv:8080",
			want: map[string]string{
				"rsshub:1200": "localhost:19998",
				"svc":         "srv:8080",
			},
		},
		{
			name: "key 大小写归一",
			raw:  "RSSHUB:1200=Localhost:19998",
			want: map[string]string{"rsshub:1200": "Localhost:19998"},
		},
		{
			name: "忽略缺少等号的项",
			raw:  "rsshub:1200=localhost:19998,baditem,svc=srv:8080",
			want: map[string]string{
				"rsshub:1200": "localhost:19998",
				"svc":         "srv:8080",
			},
		},
		{
			name: "空字符串返回 nil",
			raw:  "",
			want: nil,
		},
		{
			name: "全无效返回 nil",
			raw:  "baditem, anotherbad",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseHostmap(tt.raw)
			if len(got) != len(tt.want) {
				t.Errorf("parseHostmap(%q) len = %d, want %d (%v)", tt.raw, len(got), len(tt.want), got)
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("parseHostmap(%q)[%q] = %q, want %q", tt.raw, k, got[k], v)
				}
			}
		})
	}
}

func TestRewriteFeedURL_FromEnv(t *testing.T) {
	// RewriteFeedURL 应读取 BLOGWATCHER_FEED_HOSTMAP 环境变量
	t.Setenv("BLOGWATCHER_FEED_HOSTMAP", "rsshub:1200=localhost:19998")
	got := RewriteFeedURL("http://rsshub:1200/path")
	if got != "http://localhost:19998/path" {
		t.Errorf("RewriteFeedURL with env = %q, want http://localhost:19998/path", got)
	}
}

func TestRewriteFeedURL_NoEnv(t *testing.T) {
	// 未设置 env 时原样返回
	t.Setenv("BLOGWATCHER_FEED_HOSTMAP", "")
	got := RewriteFeedURL("http://rsshub:1200/path")
	if got != "http://rsshub:1200/path" {
		t.Errorf("expected unchanged, got %q", got)
	}
}
