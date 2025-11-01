package output

import (
	"encoding/json"
	"os"

	"github.com/iota-uz/cc-token/internal/config"
	"github.com/iota-uz/cc-token/internal/pricing"
	"github.com/iota-uz/cc-token/internal/processor"
)

// JSONFormatter formats output as JSON
type JSONFormatter struct {
	pricingService *pricing.Pricer
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter(pricingService *pricing.Pricer) *JSONFormatter {
	return &JSONFormatter{pricingService: pricingService}
}

// Format outputs results in JSON format
func (f *JSONFormatter) Format(results []*processor.Result, cfg *config.Config) error {
	output := make([]map[string]interface{}, 0, len(results))

	for _, result := range results {
		item := map[string]interface{}{
			"path":   result.Path,
			"tokens": result.Tokens,
		}

		if result.Error != nil {
			item["error"] = result.Error.Error()
		}

		if result.Cached {
			item["cached"] = true
		}

		if result.IsDir {
			item["type"] = "directory"
			item["files"] = result.CountFiles()
		} else {
			item["type"] = "file"
			// Add line metrics for files
			if result.LineCount > 0 {
				item["line_count"] = result.LineCount
				item["avg_tokens_per_line"] = result.AvgTokensPerLine
			}
		}

		if cfg.ShowCost {
			item["estimated_cost"] = f.pricingService.CalculateCost(result.Tokens, cfg.Model)
		}

		output = append(output, item)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
