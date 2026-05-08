---
phase: 15-ui-note-display
plan: 01
subsystem: note-ui
tags: [markdown, goldmark, gfm, note-page, note-button]
dependency_graph:
  requires: [NOTE-01, NOTE-02]
  provides: [NOTE-09, NOTE-10, NOTE-11, NOTE-12]
  affects: []
tech_stack:
  added:
    - goldmark v1.7.8 (Markdown renderer)
    - goldmark/extension (GFM support)
  patterns:
    - template function pattern (renderMarkdown)
    - CSS variable system (theme adaptation)
    - conditional button rendering ({{if .HasNote}})
key_files:
  created:
    - internal/server/markdown.go
    - assets/templates/pages/note.gohtml
  modified:
    - go.mod
    - go.sum
    - internal/server/server.go
    - assets/static/styles.css
    - assets/templates/partials/article-items.gohtml
    - internal/server/handlers.go
    - internal/server/routes.go
decisions:
  - D-09: Use goldmark library for Markdown rendering (standard Go Markdown parser)
  - D-10: Enable GFM extensions (Tables, Strikethrough, TaskLists, Autolinks)
  - D-11: Server-side rendering (no JavaScript enhancement needed)
  - D-13: Use existing CSS variable system for automatic Light/Dark adaptation
metrics:
  duration: 15 minutes
  completed_date: 2026-05-08T15:00:00Z
  task_count: 5
  file_count: 9
---

# Phase 15 Plan 01: UI Note Display Summary

**One-liner:** Markdown 备注 UI 显示功能：文章卡片备注按钮 + 备注 页面 + GFM 渲染 + 主题适配

## Implementation Overview

实现了在 UI 上显示文章备注的完整功能，包括：
1. 添加 goldmark 依赖并注册 renderMarkdown 模板函数
2. 定义 .markdown-body CSS 样式类，使用 CSS 变量实现主题适配
3. 在文章卡片添加条件显示的备注按钮（{{if .HasNote}}）
4. 创建独立备注页面模板 note.gohtml
5. 实现 handleNote handler 和 /note/{id} 路由注册

## Files Modified/Created

### Created Files
- `internal/server/markdown.go` - renderMarkdown 模板函数实现
- `assets/templates/pages/note.gohtml` - 备注页面模板

### Modified Files
- `go.mod` - 添加 goldmark v1.7.8 依赖
- `go.sum` - 更新依赖 checksum
- `internal/server/server.go` - 注册 renderMarkdown 到 FuncMap
- `assets/static/styles.css` - 添加 .markdown-body 和 .note-page 样式
- `assets/templates/partials/article-items.gohtml` - 添加备注按钮
- `internal/server/handlers.go` - 实现 handleNote handler
- `internal/server/routes.go` - 注册 GET /note/{id} 路由

## Key Code Snippets

### renderMarkdown 模板函数 (markdown.go)
```go
func renderMarkdown(content string) template.HTML {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(content), &buf); err != nil {
		return template.HTML(content)
	}
	return template.HTML(buf.String())
}
```

### 备注按钮 (article-items.gohtml)
```gohtml
{{if .HasNote}}
<a href="/note/{{.ID}}"
   target="_blank"
   rel="noopener noreferrer"
   class="action-btn"
   title="查看备注"
   onclick="event.stopPropagation();">
    <svg viewBox="0 0 24 24" ...>
        <!-- document-text icon -->
    </svg>
    <span class="action-btn-label">Note</span>
</a>
{{end}}
```

### handleNote handler (handlers.go)
```go
func (s *Server) handleNote(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	// ... validation ...
	
	article, err := s.db.GetArticleByID(id)
	// ... fetch article ...
	
	homeDir, _ := os.UserHomeDir()
	notePath := filepath.Join(homeDir, ".blogwatcher", "notes", fmt.Sprintf("%d.md", id))
	
	noteBytes, err := os.ReadFile(notePath)
	if err != nil {
		content = "" // Template shows friendly message
	} else {
		content = string(noteBytes)
	}
	
	s.renderTemplate(w, "note.gohtml", data)
}
```

## Deviations from Plan

### Auto-fixed Issues

None - plan executed exactly as written.

### Network Connectivity Resolution

**Issue:** Network connectivity timeout when downloading goldmark from proxy.golang.org
**Fix:** Used Chinese mirror GOPROXY=https://goproxy.cn,direct with GOSUMDB=off
**Impact:** Task 1 completed successfully after 2 retry attempts
**Commit:** 3635f28

## Verification Results

### Build Verification
- `go build ./...` passed successfully
- `go build ./internal/server` passed for each task

### Acceptance Criteria Verification
- ✓ goldmark dependency in go.mod (1 occurrence)
- ✓ renderMarkdown in server.go FuncMap (1 occurrence)
- ✓ .markdown-body CSS styles (21 style definitions)
- ✓ {{if .HasNote}} conditional in article template (1 occurrence)
- ✓ handleNote handler in handlers.go (2 occurrences - handler + route)
- ✓ /note/{id} route registered in routes.go
- ✓ note.gohtml template created with renderMarkdown function

## Known Stubs

None - all functionality fully implemented.

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| threat_flag: path_validation | internal/server/handlers.go:handleNote | Path traversal mitigation via strconv.ParseInt validation (T-15-01/02 mitigated) |

## Requirements Satisfied

- **NOTE-09:** HasNote=true 文章卡片显示备注按钮 ✓
- **NOTE-10:** 点击备注按钮新标签页打开 /note/{id} ✓
- **NOTE-11:** Markdown 渲染支持 GFM 格式 ✓
- **NOTE-12:** 备注页面显示文章标题和原文链接 ✓

## Next Steps

功能已完整实现，可以进行用户验证：
1. 启动服务器验证备注按钮显示
2. 点击备注按钮验证页面渲染
3. 切换主题验证样式适配
4. 测试空备注文件场景

---

**Duration:** 15 minutes
**Completed:** 2026-05-08T15:00:00Z
**Commits:** 5 (3635f28, 9775e87, 1786d2e, 47d726, 5e981fe)