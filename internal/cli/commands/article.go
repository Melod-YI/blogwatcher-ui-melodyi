// ABOUTME: article 子命令定义
// ABOUTME: 提供文章列表、标记已读/未读等 CLI 命令
package commands

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/cli/flags"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/cli/output"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
	"github.com/spf13/cobra"
)

// 常量定义
const (
	DefaultLimit   = 20  // 默认返回数量
	MaxLimit       = 100 // 最大返回数量
	MaxInputLength = 50  // 博客/分类名称最大长度
)

// NewArticleCmd 创建 article 命令（命令组）
// article 是一个命令组，包含 list、mark-read、mark-unread 子命令
func NewArticleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "article",
		Short: "文章管理命令",
		Long: `文章管理命令，提供文章列表和状态管理功能。

子命令：
  list       列出文章，支持筛选和多种输出格式
  mark-read  标记文章已读（单篇或全部）
  mark-unread 标记文章未读`,
	}

	// 添加子命令
	cmd.AddCommand(NewListCmd())
	cmd.AddCommand(NewMarkReadCmd())
	cmd.AddCommand(NewMarkUnreadCmd())

	return cmd
}

// NewListCmd 创建 list 子命令
// 支持筛选参数：--blog、--unread/--read、--noted/--not-noted、--after、--limit、--offset、--format
func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出文章",
		Long: `列出文章，支持筛选参数和多种输出格式。

筛选参数：
  --blog <name>    按博客名称筛选
  --category <name> 按分类名称筛选
  --unread         仅显示未读文章
  --read           仅显示已读文章
  --noted          仅显示有备注文章
  --not-noted      仅显示无备注文章
  --after <date>   显示指定日期之后的文章（格式 YYYY-MM-DD）
  --limit <n>      最多返回 n 条结果（默认 20，最大 100，0 表示无限制）
  --offset <n>     跳过前 n 条结果（用于翻页）

输出格式：
  --format table   表格格式（默认）
  --format json    JSON 格式
  --format simple  简洁格式

示例：
  blogwatcher article list
  blogwatcher article list --unread
  blogwatcher article list --noted
  blogwatcher article list --not-noted
  blogwatcher article list --noted --unread
  blogwatcher article list --category tech --unread
  blogwatcher article list --blog "Tech Blog" --unread --after 2026-01-01
  blogwatcher article list --unread --limit 10
  blogwatcher article list --limit 20 --offset 20  # 第二页
  blogwatcher article list --format json`,
		Run: runList,
	}

	// 添加筛选 flags
	cmd.Flags().String("blog", "", "博客名称筛选")
	cmd.Flags().String("category", "", "分类名称筛选")
	cmd.Flags().Bool("unread", false, "仅未读文章")
	cmd.Flags().Bool("read", false, "仅已读文章")
	cmd.Flags().Bool("noted", false, "仅有备注文章")
	cmd.Flags().Bool("not-noted", false, "仅无备注文章")
	cmd.Flags().String("after", "", "日期筛选（格式 YYYY-MM-DD）")
	cmd.Flags().Int("limit", DefaultLimit, fmt.Sprintf("返回结果数量限制（默认 %d，最大 %d，0 表示无限制）", DefaultLimit, MaxLimit))
	cmd.Flags().Int("offset", 0, "结果偏移量（用于翻页）")
	cmd.Flags().String("format", "table", "输出格式（table|json|simple）")

	// 标记互斥 flags
	cmd.MarkFlagsMutuallyExclusive("unread", "read")
	cmd.MarkFlagsMutuallyExclusive("noted", "not-noted")

	return cmd
}

// NewMarkReadCmd 创建 mark-read 子命令
// 支持单篇文章（参数 id）或批量（--all）
func NewMarkReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mark-read [id]",
		Short: "标记文章已读",
		Long: `标记指定文章已读，或使用 --all 标记所有未读文章。

示例：
  blogwatcher article mark-read 1      # 标记文章 ID 1 为已读
  blogwatcher article mark-read --all  # 标记所有未读文章为已读`,
		Args: cobra.MaximumNArgs(1),
		Run:  runMarkRead,
	}

	cmd.Flags().Bool("all", false, "标记所有未读文章为已读")

	return cmd
}

// NewMarkUnreadCmd 创建 mark-unread 子命令
// 必须提供文章 ID 参数
func NewMarkUnreadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mark-unread <id>",
		Short: "标记文章未读",
		Long: `标记指定文章为未读。

示例：
  blogwatcher article mark-unread 1  # 标记文章 ID 1 为未读`,
		Args: cobra.ExactArgs(1),
		Run:  runMarkUnread,
	}

	return cmd
}

