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

// OutputMode represents the different visualization output formats available.
// Each mode provides a different way to view and interact with tokenization results.
type OutputMode string

const (
	// ModeJSON outputs structured JSON data for machine-readable processing and LLM consumption
	ModeJSON OutputMode = "json"
	// ModePlain outputs plain text with pipe delimiters, suitable for piping to other commands
	ModePlain OutputMode = "plain"
	// ModeBasic displays colored tokens in the terminal using ANSI colors
	ModeBasic OutputMode = "basic"
	// ModeInteractive launches a web server with an interactive browser-based UI
	ModeInteractive OutputMode = "interactive"
	// ModeHTML exports visualization to a self-contained static HTML file
	ModeHTML OutputMode = "html"
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

// isNonInteractiveMode checks if the current configuration uses a non-interactive output mode
func isNonInteractiveMode(cfg *config.Config, mode string) bool {
	return cfg.JSONOutput || cfg.Plain || mode == "json" || mode == "plain" || mode == "html"
}

// ShouldSkipConfirmation determines if cost confirmation should be skipped
func ShouldSkipConfirmation(cfg *config.Config, mode string) bool {
	// Skip confirmation if --yes flag is set
	if cfg.SkipConfirmation {
		return true
	}

	// Auto-skip for non-interactive modes (JSON, plain text, HTML export)
	return isNonInteractiveMode(cfg, mode)
}
