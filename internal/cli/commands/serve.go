// ABOUTME: serve 子命令定义
// ABOUTME: 启动 BlogWatcher UI HTTP 服务器
package commands

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/esttorhe/blogwatcher-ui/v2/assets"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/cli/flags"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/server"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/version"
	"github.com/spf13/cobra"
)

// NewServeCmd 创建 serve 子命令
// 该命令启动 HTTP 服务器，提供 BlogWatcher UI
func NewServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "启动 UI 服务器",
		Long: `启动 BlogWatcher Web UI 服务器。

服务器提供博客文章浏览和管理界面，
可通过浏览器访问 http://localhost:8080（默认端口）。`,
		Run: runServe,
	}
}

// runServe 执行 serve 命令
// 启动 HTTP 服务器并处理 graceful shutdown
func runServe(cmd *cobra.Command, args []string) {
	// 创建带有信号处理的 context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 运行服务器
	if err := run(ctx); err != nil {
		log.Fatalf("服务器错误: %v", err)
	}
}

// run 启动并管理 HTTP 服务器生命周期
func run(ctx context.Context) error {
	// 获取数据库路径：如果 flags.DBPath() 为空，使用默认路径
	dbPath := flags.DBPath()
	if dbPath == "" {
		// OpenDatabase("") 会自动使用 DefaultDBPath()
		dbPath = ""
	}

	// 打开数据库
	db, err := storage.OpenDatabase(dbPath)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}
	defer db.Close()

	// 从嵌入的 FS 中提取静态文件
	staticFiles, err := fs.Sub(assets.StaticFS, "static")
	if err != nil {
		return fmt.Errorf("提取静态文件失败: %w", err)
	}

	// 从嵌入的 FS 中提取模板文件
	templateFiles, err := fs.Sub(assets.TemplateFS, "templates")
	if err != nil {
		return fmt.Errorf("提取模板文件失败: %w", err)
	}

	// 创建服务器
	handler, err := server.NewServerWithFS(db, templateFiles, staticFiles, version.Version)
	if err != nil {
		return fmt.Errorf("创建服务器失败: %w", err)
	}

	// 配置端口：从环境变量 PORT 或默认 "8080"
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 配置 HTTP 服务器超时
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 在 goroutine 中启动服务器
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("服务器启动在 %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// 等待 context 取消或服务器错误
	select {
	case <-ctx.Done():
		log.Println("收到关闭信号")
	case err := <-serverErr:
		return err
	}

	// Graceful shutdown：10 秒超时
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("服务器关闭失败: %w", err)
	}

	log.Println("服务器优雅关闭")
	return nil
}