package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iota-uz/cc-token/internal/config"
	"github.com/iota-uz/cc-token/internal/pricing"
	"github.com/iota-uz/cc-token/internal/processor"
)

// TreeFormatter formats output as a tree view
type TreeFormatter struct {
	pricer *pricing.Pricer
}

// NewTreeFormatter creates a new tree formatter
func NewTreeFormatter(pricer *pricing.Pricer) *TreeFormatter {
	return &TreeFormatter{pricer: pricer}
}

// Format outputs results in tree format
func (f *TreeFormatter) Format(results []*processor.Result, cfg *config.Config) error {
	totalTokens := 0
	totalFiles := 0

	for _, result := range results {
		if result.IsDir {
			printTreeNode(result, "", cfg.Verbose)
			totalTokens += result.Tokens
			totalFiles += result.CountFiles()
		} else {
			if result.Error != nil {
				fmt.Fprintf(os.Stderr, "%s: ERROR - %v\n", result.Path, result.Error)
			} else {
				cachedMark := ""
				if cfg.Verbose && result.Cached {
					cachedMark = " (cached)"
				}
				fmt.Printf("%s: %d tokens%s\n", result.Path, result.Tokens, cachedMark)
				totalTokens += result.Tokens
				totalFiles++
			}
		}
	}

	// Print summary
	if len(results) > 1 || (len(results) == 1 && results[0].IsDir) {
		fmt.Println(strings.Repeat("-", 50))
		fmt.Printf("Total: %d tokens across %d files\n", totalTokens, totalFiles)

		if cfg.ShowCost {
			cost := f.pricer.CalculateCost(totalTokens, cfg.Model)
			fmt.Printf("Estimated cost: $%.6f\n", cost)
		}
	} else if cfg.ShowCost && totalTokens > 0 {
		cost := f.pricer.CalculateCost(totalTokens, cfg.Model)
		fmt.Printf("Estimated cost: $%.6f\n", cost)
	}

	return nil
}

func printTreeNode(node *processor.Result, prefix string, verbose bool) {
	basePath := filepath.Base(node.Path)
	if node.IsDir && len(node.Children) > 0 {
		fmt.Printf("%s%s/\n", prefix, basePath)

		for i, child := range node.Children {
			isLast := i == len(node.Children)-1
			childPrefix := prefix + "  "

			if child.Error != nil {
				fmt.Fprintf(os.Stderr, "%s%s: ERROR - %v\n", childPrefix, filepath.Base(child.Path), child.Error)
			} else {
				cachedMark := ""
				if verbose && child.Cached {
					cachedMark = " (cached)"
				}

				connector := "├─"
				if isLast {
					connector = "└─"
				}

				fmt.Printf("%s%s %s: %d tokens%s\n", prefix, connector, filepath.Base(child.Path), child.Tokens, cachedMark)
			}
		}
	}
}
