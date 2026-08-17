// ABOUTME: note 子命令定义
// ABOUTME: 提供备注写入和删除功能
package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/cli/flags"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
	"github.com/spf13/cobra"
)

// NewNoteCmd 创建 note 命令（命令组）
// note 是一个命令组，包含顶层写入命令和 delete / show 子命令
func NewNoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "note",
		Short: "备注管理命令",
		Long: `备注管理命令，提供备注写入、查看和删除功能。

子命令：
  delete    删除文章备注
  show      查看文章备注内容

示例：
  blogwatcher note --article-id 1 --file ~/note.md
  blogwatcher note delete --article-id 1
  blogwatcher note show --article-id 1`,
	}

	// 添加顶层命令 flags（用于写入备注）
	cmd.Flags().Int64("article-id", 0, "文章 ID（必填）")
	cmd.Flags().String("file", "", "备注文件路径（必填）")
	cmd.MarkFlagRequired("article-id")
	cmd.MarkFlagRequired("file")

	// 添加子命令
	cmd.AddCommand(NewNoteDeleteCmd())
	cmd.AddCommand(NewNoteShowCmd())

	// 设置顶层命令的 Run 函数
	cmd.Run = runNote

	return cmd
}

// NewNoteDeleteCmd 创建 delete 子命令
func NewNoteDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete --article-id <id>",
		Short: "删除文章备注",
		Long: `删除指定文章的备注文件并更新数据库状态。

示例：
  blogwatcher note delete --article-id 1`,
		Run: runNoteDelete,
	}

	cmd.Flags().Int64("article-id", 0, "文章 ID（必填）")
	cmd.MarkFlagRequired("article-id")

	return cmd
}

// NewNoteShowCmd 创建 show 子命令
func NewNoteShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show --article-id <id>",
		Short: "查看文章备注内容",
		Long: `输出指定文章的备注文件内容到标准输出（markdown 原文）。

示例：
  blogwatcher note show --article-id 1`,
		Run: runNoteShow,
	}

	cmd.Flags().Int64("article-id", 0, "文章 ID（必填）")
	cmd.MarkFlagRequired("article-id")

	return cmd
}

// runNote 执行 note 命令（写入备注）
func runNote(cmd *cobra.Command, args []string) {
	// 获取数据库路径
	dbPath := flags.DBPath()
	if dbPath == "" {
		var err error
		dbPath, err = storage.DefaultDBPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取默认数据库路径失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 打开数据库
	db, err := storage.OpenDatabase(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// 解析 flags
	articleID, _ := cmd.Flags().GetInt64("article-id")
	filePath, _ := cmd.Flags().GetString("file")

	// 验证文章 ID 存在（per D-05, D-06）
	article, err := db.GetArticleByID(articleID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "验证文章失败: %v\n", err)
		os.Exit(1)
	}
	if article == nil {
		fmt.Fprintf(os.Stderr, "文章 ID %d 不存在\n", articleID)
		os.Exit(1)
	}

	// 获取备注目录路径（per D-13）
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取用户目录失败: %v\n", err)
		os.Exit(1)
	}
	notesDir := filepath.Join(home, ".blogwatcher", "notes")

	// 创建目录（幂等，per D-13, D-14）
	if err := os.MkdirAll(notesDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "创建备注目录失败: %v\n", err)
		os.Exit(1)
	}

	// 文件路径（per D-06）
	notePath := filepath.Join(notesDir, fmt.Sprintf("%d.md", articleID))

	// 读取源文件（per D-08）
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取文件失败: %v\n", err)
		os.Exit(1)
	}

	// 写入目标文件（覆盖，per D-11）
	if err := os.WriteFile(notePath, content, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "写入备注失败: %v\n", err)
		os.Exit(1)
	}

	// 更新数据库状态（per NOTE-08）
	if err := db.UpdateArticleHasNote(articleID, true); err != nil {
		fmt.Fprintf(os.Stderr, "更新备注状态失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("已成功写入文章 %d 的备注\n", articleID)
}

// runNoteDelete 执行 note delete 命令（删除备注）
func runNoteDelete(cmd *cobra.Command, args []string) {
	// 获取数据库路径
	dbPath := flags.DBPath()
	if dbPath == "" {
		var err error
		dbPath, err = storage.DefaultDBPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取默认数据库路径失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 打开数据库
	db, err := storage.OpenDatabase(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// 解析 flags
	articleID, _ := cmd.Flags().GetInt64("article-id")

	// 验证文章 ID 存在
	article, err := db.GetArticleByID(articleID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "验证文章失败: %v\n", err)
		os.Exit(1)
	}
	if article == nil {
		fmt.Fprintf(os.Stderr, "文章 ID %d 不存在\n", articleID)
		os.Exit(1)
	}

	// 获取备注目录路径
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取用户目录失败: %v\n", err)
		os.Exit(1)
	}
	notesDir := filepath.Join(home, ".blogwatcher", "notes")
	notePath := filepath.Join(notesDir, fmt.Sprintf("%d.md", articleID))

	// 删除备注文件（per D-17）
	err = os.Remove(notePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "文章 %d 没有备注\n", articleID)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "删除备注失败: %v\n", err)
		os.Exit(1)
	}

	// 更新数据库状态
	if err := db.UpdateArticleHasNote(articleID, false); err != nil {
		fmt.Fprintf(os.Stderr, "更新备注状态失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("已成功删除文章 %d 的备注\n", articleID)
}

// runNoteShow 执行 note show 命令（查看备注）
// 校验文章存在后，读取备注文件并输出原文
func runNoteShow(cmd *cobra.Command, args []string) {
	// 获取数据库路径
	dbPath := flags.DBPath()
	if dbPath == "" {
		var err error
		dbPath, err = storage.DefaultDBPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取默认数据库路径失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 打开数据库
	db, err := storage.OpenDatabase(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "打开数据库失败: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// 解析 flags
	articleID, _ := cmd.Flags().GetInt64("article-id")

	// 验证文章 ID 存在
	article, err := db.GetArticleByID(articleID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "验证文章失败: %v\n", err)
		os.Exit(1)
	}
	if article == nil {
		fmt.Fprintf(os.Stderr, "文章 ID %d 不存在\n", articleID)
		os.Exit(1)
	}

	// 获取备注目录路径
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取用户目录失败: %v\n", err)
		os.Exit(1)
	}
	notesDir := filepath.Join(home, ".blogwatcher", "notes")
	notePath := filepath.Join(notesDir, fmt.Sprintf("%d.md", articleID))

	// 读取备注文件
	content, err := os.ReadFile(notePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "文章 %d 没有备注\n", articleID)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "读取备注失败: %v\n", err)
		os.Exit(1)
	}

	// 输出备注原文（markdown，不加 --format）
	if _, err := os.Stdout.Write(content); err != nil {
		fmt.Fprintf(os.Stderr, "输出备注失败: %v\n", err)
		os.Exit(1)
	}
}