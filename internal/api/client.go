package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	countTokensURL   = "https://api.anthropic.com/v1/messages/count_tokens"
	streamingURL     = "https://api.anthropic.com/v1/messages"
	apiVersion       = "2023-06-01"
	defaultTimeout   = 30 * time.Second
	streamingTimeout = 120 * time.Second
)

// Client handles HTTP communication with Anthropic API
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new API client with the given API key
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
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

// ExtractTokensViaStreaming uses the streaming API to extract individual token boundaries.
// It sends a message asking Claude to echo the content, and parses the streaming response
// to identify token boundaries based on the text deltas.
func (c *Client) ExtractTokensViaStreaming(content, model string) ([]Token, error) {
	// Create streaming request
	// We ask Claude to repeat the content to extract tokens
	reqBody := StreamingRequest{
		Model: model,
		Messages: []MessageInput{
			{
				Role:    "user",
				Content: "Please repeat the following text exactly as it appears, character by character, without any changes:\n\n" + content,
			},
		},
		MaxTokens:   8192, // Reasonable limit
		Stream:      true,
		Temperature: 0, // Deterministic output
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal streaming request: %w", err)
	}

	// Create HTTP client with longer timeout for streaming
	streamClient := &http.Client{
		Timeout: streamingTimeout,
	}

	req, err := http.NewRequest("POST", streamingURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create streaming request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", apiVersion)
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := streamClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("streaming API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("streaming API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse SSE stream
	tokens, err := parseStreamingResponse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse streaming response: %w", err)
	}

	return tokens, nil
}
