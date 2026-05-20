// ABOUTME: CLI 输出格式公共类型定义
// ABOUTME: 提供分页元信息等共享类型，供 table/json/simple 输出格式使用
package output

import "fmt"

// PaginationMeta 分页元信息，用于各输出格式显示分页状态
type PaginationMeta struct {
	Total   int64 // 符合条件的总条数
	Count   int   // 当前返回的条数
	Offset  int   // 当前偏移量
	Limit   int   // 请求的限制数（0 表示无限制）
	HasMore bool  // 是否有更多数据
}

// formatPaginationFooter 格式化分页信息为底部说明
// 供 table 和 simple 格式复用
func formatPaginationFooter(meta PaginationMeta) string {
	limitStr := "无限制"
	if meta.Limit > 0 {
		limitStr = fmt.Sprintf("%d", meta.Limit)
	}

	base := fmt.Sprintf("总计: %d 条 | 返回: %d 条 | offset: %d | limit: %s",
		meta.Total, meta.Count, meta.Offset, limitStr)

	if meta.HasMore {
		nextOffset := meta.Offset + meta.Count
		return fmt.Sprintf("%s | 更多数据: --offset %d", base, nextOffset)
	}

	return base
}