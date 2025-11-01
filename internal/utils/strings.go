package utils

// Truncate shortens a string to maxLen characters, adding "..." if truncated
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// EstimateTokens provides a rough token count estimate using the 4-char-per-token heuristic
// This is less accurate than API tokenization but useful for quick estimates
func EstimateTokens(text string) int {
	return (len(text) + 3) / 4
}
