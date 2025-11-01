package cmd

import (
	"fmt"
	"os"

	"github.com/iota-uz/cc-token/internal/analyzer"
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
  cc-token count file1.txt file2.txt dir1/

  # Analyze token optimization opportunities
  cc-token count --analyze document.txt`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle --analyze flag (files only)
		if cfg.Analyze {
			if len(args) != 1 {
				return fmt.Errorf("--analyze flag requires exactly one file argument")
			}

			path := args[0]

			// Check if it's stdin
			if path == "-" {
				return fmt.Errorf("--analyze flag does not support stdin input")
			}

			// Check if it's a file (not directory)
			info, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("failed to access %s: %w", path, err)
			}
			if info.IsDir() {
				return fmt.Errorf("--analyze flag only works with individual files, not directories")
			}

			// Read file content
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			// Get accurate token count from API
			tokens, err := apiClient.CountTokens(string(content), cfg.Model)
			if err != nil {
				return fmt.Errorf("failed to count tokens: %w", err)
			}

			// Perform analysis
			analysis, err := analyzer.AnalyzeFile(string(content), tokens, apiClient)
			if err != nil {
				return fmt.Errorf("failed to analyze file: %w", err)
			}

			// Format and output analysis
			formatter := output.NewAnalysisFormatter(!cfg.Plain)
			return formatter.FormatAnalysis(analysis, path, cfg)
		}

		// Normal count mode
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
		return output.OutputResults(results, cfg, pricingService)
	},
}

func init() {
	rootCmd.AddCommand(countCmd)
}
