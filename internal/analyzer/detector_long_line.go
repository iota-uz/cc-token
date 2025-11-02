package analyzer

import (
	"github.com/iota-uz/cc-token/internal/utils"
)

// LongLineDetector finds lines that are unusually long
type LongLineDetector struct {
	issues []*LongLine
}

// NewLongLineDetector creates a new long line detector
func NewLongLineDetector() *LongLineDetector {
	return &LongLineDetector{
		issues: make([]*LongLine, 0),
	}
}

// Name returns the detector's identifier
func (d *LongLineDetector) Name() string {
	return "long_line"
}

// Priority returns execution priority (lower values execute first)
func (d *LongLineDetector) Priority() int {
	return 14
}

// Issues returns the detected issues
func (d *LongLineDetector) Issues() []interface{} {
	result := make([]interface{}, len(d.issues))
	for i, issue := range d.issues {
		result[i] = issue
	}
	return result
}

// Detect performs long line detection
func (d *LongLineDetector) Detect(ctx *DetectionContext) error {
	d.issues = make([]*LongLine, 0)

	threshold := defaultLongLineThreshold

	for _, insight := range ctx.LineInsights {
		if insight.Chars > threshold && insight.Tokens > 0 {
			issue := &LongLine{
				LineNumber: insight.LineNumber,
				Length:     insight.Chars,
				Tokens:     insight.Tokens,
				Content:    utils.Truncate(insight.Content, 100),
			}
			d.issues = append(d.issues, issue)
		}
	}

	return nil
}
