// ABOUTME: hn 子命令定义
// ABOUTME: 提供 HN 讨论链接同步 CLI 命令
package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/cli/flags"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/hn"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
	"github.com/spf13/cobra"
)

// NewHNCmd 创建 hn 命令组
func NewHNCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hn",
		Short: "Hacker News 讨论链接管理",
		Long: `Hacker News 讨论链接管理命令。

提供同步命令为文章搜索对应的 HN 讨论。`,
	}
	cmd.AddCommand(NewHNSyncCmd())
	return cmd
}

// NewHNSyncCmd 创建 sync 子命令
func NewHNSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "同步 HN 讨论链接",
		Long: `为文章搜索 Hacker News 上的讨论链接。

默认只搜索未搜索的文章（hn_status='not_searched'）。

参数：
  --all          强制重新搜索所有文章（覆盖已有状态）
  --failed       仅重新搜索失败的文章
  --blog <name>  仅搜索指定博客的文章
  --limit <n>    限制搜索数量（默认无限制）

示例：
  blogwatcher hn sync
  blogwatcher hn sync --all
  blogwatcher hn sync --failed
  blogwatcher hn sync --blog "Tech Blog"
  blogwatcher hn sync --limit 50`,
		Run: runHNSync,
	}

	cmd.Flags().Bool("all", false, "重新搜索所有文章")
	cmd.Flags().Bool("failed", false, "仅重新搜索失败的文章")
	cmd.Flags().String("blog", "", "指定博客名称")
	cmd.Flags().Int("limit", 0, "搜索数量限制（0 表示无限制）")

	cmd.MarkFlagsMutuallyExclusive("all", "failed")

	return cmd
}

// runHNSync 执行 HN 同步命令
func runHNSync(cmd *cobra.Command, args []string) {
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

	// 解析参数
	all, _ := cmd.Flags().GetBool("all")
	failed, _ := cmd.Flags().GetBool("failed")
	blogName, _ := cmd.Flags().GetString("blog")
	limit, _ := cmd.Flags().GetInt("limit")

	// 确定搜索模式
	mode := "not_searched"
	if all {
		mode = "all"
	} else if failed {
		mode = "failed"
	}

	// 获取需要搜索的文章
	articles, err := db.GetArticlesForHNSync(mode, blogName, limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查询文章失败: %v\n", err)
		os.Exit(1)
	}

	if len(articles) == 0 {
		fmt.Println("没有需要搜索的文章")
		return
	}

	fmt.Printf("开始搜索 %d 篇文章的 HN 讨论...\n", len(articles))

	// 统计结果
	var searched, foundExact, foundFuzzy, notFound, failedCount int

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// 逐篇搜索
	for i, article := range articles {
		searched++
		fmt.Printf("[%d/%d] 搜索文章 ID %d...\n", i+1, len(articles), article.ID)

		match, err := hn.SearchHNDiscussion(ctx, article.URL)
		if err != nil {
			failedCount++
			fmt.Printf("  失败: %v\n", err)
			_ = db.UpdateArticleHNStatus(article.ID, "", model.HNStatusFailed)
			continue
		}

		switch match.Status {
		case model.HNStatusExact:
			foundExact++
			fmt.Printf("  精确匹配: %s\n", match.HNURL)
		case model.HNStatusFuzzy:
			foundFuzzy++
			fmt.Printf("  模糊匹配: %s\n", match.HNURL)
		case model.HNStatusNotFound:
			notFound++
			fmt.Printf("  未找到讨论\n")
		}

		_ = db.UpdateArticleHNStatus(article.ID, match.HNURL, match.Status)
	}

	// 输出统计
	fmt.Println("\nHN 链接同步完成")
	fmt.Printf("搜索: %d 篇文章\n", searched)
	fmt.Printf("找到: %d 篇（精确: %d, 模糊: %d）\n", foundExact+foundFuzzy, foundExact, foundFuzzy)
	fmt.Printf("未找到: %d 篇\n", notFound)
	fmt.Printf("失败: %d 篇\n", failedCount)
}