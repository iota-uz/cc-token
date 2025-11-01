package visualizer

import (
	"fmt"
	"os"
	"strings"
)

// PlainRenderer outputs token visualization as plain text (no ANSI colors)
type PlainRenderer struct{}

// Render outputs the result as plain text with pipe delimiters
func (r *PlainRenderer) Render(result *Result) error {
	if result == nil {
		return fmt.Errorf("result is nil")
	}

	// Header
	fmt.Fprintln(os.Stdout, "Token Visualization (Plain Text)")
	fmt.Fprintln(os.Stdout, strings.Repeat("=", 80))
	fmt.Fprintf(os.Stdout, "Content Tokens: %d | API Tokens: %d (includes message overhead)\n",
		result.TotalTokens, result.APITokens)
	fmt.Fprintf(os.Stdout, "Characters: %d    Model: %s\n", len(result.Content), result.Model)
	fmt.Fprintf(os.Stdout, "Estimated Cost: $%.6f\n", result.Cost)
	fmt.Fprintln(os.Stdout, strings.Repeat("=", 80))
	fmt.Fprintln(os.Stdout)

	// Tokens with pipe delimiters
	fmt.Fprintln(os.Stdout, "Tokenized Text:")
	fmt.Fprintln(os.Stdout, strings.Repeat("-", 80))

	var tokenParts []string
	for _, token := range result.Tokens {
		tokenParts = append(tokenParts, token.Text)
	}

	// Join tokens with pipe delimiter
	fmt.Fprintln(os.Stdout, strings.Join(tokenParts, "|"))

	fmt.Fprintln(os.Stdout, strings.Repeat("-", 80))
	fmt.Fprintln(os.Stdout)

	// Detailed token list
	fmt.Fprintln(os.Stdout, "Token Details:")
	fmt.Fprintln(os.Stdout, strings.Repeat("-", 80))

	for i, token := range result.Tokens {
		// Escape special characters for display
		displayText := strings.ReplaceAll(token.Text, "\n", "\\n")
		displayText = strings.ReplaceAll(displayText, "\t", "\\t")
		displayText = strings.ReplaceAll(displayText, "\r", "\\r")

		fmt.Fprintf(os.Stdout, "[%d] \"%s\" (pos: %d, len: %d)\n",
			i, displayText, token.Position, token.Length)
	}

	fmt.Fprintln(os.Stdout, strings.Repeat("-", 80))
	fmt.Fprintf(os.Stdout, "\nContent Tokens: %d  |  API Tokens: %d  |  Overhead: %d\n",
		result.TotalTokens, result.APITokens, result.APITokens-result.TotalTokens)

	return nil
}
