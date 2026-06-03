// ABOUTME: Main entry point for the BlogWatcher HTTP server
// ABOUTME: Handles graceful shutdown and server lifecycle management
package main

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/esttorhe/blogwatcher-ui/v2/assets"
	_ "github.com/esttorhe/blogwatcher-ui/v2/internal/processor/setup"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/server"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/storage"
	"github.com/esttorhe/blogwatcher-ui/v2/internal/version"
)

func run(ctx context.Context) error {
	// Open database
	db, err := storage.OpenDatabase("")
	if err != nil {
		return err
	}
	defer db.Close()

	// Extract static files from embedded FS
	staticFiles, err := fs.Sub(assets.StaticFS, "static")
	if err != nil {
		return err
	}

	// Extract templates from embedded FS
	templateFiles, err := fs.Sub(assets.TemplateFS, "templates")
	if err != nil {
		return err
	}

	// Create server with embedded filesystems
	handler, err := server.NewServerWithFS(db, templateFiles, staticFiles, version.Version)
	if err != nil {
		return err
	}

	// Configure HTTP server with timeouts
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		log.Println("Shutdown signal received")
	case err := <-serverErr:
		return err
	}

	// Graceful shutdown with 10s timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	log.Println("Server stopped gracefully")
	return nil
}

func main() {
	// Create context with signal handling for SIGINT and SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
