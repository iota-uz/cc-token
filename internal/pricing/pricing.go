// Package pricing handles model pricing, cost calculation, and model alias resolution for cc-token.
package pricing

import "strings"

// Model pricing (USD per 1M tokens - input pricing)
// Source: https://www.anthropic.com/pricing (as of 2025-11-01)
var modelPricing = map[string]float64{
	// Claude 4.x models
	"claude-sonnet-4-5": 3.00,  // Claude Sonnet 4.5
	"claude-sonnet-4.5": 3.00,  // Alternate format
	"claude-haiku-4-5":  1.00,  // Claude Haiku 4.5
	"claude-haiku-4.5":  1.00,  // Alternate format
	"claude-opus-4-1":   15.00, // Claude Opus 4.1
	"claude-opus-4.1":   15.00, // Alternate format
	"claude-sonnet-4":   3.00,  // Claude Sonnet 4
	"claude-4-sonnet":   3.00,  // Alternate format
	"claude-opus-4":     15.00, // Generic Claude Opus 4 (fallback to 4.1 pricing)
	"claude-haiku-4":    1.00,  // Generic Claude Haiku 4 (fallback to 4.5 pricing)

	// Claude 3.x models
	"claude-haiku-3-5":  0.80,  // Claude Haiku 3.5
	"claude-3-5-haiku":  0.80,  // Alternate format
	"claude-haiku-3.5":  0.80,  // Alternate format
	"claude-sonnet-3-7": 3.00,  // Claude Sonnet 3.7 (legacy)
	"claude-3-7-sonnet": 3.00,  // Alternate format
	"claude-sonnet-3.7": 3.00,  // Alternate format
	"claude-3-5-sonnet": 3.00,  // Claude Sonnet 3.5 (legacy, same as 3.7)
	"claude-sonnet-3-5": 3.00,  // Alternate format
	"claude-sonnet-3.5": 3.00,  // Alternate format
	"claude-opus-3":     15.00, // Claude Opus 3 (legacy)
	"claude-3-opus":     15.00, // Alternate format
	"claude-haiku-3":    0.25,  // Claude Haiku 3 (legacy)
	"claude-3-haiku":    0.25,  // Alternate format
	"claude-sonnet-3":   3.00,  // Claude Sonnet 3 (legacy)
	"claude-3-sonnet":   3.00,  // Alternate format
}

const (
	// DefaultModel is the default model to use for token counting
	DefaultModel = "claude-sonnet-4-5"
)

// Pricer handles cost calculations for token counts
type Pricer struct{}

// New creates a new Pricer instance
func New() *Pricer {
	return &Pricer{}
}

// CalculateCost estimates the API cost for the given number of tokens using the specified model.
// It returns the cost in USD based on the model's pricing per million input tokens.
func (p *Pricer) CalculateCost(tokens int, model string) float64 {
	pricePerMillion, ok := modelPricing[model]
	if !ok {
		pricePerMillion = 3.00 // Default to Sonnet pricing
	}
	return float64(tokens) * pricePerMillion / 1_000_000
}

// ResolveModelAlias converts short model aliases (haiku, sonnet, opus) to their full
// model names. It performs case-insensitive matching and returns the original model
// name if no alias is found.
func (p *Pricer) ResolveModelAlias(model string) string {
	// Map of short aliases to full model names (latest versions)
	aliases := map[string]string{
		"sonnet": "claude-sonnet-4-5", // Latest Sonnet (Claude 4.5)
		"haiku":  "claude-haiku-4-5",  // Latest Haiku (Claude 4.5)
		"opus":   "claude-opus-4-1",   // Latest Opus (Claude 4.1)
	}

	// Convert to lowercase for case-insensitive matching
	modelLower := strings.ToLower(strings.TrimSpace(model))

	if resolved, ok := aliases[modelLower]; ok {
		return resolved
	}

	// Return original if no alias found
	return model
}
