// ABOUTME: tag 子命令定义
// ABOUTME: 提供标签列表、重命名、删除功能
package commands

import (
	"fmt"
	"os"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/cli/flags"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/cli/output"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
	"github.com/spf13/cobra"
)

// NewTagCmd 创建 tag 命令组
// tag 是命令组，不直接执行，需要挂载子命令
func NewTagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "标签管理命令",
		Long: `标签管理命令组。

子命令：
  list                 列出所有标签及文章计数
  rename <old> <new>  重命名标签
  delete <name>        删除标签（级联解除关联）`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(NewTagListCmd())
	cmd.AddCommand(NewTagRenameCmd())
	cmd.AddCommand(NewTagDeleteCmd())
	return cmd
}

// NewTagListCmd 创建 list 子命令
func NewTagListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有标签",
		Long: `列出所有标签及文章计数。

输出格式：
  --format table   表格格式（默认）
  --format json    JSON 格式`,
		Run: runTagList,
	}
	cmd.Flags().String("format", "table", "输出格式 (table/json)")
	return cmd
}

// runTagList 执行 tag list 命令
func runTagList(cmd *cobra.Command, args []string) {
	db := openTagCmdDB()
	defer db.Close()

	tags, err := db.ListTags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取标签列表失败: %v\n", err)
		os.Exit(1)
	}

	format, _ := cmd.Flags().GetString("format")
	switch format {
	case "json":
		fmt.Println(output.FormatTagJSON(tags))
	default:
		fmt.Println(output.FormatTagTable(tags))
	}
}

// NewTagRenameCmd 创建 rename 子命令
func NewTagRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <oldName> <newName>",
		Short: "重命名标签",
		Long: `重命名标签。

示例：
  blogwatcher tag rename Go Golang  # 将标签 "Go" 重命名为 "Golang"`,
		Args: cobra.ExactArgs(2),
		Run:  runTagRename,
	}
}

// runTagRename 执行 tag rename 命令
func runTagRename(cmd *cobra.Command, args []string) {
	db := openTagCmdDB()
	defer db.Close()

	oldName, newName := args[0], args[1]
	tag, err := db.GetTagByName(oldName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查询标签失败: %v\n", err)
		os.Exit(1)
	}
	if tag == nil {
		fmt.Fprintf(os.Stderr, "标签 '%s' 不存在\n", oldName)
		os.Exit(1)
	}
	if err := db.RenameTag(tag.ID, newName); err != nil {
		fmt.Fprintf(os.Stderr, "重命名标签失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("已重命名标签: %s -> %s\n", oldName, newName)
}

// NewTagDeleteCmd 创建 delete 子命令
func NewTagDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "删除标签（级联解除关联）",
		Long: `删除标签，同时解除所有文章上的该标签关联。

示例：
  blogwatcher tag delete Go  # 删除标签 "Go"`,
		Args: cobra.ExactArgs(1),
		Run:  runTagDelete,
	}
}

// runTagDelete 执行 tag delete 命令
func runTagDelete(cmd *cobra.Command, args []string) {
	db := openTagCmdDB()
	defer db.Close()

	name := args[0]
	tag, err := db.GetTagByName(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查询标签失败: %v\n", err)
		os.Exit(1)
	}
	if tag == nil {
		fmt.Fprintf(os.Stderr, "标签 '%s' 不存在\n", name)
		os.Exit(1)
	}
	affected, err := db.DeleteTag(tag.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "删除标签失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("已删除标签: %s（解除 %d 篇文章关联）\n", name, affected)
}

// openTagCmdDB 打开命令用的数据库
// 复用 article.go runFavorite 的 db 路径处理逻辑（flags.DBPath -> DefaultDBPath 兜底）
func openTagCmdDB() *storage.Database {
	dbPath := flags.DBPath()
	if dbPath == "" {
		var err error
		dbPath, err = storage.DefaultDBPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取默认数据库路径失败: %v\n", err)
			os.Exit(1)
		}
	}
	db, err := storage.OpenDatabase(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开数据库失败: %v\n", err)
		os.Exit(1)
	}
	return db
}
