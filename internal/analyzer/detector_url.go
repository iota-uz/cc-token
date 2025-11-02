package analyzer

import (
	"regexp"
	"strings"

	"github.com/iota-uz/cc-token/internal/utils"
)

// URLDetector finds URLs in text that can affect tokenization
type URLDetector struct {
	issues []*URLIssue
}

// NewURLDetector creates a new URL detector
func NewURLDetector() *URLDetector {
	return &URLDetector{
		issues: make([]*URLIssue, 0),
	}
}

// Name returns the detector's identifier
func (d *URLDetector) Name() string {
	return "url"
}

// Priority returns execution priority (lower values execute first)
// URLs are detected after all LLM safety detectors (priority 12)
func (d *URLDetector) Priority() int {
	return 12
}

// Issues returns the detected issues
func (d *URLDetector) Issues() []interface{} {
	result := make([]interface{}, len(d.issues))
	for i, issue := range d.issues {
		result[i] = issue
	}
	return result
}

// Detect performs URL detection
func (d *URLDetector) Detect(ctx *DetectionContext) error {
	d.issues = make([]*URLIssue, 0)

	urlMap := make(map[string]*URLIssue)

	for i, line := range ctx.Lines {
		matches := urlRegexPattern.FindAllString(line, -1)
		for _, url := range matches {
			// Clean up URL (remove trailing punctuation)
			url = strings.TrimRight(url, ".,;:!?")

			if existing, found := urlMap[url]; found {
				existing.Occurrences++
				existing.LineNumbers = append(existing.LineNumbers, i+1)
			} else {
				urlMap[url] = &URLIssue{
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
	for _, issue := range urlMap {
		d.issues = append(d.issues, issue)
	}

	return nil
}

// URLIssue represents a detected URL that affects tokenization
type URLIssue struct {
	URL         string
	Length      int
	Occurrences int
	LineNumbers []int
	TokenCost   int // Estimated tokens this URL costs
}

var (
	// urlRegexPattern matches HTTP and HTTPS URLs
	urlRegexPattern = regexp.MustCompile(`https?://[^\s\)]+`)
)
