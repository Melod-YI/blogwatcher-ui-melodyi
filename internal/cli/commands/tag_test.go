// ABOUTME: Tests for tag CLI command group
// ABOUTME: 采用子进程模式端到端验证 stdout/stderr/exit code（run 函数直接 os.Exit，需子进程才能断言非 0 退出）
package commands

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
)

// tagSubprocEnv 置位时，TestMain 不跑测试，而是把本测试二进制当作 blogwatcher CLI 执行。
const tagSubprocEnv = "BLOGWATCHER_TAG_SUBPROC=1"

// TestMain：当环境变量 tagSubprocEnv 置位时，走 CLI 入口；否则正常跑测试。
// 这样可以在 run 函数直接调用 os.Exit 的情况下，仍能通过子进程断言退出码与输出。
func TestMain(m *testing.M) {
	if os.Getenv("BLOGWATCHER_TAG_SUBPROC") == "1" {
		Execute()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// runTagCLI 重新执行测试二进制以调用 CLI。
// 为了规避 std flag.Parse 对未知 flag 的报错，子命令置于 --db 之前。
func runTagCLI(t *testing.T, dbPath string, args ...string) (string, string, int) {
	t.Helper()
	// 形如: <binary> tag list --db <path>
	full := append(append([]string{}, args...), "--db", dbPath)
	cmd := exec.Command(os.Args[0], full...)
	cmd.Env = append(os.Environ(), tagSubprocEnv)
	var out, errB bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errB
	err := cmd.Run()
	exitCode := 0
	if ee, ok := err.(*exec.ExitError); ok {
		exitCode = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run cli: %v", err)
	}
	return out.String(), errB.String(), exitCode
}

// newTagTestDB 创建一个空临时 DB 并返回路径。
func newTagTestDB(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "blogwatcher.db")
	db, err := storage.OpenDatabase(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.Close()
	return path
}

func openTagTestDB(t *testing.T, path string) *storage.Database {
	t.Helper()
	db, err := storage.OpenDatabase(path)
	if err != nil {
		t.Fatalf("reopen db: %v", err)
	}
	return db
}

func TestTagList_Table(t *testing.T) {
	path := newTagTestDB(t)
	db := openTagTestDB(t, path)
	if _, err := db.CreateTag("Go"); err != nil {
		t.Fatalf("create tag: %v", err)
	}
	if _, err := db.CreateTag("DB"); err != nil {
		t.Fatalf("create tag: %v", err)
	}
	db.Close()

	stdout, stderr, code := runTagCLI(t, path, "tag", "list")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Go") || !strings.Contains(stdout, "DB") {
		t.Fatalf("expected tags in table output, got: %s", stdout)
	}
}

func TestTagList_JSON(t *testing.T) {
	path := newTagTestDB(t)
	db := openTagTestDB(t, path)
	if _, err := db.CreateTag("Go"); err != nil {
		t.Fatalf("create tag: %v", err)
	}
	db.Close()

	stdout, _, code := runTagCLI(t, path, "tag", "list", "--format", "json")
	if code != 0 {
		t.Fatalf("expected exit 0")
	}
	if !strings.Contains(stdout, `"name": "Go"`) {
		t.Fatalf("expected json name field, got: %s", stdout)
	}
	if !strings.Contains(stdout, `"article_count": 0`) {
		t.Fatalf("expected article_count field, got: %s", stdout)
	}
}

func TestTagList_Empty(t *testing.T) {
	path := newTagTestDB(t)
	stdout, _, code := runTagCLI(t, path, "tag", "list")
	if code != 0 {
		t.Fatalf("expected exit 0")
	}
	if !strings.Contains(stdout, "没有标签") {
		t.Fatalf("expected empty hint, got: %s", stdout)
	}
}

func TestTagRename_Success(t *testing.T) {
	path := newTagTestDB(t)
	db := openTagTestDB(t, path)
	if _, err := db.CreateTag("Go"); err != nil {
		t.Fatalf("create tag: %v", err)
	}
	db.Close()

	stdout, stderr, code := runTagCLI(t, path, "tag", "rename", "Go", "Golang")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Go -> Golang") {
		t.Fatalf("expected rename message, got: %s", stdout)
	}
	db = openTagTestDB(t, path)
	got, err := db.GetTagByName("Golang")
	if err != nil || got == nil {
		t.Fatalf("expected renamed tag in db, err=%v got=%+v", err, got)
	}
	if old, _ := db.GetTagByName("Go"); old != nil {
		t.Fatalf("old name should not exist anymore")
	}
	db.Close()
}

func TestTagRename_NotFound(t *testing.T) {
	path := newTagTestDB(t)
	_, stderr, code := runTagCLI(t, path, "tag", "rename", "Missing", "New")
	if code == 0 {
		t.Fatalf("expected non-zero exit for missing tag")
	}
	if !strings.Contains(stderr, "不存在") {
		t.Fatalf("expected not-found error in stderr, got: %s", stderr)
	}
}

func TestTagDelete_Success(t *testing.T) {
	path := newTagTestDB(t)
	db := openTagTestDB(t, path)
	tag, err := db.CreateTag("Go")
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}
	// 建一篇文章并关联该标签，验证 affected 计数
	blog, err := db.AddBlog(model.Blog{Name: "B", URL: "https://example.com"})
	if err != nil {
		t.Fatalf("add blog: %v", err)
	}
	if _, _, err := db.AddArticlesBulk([]model.Article{{
		BlogID: blog.ID, Title: "A", URL: "https://example.com/a", HNStatus: model.HNStatusNotSearch,
	}}); err != nil {
		t.Fatalf("add articles: %v", err)
	}
	arts, err := db.ListArticles(false, nil)
	if err != nil || len(arts) != 1 {
		t.Fatalf("list articles: err=%v len=%d", err, len(arts))
	}
	if err := db.AddArticleTag(arts[0].ID, tag.ID); err != nil {
		t.Fatalf("add article tag: %v", err)
	}
	db.Close()

	stdout, stderr, code := runTagCLI(t, path, "tag", "delete", "Go")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "解除 1 篇文章关联") {
		t.Fatalf("expected affected count message, got: %s", stdout)
	}
	db = openTagTestDB(t, path)
	if got, _ := db.GetTagByName("Go"); got != nil {
		t.Fatalf("tag should be deleted, got: %+v", got)
	}
	db.Close()
}

func TestTagDelete_NotFound(t *testing.T) {
	path := newTagTestDB(t)
	_, stderr, code := runTagCLI(t, path, "tag", "delete", "Missing")
	if code == 0 {
		t.Fatalf("expected non-zero exit for missing tag")
	}
	if !strings.Contains(stderr, "不存在") {
		t.Fatalf("expected not-found error in stderr, got: %s", stderr)
	}
}

func TestTag_NoSubcommand_Help(t *testing.T) {
	path := newTagTestDB(t)
	stdout, _, code := runTagCLI(t, path, "tag")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "标签管理命令") {
		t.Fatalf("expected help output containing description, got: %s", stdout)
	}
	if !strings.Contains(stdout, "list") || !strings.Contains(stdout, "rename") || !strings.Contains(stdout, "delete") {
		t.Fatalf("expected subcommands in help, got: %s", stdout)
	}
}
