package visualizer

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// JSONRenderer outputs token visualization in JSON format (LLM-friendly)
type JSONRenderer struct{}

// TokenJSON represents a single token in JSON output
type TokenJSON struct {
	Index    int    `json:"index"`     // Token index (0-based)
	Text     string `json:"text"`      // Token text content
	Position int    `json:"position"`  // Start position in original content
	Length   int    `json:"length"`    // Length in characters
	ByteSize int    `json:"byte_size"` // Length in bytes
}

// ResultJSON represents the complete visualization result in JSON format
type ResultJSON struct {
	Content       string      `json:"content"`         // Original content
	Model         string      `json:"model"`           // Model used for tokenization
	ContentTokens int         `json:"content_tokens"`  // Content-only tokens (visualized)
	APITokens     int         `json:"api_tokens"`      // API token count (includes overhead)
	TotalChars    int         `json:"total_chars"`     // Total characters
	TotalBytes    int         `json:"total_bytes"`     // Total bytes
	TotalLines    int         `json:"total_lines"`     // Total lines in content
	TokensPerLine float64     `json:"tokens_per_line"` // Average tokens per line
	Cost          float64     `json:"cost"`            // Estimated cost in USD
	Tokens        []TokenJSON `json:"tokens"`          // Array of individual tokens
}

// Render outputs the result as formatted JSON
func (r *JSONRenderer) Render(result *Result) error {
	if result == nil {
		return fmt.Errorf("result is nil")
	}

	// Build tokens array
	tokens := make([]TokenJSON, len(result.Tokens))
	for i, token := range result.Tokens {
		tokens[i] = TokenJSON{
			Index:    i,
			Text:     token.Text,
			Position: token.Position,
			Length:   token.Length,
			ByteSize: len(token.Text),
		}
	}

	// Calculate line count and tokens per line
	lineCount := strings.Count(result.Content, "\n")
	if len(result.Content) > 0 && !strings.HasSuffix(result.Content, "\n") {
		lineCount++ // Count last line if content doesn't end with newline
	}
	if lineCount == 0 {
		lineCount = 1 // Minimum one line for non-empty content
	}

	tokensPerLine := 0.0
	if lineCount > 0 {
		tokensPerLine = float64(result.TotalTokens) / float64(lineCount)
	}

	// Build result structure
	output := ResultJSON{
		Content:       result.Content,
		Model:         result.Model,
		ContentTokens: result.TotalTokens,
		APITokens:     result.APITokens,
		TotalChars:    len(result.Content),
		TotalBytes:    len([]byte(result.Content)),
		TotalLines:    lineCount,
		TokensPerLine: tokensPerLine,
		Cost:          result.Cost,
		Tokens:        tokens,
	}

	// Marshal to JSON with indentation for readability
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to stdout
	fmt.Fprintln(os.Stdout, string(jsonData))

	return nil
}
