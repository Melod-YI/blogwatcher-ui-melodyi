# Phase 13: CLI Notes Infrastructure - Pattern Map

**Mapped:** 2026-05-07
**Files analyzed:** 3 (2 new, 1 modified)
**Analogs found:** 3 / 3

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/cli/commands/note.go` | CLI command | file-I/O + CRUD | `internal/cli/commands/article.go` | exact |
| `internal/storage/database.go` | database method | CRUD | `internal/storage/database.go` (existing) | exact |
| `internal/model/model.go` | model | N/A | `internal/model/model.go` (existing) | exact |

## Pattern Assignments

### `internal/cli/commands/note.go` (CLI command, file-I/O + CRUD)

**Analog:** `internal/cli/commands/article.go` + `internal/cli/commands/blog.go`

**Package imports pattern** (article.go:1-15):
```go
// ABOUTME: note 子命令定义
// ABOUTME: 提供备注写入和删除功能
package commands

import (
	"fmt"
	"os"
	"strconv"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/cli/flags"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
	"github.com/spf13/cobra"
)
```

**命令组注册模式** (root.go:27-37):
在 `init()` 函数中注册顶层命令：
```go
func init() {
	// 其他子命令...
	rootCmd.AddCommand(NewNoteCmd())
}
```

**命令组定义模式** (article.go:18-37):
```go
// NewNoteCmd 创建 note 命令（命令组）
// note 是一个命令组，包含顶层写入命令和 delete 子命令
func NewNoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "note",
		Short: "备注管理命令",
		Long: `备注管理命令，提供备注写入和删除功能。

	子命令：
	  delete    删除文章备注`,
	}

	// 添加子命令
	cmd.AddCommand(NewNoteDeleteCmd())

	return cmd
}
```

**顶层命令 Run 函数模式** (article.go:118-203):
```go
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

	// 执行备注操作（文件复制 + 数据库更新）
	// ... 具体实现 ...
}
```

**子命令定义模式** (article.go:101-114):
```go
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
```

**错误处理模式** (article.go:125-136, 207-238):
```go
// 所有错误都使用相同格式：
// fmt.Fprintf(os.Stderr, "错误描述: %v\n", err)
// os.Exit(1)

if err != nil {
	fmt.Fprintf(os.Stderr, "操作失败: %v\n", err)
	os.Exit(1)
}

// 成功时的简单输出：
fmt.Printf("成功消息\n")
```

**文件操作模式** (参考 database.go:21-27 + io.Copy 或 os.ReadFile/os.WriteFile):
```go
import (
	"io"
	"os"
	"path/filepath"
)

// 获取备注目录路径
home, err := os.UserHomeDir()
if err != nil {
	return "", err
}
notesDir := filepath.Join(home, ".blogwatcher", "notes")

// 创建目录（幂等）
if err := os.MkdirAll(notesDir, 0o755); err != nil {
	return fmt.Errorf("创建备注目录失败: %w", err)
}

// 文件路径
notePath := filepath.Join(notesDir, fmt.Sprintf("%d.md", articleID))

// 读取源文件
content, err := os.ReadFile(filePath)
if err != nil {
	return fmt.Errorf("读取文件失败: %w", err)
}

// 写入目标文件（覆盖）
if err := os.WriteFile(notePath, content, 0o644); err != nil {
	return fmt.Errorf("写入备注失败: %w", err)
}

// 删除文件
if err := os.Remove(notePath); err != nil && !os.IsNotExist(err) {
	return fmt.Errorf("删除备注失败: %w", err)
}
```

**Flag 定义模式** (article.go:66-74):
```go
// 在 NewNoteCmd 中添加顶层命令 flags
cmd.Flags().Int64("article-id", 0, "文章 ID（必填）")
cmd.Flags().String("file", "", "备注文件路径（必填）")
cmd.MarkFlagRequired("article-id")
cmd.MarkFlagRequired("file")
```

---

### `internal/storage/database.go` (database method, CRUD)

**Analog:** `internal/storage/database.go` (existing methods)

**Schema 迁移模式** (database.go:109-167):
在 `ensureMigrations()` 函数中添加新字段迁移：
```go
// 在 ensureMigrations() 函数末尾添加：

// Add has_note column if it doesn't exist
if !db.columnExists("articles", "has_note") {
	if _, err := db.conn.Exec(`ALTER TABLE articles ADD COLUMN has_note BOOLEAN DEFAULT FALSE`); err != nil {
		return fmt.Errorf("failed to add has_note column: %w", err)
	}
}
```

**GetByID 查询模式** (database.go:606-616):
```go
// GetArticleByID returns an article by its ID, or nil if not found.
func (db *Database) GetArticleByID(id int64) (*model.Article, error) {
	row := db.conn.QueryRow(`SELECT id, blog_id, title, url, thumbnail_url, published_date, discovered_date, is_read FROM articles WHERE id = ?`, id)
	return scanArticle(row)
}
```

**Update 方法模式** (database.go:567-589, 1050-1053):
```go
// UpdateArticleHasNote updates the has_note field for an article.
func (db *Database) UpdateArticleHasNote(id int64, hasNote bool) error {
	_, err := db.conn.Exec(`UPDATE articles SET has_note = ? WHERE id = ?`, hasNote, id)
	return err
}

