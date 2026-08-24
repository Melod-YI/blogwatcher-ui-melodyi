// ABOUTME: CLI 时间过滤参数解析
// ABOUTME: 支持 --after/--before 接受 日期 / 精确时间 / 相对现在（now-<dur>）
package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// parseTimeFilter 解析 --after/--before 的值，返回绝对时间。
// 支持两类输入：
//  1. 相对现在：now、now-<dur>（dur 由 parseRelDuration 解析，支持 d/h/m/s 及组合）
//     仅支持过去方向（now-）；文章不会在未来发布，故不提供 now+。
//  2. 绝对时间：RFC3339Nano / RFC3339 / 2006-01-02T15:04:05 / 2006-01-02 15:04:05 / 2006-01-02 15:04
//  3. 日期：2006-01-02（解析为该日 UTC 00:00）
//
// now 由调用方注入，便于测试。
func parseTimeFilter(s string, now time.Time) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("时间为空")
	}

	// 相对现在（仅过去方向）
	if s == "now" {
		return now, nil
	}
	if strings.HasPrefix(s, "now-") {
		dur, err := parseRelDuration(s[4:]) // 去掉 "now-"
		if err != nil {
			return time.Time{}, fmt.Errorf("相对时间格式错误: %q（示例 now-24h、now-3d、now-2h30m）: %w", s, err)
		}
		return now.Add(-dur), nil
	}

	// 绝对时间：依次尝试多种 layout
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	var lastErr error
	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}
	return time.Time{}, fmt.Errorf("无法解析时间 %q（支持 日期 2026-01-01 / 精确 2026-01-01T12:00:00[ Z|+08:00] / 相对 now-24h）: %w", s, lastErr)
}

// parseRelDuration 解析相对时长，支持 d/h/m/s 及组合（如 2d4h、12h30m、90m、3600s）。
// 与 time.ParseDuration 的差别：支持天 d，且不接受负号（符号由 now-/now+ 承担）。
func parseRelDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("时长为空")
	}

	var total time.Duration
	i := 0
	for i < len(s) {
		// 解析数字段
		j := i
		for j < len(s) && s[j] >= '0' && s[j] <= '9' {
			j++
		}
		if j == i {
			return 0, fmt.Errorf("位置 %d 处期望数字，得到 %q", i, s[i:])
		}
		n, err := strconv.Atoi(s[i:j])
		if err != nil {
			return 0, err
		}
		// 解析单位
		if j >= len(s) {
			return 0, fmt.Errorf("数字 %d 后缺少单位（d/h/m/s）", n)
		}
		unit := s[j]
		k := j + 1
		switch unit {
		case 'd':
			total += time.Duration(n) * 24 * time.Hour
		case 'h':
			total += time.Duration(n) * time.Hour
		case 'm':
			total += time.Duration(n) * time.Minute
		case 's':
			total += time.Duration(n) * time.Second
		default:
			return 0, fmt.Errorf("未知单位 %q（仅支持 d/h/m/s）", string(unit))
		}
		i = k
	}
	return total, nil
}
