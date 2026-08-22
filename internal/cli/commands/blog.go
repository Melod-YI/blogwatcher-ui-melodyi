// ABOUTME: blog 子命令定义
// ABOUTME: 提供博客扫描和管理功能
package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/cli/flags"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/cli/output"
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
	// 添加 list 子命令
	blogCmd.AddCommand(NewBlogListCmd())

	return blogCmd
}

// NewScanCmd 创建 scan 子命令
// 支持扫描所有博客、指定博客或指定分类下的博客
func NewScanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan [name]",
		Short: "扫描博客获取新文章",
		Long: `扫描博客获取新文章。

不提供 name 参数时扫描所有博客，
提供 name 参数时扫描指定博客，
使用 --category 参数时扫描指定分类下的所有博客。

参数：
  name            博客名称（可选）
  --category <name> 分类名称（可选）

name 参数和 --category 参数不能同时使用。`,
		Args: cobra.MaximumNArgs(1), // 最多接受 1 个位置参数
		Run:  runScan,
	}

	cmd.Flags().String("category", "", "分类名称筛选")
	cmd.MarkFlagsMutuallyExclusive("category")

	return cmd
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

	// 获取 --category 参数
	categoryName, _ := cmd.Flags().GetString("category")

	// 根据参数决定扫描方式
	if categoryName != "" {
		// 按分类扫描
		scanBlogsByCategory(ctx, db, categoryName)
	} else if len(args) == 0 {
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

// scanBlogsByCategory 扫描指定分类下的所有博客并输出结果
func scanBlogsByCategory(ctx context.Context, db *storage.Database, categoryName string) {
	// 验证分类存在
	category, err := db.GetCategoryByName(categoryName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查询分类失败: %v\n", err)
		os.Exit(1)
	}

	if category == nil {
		fmt.Fprintf(os.Stderr, "分类 '%s' 不存在\n", categoryName)
		os.Exit(1)
	}

	// 扫描分类下的博客
	results, err := scanner.ScanBlogsByCategory(ctx, db, category.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "扫描失败: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Printf("分类 '%s' 下没有博客\n", categoryName)
		return
	}

	// 输出扫描结果
	for _, result := range results {
		outputScanResult(result)
	}
}

// NewBlogListCmd 创建 list 子命令
func NewBlogListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出博客",
		Long: `列出博客，支持按分类筛选。

筛选参数：
  --category <name> 按分类名称筛选

输出格式：
  --format tsv     TSV 格式（默认，表头+数据行）
  --format json    JSON 格式

示例：
  blogwatcher blog list
  blogwatcher blog list --category tech
  blogwatcher blog list --format json`,
		Run: runBlogList,
	}

	cmd.Flags().String("category", "", "分类名称筛选")
	cmd.Flags().String("format", "tsv", "输出格式 (tsv|json)")

	return cmd
}

// runBlogList 执行 blog list 命令
func runBlogList(cmd *cobra.Command, args []string) {
	// 获取数据库路径
	dbPath := flags.DBPath()

	// 打开数据库
	db, err := storage.OpenDatabase(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// 获取 --category 参数
	categoryName, _ := cmd.Flags().GetString("category")

	var blogs []storage.BlogWithCount

	if categoryName != "" {
		// 验证分类存在
		category, err := db.GetCategoryByName(categoryName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "查询分类失败: %v\n", err)
			os.Exit(1)
		}

		if category == nil {
			fmt.Fprintf(os.Stderr, "分类 '%s' 不存在\n", categoryName)
			os.Exit(1)
		}

		// 查询分类下的博客
		blogs, err = db.ListBlogsWithCountsByCategoryID(category.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取博客列表失败: %v\n", err)
			os.Exit(1)
		}
	} else {
		// 查询所有博客
		blogs, err = db.ListBlogsWithCounts()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取博客列表失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 根据格式输出
	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "json":
		fmt.Println(output.FormatBlogJSON(blogs))
	default:
		fmt.Println(output.FormatBlogTSV(blogs))
	}
}