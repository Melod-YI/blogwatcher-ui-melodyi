// ABOUTME: Cobra 根命令定义
// ABOUTME: 提供统一的 CLI 入口点，支持版本显示和帮助信息
package commands

import (
	"fmt"
	"os"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/cli/flags"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/version"
	"github.com/spf13/cobra"
)

// rootCmd 是 CLI 的根命令
// 所有子命令都挂载在 rootCmd 上
var rootCmd = &cobra.Command{
	Use:     "blogwatcher",
	Short:   "博客文章管理工具",
	Long: `BlogWatcher 是一个博客文章管理工具。

它提供 Web UI 用于浏览和管理博客文章，
同时提供 CLI 命令用于扫描博客和管理文章状态。`,
	Version: version.Version,
}

// 初始化根命令（设置版本模板等）
func init() {
	// 设置版本输出模板
	rootCmd.SetVersionTemplate(`{{.Name}} {{.Version}}
`)
	// 添加 serve 子命令
	rootCmd.AddCommand(NewServeCmd())
}

// Execute 执行根命令
// 这是 CLI 的主入口点，由 main.go 调用
func Execute() {
	// 添加全局 flags
	flags.SetGlobalFlags(rootCmd)

	// 执行命令
	if err := rootCmd.Execute(); err != nil {
		// cobra 已经打印了错误信息，我们只需要退出
		os.Exit(1)
	}
}

// AddCommand 添加子命令到根命令
// 供其他包注册子命令
func AddCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}

// GetVersion 获取当前版本信息
// 用于测试和调试
func GetVersion() string {
	return fmt.Sprintf("%s %s", rootCmd.Use, version.Version)
}