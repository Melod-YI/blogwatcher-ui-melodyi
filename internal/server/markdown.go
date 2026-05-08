// ABOUTME: Defines renderMarkdown template function for HTML rendering.
// ABOUTME: Uses goldmark library with GFM extensions for Markdown rendering.
package server

import (
	"bytes"
	"html/template"
	"log"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

// renderMarkdown converts Markdown content to HTML using goldmark with GFM extensions.
// Returns template.HTML to avoid auto-escaping in Go templates.
// GFM extensions enable: Tables, Strikethrough, TaskLists, Autolinks.
func renderMarkdown(content string) template.HTML {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(content), &buf); err != nil {
		log.Printf("Error rendering markdown: %v", err)
		// Return escaped content to prevent XSS
		return template.HTML(template.HTMLEscapeString(content))
	}
	return template.HTML(buf.String())
}
