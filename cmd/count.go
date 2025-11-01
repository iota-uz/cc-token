package cmd

import (
	"fmt"

	"github.com/iota-uz/cc-token/internal/output"
	"github.com/iota-uz/cc-token/internal/processor"
	"github.com/spf13/cobra"
)

var countCmd = &cobra.Command{
	Use:   "count [paths...]",
	Short: "Count tokens in files or directories",
	Long: `Count tokens in one or more files or directories using Claude's token counting API.

The count command processes files and directories, respecting .gitignore patterns and applying
configured filters. Results can be displayed in tree format or JSON.`,
	Example: `  # Count tokens in a single file
  cc-token count document.txt

  # Count tokens in current directory
  cc-token count .

  # Count only Go and Markdown files
  cc-token count --ext .go,.md src/

  # Use Haiku model (fastest/cheapest)
  cc-token count --model haiku file.txt

  # Output JSON format
  cc-token count --json . > tokens.json

  # Read from stdin
  cat file.txt | cc-token count -

  # Process multiple paths
  cc-token count file1.txt file2.txt dir1/`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create processor
		proc := processor.New(apiClient, cacheInst, cfg)

		// Process each path
		var results []*processor.Result
		for _, path := range args {
			result, err := proc.ProcessPath(path)
			if err != nil {
				return fmt.Errorf("failed to process %s: %w", path, err)
			}
			results = append(results, result)
		}

		// Output results
		return output.OutputResults(results, cfg, pricer)
	},
}

func init() {
	rootCmd.AddCommand(countCmd)
}
