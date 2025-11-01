// Package output provides output formatting for token counting results.
package output

import (
	"github.com/iota-uz/cc-token/internal/config"
	"github.com/iota-uz/cc-token/internal/pricing"
	"github.com/iota-uz/cc-token/internal/processor"
)

// Formatter interface for output formatting
type Formatter interface {
	Format(results []*processor.Result, cfg *config.Config) error
}

// OutputResults formats and outputs the token counting results in either tree or JSON format
// based on the configuration.
func OutputResults(results []*processor.Result, cfg *config.Config, pricer *pricing.Pricer) error {
	var formatter Formatter

	if cfg.JSONOutput {
		formatter = NewJSONFormatter(pricer)
	} else {
		formatter = NewTreeFormatter(pricer)
	}

	return formatter.Format(results, cfg)
}
