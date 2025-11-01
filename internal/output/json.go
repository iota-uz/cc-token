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
	pricer *pricing.Pricer
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter(pricer *pricing.Pricer) *JSONFormatter {
	return &JSONFormatter{pricer: pricer}
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
			item["files"] = countFilesForJSON(result)
		} else {
			item["type"] = "file"
		}

		if cfg.ShowCost {
			item["estimated_cost"] = f.pricer.CalculateCost(result.Tokens, cfg.Model)
		}

		output = append(output, item)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func countFilesForJSON(result *processor.Result) int {
	if !result.IsDir {
		if result.Error == nil {
			return 1
		}
		return 0
	}

	count := 0
	for _, child := range result.Children {
		count += countFilesForJSON(child)
	}
	return count
}
