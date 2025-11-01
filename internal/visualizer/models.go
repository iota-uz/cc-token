// Package visualizer provides token visualization capabilities for cc-token.
package visualizer

import "github.com/iota-uz/cc-token/internal/api"

// Result holds tokenization data for visualization
type Result struct {
	Content     string
	Tokens      []api.Token
	TotalTokens int // Total number of content tokens (from visualization)
	APITokens   int // API token count (includes message overhead)
	Model       string
	Cost        float64 // Estimated cost in USD
}
