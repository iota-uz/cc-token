package analyzer

import (
	"regexp"
	"strings"

	"github.com/iota-uz/cc-token/internal/utils"
)

const (
	// Minimum consecutive empty lines to detect
	minConsecutiveEmptyLines = 2
	// Default long line threshold in characters
	defaultLongLineThreshold = 120
)

// AdvancedPatterns holds detected advanced patterns
type AdvancedPatterns struct {
	URLs             []*URLPattern
	ConsecutiveEmpty []*ConsecutiveEmptyLines
	LongLines        []*LongLine
}

// URLPattern represents a detected URL
type URLPattern struct {
	URL         string
	Length      int
	Occurrences int
	LineNumbers []int
	TokenCost   int
}

// ConsecutiveEmptyLines represents a run of empty lines
type ConsecutiveEmptyLines struct {
	StartLine int
	EndLine   int
	Count     int
}

// LongLine represents a line that's unusually long
type LongLine struct {
	LineNumber int
	Length     int
	Tokens     int
	Content    string
}

var (
	urlRegex = regexp.MustCompile(`https?://[^\s\)]+`)
)

// DetectURLs finds all URLs in the content and analyzes them
func DetectURLs(lines []string) []*URLPattern {
	urlMap := make(map[string]*URLPattern)

	for i, line := range lines {
		matches := urlRegex.FindAllString(line, -1)
		for _, url := range matches {
			// Clean up URL (remove trailing punctuation)
			url = strings.TrimRight(url, ".,;:!?")

			if existing, found := urlMap[url]; found {
				existing.Occurrences++
				existing.LineNumbers = append(existing.LineNumbers, i+1)
			} else {
				urlMap[url] = &URLPattern{
					URL:         url,
					Length:      len(url),
					Occurrences: 1,
					LineNumbers: []int{i + 1},
					TokenCost:   utils.EstimateTokens(url),
				}
			}
		}
	}

	// Convert to slice
	result := make([]*URLPattern, 0, len(urlMap))
	for _, pattern := range urlMap {
		result = append(result, pattern)
	}

	return result
}

// DetectConsecutiveEmpty finds runs of consecutive empty lines
func DetectConsecutiveEmpty(insights []*LineInsight, minCount int) []*ConsecutiveEmptyLines {
	if minCount < minConsecutiveEmptyLines {
		minCount = minConsecutiveEmptyLines
	}

	result := make([]*ConsecutiveEmptyLines, 0)
	var currentRun *ConsecutiveEmptyLines

	for _, insight := range insights {
		if insight.IsEmpty {
			if currentRun == nil {
				currentRun = &ConsecutiveEmptyLines{
					StartLine: insight.LineNumber,
					EndLine:   insight.LineNumber,
					Count:     1,
				}
			} else {
				currentRun.EndLine = insight.LineNumber
				currentRun.Count++
			}
		} else {
			if currentRun != nil && currentRun.Count >= minCount {
				result = append(result, currentRun)
			}
			currentRun = nil
		}
	}

	// Don't forget the last run
	if currentRun != nil && currentRun.Count >= minCount {
		result = append(result, currentRun)
	}

	return result
}

// DetectLongLines finds lines that are unusually long
func DetectLongLines(insights []*LineInsight, threshold int) []*LongLine {
	if threshold == 0 {
		threshold = defaultLongLineThreshold
	}

	result := make([]*LongLine, 0)
	for _, insight := range insights {
		if insight.Chars > threshold && insight.Tokens > 0 {
			result = append(result, &LongLine{
				LineNumber: insight.LineNumber,
				Length:     insight.Chars,
				Tokens:     insight.Tokens,
				Content:    utils.Truncate(insight.Content, 100),
			})
		}
	}

	return result
}

// DetectAdvancedPatterns runs all advanced pattern detections
func DetectAdvancedPatterns(lines []string, insights []*LineInsight) *AdvancedPatterns {
	return &AdvancedPatterns{
		URLs:             DetectURLs(lines),
		ConsecutiveEmpty: DetectConsecutiveEmpty(insights, minConsecutiveEmptyLines),
		LongLines:        DetectLongLines(insights, defaultLongLineThreshold),
	}
}
