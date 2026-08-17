package output

import (
	"strings"
	"testing"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
)

// TestArticleOutputTags 验证文章标签在 table/json/simple 三种格式中正确渲染。
// 覆盖 article get / article list 输出对标签的展示集成。
func TestArticleOutputTags(t *testing.T) {
	articles := []model.ArticleWithBlog{
		{
			ID: 1, Title: "Tagged", BlogName: "B",
			Tags: []model.Tag{{ID: 1, Name: "Go"}, {ID: 2, Name: "DB"}},
		},
	}
	meta := PaginationMeta{Total: 1, Count: 1}

	// table：Tags 列含两个标签名（逗号拼接）
	tbl := FormatTable(articles, meta)
	if !strings.Contains(tbl, "Tags") {
		t.Fatalf("table header missing Tags column:\n%s", tbl)
	}
	if !strings.Contains(tbl, "Go") || !strings.Contains(tbl, "DB") {
		t.Fatalf("table row missing tag names:\n%s", tbl)
	}

	// json：tags 字段含两个标签名
	js := FormatJSON(articles, meta)
	if !strings.Contains(js, `"tags"`) || !strings.Contains(js, `"Go"`) || !strings.Contains(js, `"DB"`) {
		t.Fatalf("json missing tags field/names:\n%s", js)
	}

	// simple：#Go #DB 后缀
	smp := FormatSimple(articles, meta)
	if !strings.Contains(smp, "#Go") || !strings.Contains(smp, "#DB") {
		t.Fatalf("simple missing tag suffix:\n%s", smp)
	}
}

// TestArticleOutputNoTags 验证无标签文章不输出标签信息（json omitempty、table/simple 空占位）。
func TestArticleOutputNoTags(t *testing.T) {
	articles := []model.ArticleWithBlog{
		{ID: 2, Title: "Plain", BlogName: "B"},
	}
	meta := PaginationMeta{Total: 1, Count: 1}

	// json：无 tags 字段（omitempty）
	js := FormatJSON(articles, meta)
	if strings.Contains(js, `"tags"`) {
		t.Fatalf("json should omit tags when empty:\n%s", js)
	}

	// table/simple：行内不应出现 "#"
	smp := FormatSimple(articles, meta)
	if strings.Contains(smp, "#") {
		t.Fatalf("simple should not render tag suffix when empty:\n%s", smp)
	}
}
