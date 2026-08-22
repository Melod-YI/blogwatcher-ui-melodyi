package output

import (
	"strings"
	"testing"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/model"
)

func TestFormatTagTSV(t *testing.T) {
	tags := []model.Tag{
		{ID: 1, Name: "Go", ArticleCount: 3},
		{ID: 2, Name: "DB", ArticleCount: 0},
	}
	out := FormatTagTSV(tags)
	// schema 表头
	if !strings.HasPrefix(out, "id\tname\tarticles") {
		t.Fatalf("expected tsv header, got:\n%s", out)
	}
	// 数据行包含名称与计数
	if !strings.Contains(out, "Go") || !strings.Contains(out, "DB") {
		t.Fatalf("expected tag names in tsv, got:\n%s", out)
	}
	if !strings.Contains(out, "3") {
		t.Fatalf("expected article count in tsv, got:\n%s", out)
	}
}

func TestFormatTagTSV_Empty(t *testing.T) {
	out := FormatTagTSV(nil)
	if !strings.Contains(out, "没有标签") {
		t.Fatalf("expected empty hint, got: %s", out)
	}
}

func TestFormatTagJSON(t *testing.T) {
	tags := []model.Tag{
		{ID: 1, Name: "Go", ArticleCount: 3},
	}
	out := FormatTagJSON(tags)
	// json.MarshalIndent 在冒号后输出空格，故断言带空格
	if !strings.Contains(out, `"name": "Go"`) {
		t.Fatalf("expected name field, got: %s", out)
	}
	if !strings.Contains(out, `"article_count": 3`) {
		t.Fatalf("expected article_count field, got: %s", out)
	}
}

func TestFormatTagJSON_Empty(t *testing.T) {
	out := FormatTagJSON(nil)
	if out != "[]" {
		t.Fatalf("expected empty array, got: %s", out)
	}
}
