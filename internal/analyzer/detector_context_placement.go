package analyzer

import (
	"strings"
)

// ContextPlacementDetector finds long-context attention issues
type ContextPlacementDetector struct {
	issues []*ContextPlacementIssue
}

// NewContextPlacementDetector creates a new context placement detector
func NewContextPlacementDetector() *ContextPlacementDetector {
	return &ContextPlacementDetector{
		issues: make([]*ContextPlacementIssue, 0),
	}
}

// Name returns the detector's identifier
func (d *ContextPlacementDetector) Name() string {
	return "context_placement"
}

// Priority returns execution priority (lower values execute first)
func (d *ContextPlacementDetector) Priority() int {
	return 10
}

// Issues returns the detected issues
func (d *ContextPlacementDetector) Issues() []interface{} {
	result := make([]interface{}, len(d.issues))
	for i, issue := range d.issues {
		result[i] = issue
	}
	return result
}

// Detect performs context placement detection
func (d *ContextPlacementDetector) Detect(ctx *DetectionContext) error {
	d.issues = make([]*ContextPlacementIssue, 0)

	// Only check for long contexts (4000+ tokens)
	if ctx.TotalTokens < 4000 {
		return nil
	}

	lines := ctx.Lines
	importantAtStart := detectContextImportantContent(lines[0:minVal(5, len(lines))])
	importantAtEnd := detectContextImportantContent(lines[maxVal(0, len(lines)-5):])

	middleStart := len(lines) / 3
	middleEnd := 2 * len(lines) / 3
	importantInMiddle := false
	if middleEnd > middleStart {
		importantInMiddle = detectContextImportantContent(lines[middleStart:middleEnd])
	}

	issue := &ContextPlacementIssue{
		TotalTokens:        ctx.TotalTokens,
		ImportantAtStart:   importantAtStart,
		ImportantAtEnd:     importantAtEnd,
		ImportantInMiddle:  importantInMiddle,
		RecommendedChanges: "Move key facts to start/end; avoid burying instructions in middle",
	}
	d.issues = append(d.issues, issue)

	return nil
}

// detectContextImportantContent checks if a slice of lines contains important keywords
func detectContextImportantContent(lines []string) bool {
	importantKeywords := []string{
		"system:",
		"instruction:",
		"important:",
		"note:",
		"critical:",
		"must:",
		"required:",
	}
	for _, line := range lines {
		lower := strings.ToLower(line)
		for _, keyword := range importantKeywords {
			if strings.Contains(lower, keyword) {
				return true
			}
		}
	}
	return false
}

// minVal returns the minimum of two integers
func minVal(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// maxVal returns the maximum of two integers
func maxVal(a, b int) int {
	if a > b {
		return a
	}
	return b
}
