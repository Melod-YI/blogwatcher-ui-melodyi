// ABOUTME: Tests for time filter parsing (parseTimeFilter / parseRelDuration)
package commands

import (
	"testing"
	"time"
)

// fixedNow 是测试基准时间，避免依赖系统时区。
var fixedNow = time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)

func TestParseTimeFilter_Relative(t *testing.T) {
	cases := []struct {
		in   string
		want time.Time
	}{
		{"now", fixedNow},
		{"now-24h", fixedNow.Add(-24 * time.Hour)},
		{"now-3d", fixedNow.Add(-3 * 24 * time.Hour)},
		{"now-2h30m", fixedNow.Add(-2*time.Hour - 30*time.Minute)},
	}
	for _, c := range cases {
		got, err := parseTimeFilter(c.in, fixedNow)
		if err != nil {
			t.Fatalf("parseTimeFilter(%q) err: %v", c.in, err)
		}
		if !got.Equal(c.want) {
			t.Errorf("parseTimeFilter(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseTimeFilter_Absolute(t *testing.T) {
	cases := []struct {
		in   string
		want time.Time
	}{
		// 日期：解析为 UTC 00:00
		{"2026-01-01", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		// RFC3339 带 Z
		{"2026-08-10T12:00:00Z", time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)},
		// RFC3339 带偏移
		{"2026-08-10T20:00:00+08:00", time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)},
		// 无时区的精确时间（time.Parse 无 location 时取 UTC）
		{"2026-08-10T12:00:00", time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)},
		// 空格分隔
		{"2026-08-10 12:00:00", time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)},
		// 空格分隔到分钟
		{"2026-08-10 12:00", time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)},
	}
	for _, c := range cases {
		got, err := parseTimeFilter(c.in, fixedNow)
		if err != nil {
			t.Fatalf("parseTimeFilter(%q) err: %v", c.in, err)
		}
		if !got.Equal(c.want) {
			t.Errorf("parseTimeFilter(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseTimeFilter_Invalid(t *testing.T) {
	bad := []string{
		"",
		"   ",
		"not-a-date",
		"2026-13-40",
		"now-",
		"now-12",   // 缺单位
		"now-12x",  // 未知单位
		"now--12h", // 双重符号
		"now12h",   // 缺分隔
		"now+90m",  // 未来方向不支持（文章不会在未来发布）
	}
	for _, in := range bad {
		if _, err := parseTimeFilter(in, fixedNow); err == nil {
			t.Errorf("expected error for %q, got nil", in)
		}
	}
}

func TestParseRelDuration(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"24h", 24 * time.Hour},
		{"3d", 3 * 24 * time.Hour},
		{"2d4h", 2*24*time.Hour + 4*time.Hour},
		{"12h30m", 12*time.Hour + 30*time.Minute},
		{"90m", 90 * time.Minute},
		{"3600s", 3600 * time.Second},
	}
	for _, c := range cases {
		got, err := parseRelDuration(c.in)
		if err != nil {
			t.Fatalf("parseRelDuration(%q) err: %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("parseRelDuration(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
