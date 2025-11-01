package visualizer

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// BasicRenderer displays tokens with colored borders in the terminal
type BasicRenderer struct{}

// Constants for basic rendering
const (
	headerWidth = 80
)

// tokenColors defines the color palette for token boundaries (rainbow colors)
var tokenColors = []*color.Color{
	color.New(color.FgCyan),
	color.New(color.FgGreen),
	color.New(color.FgYellow),
	color.New(color.FgBlue),
	color.New(color.FgMagenta),
	color.New(color.FgRed),
}

// Render displays tokens with colored borders in the terminal
func (r *BasicRenderer) Render(result *Result) error {
	if result == nil {
		return fmt.Errorf("result is nil")
	}

	// Header
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, "Token Visualization")
	fmt.Fprintln(os.Stdout, strings.Repeat("=", headerWidth))
	fmt.Fprintf(os.Stdout, "Tokens: %d    Characters: %d    Model: %s\n",
		result.TotalTokens, len(result.Content), result.Model)
	if result.Cost > 0 {
		fmt.Fprintf(os.Stdout, "Estimated Cost: $%.6f\n", result.Cost)
	}
	fmt.Fprintln(os.Stdout, strings.Repeat("=", headerWidth))
	fmt.Fprintln(os.Stdout)

	// Render tokens with alternating colors
	for i, token := range result.Tokens {
		colorIndex := i % len(tokenColors)
		c := tokenColors[colorIndex]

		// Use bold color with brackets for token boundaries
		styledToken := c.Add(color.Bold).Sprintf("⎡%s⎦", token.Text)
		fmt.Fprint(os.Stdout, styledToken)
	}
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout)

	// Footer
	fmt.Fprintln(os.Stdout, strings.Repeat("=", headerWidth))
	fmt.Fprintf(os.Stdout, "Total: %d tokens\n", result.TotalTokens)
	fmt.Fprintln(os.Stdout, strings.Repeat("=", headerWidth))

	return nil
}

// RenderBasic is a legacy function that wraps BasicRenderer for backward compatibility
// Deprecated: Use BasicRenderer.Render() instead
func RenderBasic(result *Result) {
	renderer := &BasicRenderer{}
	if err := renderer.Render(result); err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering: %v\n", err)
	}
}
