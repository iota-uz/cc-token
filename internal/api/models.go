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

// Token represents a single token with its text and position
type Token struct {
	Text     string
	Position int // Character position in the original text
	Length   int // Character length
}
