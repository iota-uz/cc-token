package visualizer

import (
	"fmt"

	"github.com/iota-uz/cc-token/internal/server"
)

// WebRenderer launches a web server for interactive visualization
type WebRenderer struct {
	NoBrowser bool // Whether to skip auto-opening browser
}

// Render starts a web server and serves the visualization
func (r *WebRenderer) Render(result *Result) error {
	if result == nil {
		return fmt.Errorf("result is nil")
	}

	// Create server instance
	srv, err := server.New()
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Convert visualizer.Result to server.Result
	serverResult := &server.Result{
		Content:     result.Content,
		Tokens:      result.Tokens,
		TotalTokens: result.TotalTokens,
		Model:       result.Model,
		Cost:        result.Cost,
	}

	// Start server (blocks until Ctrl+C)
	openBrowser := !r.NoBrowser
	if err := srv.Start(serverResult, openBrowser); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
