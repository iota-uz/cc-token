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

	// Extract tokens via streaming
	fmt.Fprintf(os.Stderr, "\nExtracting tokens via streaming API...\n")
	tokens, err := v.apiClient.ExtractTokensViaStreaming(content, cfg.Model)
	if err != nil {
		return fmt.Errorf("failed to extract tokens: %w", err)
	}

	// Calculate cost (input + output tokens)
	totalTokens := len(tokens)
	cost := v.pricer.CalculateStreamingCost(estimatedTokens, totalTokens, cfg.Model)

	result := &Result{
		Content:     content,
		Tokens:      tokens,
		TotalTokens: totalTokens,
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
// given the cost difference compared to simple token counting
func (v *Visualizer) confirmVisualization(estimatedTokens int, model string) bool {
	// Estimate costs
	countingCost := v.pricer.CalculateCost(estimatedTokens, model)
	// For streaming, we estimate output will be similar to input
	streamingCost := v.pricer.CalculateStreamingCost(estimatedTokens, estimatedTokens, model)

	fmt.Fprintf(os.Stderr, "\n⚠️  Token Visualization Cost Warning\n")
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Fprintf(os.Stderr, "Token counting (current):  $%.6f\n", countingCost)
	fmt.Fprintf(os.Stderr, "Token visualization:       $%.6f\n", streamingCost)
	fmt.Fprintf(os.Stderr, "Cost difference:           $%.6f (%.1fx more expensive)\n",
		streamingCost-countingCost, streamingCost/countingCost)
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	fmt.Fprintf(os.Stderr, "Visualization uses the streaming API to extract individual tokens.\n")
	fmt.Fprintf(os.Stderr, "This requires a full message generation, which costs more than counting.\n\n")
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
