package analyzer

import (
	"regexp"
	"strings"

	"github.com/iota-uz/cc-token/internal/api"
	"github.com/iota-uz/cc-token/internal/utils"
)

// CategoryBreakdown shows token distribution by category
type CategoryBreakdown struct {
	Prose      int // Regular text content
	CodeBlocks int // Code blocks (``` markers)
	URLs       int // URLs
	Formatting int // Markdown formatting symbols
	Whitespace int // Empty lines and whitespace
	Total      int
}

// CategoryStats provides percentage breakdown
type CategoryStats struct {
	Prose      float64
	CodeBlocks float64
	URLs       float64
	Formatting float64
	Whitespace float64
}

var (
	codeBlockRegex      = regexp.MustCompile("```")
	markdownHeaderRegex = regexp.MustCompile(`^#{1,6}\s`)
	markdownListRegex   = regexp.MustCompile(`^[\s]*[-*+]\s`)
	markdownBoldRegex   = regexp.MustCompile(`\*\*[^*]+\*\*|__[^_]+__`)
	markdownItalicRegex = regexp.MustCompile(`\*[^*]+\*|_[^_]+_`)
	markdownLinkRegex   = regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`)
	markdownCodeRegex   = regexp.MustCompile("`[^`]+`")
)

// CategorizeTokens classifies tokens into categories
func CategorizeTokens(lines []string, tokens []api.Token, insights []*LineInsight) *CategoryBreakdown {
	breakdown := &CategoryBreakdown{}

	// Track code block state
	inCodeBlock := false
	codeBlockTokens := 0

	// Map tokens to lines
	lineTokenMap := make(map[int][]api.Token)
	lineStarts := utils.CalculateLineStarts(lines)

	for _, token := range tokens {
		lineIdx := utils.FindLineForPosition(token.Position, lineStarts)
		if lineIdx >= 0 && lineIdx < len(lines) {
			lineTokenMap[lineIdx] = append(lineTokenMap[lineIdx], token)
		}
	}

	// Process each line
	for i, line := range lines {
		lineTokens := lineTokenMap[i]
		lineTokenCount := len(lineTokens)

		// Check if this is an empty line
		if i < len(insights) && insights[i].IsEmpty {
			breakdown.Whitespace += lineTokenCount
			continue
		}

		// Check for code block markers
		if codeBlockRegex.MatchString(line) {
			inCodeBlock = !inCodeBlock
			breakdown.Formatting += lineTokenCount
			continue
		}

		// If in code block, count as code
		if inCodeBlock {
			codeBlockTokens += lineTokenCount
			continue
		}

		// Categorize tokens on this line
		urlTokens := 0
		formattingTokens := 0

		// URLs
		if urlMatches := urlRegex.FindAllString(line, -1); len(urlMatches) > 0 {
			for _, url := range urlMatches {
				urlTokens += utils.EstimateTokens(url)
			}
		}

		// Markdown formatting
		formattingChars := 0

		// Headers (#)
		if markdownHeaderRegex.MatchString(line) {
			headerMatch := markdownHeaderRegex.FindString(line)
			formattingChars += len(headerMatch)
		}

		// Lists (-, *, +)
		if markdownListRegex.MatchString(line) {
			listMatch := markdownListRegex.FindString(line)
			formattingChars += len(listMatch)
		}

		// Bold/italic markers
		formattingChars += countFormattingChars(line, markdownBoldRegex)
		formattingChars += countFormattingChars(line, markdownItalicRegex)

		// Links [text](url)
		formattingChars += countLinkFormatting(line)

		// Inline code `code`
		formattingChars += countFormattingChars(line, markdownCodeRegex)

		// Estimate formatting tokens (rough)
		formattingTokens = formattingChars / 4
		if formattingTokens > lineTokenCount {
			formattingTokens = lineTokenCount / 4 // Cap at 25% of line
		}

		// Cap URL tokens
		if urlTokens > lineTokenCount {
			urlTokens = lineTokenCount / 2
		}

		// Distribute tokens
		proseTokens := lineTokenCount - urlTokens - formattingTokens
		if proseTokens < 0 {
			proseTokens = lineTokenCount / 2
			urlTokens = lineTokenCount / 4
			formattingTokens = lineTokenCount - proseTokens - urlTokens
		}

		breakdown.URLs += urlTokens
		breakdown.Formatting += formattingTokens
		breakdown.Prose += proseTokens
	}

	// Add code block tokens
	breakdown.CodeBlocks = codeBlockTokens

	// Calculate total
	breakdown.Total = breakdown.Prose + breakdown.CodeBlocks + breakdown.URLs +
		breakdown.Formatting + breakdown.Whitespace

	return breakdown
}

// GetStats calculates percentage breakdown
func (c *CategoryBreakdown) GetStats() *CategoryStats {
	if c.Total == 0 {
		return &CategoryStats{}
	}

	return &CategoryStats{
		Prose:      float64(c.Prose) / float64(c.Total) * 100,
		CodeBlocks: float64(c.CodeBlocks) / float64(c.Total) * 100,
		URLs:       float64(c.URLs) / float64(c.Total) * 100,
		Formatting: float64(c.Formatting) / float64(c.Total) * 100,
		Whitespace: float64(c.Whitespace) / float64(c.Total) * 100,
	}
}

// Helper functions

func countFormattingChars(line string, regex *regexp.Regexp) int {
	matches := regex.FindAllString(line, -1)
	count := 0
	for _, match := range matches {
		// Count only the formatting characters, not content
		// For **bold** count 4 stars, for *italic* count 2 stars
		if strings.HasPrefix(match, "**") || strings.HasPrefix(match, "__") {
			count += 4 // ** at start and end
		} else if strings.HasPrefix(match, "*") || strings.HasPrefix(match, "_") {
			count += 2 // * at start and end
		} else if strings.HasPrefix(match, "`") {
			count += 2 // ` at start and end
		}
	}
	return count
}

func countLinkFormatting(line string) int {
	matches := markdownLinkRegex.FindAllString(line, -1)
	count := 0
	for range matches {
		// Count brackets and parentheses: []()
		count += 4
	}
	return count
}
