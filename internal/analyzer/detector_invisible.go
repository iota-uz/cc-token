package analyzer

import (
	"strings"
)

// InvisibleCharDetector finds zero-width and control characters
type InvisibleCharDetector struct {
	issues []*InvisibleCharIssue
}

// NewInvisibleCharDetector creates a new invisible character detector
func NewInvisibleCharDetector() *InvisibleCharDetector {
	return &InvisibleCharDetector{
		issues: make([]*InvisibleCharIssue, 0),
	}
}

// Name returns the detector's identifier
func (d *InvisibleCharDetector) Name() string {
	return "invisible_char"
}

// Priority returns execution priority (lower values execute first)
func (d *InvisibleCharDetector) Priority() int {
	return 2
}

// Issues returns the detected issues
func (d *InvisibleCharDetector) Issues() []interface{} {
	result := make([]interface{}, len(d.issues))
	for i, issue := range d.issues {
		result[i] = issue
	}
	return result
}

// Detect performs invisible character detection
func (d *InvisibleCharDetector) Detect(ctx *DetectionContext) error {
	d.issues = make([]*InvisibleCharIssue, 0)

	for lineNum, line := range ctx.Lines {
		for pos, r := range line {
			if charType, exists := zeroWidthCharMap[r]; exists {
				// Get context: invisibleCharContextSize chars before and after
				context := extractContext(line, pos)

				issue := &InvisibleCharIssue{
					CharType:   charType,
					LineNumber: lineNum + 1,
					Position:   pos,
					Context:    context,
					Count:      1,
					IsEvasion:  isLikelyEvasion(line, pos),
				}

				// Check for existing issue on same line to merge
				merged := tryMergeInvisibleCharIssue(d.issues, lineNum+1, charType)
				if merged == nil {
					d.issues = append(d.issues, issue)
				} else {
					merged.Count++
				}
			}
		}
	}

	return nil
}

// tryMergeInvisibleCharIssue attempts to find an existing issue to merge with
func tryMergeInvisibleCharIssue(issues []*InvisibleCharIssue, lineNum int, charType string) *InvisibleCharIssue {
	for _, existing := range issues {
		if existing.LineNumber == lineNum && existing.CharType == charType {
			return existing
		}
	}
	return nil
}

// isInvisibleChar checks if a rune is an invisible/zero-width character
func isInvisibleChar(r rune) bool {
	_, exists := zeroWidthCharMap[r]
	return exists
}

// getInvisibleCharType returns the type of invisible character
func getInvisibleCharType(r rune) string {
	if charType, exists := zeroWidthCharMap[r]; exists {
		return charType
	}
	return "unknown"
}

// isLikelyEvasion checks if invisible chars are used for evasion
func isLikelyEvasion(line string, pos int) bool {
	// Check context for suspicious patterns using set lookup
	lowerLine := strings.ToLower(line)
	for pattern := range suspiciousPatternSet {
		if strings.Contains(lowerLine, pattern) {
			return true
		}
	}
	return false
}
