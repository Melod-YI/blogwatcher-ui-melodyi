// ABOUTME: CLI 输出格式公共类型定义
// ABOUTME: 提供分页元信息等共享类型，供 tsv/json 输出格式使用
package output

// PaginationMeta 分页元信息，用于各输出格式显示分页状态
type PaginationMeta struct {
	Total   int64 // 符合条件的总条数
	Count   int   // 当前返回的条数
	Offset  int   // 当前偏移量
	Limit   int    // 请求的限制数（0 表示无限制）
	HasMore bool  // 是否有更多数据
}
