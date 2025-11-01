// Package cmd provides the command-line interface for cc-token using Cobra.
package cmd

import (
	"fmt"
	"os"

	"github.com/iota-uz/cc-token/internal/api"
	"github.com/iota-uz/cc-token/internal/cache"
	"github.com/iota-uz/cc-token/internal/config"
	"github.com/iota-uz/cc-token/internal/pricing"
	"github.com/spf13/cobra"
)

const (
	version = "1.1.1" // Fixed tokenization discrepancy with client-side tokenizer

	// Default configuration values
	defaultMaxFileSize = 2 * 1024 * 1024 // 2MB
	defaultConcurrency = 5
)

var (
	cfg       *config.Config
	apiClient *api.Client
	cacheInst *cache.Cache
	pricer    *pricing.Pricer
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "cc-token",
	Short: "Claude token counting tool",
	Long: `cc-token is a CLI tool for counting tokens in files and directories using Anthropic's Claude API.
It supports caching, parallel processing, and multiple output formats.`,
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Normalize extensions to have leading dots
		for i, ext := range cfg.Extensions {
			if ext != "" && ext[0] != '.' {
				cfg.Extensions[i] = "." + ext
			}
		}

		// Validate configuration
		if err := cfg.Validate(); err != nil {
			return err
		}

		// Resolve model alias
		pricer = pricing.New()
		cfg.Model = pricer.ResolveModelAlias(cfg.Model)

		// Validate API key (except for cache clear command)
		if cmd.Name() != "clear" {
			apiKey := os.Getenv("ANTHROPIC_API_KEY")
			if apiKey == "" {
				return fmt.Errorf("ANTHROPIC_API_KEY environment variable is not set.\nGet your API key from: https://console.anthropic.com/")
			}

			// Initialize API client
			apiClient = api.NewClient(apiKey)
		}

		// Initialize cache
		if !cfg.NoCache && cmd.Name() != "clear" {
			var err error
			cacheInst, err = cache.Load()
			if err != nil && cfg.Verbose {
				fmt.Fprintf(os.Stderr, "Warning: Failed to load cache: %v\n", err)
			}
		}

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		// Save cache
		if cacheInst != nil && cmd.Name() != "clear" {
			if err := cacheInst.Save(); err != nil && cfg.Verbose {
				fmt.Fprintf(os.Stderr, "Warning: Failed to save cache: %v\n", err)
			}
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cfg = &config.Config{}

	// Global flags available to all commands
	rootCmd.PersistentFlags().StringVarP(&cfg.Model, "model", "m", pricing.DefaultModel, "Model to use for token counting (supports aliases: sonnet, haiku, opus)")
	rootCmd.PersistentFlags().StringSliceVarP(&cfg.Extensions, "ext", "e", []string{}, "File extensions to include (e.g., .go,.txt,.md)")
	rootCmd.PersistentFlags().Int64Var(&cfg.MaxSize, "max-size", defaultMaxFileSize, "Maximum file size in bytes (default: 2MB)")
	rootCmd.PersistentFlags().IntVarP(&cfg.Concurrency, "concurrency", "c", defaultConcurrency, "Number of concurrent API requests for directories")
	rootCmd.PersistentFlags().BoolVar(&cfg.ShowCost, "show-cost", true, "Show estimated API cost")
	rootCmd.PersistentFlags().BoolVarP(&cfg.JSONOutput, "json", "j", false, "Output results in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&cfg.NoCache, "no-cache", false, "Disable caching")
	rootCmd.PersistentFlags().BoolVarP(&cfg.SkipConfirmation, "yes", "y", false, "Skip confirmation prompts (for automation)")
	rootCmd.PersistentFlags().BoolVar(&cfg.Plain, "plain", false, "Use plain text output without ANSI colors")
	rootCmd.PersistentFlags().StringVarP(&cfg.OutputFile, "output", "o", "", "Output file path for HTML export")
	rootCmd.PersistentFlags().BoolVar(&cfg.NoBrowser, "no-browser", false, "Skip auto-opening browser for web visualization")
}
