package visualizer

import (
	"encoding/json"
	"fmt"
	"os"
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
	Content     string      `json:"content"`      // Original content
	Model       string      `json:"model"`        // Model used for tokenization
	TotalTokens int         `json:"total_tokens"` // Total number of tokens
	TotalChars  int         `json:"total_chars"`  // Total characters
	TotalBytes  int         `json:"total_bytes"`  // Total bytes
	Cost        float64     `json:"cost"`         // Estimated cost in USD
	Tokens      []TokenJSON `json:"tokens"`       // Array of individual tokens
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

	// Build result structure
	output := ResultJSON{
		Content:     result.Content,
		Model:       result.Model,
		TotalTokens: result.TotalTokens,
		TotalChars:  len(result.Content),
		TotalBytes:  len([]byte(result.Content)),
		Cost:        result.Cost,
		Tokens:      tokens,
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
