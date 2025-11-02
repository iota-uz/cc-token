package analyzer

import (
	"regexp"
	"strings"
)

// NumberFormattingDetector finds unformatted large numbers that hurt arithmetic
type NumberFormattingDetector struct {
	issues []*NumberFormatIssue
}

// NewNumberFormattingDetector creates a new number formatting detector
func NewNumberFormattingDetector() *NumberFormattingDetector {
	return &NumberFormattingDetector{
		issues: make([]*NumberFormatIssue, 0),
	}
}

// Name returns the detector's identifier
func (d *NumberFormattingDetector) Name() string {
	return "number_formatting"
}

// Priority returns execution priority (lower values execute first)
func (d *NumberFormattingDetector) Priority() int {
	return 3
}

// Issues returns the detected issues
func (d *NumberFormattingDetector) Issues() []interface{} {
	result := make([]interface{}, len(d.issues))
	for i, issue := range d.issues {
		result[i] = issue
	}
	return result
}

// Detect performs number formatting detection
func (d *NumberFormattingDetector) Detect(ctx *DetectionContext) error {
	d.issues = make([]*NumberFormatIssue, 0)

	// Match unformatted numbers with 4+ digits (no commas)
	numberPattern := regexp.MustCompile(`\b\d{4,}\b`)

	for lineNum, line := range ctx.Lines {
		matches := numberPattern.FindAllString(line, -1)
		for _, match := range matches {
			// Check if it's already formatted
			if strings.Contains(match, ",") {
				continue // Already formatted
			}

			// Check if it has proper comma formatting (if it should)
			formatted := addCommasToNumber(match)
			if formatted != match {
				suggestion := formatted
				saveEstimate := estimateNumberFormatTokenSave(match, formatted)

				issue := &NumberFormatIssue{
					Number:       match,
					IsFormatted:  false,
					LineNumber:   lineNum + 1,
					LineContent:  line,
					TokenCost:    len(strings.Split(match, "")), // Rough estimate
					Suggestion:   suggestion,
					SaveEstimate: saveEstimate,
				}

				d.issues = append(d.issues, issue)
			}
		}
	}

	return nil
}

// addCommasToNumber adds comma grouping to a number string
func addCommasToNumber(numStr string) string {
	// Remove any existing formatting
	numStr = strings.ReplaceAll(numStr, ",", "")

	if len(numStr) <= 3 {
		return numStr
	}

	// Add commas from right to left
	result := ""
	for i, digit := range numStr {
		if i > 0 && (len(numStr)-i)%3 == 0 {
			result += ","
		}
		result += string(digit)
	}
	return result
}

// estimateNumberFormatTokenSave estimates token savings from number formatting
func estimateNumberFormatTokenSave(original, formatted string) int {
	// Rough heuristic: formatting can save ~1-2 tokens per large number
	if len(original) > 10 {
		return 2
	}
	if len(original) > 7 {
		return 1
	}
	return 0
}
