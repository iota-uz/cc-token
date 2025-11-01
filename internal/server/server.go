// Package server provides HTTP server functionality for web-based token visualization.
package server

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iota-uz/cc-token/internal/api"
	"github.com/pkg/browser"
)

//go:embed templates/* static/*
var content embed.FS

// Result holds tokenization data for web visualization
type Result struct {
	Content     string
	Tokens      []api.Token
	TotalTokens int
	Model       string
	Cost        float64
}

// Server handles HTTP requests for token visualization
type Server struct {
	addr   string
	tmpl   *template.Template
	result *Result
}

// New creates a new Server instance with an available port
func New() (*Server, error) {
	port, err := findAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("failed to find available port: %w", err)
	}

	tmpl, err := template.New("visualize.html").Funcs(template.FuncMap{
		"colorIndex": func(i int) int {
			return i % 6
		},
		"colorName": func(i int) string {
			colors := []string{"cyan", "green", "yellow", "blue", "magenta", "red"}
			return colors[i%6]
		},
	}).ParseFS(content, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Server{
		addr: fmt.Sprintf("localhost:%d", port),
		tmpl: tmpl,
	}, nil
}

// Start launches the HTTP server and opens the browser
func (s *Server) Start(result *Result, openBrowser bool) error {
	s.result = result

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.Handle("/static/", http.FileServer(http.FS(content)))

	// Create server with graceful shutdown
	srv := &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	// Channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		fmt.Fprintf(os.Stderr, "\n✓ Visualization server started at http://%s\n", s.addr)
		fmt.Fprintf(os.Stderr, "✓ Press Ctrl+C to stop the server\n\n")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		}
	}()

	// Open browser if requested
	if openBrowser {
		time.Sleep(500 * time.Millisecond) // Give server time to start
		url := fmt.Sprintf("http://%s", s.addr)
		if err := browser.OpenURL(url); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Failed to open browser automatically: %v\n", err)
			fmt.Fprintf(os.Stderr, "   Please open manually: %s\n\n", url)
		}
	}

	// Wait for interrupt signal
	<-stop
	fmt.Fprintf(os.Stderr, "\n⏳ Shutting down server...\n")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✓ Server stopped\n")
	return nil
}

// handleIndex serves the main visualization page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := struct {
		Result *Result
	}{
		Result: s.result,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, "visualize.html", data); err != nil {
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
	}
}

// findAvailablePort finds an available port starting from 8080
func findAvailablePort() (int, error) {
	startPort := 8080
	maxAttempts := 100

	for i := 0; i < maxAttempts; i++ {
		port := startPort + i
		addr := fmt.Sprintf("localhost:%d", port)

		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available ports found in range %d-%d", startPort, startPort+maxAttempts)
}
