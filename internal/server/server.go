// ABOUTME: HTTP server package with dependency injection pattern
// ABOUTME: Manages server lifecycle, template parsing, and request routing
package server

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/esttorhe/blogwatcher-ui/v2/internal/service"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
)

// Server represents the HTTP server with all dependencies
type Server struct {
	db          *storage.Database
	blogService *service.BlogService
	templates   *template.Template
	mux         *http.ServeMux
	staticFS    fs.FS
	version     string
}

// NewServer creates a new HTTP server with dependency injection
// Parses all templates at startup and registers routes
// Deprecated: Use NewServerWithFS instead to provide embedded filesystems
func NewServer(db *storage.Database) (http.Handler, error) {
	return nil, fmt.Errorf("NewServer is deprecated, use NewServerWithFS with embedded filesystems")
}

// NewServerWithFS creates a new HTTP server with embedded filesystems
// Parses all templates at startup and registers routes
func NewServerWithFS(db *storage.Database, templateFS fs.FS, staticFS fs.FS, version string) (http.Handler, error) {
	// Register template functions BEFORE parsing templates
	funcMap := template.FuncMap{
		"timeAgo":       timeAgo,
		"faviconURL":    faviconURL,
		"smryURL":       smryURL,
		"renderMarkdown": renderMarkdown,
	}

	// Parse all templates once at startup from embedded filesystem
	tmpl := template.New("").Funcs(funcMap)

	// Walk the embedded template filesystem and parse all templates
	err := fs.WalkDir(templateFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := fs.ReadFile(templateFS, path)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", path, err)
		}
		_, err = tmpl.Parse(string(data))
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", path, err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	// Create server with dependencies
	s := &Server{
		db:          db,
		blogService: service.NewBlogService(db),
		templates:   tmpl,
		mux:         http.NewServeMux(),
		staticFS:    staticFS,
		version:     version,
	}

	// Register all routes
	s.registerRoutes()

	return s, nil
}

// ServeHTTP implements http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
