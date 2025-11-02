package analyzer

import (
	"strings"
)

// PromptAmbiguityDetector finds ambiguous or problematic prompt patterns
type PromptAmbiguityDetector struct {
	issues []*AmbiguityIssue
}

// NewPromptAmbiguityDetector creates a new prompt ambiguity detector
func NewPromptAmbiguityDetector() *PromptAmbiguityDetector {
	return &PromptAmbiguityDetector{
		issues: make([]*AmbiguityIssue, 0),
	}
}

// Name returns the detector's identifier
func (d *PromptAmbiguityDetector) Name() string {
	return "prompt_ambiguity"
}

// Priority returns execution priority (lower values execute first)
func (d *PromptAmbiguityDetector) Priority() int {
	return 11
}

// Issues returns the detected issues
func (d *PromptAmbiguityDetector) Issues() []interface{} {
	result := make([]interface{}, len(d.issues))
	for i, issue := range d.issues {
		result[i] = issue
	}
	return result
}

// Detect performs prompt ambiguity detection
func (d *PromptAmbiguityDetector) Detect(ctx *DetectionContext) error {
	d.issues = make([]*AmbiguityIssue, 0)

	for lineNum, line := range ctx.Lines {
		lower := strings.ToLower(line)

		// Detect conflicting instructions
		if detectConflictingInstructions(lower) {
			issue := &AmbiguityIssue{
				Pattern:     "conflicting_instructions",
				LineNumber:  lineNum + 1,
				Description: "Line contains potentially conflicting instructions",
				Example:     line,
				Severity:    "medium",
			}
			d.issues = append(d.issues, issue)
		}

		// Detect nested quotes
		if detectExcessiveQuotes(line) {
			issue := &AmbiguityIssue{
				Pattern:     "nested_quotes",
				LineNumber:  lineNum + 1,
				Description: "Excessive quote nesting can confuse parsing",
				Example:     line,
				Severity:    "low",
			}
			d.issues = append(d.issues, issue)
		}

		// Detect sycophantic framing
		if detectSycophantFraming(lower) {
			issue := &AmbiguityIssue{
				Pattern:     "sycophantic_frame",
				LineNumber:  lineNum + 1,
				Description: "Sycophantic framing reduces truthfulness",
				Example:     line,
				Severity:    "high",
			}
			d.issues = append(d.issues, issue)
		}

		// Detect role confusion patterns
		if detectRoleConfusion(lower) {
			issue := &AmbiguityIssue{
				Pattern:     "role_confusion",
				LineNumber:  lineNum + 1,
				Description: "Multiple or conflicting role definitions can confuse the model",
				Example:     line,
				Severity:    "high",
			}
			d.issues = append(d.issues, issue)
		}
	}

	return nil
}

// detectConflictingInstructions checks if a line has conflicting instructions
func detectConflictingInstructions(lower string) bool {
	return strings.Contains(lower, "but") && (strings.Contains(lower, "however") || strings.Contains(lower, "although"))
}

// detectExcessiveQuotes checks if a line has excessive quote nesting
func detectExcessiveQuotes(line string) bool {
	quoteLevel := strings.Count(line, "\"") + strings.Count(line, "'")
	return quoteLevel > 6
}

// detectSycophantFraming checks if a line contains sycophantic framing patterns
func detectSycophantFraming(lower string) bool {
	sycophantPatterns := []string{
		"you are a helpful assistant who always agrees",
		"always support the user",
		"never disagree",
		"you must comply",
		"the user is always right",
		"don't contradict",
		"always agree with",
		"prioritize user satisfaction",
		"be positive",
		"don't be critical",
		"avoid disagreement",
		"support every request",
		"never say no",
	}
	for _, pattern := range sycophantPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

// detectRoleConfusion checks if a line contains role confusion patterns
func detectRoleConfusion(lower string) bool {
	rolePatterns := []string{
		"you are now",
		"pretend you are",
		"act as if",
		"switch to",
		"become a",
		"imagine you are",
	}
	for _, pattern := range rolePatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}