// runList 执行 list 命令
// 解析筛选参数，调用 storage 方法，输出格式化结果
func runList(cmd *cobra.Command, args []string) {
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

	// 解析筛选参数
	blogName, _ := cmd.Flags().GetString("blog")
	categoryName, _ := cmd.Flags().GetString("category")
	unread, _ := cmd.Flags().GetBool("unread")
	read, _ := cmd.Flags().GetBool("read")
	noted, _ := cmd.Flags().GetBool("noted")
	notNoted, _ := cmd.Flags().GetBool("not-noted")
	afterStr, _ := cmd.Flags().GetString("after")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	format, _ := cmd.Flags().GetString("format")

	// 验证输入长度
	if len(blogName) > MaxInputLength {
		fmt.Fprintf(os.Stderr, "博客名称长度超过最大值 %d\n", MaxInputLength)
		os.Exit(1)
	}
	if len(categoryName) > MaxInputLength {
		fmt.Fprintf(os.Stderr, "分类名称长度超过最大值 %d\n", MaxInputLength)
		os.Exit(1)
	}

	// 验证并处理 limit 参数
	if limit < 0 {
		fmt.Fprintf(os.Stderr, "limit 参数不能为负数\n")
		os.Exit(1)
	}
	if limit > MaxLimit {
		fmt.Fprintf(os.Stderr, "limit 参数超过最大值 %d\n", MaxLimit)
		os.Exit(1)
	}

	// 验证 offset 参数
	if offset < 0 {
		fmt.Fprintf(os.Stderr, "offset 参数不能为负数\n")
		os.Exit(1)
	}

	// 构建筛选选项
	opts := storage.ListFilterOptions{
		BlogName:     blogName,
		CategoryName: categoryName,
		Limit:        limit,
		Offset:       offset,
	}

	// 设置 IsRead 状态筛选
	if unread {
		isRead := false
		opts.IsRead = &isRead
	} else if read {
		isRead := true
		opts.IsRead = &isRead
	}
	// 如果都没有设置，opts.IsRead 为 nil（所有状态）

	// 设置 HasNote 状态筛选
	if noted {
		hasNote := true
		opts.HasNote = &hasNote
	} else if notNoted {
		hasNote := false
		opts.HasNote = &hasNote
	}
	// 如果都没有设置，opts.HasNote 为 nil（所有状态）

	// 解析日期筛选
	if afterStr != "" {
		afterDate, err := time.Parse("2006-01-02", afterStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "日期格式错误: %v（格式应为 YYYY-MM-DD）\n", err)
			os.Exit(1)
		}
		opts.AfterDate = &afterDate
	}

	// 查询文章
	articles, err := db.ListArticlesWithFilters(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查询文章失败: %v\n", err)
		os.Exit(1)
	}

	// 查询总数
	total, err := db.CountArticlesWithFilters(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查询总数失败: %v\n", err)
		os.Exit(1)
	}

	// 检查博客名称是否存在（如果指定了）
	if blogName != "" && len(articles) == 0 {
		// 验证博客是否存在
		blog, err := db.GetBlogByName(blogName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "查询博客失败: %v\n", err)
			os.Exit(1)
		}
		if blog == nil {
			fmt.Fprintf(os.Stderr, "博客 '%s' 不存在\n", blogName)
			os.Exit(1)
		}
	}

	// 检查分类名称是否存在（如果指定了）
	if categoryName != "" && len(articles) == 0 {
		// 验证分类是否存在
		category, err := db.GetCategoryByName(categoryName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "查询分类失败: %v\n", err)
			os.Exit(1)
		}
		if category == nil {
			fmt.Fprintf(os.Stderr, "分类 '%s' 不存在\n", categoryName)
			os.Exit(1)
		}
	}

	// 构建分页元信息
	meta := output.PaginationMeta{
		Total:   total,
		Count:   len(articles),
		Offset:  offset,
		Limit:   limit,
		HasMore: int64(offset+len(articles)) < total,
	}

	// 根据格式输出结果
	var result string
	switch format {
	case "json":
		result = output.FormatJSON(articles, meta)
	case "simple":
		result = output.FormatSimple(articles, meta)
	default:
		result = output.FormatTable(articles, meta)
	}

	fmt.Println(result)
}

// runMarkRead 执行 mark-read 命令
// 支持单篇文章或批量标记
func runMarkRead(cmd *cobra.Command, args []string) {
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

	// 判断 --all flag
	markAll, _ := cmd.Flags().GetBool("all")

	if markAll {
		// 标记所有未读文章为已读
		err := db.MarkAllUnreadArticlesRead(nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "标记全部已读失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("已标记所有未读文章为已读")
		return
	}

	// 检查是否提供了文章 ID
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "需要提供文章 ID 或使用 --all\n")
		os.Exit(1)
	}

	// 解析文章 ID
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "文章 ID 格式错误: %v\n", err)
		os.Exit(1)
	}

	// 标记单篇文章已读
	success, err := db.MarkArticleRead(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "标记文章已读失败: %v\n", err)
		os.Exit(1)
	}

	if success {
		fmt.Printf("已标记文章 %d 为已读\n", id)
	} else {
		fmt.Fprintf(os.Stderr, "文章 %d 不存在或已标记为已读\n", id)
	}
}

// runMarkUnread 执行 mark-unread 命令
// 必须提供文章 ID 参数
func runMarkUnread(cmd *cobra.Command, args []string) {
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

	// 解析文章 ID
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "文章 ID 格式错误: %v\n", err)
		os.Exit(1)
	}

	// 标记文章未读
	success, err := db.MarkArticleUnread(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "标记文章未读失败: %v\n", err)
		os.Exit(1)
	}

	if success {
		fmt.Printf("已标记文章 %d 为未读\n", id)
	} else {
		fmt.Fprintf(os.Stderr, "文章 %d 不存在或已标记为未读\n", id)
	}
}