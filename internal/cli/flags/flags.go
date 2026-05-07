// ABOUTME: 全局 CLI flags 定义和解析
// ABOUTME: 提供数据库路径等全局配置选项，供所有子命令使用
package flags

import (
	"github.com/spf13/cobra"
)

// dbPath 存储全局数据库路径 flag 值
// 如果为空字符串，命令执行时会使用 storage.DefaultDBPath() 获取默认路径
var dbPath string

// SetGlobalFlags 为根命令添加全局 persistent flags
// 这些 flags 对所有子命令都可用
func SetGlobalFlags(rootCmd *cobra.Command) {
	// --db flag：指定数据库路径
	// 默认值空字符串，实际使用时由 storage.DefaultDBPath() 提供
	rootCmd.PersistentFlags().StringVar(
		&dbPath,
		"db",
		"",
		"数据库路径（默认 ~/.blogwatcher/blogwatcher.db）",
	)
}

// DBPath 返回当前设置的数据库路径
// 供子命令在执行时获取全局 flag 值
func DBPath() string {
	return dbPath
}