package analyzer

// BiDiControlDetector finds bidirectional text control characters (Trojan Source attacks)
type BiDiControlDetector struct {
	issues []*BiDiControlIssue
}

// NewBiDiControlDetector creates a new bidirectional control detector
func NewBiDiControlDetector() *BiDiControlDetector {
	return &BiDiControlDetector{
		issues: make([]*BiDiControlIssue, 0),
	}
}

// Name returns the detector's identifier
func (d *BiDiControlDetector) Name() string {
	return "bidi_control"
}

// Priority returns execution priority (lower values execute first)
func (d *BiDiControlDetector) Priority() int {
	return 5
}

// Issues returns the detected issues
func (d *BiDiControlDetector) Issues() []interface{} {
	result := make([]interface{}, len(d.issues))
	for i, issue := range d.issues {
		result[i] = issue
	}
	return result
}

// Detect performs bidirectional control character detection
func (d *BiDiControlDetector) Detect(ctx *DetectionContext) error {
	d.issues = make([]*BiDiControlIssue, 0)

	for lineNum, line := range ctx.Lines {
		runes := []rune(line)
		for pos, r := range runes {
			if controlType, exists := bidiControlCharMap[r]; exists {
				context := extractContext(line, pos)
				isTrojanSource := detectTrojanSourcePattern(line)

				issue := &BiDiControlIssue{
					ControlType:    controlType,
					LineNumber:     lineNum + 1,
					Position:       pos,
					Context:        context,
					Count:          1,
					IsTrojanSource: isTrojanSource,
				}

				merged := tryMergeIssueByLineAndType(
					d.issues,
					lineNum+1,
					func(e *BiDiControlIssue, line int) bool {
						return e.LineNumber == line && e.ControlType == controlType
					},
					func(e *BiDiControlIssue) { e.Count++ },
				)

				if !merged {
					d.issues = append(d.issues, issue)
				}
			}
		}
	}

	return nil
}

// detectTrojanSourcePattern checks if a line contains both LTR and RTL control characters,
// which indicates a potential Trojan Source attack (CVE-2021-42574)
func detectTrojanSourcePattern(line string) bool {
	hasLTR, hasRTL := false, false
	for _, r := range line {
		switch r {
		case 0x202A, 0x202D, 0x2066: // LRE, LRO, LRI
			hasLTR = true
		case 0x202B, 0x202E, 0x2067: // RLE, RLO, RLI
			hasRTL = true
		}
	}
	return hasLTR && hasRTL
}
