// Package config provides configuration structures for cc-token CLI tool.
package config

import "fmt"

var (
	// ValidVisualizationModes defines all supported visualization modes
	ValidVisualizationModes = map[string]bool{
		"basic":       true,
		"interactive": true,
		"html":        true,
		"json":        true,
		"plain":       true,
	}
)

// Config holds CLI configuration
type Config struct {
	Model            string
	Extensions       []string
	MaxSize          int64
	Concurrency      int
	ShowCost         bool
	JSONOutput       bool
	Verbose          bool
	NoCache          bool
	Visualize        string // "basic", "interactive", "html", "json", "plain", or empty string
	SkipConfirmation bool   // Skip cost confirmation prompts (for automation)
	Plain            bool   // Use plain text output (no ANSI colors)
	OutputFile       string // Output file path for HTML export
	NoBrowser        bool   // Skip auto-opening browser for web modes
}

// IsValidVisualizationMode checks if the given mode is a valid visualization mode
func IsValidVisualizationMode(mode string) bool {
	return ValidVisualizationModes[mode]
}

// Validate checks if the configuration values are valid
func (c *Config) Validate() error {
	if c.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be greater than 0")
	}
	if c.MaxSize <= 0 {
		return fmt.Errorf("max-size must be greater than 0")
	}
	if c.Visualize != "" && !IsValidVisualizationMode(c.Visualize) {
		return fmt.Errorf("invalid visualization mode: %s (must be 'basic', 'interactive', 'html', 'json', or 'plain')", c.Visualize)
	}
	return nil
}
