// ABOUTME: Tests for article tag/untag/list --tag CLI subcommands
// ABOUTME: 复用 tag_test.go 的子进程范式（TestMain + runTagCLI 已在同包定义）
package commands

import (
	"strconv"
	"strings"
	"testing"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
)

// itoaID 将 int64 文章 ID 转为字符串（CLI 参数）。
func itoaID(id int64) string { return strconv.FormatInt(id, 10) }

// newArticleWithTagTestDB 创建临时 DB 并插入一篇文章，返回 DB 路径与文章 ID。
func newArticleWithTagTestDB(t *testing.T) (path string, articleID int64) {
	t.Helper()
	path = newTagTestDB(t)
	db := openTagTestDB(t, path)
	defer db.Close()

	blog, err := db.AddBlog(model.Blog{Name: "TestBlog", URL: "https://example.com"})
	if err != nil {
		t.Fatalf("add blog: %v", err)
	}
	if _, _, err := db.AddArticlesBulk([]model.Article{{
		BlogID: blog.ID, Title: "TaggedArticle", URL: "https://example.com/a", HNStatus: model.HNStatusNotSearch,
	}, {
		BlogID: blog.ID, Title: "UntaggedArticle", URL: "https://example.com/b", HNStatus: model.HNStatusNotSearch,
	}}); err != nil {
		t.Fatalf("add articles: %v", err)
	}
	arts, err := db.ListArticles(false, nil)
	if err != nil || len(arts) != 2 {
		t.Fatalf("list articles: err=%v len=%d", err, len(arts))
	}
	// 返回第一篇（TaggedArticle）的 ID
	for _, a := range arts {
		if a.Title == "TaggedArticle" {
			return path, a.ID
		}
	}
	t.Fatalf("TaggedArticle not found in list")
	return path, 0
}

func TestArticleTag_Success(t *testing.T) {
	path, id := newArticleWithTagTestDB(t)

	stdout, stderr, code := runTagCLI(t, path, "article", "tag", itoaID(id), "Go")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "添加标签") || !strings.Contains(stdout, "Go") {
		t.Fatalf("expected add-tag message, got: %s", stdout)
	}

	// 重开 DB 验证关联与标签
	db := openTagTestDB(t, path)
	defer db.Close()
	tag, err := db.GetTagByName("Go")
	if err != nil || tag == nil {
		t.Fatalf("expected tag 'Go' in db, err=%v tag=%+v", err, tag)
	}
	tags, err := db.GetArticleTags(id)
	if err != nil {
		t.Fatalf("get article tags: %v", err)
	}
	if len(tags) != 1 || tags[0].Name != "Go" {
		t.Fatalf("expected 1 tag 'Go' on article, got: %+v", tags)
	}
}

func TestArticleTag_Idempotent(t *testing.T) {
	path, id := newArticleWithTagTestDB(t)

	// 第一次打标签
	if _, stderr, code := runTagCLI(t, path, "article", "tag", itoaID(id), "Go"); code != 0 {
		t.Fatalf("first tag exit %d stderr=%s", code, stderr)
	}
	// 第二次重复打同一标签：应幂等不报错，关联仍 1 条
	stdout, stderr, code := runTagCLI(t, path, "article", "tag", itoaID(id), "Go")
	if code != 0 {
		t.Fatalf("second tag expected exit 0, got %d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "添加标签") {
		t.Fatalf("expected add-tag message, got: %s", stdout)
	}

	db := openTagTestDB(t, path)
	defer db.Close()
	tags, err := db.GetArticleTags(id)
	if err != nil {
		t.Fatalf("get article tags: %v", err)
	}
	if len(tags) != 1 {
		t.Fatalf("expected 1 tag after idempotent re-tag, got %d: %+v", len(tags), tags)
	}
}

func TestArticleTag_ArticleNotFound(t *testing.T) {
	path := newTagTestDB(t)

	_, stderr, code := runTagCLI(t, path, "article", "tag", "9999", "Go")
	if code == 0 {
		t.Fatalf("expected non-zero exit for missing article")
	}
	if !strings.Contains(stderr, "不存在") {
		t.Fatalf("expected not-found error in stderr, got: %s", stderr)
	}
}

func TestArticleUntag_Success(t *testing.T) {
	path, id := newArticleWithTagTestDB(t)

	// 先打标签
	if _, _, code := runTagCLI(t, path, "article", "tag", itoaID(id), "Go"); code != 0 {
		t.Fatalf("pre-tag exit %d", code)
	}
	// 再移除
	stdout, stderr, code := runTagCLI(t, path, "article", "untag", itoaID(id), "Go")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "移除") || !strings.Contains(stdout, "Go") {
		t.Fatalf("expected remove-tag message, got: %s", stdout)
	}

	// 重开 DB 验证关联已清空，标签本身仍在
	db := openTagTestDB(t, path)
	defer db.Close()
	tags, err := db.GetArticleTags(id)
	if err != nil {
		t.Fatalf("get article tags: %v", err)
	}
	if len(tags) != 0 {
		t.Fatalf("expected no tags on article after untag, got: %+v", tags)
	}
	tag, err := db.GetTagByName("Go")
	if err != nil || tag == nil {
		t.Fatalf("expected tag 'Go' still exists, err=%v tag=%+v", err, tag)
	}
}

func TestArticleUntag_TagNotFound(t *testing.T) {
	path, id := newArticleWithTagTestDB(t)

	_, stderr, code := runTagCLI(t, path, "article", "untag", itoaID(id), "Missing")
	if code == 0 {
		t.Fatalf("expected non-zero exit for missing tag")
	}
	if !strings.Contains(stderr, "不存在") {
		t.Fatalf("expected not-found error in stderr, got: %s", stderr)
	}
}

func TestArticleList_TagFilter(t *testing.T) {
	path, id := newArticleWithTagTestDB(t)

	// 给 TaggedArticle 打标签
	if _, _, code := runTagCLI(t, path, "article", "tag", itoaID(id), "Go"); code != 0 {
		t.Fatalf("pre-tag exit %d", code)
	}

	// --tag Go 只返回 TaggedArticle，不含 UntaggedArticle
	stdout, stderr, code := runTagCLI(t, path, "article", "list", "--tag", "Go")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "TaggedArticle") {
		t.Fatalf("expected TaggedArticle in output, got: %s", stdout)
	}
	if strings.Contains(stdout, "UntaggedArticle") {
		t.Fatalf("UntaggedArticle should not appear, got: %s", stdout)
	}
}
