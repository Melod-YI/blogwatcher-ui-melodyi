// ABOUTME: Route registration for the HTTP server
// ABOUTME: Uses Go 1.22+ method routing with http.ServeMux
package server

import (
	"net/http"
)

// registerRoutes sets up all HTTP routes for the server
func (s *Server) registerRoutes() {
	// Static files from embedded filesystem (already extracted in main.go)
	s.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(s.staticFS))))

	// Pages
	s.mux.HandleFunc("GET /", s.handleIndex)
	s.mux.HandleFunc("GET /articles", s.handleArticleList)
	s.mux.HandleFunc("GET /blogs", s.handleBlogList)
	s.mux.HandleFunc("GET /blogs/grouped", s.handleBlogListGrouped)
	s.mux.HandleFunc("GET /settings", s.handleSettings)
	s.mux.HandleFunc("GET /note/{id}", s.handleNote)

	// Article management actions
	s.mux.HandleFunc("POST /articles/{id}/read", s.handleMarkRead)
	s.mux.HandleFunc("POST /articles/{id}/unread", s.handleMarkUnread)
	s.mux.HandleFunc("POST /articles/mark-all-read", s.handleMarkAllRead)

	// Sync
	s.mux.HandleFunc("POST /sync", s.handleSync)
	s.mux.HandleFunc("POST /api/sync", s.handleAPISync)

	// Blog management
	s.mux.HandleFunc("POST /blogs/add", s.handleAddBlog)
	s.mux.HandleFunc("POST /blogs/preview", s.handleBlogPreview) // PREV-02~04: Preview feed before adding
	s.mux.HandleFunc("GET /blogs/{id}", s.handleGetBlog)
	s.mux.HandleFunc("GET /blogs/{id}/edit", s.handleEditBlog)
	s.mux.HandleFunc("PUT /blogs/{id}", s.handleUpdateBlogName)
	s.mux.HandleFunc("DELETE /blogs/{id}", s.handleDeleteBlog)

	// Category management
	s.mux.HandleFunc("GET /categories/new", s.handleCategoriesNew)
	s.mux.HandleFunc("GET /categories/list", s.handleCategoriesList)
	s.mux.HandleFunc("GET /categories/{id}", s.handleGetCategory)
	s.mux.HandleFunc("POST /categories", s.handleCategoriesCreate)
	s.mux.HandleFunc("GET /categories/{id}/edit", s.handleCategoryEdit)
	s.mux.HandleFunc("PUT /categories/{id}", s.handleCategoryUpdate)
	s.mux.HandleFunc("DELETE /categories/{id}", s.handleCategoryDelete)
}
