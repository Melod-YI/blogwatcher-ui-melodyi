package output

import (
	"strings"
	"testing"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
)

// TestArticleOutputTags 验证文章标签在 TSV/JSON 两种格式中正确渲染。
func TestArticleOutputTags(t *testing.T) {
	articles := []model.ArticleWithBlog{
		{
			ID: 1, Title: "Tagged", BlogName: "B",
			Tags: []model.Tag{{ID: 1, Name: "Go"}, {ID: 2, Name: "DB"}},
		},
	}
	meta := PaginationMeta{Total: 1, Count: 1}

	// tsv：schema 含 tags 列，数据行 tags 单元格逗号拼接
	tsv := FormatTSV(articles, meta)
	if !strings.Contains(tsv, "id\ttitle\tblog\tstatus\tfav\ttags\tpublished") {
		t.Fatalf("tsv schema missing tags column:\n%s", tsv)
	}
	if !strings.Contains(tsv, "Go,DB") {
		t.Fatalf("tsv row missing comma-joined tags:\n%s", tsv)
	}

	// json：tags 字段含两个标签名
	js := FormatJSON(articles, meta)
	if !strings.Contains(js, `"tags"`) || !strings.Contains(js, `"Go"`) || !strings.Contains(js, `"DB"`) {
		t.Fatalf("json missing tags field/names:\n%s", js)
	}
}

// TestArticleOutputNoTags 验证无标签文章：tsv tags 单元格为空、json 省略 tags 字段。
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

	// tsv：schema 仍含 tags 列，但数据行该单元格为空
	tsv := FormatTSV(articles, meta)
	if !strings.Contains(tsv, "\ttags\t") {
		t.Fatalf("tsv schema should still have tags column:\n%s", tsv)
	}
	// 数据行不应出现 "Go" 之类标签名
	if strings.Contains(tsv, "Plain\tB\t未读\t\t") == false {
		// 仅做结构性断言：fav 与 tags 两列均为空（连续两个 \t）
		t.Fatalf("tsv row should have empty fav and tags cells:\n%s", tsv)
	}
}

// TestArticleTSV_Pagination 验证补充信息行：count/total/has_more。
func TestArticleTSV_Pagination(t *testing.T) {
	articles := []model.ArticleWithBlog{
		{ID: 1, Title: "A", BlogName: "B"},
	}
	// total > count：应给出 total；has_more=true
	meta := PaginationMeta{Total: 15, Count: 1, Offset: 0, HasMore: true}
	out := FormatTSV(articles, meta)
	if !strings.Contains(out, "count=1") {
		t.Fatalf("expected count line:\n%s", out)
	}
	if !strings.Contains(out, "total=15") {
		t.Fatalf("expected total line when total>count:\n%s", out)
	}
	if !strings.Contains(out, "has_more=true") {
		t.Fatalf("expected has_more line:\n%s", out)
	}

	// total == count：不应给出 total（与 count 重复）
	meta2 := PaginationMeta{Total: 1, Count: 1, HasMore: false}
	out2 := FormatTSV(articles, meta2)
	if strings.Contains(out2, "total=") {
		t.Fatalf("should omit total when total==count:\n%s", out2)
	}
	if !strings.Contains(out2, "has_more=false") {
		t.Fatalf("expected has_more=false:\n%s", out2)
	}
}
