// ABOUTME: BlogWatcher CLI 统一入口点
// ABOUTME: 替代原 cmd/server 入口，支持 go install 安装
package main

import (
	"os"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/cli/commands"
)

func main() {
	// 执行 CLI 命令
	// Execute() 会处理所有命令解析和执行
	// 如果出错，Execute() 会打印错误并退出
	commands.Execute()

	// 如果 Execute() 返回，说明命令成功执行
	// 但 cobra Execute() 内部处理了错误退出，我们不需要在这里处理
	// 只有当 Execute() 返回错误时才需要手动退出
	// 这里保留 os.Exit(0) 以确保干净的退出状态
	os.Exit(0)
}