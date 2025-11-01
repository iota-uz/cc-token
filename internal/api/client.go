package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hupe1980/go-tiktoken"
)

const (
	countTokensURL = "https://api.anthropic.com/v1/messages/count_tokens"
	apiVersion     = "2023-06-01"
	defaultTimeout = 30 * time.Second
)

// Client handles HTTP communication with Anthropic API and token encoding
type Client struct {
	apiKey     string
	httpClient *http.Client
	encoding   *tiktoken.Encoding
}

// NewClient creates a new API client with the given API key and initializes the Claude tokenizer
func NewClient(apiKey string) *Client {
	// Initialize Claude tokenizer for client-side token extraction
	codec, err := tiktoken.NewClaude()
	if err != nil {
		// Fallback to nil encoding if initialization fails
		// Token visualization will not work but token counting will
		fmt.Fprintf(os.Stderr, "Warning: Failed to initialize Claude tokenizer codec: %v\n", err)
		fmt.Fprintf(os.Stderr, "Token visualization features will be unavailable.\n")
		return &Client{
			apiKey: apiKey,
			httpClient: &http.Client{
				Timeout: defaultTimeout,
			},
			encoding: nil,
		}
	}

	encoding, err := tiktoken.NewEncoding(codec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to initialize tokenizer encoding: %v\n", err)
		fmt.Fprintf(os.Stderr, "Token visualization features will be unavailable.\n")
		return &Client{
			apiKey: apiKey,
			httpClient: &http.Client{
				Timeout: defaultTimeout,
			},
			encoding: nil,
		}
	}

	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: defaultTimeout},
		encoding:   encoding,
	}
}

// CountTokens calls the Anthropic API to count tokens in the given content using the specified model.
// It returns the number of input tokens or an error if the API request fails.
func (c *Client) CountTokens(content, model string) (int, error) {
	reqBody := Request{
		Model: model,
		Messages: []MessageInput{
			{Role: "user", Content: content},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", countTokensURL, bytes.NewReader(jsonData))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", apiVersion)
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return 0, fmt.Errorf("API returned status %d (failed to read response body: %w)", resp.StatusCode, readErr)
		}
		return 0, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp Response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return apiResp.InputTokens, nil
}

// ExtractTokensClientSide uses the client-side Claude tokenizer to extract individual tokens
// without making API calls. This is faster, cheaper, and works offline.
func (c *Client) ExtractTokensClientSide(content string) ([]Token, error) {
	if c.encoding == nil {
		return nil, fmt.Errorf("tokenizer not initialized")
	}

	// Encode the content to get token IDs and token strings
	_, tokenStrings, err := c.encoding.Encode(content, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to encode content: %w", err)
	}

	// Build Token structs with position and length information
	tokens := make([]Token, 0, len(tokenStrings))
	position := 0

	for _, tokenText := range tokenStrings {
		// Note: Token text might not match exactly in content due to BPE encoding (spaces, special chars)
		remainingContent := content[position:]
		idx := strings.Index(remainingContent, tokenText)

		if idx == -1 {
			// Token text doesn't appear literally in content (common with BPE encoding)
			tokenLength := len(tokenText)
			tokens = append(tokens, Token{
				Text:     tokenText,
				Position: position,
				Length:   tokenLength,
			})
			position += tokenLength
		} else {
			tokens = append(tokens, Token{
				Text:     tokenText,
				Position: position + idx,
				Length:   len(tokenText),
			})
			position += idx + len(tokenText)
		}
	}

	return tokens, nil
}
