// Package visualizer provides token visualization renderers.
package visualizer

import (
	"fmt"

	"github.com/iota-uz/cc-token/internal/config"
)

// Renderer defines the interface for rendering token visualization results.
type Renderer interface {
	// Render outputs the visualization result to stdout
	Render(result *Result) error
}

// OutputMode represents different output rendering modes
type OutputMode string

const (
	ModeJSON        OutputMode = "json"        // JSON structured output
	ModePlain       OutputMode = "plain"       // Plain text output (no ANSI)
	ModeBasic       OutputMode = "basic"       // Colored terminal output
	ModeInteractive OutputMode = "interactive" // Web-based interactive viewer
	ModeHTML        OutputMode = "html"        // Static HTML export
)

// SelectRenderer chooses the appropriate renderer based on configuration
func SelectRenderer(cfg *config.Config, mode string) (Renderer, error) {
	// Priority: JSON flag > Plain flag > specified mode
	if cfg.JSONOutput {
		return &JSONRenderer{}, nil
	}

	if cfg.Plain {
		return &PlainRenderer{}, nil
	}

	// Use mode from command argument
	switch mode {
	case "basic":
		return &BasicRenderer{}, nil
	case "interactive":
		return &WebRenderer{NoBrowser: cfg.NoBrowser}, nil
	case "html":
		// For HTML mode, OutputFile must be provided (validated in cmd layer)
		return &HTMLRenderer{
			OutputFile:  cfg.OutputFile,
			OpenBrowser: false, // Manual control via shell commands
		}, nil
	case "json":
		return &JSONRenderer{}, nil
	case "plain":
		return &PlainRenderer{}, nil
	default:
		return nil, fmt.Errorf("unknown visualization mode: %s", mode)
	}
}

// ShouldSkipConfirmation determines if cost confirmation should be skipped
func ShouldSkipConfirmation(cfg *config.Config, mode string) bool {
	// Skip confirmation if --yes flag is set
	if cfg.SkipConfirmation {
		return true
	}

	// Auto-skip for non-interactive modes (JSON, plain text, HTML export)
	if cfg.JSONOutput || cfg.Plain || mode == "json" || mode == "plain" || mode == "html" {
		return true
	}

	return false
}
