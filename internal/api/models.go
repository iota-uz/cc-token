// Package api provides HTTP client functionality for interacting with Anthropic's Claude API.
package api

// Request represents the token counting API request
type Request struct {
	Model    string         `json:"model"`
	Messages []MessageInput `json:"messages"`
}

// MessageInput represents a message in the API request
type MessageInput struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Response represents the token counting API response
type Response struct {
	InputTokens int `json:"input_tokens"`
}

// StreamingRequest represents the streaming messages API request
type StreamingRequest struct {
	Model       string         `json:"model"`
	Messages    []MessageInput `json:"messages"`
	MaxTokens   int            `json:"max_tokens"`
	Stream      bool           `json:"stream"`
	Temperature float64        `json:"temperature,omitempty"`
}

// StreamEvent represents a server-sent event from the streaming API
type StreamEvent struct {
	Type  string `json:"type"`
	Index int    `json:"index,omitempty"`
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta,omitempty"`
}

// Token represents a single token with its text and position
type Token struct {
	Text     string
	Position int // Character position in the original text
	Length   int // Character length
}
