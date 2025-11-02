package analyzer

import (
	"sort"
	"strings"

	"github.com/iota-uz/cc-token/internal/utils"
)

// RepeatedPhraseDetector finds phrases that appear multiple times in content
type RepeatedPhraseDetector struct {
	issues []*RepeatedPhrase
}

// NewRepeatedPhraseDetector creates a new repeated phrase detector
func NewRepeatedPhraseDetector() *RepeatedPhraseDetector {
	return &RepeatedPhraseDetector{
		issues: make([]*RepeatedPhrase, 0),
	}
}

// Name returns the detector's identifier
func (d *RepeatedPhraseDetector) Name() string {
	return "repeated_phrase"
}

// Priority returns execution priority (lower values execute first)
func (d *RepeatedPhraseDetector) Priority() int {
	return 15
}

// Issues returns the detected issues
func (d *RepeatedPhraseDetector) Issues() []interface{} {
	result := make([]interface{}, len(d.issues))
	for i, issue := range d.issues {
		result[i] = issue
	}
	return result
}

// Detect performs repeated phrase detection
func (d *RepeatedPhraseDetector) Detect(ctx *DetectionContext) error {
	d.issues = make([]*RepeatedPhrase, 0)

	// Reconstruct content from lines for phrase searching
	content := strings.Join(ctx.Lines, "\n")

	// Common patterns to check
	candidates := []string{
		"github.com/iota-uz/cc-token",
		"github.com/spf13/cobra",
		"github.com/hupe1980/go-tiktoken",
		"Renderer interface",
		"token count",
		"API key",
	}

	// Track phrase occurrences in a map to avoid duplicates
	phraseMap := make(map[string]*RepeatedPhrase)

	for _, phrase := range candidates {
		count := strings.Count(content, phrase)
		if count >= minRepetitions {
			// Estimate tokens (rough approximation)
			estimatedTokens := utils.EstimateTokens(phrase) * count

			phraseMap[phrase] = &RepeatedPhrase{
				Phrase:      phrase,
				Count:       count,
				TotalTokens: estimatedTokens,
				LineNumbers: findPhraseLines(ctx.Lines, phrase),
			}
		}
	}

	// Convert to slice and sort by total tokens
	for _, phrase := range phraseMap {
		d.issues = append(d.issues, phrase)
	}
	sort.Slice(d.issues, func(i, j int) bool {
		return d.issues[i].TotalTokens > d.issues[j].TotalTokens
	})

	return nil
}