// 返回受影响行数检查（可选）：
func (db *Database) UpdateArticleHasNote(id int64, hasNote bool) (bool, error) {
	result, err := db.conn.Exec(`UPDATE articles SET has_note = ? WHERE id = ?`, hasNote, id)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}
```

**事务模式** (database.go:659-691, 217-277):
如需原子操作（写入文件 + 更新数据库），参考事务模式：
```go
func (db *Database) UpdateArticleHasNote(id int64, hasNote bool) error {
	_, err := db.conn.Exec(`UPDATE articles SET has_note = ? WHERE id = ?`, hasNote, id)
	return err
}
// 如果需要事务：
tx, err := db.conn.Begin()
if err != nil {
	return err
}
defer func() {
	if err != nil {
		_ = tx.Rollback()
	}
}()
// ... 操作 ...
if err = tx.Commit(); err != nil {
	return fmt.Errorf("commit: %w", err)
}
```

---

### `internal/model/model.go` (model, N/A)

**Analog:** `internal/model/model.go` (existing Article struct)

**结构体扩展模式** (model.go:16-25):
```go
type Article struct {
	ID             int64
	BlogID         int64
	Title          string
	URL            string
	ThumbnailURL   string
	PublishedDate  *time.Time
	DiscoveredDate *time.Time
	IsRead         bool
	HasNote        bool  // 新增字段
}

// 如果 ArticleWithBlog 也需要显示备注状态：
type ArticleWithBlog struct {
	ID             int64
	BlogID         int64
	Title          string
	URL            string
	ThumbnailURL   string
	PublishedDate  *time.Time
	DiscoveredDate *time.Time
	IsRead         bool
	BlogName       string
	BlogURL        string
	HasNote        bool  // 新增字段
}
```

**Scan 方法更新** (database.go:854-892):
需要更新 `scanArticle` 和 `scanArticleWithBlog` 以包含 has_note 字段：
```go
func scanArticle(scanner interface{ Scan(dest ...any) error }) (*model.Article, error) {
	var (
		id            int64
		blogID        int64
		title         string
		url           string
		thumbnailURL  sql.NullString
		publishedDate sql.NullString
		discovered    sql.NullString
		isRead        bool
		hasNote       bool  // 新增
	)
	if err := scanner.Scan(&id, &blogID, &title, &url, &thumbnailURL, &publishedDate, &discovered, &isRead, &hasNote); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	article := &model.Article{
		ID:           id,
		BlogID:       blogID,
		Title:        title,
		URL:          url,
		ThumbnailURL: thumbnailURL.String,
		IsRead:       isRead,
		HasNote:      hasNote,  // 新增
	}
	// ... 时间字段处理 ...
	return article, nil
}
```

**注意：** 添加字段后需要更新所有使用 `scanArticle` 的查询语句，在 SELECT 中添加 `has_note` 字段。

---

## Shared Patterns

### 数据库连接模式
**Source:** `internal/cli/commands/article.go` lines 119-136
**Apply to:** 所有 CLI 命令的 Run 函数
```go
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
```

### 错误处理与退出模式
**Source:** `internal/cli/commands/article.go` lines 125-136, 207-238
**Apply to:** 所有 CLI 命令
```go
// 错误消息格式
if err != nil {
	fmt.Fprintf(os.Stderr, "操作描述失败: %v\n", err)
	os.Exit(1)
}

// 成功消息格式
fmt.Printf("成功消息\n")
// 或
fmt.Printf("已成功操作文章 %d\n", id)
```

### 幂等迁移模式
**Source:** `internal/storage/database.go` lines 109-167
**Apply to:** 所有 schema 变更
```go
// 在 ensureMigrations() 中使用 columnExists() 检查
if !db.columnExists("articles", "has_note") {
	if _, err := db.conn.Exec(`ALTER TABLE articles ADD COLUMN has_note BOOLEAN DEFAULT FALSE`); err != nil {
		return fmt.Errorf("failed to add has_note column: %w", err)
	}
}
```

### 参数验证模式
**Source:** `internal/cli/commands/article.go` lines 241-252, 290-295
**Apply to:** 所有需要参数的命令
```go
// 使用 cmd.MarkFlagRequired() 标记必填参数
cmd.Flags().Int64("article-id", 0, "文章 ID（必填）")
cmd.MarkFlagRequired("article-id")

// 或在 Run 函数中手动验证
if articleID == 0 {
	fmt.Fprintf(os.Stderr, "必须提供文章 ID\n")
	os.Exit(1)
}
```

---

## No Analog Found

所有文件都有明确的现有类比，无需额外研究。

---

## Metadata

**Analog search scope:**
- `internal/cli/commands/` (所有 .go 文件)
- `internal/storage/database.go`
- `internal/model/model.go`
- `internal/cli/flags/flags.go`

**Files scanned:** 8 Go files

**Pattern extraction date:** 2026-05-07

**Key insights:**
1. CLI 命令遵循统一的 Cobra 模式：NewXxxCmd() + Run 函数 + 错误处理
2. 所有命令使用全局 `--db` flag (通过 flags.DBPath())
3. Schema 迁移使用幂等的 `columnExists()` 检查
4. 文件操作使用标准 Go 库 (os, io, filepath)
5. 数据库操作遵循现有方法的查询/更新模式