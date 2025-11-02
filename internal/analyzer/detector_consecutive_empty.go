package analyzer

// ConsecutiveEmptyDetector finds consecutive empty lines that can harm readability and tokenization
type ConsecutiveEmptyDetector struct {
	issues []*ConsecutiveEmptyLines
}

// NewConsecutiveEmptyDetector creates a new consecutive empty detector
func NewConsecutiveEmptyDetector() *ConsecutiveEmptyDetector {
	return &ConsecutiveEmptyDetector{
		issues: make([]*ConsecutiveEmptyLines, 0),
	}
}

// Name returns the detector's identifier
func (d *ConsecutiveEmptyDetector) Name() string {
	return "consecutive_empty"
}

// Priority returns execution priority (lower values execute first)
func (d *ConsecutiveEmptyDetector) Priority() int {
	return 13
}

// Issues returns the detected issues
func (d *ConsecutiveEmptyDetector) Issues() []interface{} {
	result := make([]interface{}, len(d.issues))
	for i, issue := range d.issues {
		result[i] = issue
	}
	return result
}

// Detect performs consecutive empty line detection
func (d *ConsecutiveEmptyDetector) Detect(ctx *DetectionContext) error {
	d.issues = make([]*ConsecutiveEmptyLines, 0)

	// Use line insights to detect consecutive empty lines
	var currentRun *ConsecutiveEmptyLines

	for _, insight := range ctx.LineInsights {
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
			if currentRun != nil && currentRun.Count >= minConsecutiveEmptyLines {
				d.issues = append(d.issues, currentRun)
			}
			currentRun = nil
		}
	}

	// Don't forget the last run
	if currentRun != nil && currentRun.Count >= minConsecutiveEmptyLines {
		d.issues = append(d.issues, currentRun)
	}

	return nil
}
