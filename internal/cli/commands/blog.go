// ABOUTME: blog 子命令定义
// ABOUTME: 提供博客扫描和管理功能
package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/cli/flags"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/scanner"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
	"github.com/spf13/cobra"
)

// NewBlogCmd 创建 blog 子命令
// blog 是命令组，不直接执行，需要挂载子命令
func NewBlogCmd() *cobra.Command {
	blogCmd := &cobra.Command{
		Use:   "blog",
		Short: "博客管理命令",
		Long: `博客管理命令组。

提供博客扫描和管理的各种子命令。`,
		// blog 命令本身不执行任何操作，只显示帮助
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// 添加 scan 子命令
	blogCmd.AddCommand(NewScanCmd())

	return blogCmd
}

// NewScanCmd 创建 scan 子命令
// 支持扫描所有博客或指定博客
func NewScanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "scan [name]",
		Short: "扫描博客获取新文章",
		Long: `扫描博客获取新文章。

不提供 name 参数时扫描所有博客，
提供 name 参数时扫描指定博客。`,
		Args: cobra.MaximumNArgs(1), // 最多接受 1 个位置参数
		Run:   runScan,
	}
}

// runScan 执行 scan 命令
// 扫描博客并输出结果
func runScan(cmd *cobra.Command, args []string) {
	// 获取数据库路径
	dbPath := flags.DBPath()
	if dbPath == "" {
		// 使用默认路径
		dbPath = ""
	}

	// 打开数据库
	db, err := storage.OpenDatabase(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// 创建 context
	ctx := context.Background()

	// 根据参数数量决定扫描方式
	if len(args) == 0 {
		// 扫描所有博客
		scanAllBlogs(ctx, db)
	} else {
		// 扫描指定博客
		scanSingleBlog(ctx, db, args[0])
	}
}

// scanAllBlogs 扫描所有博客并输出结果
func scanAllBlogs(ctx context.Context, db *storage.Database) {
	results, err := scanner.ScanAllBlogs(ctx, db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "扫描失败: %v\n", err)
		os.Exit(1)
	}

	// 输出扫描结果
	for _, result := range results {
		outputScanResult(result)
	}
}

// scanSingleBlog 扫描指定博客并输出结果
func scanSingleBlog(ctx context.Context, db *storage.Database, name string) {
	result, err := scanner.ScanBlogByName(ctx, db, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "扫描失败: %v\n", err)
		os.Exit(1)
	}

	// 博客不存在
	if result == nil {
		fmt.Printf("博客 '%s' 不存在\n", name)
		os.Exit(1)
	}

	// 输出扫描结果
	outputScanResult(*result)
}

// outputScanResult 格式化输出单个扫描结果
func outputScanResult(result scanner.ScanResult) {
	if result.Error != "" {
		// 扫描失败
		fmt.Printf("[%s] 错误: %s\n", result.BlogName, result.Error)
	} else {
		// 扫描成功
		fmt.Printf("[%s] 新文章: %d, 总文章: %d, 来源: %s\n",
			result.BlogName,
			result.NewArticles,
			result.TotalFound,
			result.Source,
		)
	}
}