// ABOUTME: category 子命令定义
// ABOUTME: 提供分类查看功能
package commands

import (
	"fmt"
	"os"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/cli/flags"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/cli/output"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
	"github.com/spf13/cobra"
)

// NewCategoryCmd 创建 category 子命令
// category 是命令组，不直接执行，需要挂载子命令
func NewCategoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "category",
		Short: "分类管理命令",
		Long: `分类管理命令组。

提供分类列表查看功能。`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(NewCategoryListCmd())
	return cmd
}

// NewCategoryListCmd 创建 list 子命令
func NewCategoryListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有分类",
		Long: `列出所有分类及其包含的博客数量。

输出格式：
  --format table   表格格式（默认）
  --format json    JSON 格式`,
		Run: runCategoryList,
	}
	cmd.Flags().String("format", "table", "输出格式 (table/json)")
	return cmd
}

// runCategoryList 执行 category list 命令
func runCategoryList(cmd *cobra.Command, args []string) {
	// 获取数据库路径
	dbPath := flags.DBPath()

	// 打开数据库
	db, err := storage.OpenDatabase(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// 获取分类列表
	categories, err := db.ListCategoriesWithBlogCount()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取分类列表失败: %v\n", err)
		os.Exit(1)
	}

	// 根据格式输出
	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "json":
		fmt.Println(output.FormatCategoryJSON(categories))
	default:
		fmt.Println(output.FormatCategoryTable(categories))
	}
}