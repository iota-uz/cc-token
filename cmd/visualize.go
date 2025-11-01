package cmd

import (
	"fmt"

	"github.com/iota-uz/cc-token/internal/config"
	"github.com/iota-uz/cc-token/internal/visualizer"
	"github.com/spf13/cobra"
)

var visualizeCmd = &cobra.Command{
	Use:   "visualize [basic|interactive|html|json|plain] <file>",
	Short: "Visualize individual tokens in a file",
	Long: `Visualize individual tokens using client-side tokenization to extract token boundaries.

This command uses a client-side tokenizer which provides approximate token boundaries (94-98%
accurate for typical files) without requiring additional API calls. Token boundaries shown are
based on the go-tiktoken library, which closely matches Claude's tokenizer but may differ by
~6-8 tokens due to special tokens and encoding differences.

Visualization Modes:
  basic       - Display colored tokens in terminal output (ANSI colors)
  interactive - Launch web server with modern interactive UI (auto-opens browser)
  html        - Export to static HTML file (use --output to specify path)
  json        - Output structured JSON data (LLM-friendly, machine-readable)
  plain       - Output plain text with pipe delimiters (no ANSI colors)

The global --json flag can also be used with 'basic' or 'interactive' modes to override
the output format to JSON.

Note: Visualization only works with single files, not directories.`,
	Example: `  # Basic colored visualization
  cc-token visualize basic document.txt

  # Interactive web viewer (launches browser)
  cc-token visualize interactive README.md

  # Interactive web viewer (no auto-open browser)
  cc-token visualize interactive --no-browser README.md

  # Export to HTML file
  cc-token visualize html --output viz.html document.txt

  # Export and open in browser
  cc-token visualize html --output viz.html document.txt && open viz.html

  # JSON output (LLM-friendly)
  cc-token visualize json document.txt

  # Plain text output (pipe-friendly)
  cc-token visualize plain document.txt

  # Use JSON with global flag
  cc-token visualize basic --json document.txt

  # Skip confirmation prompt
  cc-token visualize basic --yes document.txt

  # Use cheaper Haiku model
  cc-token visualize json --model haiku code.py

  # Visualize from stdin
  echo "Hello, world!" | cc-token visualize json -`,
	Args: cobra.ExactArgs(2),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Set visualization mode from first argument
		mode := args[0]
		if !config.IsValidVisualizationMode(mode) {
			return fmt.Errorf("invalid mode: %s (must be 'basic', 'interactive', 'html', 'json', or 'plain')", mode)
		}
		cfg.Visualize = mode

		// Validate --output flag for html mode
		if mode == "html" && cfg.OutputFile == "" {
			return fmt.Errorf("html mode requires --output flag to specify the output file path")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// args[0] is the mode, args[1] is the file path
		path := args[1]

		// Create visualizer
		viz := visualizer.New(apiClient, pricingService)

		// Run visualization
		return viz.Run(path, cfg)
	},
}

func init() {
	rootCmd.AddCommand(visualizeCmd)
}
