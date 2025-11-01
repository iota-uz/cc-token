package api

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

// parseStreamingResponse parses the SSE stream and extracts tokens based on text deltas
func parseStreamingResponse(reader io.Reader) ([]Token, error) {
	var tokens []Token
	scanner := bufio.NewScanner(reader)
	position := 0

	for scanner.Scan() {
		line := scanner.Text()

		// SSE format: "data: {...}"
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Remove "data: " prefix
		jsonData := strings.TrimPrefix(line, "data: ")

		// Skip ping events
		if jsonData == "[DONE]" {
			break
		}

		// Parse event
		var event StreamEvent

		if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
			continue // Skip malformed events
		}

		// Extract text deltas (each delta typically represents one token)
		if event.Type == "content_block_delta" && event.Delta.Type == "text_delta" {
			text := event.Delta.Text
			if text != "" {
				tokens = append(tokens, Token{
					Text:     text,
					Position: position,
					Length:   len(text),
				})
				position += len(text)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return tokens, nil
}
