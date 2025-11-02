package analyzer

import (
	"regexp"
	"strings"
	"unicode"
)

// detector_helpers.go contains shared helper functions and constants used by multiple detectors

// OOVPatternDetector defines a detector for out-of-vocabulary patterns
type OOVPatternDetector struct {
	Pattern   *regexp.Regexp
	Type      string
	MatchFunc func(string, int, string) *OOVStringIssue
}

const (
	invisibleCharContextSize = 20  // Context window size for invisible character detection
	minHashLength            = 32  // Minimum length for hash detection (MD5)
	hexCharRatioThreshold    = 0.8 // Hex character ratio to classify as hash
	minURLLengthForOOV       = 50  // Minimum URL length to flag as OOV
)

// zeroWidthCharMap maps zero-width and invisible Unicode characters to their type names
var zeroWidthCharMap = map[rune]string{
	0x200B: "zwsp", // Zero-Width Space
	0x200C: "zwnj", // Zero-Width Non-Joiner
	0x200D: "zwj",  // Zero-Width Joiner
	0x200E: "lrm",  // Left-to-Right Mark
	0x200F: "rlm",  // Right-to-Left Mark
	0x061C: "alm",  // Arabic Letter Mark
	0x00AD: "shy",  // Soft Hyphen
	0xFEFF: "bom",  // Byte Order Mark / Zero-Width No-Break Space
	0x2060: "wj",   // Word Joiner
}

// bidiControlCharMap maps bidirectional text control characters (Trojan Source)
var bidiControlCharMap = map[rune]string{
	0x202A: "lre", // Left-to-Right Embedding
	0x202B: "rle", // Right-to-Left Embedding
	0x202C: "pdf", // Pop Directional Formatting
	0x202D: "lro", // Left-to-Right Override
	0x202E: "rlo", // Right-to-Left Override
	0x2066: "lri", // Left-to-Right Isolate
	0x2067: "rli", // Right-to-Left Isolate
	0x2068: "fsi", // First Strong Isolate
	0x2069: "pdi", // Pop Directional Isolate
}

// suspiciousPatternSet contains patterns that may indicate evasion attempts
var suspiciousPatternSet = map[string]bool{
	"system:":     true,
	"prompt:":     true,
	"instruction": true,
	"should:":     true,
	"refuse":      true,
	"reject":      true,
	"don't":       true,
	"never":       true,
	"always":      true,
	"bypass":      true,
	"ignore":      true,
	"override":    true,
	"secret":      true,
}

// tryMergeIssueByLineAndType attempts to merge a new issue into existing issues by line number and type matcher
// Returns true if merged, false if should append as new
func tryMergeIssueByLineAndType[T any](issues []*T, lineNum int, typeMatcher func(*T, int) bool, incrementer func(*T)) bool {
	for _, existing := range issues {
		if typeMatcher(existing, lineNum) {
			incrementer(existing)
			return true
		}
	}
	return false
}

// extractContext extracts a context window around a position in a line
func extractContext(line string, pos int) string {
	start := pos - invisibleCharContextSize
	if start < 0 {
		start = 0
	}
	end := pos + invisibleCharContextSize
	if end > len(line) {
		end = len(line)
	}
	return line[start:end]
}

// detectLeetspeak checks if text contains leetspeak patterns
func detectLeetspeak(line string) bool {
	leetPatterns := []string{"1337", "h4x", "l33t", "w4nn4", "n00b", "pwn", "0wn", "d00d"}
	lower := strings.ToLower(line)
	for _, pattern := range leetPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

// deLeetspeak converts leetspeak text to normal text
func deLeetspeak(text string) string {
	replacements := map[string]string{
		"1": "l", "3": "e", "4": "a", "0": "o", "7": "t", "$": "s", "@": "a", "!": "i",
	}
	result := text
	for leet, normal := range replacements {
		result = strings.ReplaceAll(result, leet, normal)
	}
	return result
}

// detectROT13 checks if text is likely ROT13 encoded
// ROT13 inverts typical vowel frequency in English text
func detectROT13(text string) bool {
	// Skip short lines
	if len(text) < 20 {
		return false
	}

	// Count vowels - ROT13 typically has unusual vowel ratios
	vowels := "aeiouAEIOU"
	vowelCount := 0
	letterCount := 0

	for _, r := range text {
		if unicode.IsLetter(r) {
			letterCount++
			if strings.ContainsRune(vowels, r) {
				vowelCount++
			}
		}
	}

	if letterCount == 0 {
		return false
	}

	vowelRatio := float64(vowelCount) / float64(letterCount)

	// Normal English: 35-45% vowels
	// ROT13 English: vowel ratio inverts, becomes consonant-heavy or vowel-heavy
	return vowelRatio < 0.25 || vowelRatio > 0.60
}

// rot13Decode performs ROT13 decoding
func rot13Decode(text string) string {
	result := make([]rune, 0, len(text))
	for _, r := range text {
		switch {
		case 'a' <= r && r <= 'z':
			result = append(result, 'a'+(r-'a'+13)%26)
		case 'A' <= r && r <= 'Z':
			result = append(result, 'A'+(r-'A'+13)%26)
		default:
			result = append(result, r)
		}
	}
	return string(result)
}

// detectASCIIArt checks if a line is likely ASCII art
// ASCII art has high density of box-drawing and ASCII art characters
func detectASCIIArt(line string) bool {
	// Skip short lines
	if len(line) < 10 {
		return false
	}

	// Box-drawing characters (Unicode)
	boxDrawingChars := "─│┌┐└┘├┤┬┴┼═║╔╗╚╝╠╣╦╩╬"
	// Common ASCII art characters
	asciiArtChars := "|-+/\\*#@_<>[]{}()"

	artCharCount := 0
	for _, r := range line {
		if strings.ContainsRune(boxDrawingChars+asciiArtChars, r) {
			artCharCount++
		}
	}

	// If more than 50% of characters are art characters, likely ASCII art
	artRatio := float64(artCharCount) / float64(len(line))
	return artRatio > 0.5
}

// estimateTokenCost estimates token cost for various issue types
func estimateTokenCost(issueType string, value string) int {
	switch issueType {
	// Emoji types
	case "zwj_sequence":
		return 3 // ZWJ sequences are more expensive
	case "skin_tone":
		return 2 // Skin tone modifiers add tokens
	case "flag":
		return 2 // Flags use regional indicators
	case "emoji":
		return 1 // Standard emoji

	// OOV string types
	case "url":
		return (len(value) + 4) / 5 // 1 token per 4-5 characters
	case "uuid":
		return 5 // Standard UUID: ~4-6 tokens
	case "hash":
		return (len(value) + 3) / 4 // 1 token per 3-4 hex characters

	default:
		return 1
	}
}
