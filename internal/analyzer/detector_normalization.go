package analyzer

import (
	"golang.org/x/text/unicode/norm"
)

// NormalizationDetector finds Unicode normalization issues
type NormalizationDetector struct {
	issues []*NormalizationIssue
}

// NewNormalizationDetector creates a new normalization detector
func NewNormalizationDetector() *NormalizationDetector {
	return &NormalizationDetector{
		issues: make([]*NormalizationIssue, 0),
	}
}

// Name returns the detector's identifier
func (d *NormalizationDetector) Name() string {
	return "normalization"
}

// Priority returns execution priority (lower values execute first)
func (d *NormalizationDetector) Priority() int {
	return 8
}

// Issues returns the detected issues
func (d *NormalizationDetector) Issues() []interface{} {
	result := make([]interface{}, len(d.issues))
	for i, issue := range d.issues {
		result[i] = issue
	}
	return result
}

// Detect performs normalization detection
func (d *NormalizationDetector) Detect(ctx *DetectionContext) error {
	d.issues = make([]*NormalizationIssue, 0)

	for lineNum, line := range ctx.Lines {
		// Check NFC (Canonical Composition)
		nfc := norm.NFC.String(line)
		if line != nfc {
			issue := &NormalizationIssue{
				OriginalText:   line,
				NormalizedText: nfc,
				FormExpected:   "NFC",
				LineNumber:     lineNum + 1,
				Position:       0,
				IssueType:      "not_nfc",
			}
			d.issues = append(d.issues, issue)
		}

		// Check NFKC (Compatibility Composition)
		nfkc := norm.NFKC.String(line)
		if line != nfkc {
			issue := &NormalizationIssue{
				OriginalText:   line,
				NormalizedText: nfkc,
				FormExpected:   "NFKC",
				LineNumber:     lineNum + 1,
				Position:       0,
				IssueType:      "not_nfkc",
			}
			d.issues = append(d.issues, issue)
		}
	}

	return nil
}
