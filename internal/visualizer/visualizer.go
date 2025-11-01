package visualizer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/iota-uz/cc-token/internal/api"
	"github.com/iota-uz/cc-token/internal/config"
	"github.com/iota-uz/cc-token/internal/pricing"
)

// Visualizer handles token visualization workflows
type Visualizer struct {
	apiClient *api.Client
	pricer    *pricing.Pricer
}

// New creates a new Visualizer instance
func New(apiClient *api.Client, pricer *pricing.Pricer) *Visualizer {
	return &Visualizer{
		apiClient: apiClient,
		pricer:    pricer,
	}
}

// Run handles the visualization workflow for a single file
func (v *Visualizer) Run(path string, cfg *config.Config) error {
	// Handle stdin
	var content string
	if path == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		content = string(data)
	} else {
		// Read file
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("failed to access %s: %w", path, err)
		}

		if info.IsDir() {
			return fmt.Errorf("visualization only supports single files, not directories")
		}

		if info.Size() > cfg.MaxSize {
			return fmt.Errorf("file too large (%d bytes, max: %d bytes)", info.Size(), cfg.MaxSize)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		content = string(data)
	}

	// Get initial token count estimate for cost calculation
	estimatedTokens, err := v.apiClient.CountTokens(content, cfg.Model)
	if err != nil {
		return fmt.Errorf("failed to count tokens: %w", err)
	}

	// Show cost warning and get confirmation (unless skipped)
	if !ShouldSkipConfirmation(cfg, cfg.Visualize) {
		if !v.confirmVisualization(estimatedTokens, cfg.Model) {
			fmt.Println("Visualization cancelled.")
			return nil
		}
	}

	// Extract tokens using client-side tokenizer
	fmt.Fprintf(os.Stderr, "\nExtracting tokens using client-side tokenizer...\n")
	tokens, err := v.apiClient.ExtractTokensClientSide(content)
	if err != nil {
		return fmt.Errorf("failed to extract tokens: %w", err)
	}

	// Calculate cost (based on API token count which includes message overhead)
	contentTokens := len(tokens)
	cost := v.pricer.CalculateCost(estimatedTokens, cfg.Model)

	result := &Result{
		Content:     content,
		Tokens:      tokens,
		TotalTokens: contentTokens,   // Content-only tokens (what we visualize)
		APITokens:   estimatedTokens, // API count (includes message overhead)
		Model:       cfg.Model,
		Cost:        cost,
	}

	// Select and use appropriate renderer
	renderer, err := SelectRenderer(cfg, cfg.Visualize)
	if err != nil {
		return fmt.Errorf("failed to select renderer: %w", err)
	}

	return renderer.Render(result)
}

// confirmVisualization prompts the user to confirm they want to proceed with visualization
func (v *Visualizer) confirmVisualization(estimatedTokens int, model string) bool {
	// Calculate cost (same as count mode since we're using client-side tokenization)
	cost := v.pricer.CalculateCost(estimatedTokens, model)

	fmt.Fprintf(os.Stderr, "\n💡 Token Visualization\n")
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Fprintf(os.Stderr, "API tokens:         %d (exact)\n", estimatedTokens)
	fmt.Fprintf(os.Stderr, "Estimated cost:     $%.6f\n", cost)
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	fmt.Fprintf(os.Stderr, "Visualization uses client-side tokenization (no additional API cost).\n")
	fmt.Fprintf(os.Stderr, "Note: Token boundaries are approximate (94-98%% accurate for typical files).\n\n")
	fmt.Fprintf(os.Stderr, "Proceed with visualization? [Y/n]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	// Default to Yes if user just presses Enter
	return response == "" || response == "y" || response == "yes"
}
