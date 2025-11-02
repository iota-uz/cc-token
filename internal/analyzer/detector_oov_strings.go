package analyzer

import (
	"regexp"
	"strings"
)

// OOVStringsDetector finds out-of-vocabulary strings that tokenize poorly
type OOVStringsDetector struct {
	issues []*OOVStringIssue
}

// NewOOVStringsDetector creates a new OOV strings detector
func NewOOVStringsDetector() *OOVStringsDetector {
	return &OOVStringsDetector{
		issues: make([]*OOVStringIssue, 0),
	}
}

// Name returns the detector's identifier
func (d *OOVStringsDetector) Name() string {
	return "oov_strings"
}

// Priority returns execution priority (lower values execute first)
func (d *OOVStringsDetector) Priority() int {
	return 4
}

// Issues returns the detected issues
func (d *OOVStringsDetector) Issues() []interface{} {
	result := make([]interface{}, len(d.issues))
	for i, issue := range d.issues {
		result[i] = issue
	}
	return result
}

// Detect performs OOV string detection
func (d *OOVStringsDetector) Detect(ctx *DetectionContext) error {
	d.issues = make([]*OOVStringIssue, 0)

	// Define all detectors in a slice for single-pass iteration
	detectors := []OOVPatternDetector{
		{
			Pattern: regexp.MustCompile(`https?://[^\s]+`),
			Type:    "url",
			MatchFunc: func(match string, lineNum int, line string) *OOVStringIssue {
				if len(match) > minURLLengthForOOV { // Long URLs are OOV
					return &OOVStringIssue{
						String:         match,
						StringType:     "url",
						LineNumber:     lineNum + 1,
						TokenCount:     estimateURLTokenCount(match),
						Context:        line,
						Recommendation: "Replace with short URL or <URL> placeholder",
					}
				}
				return nil
			},
		},
		{
			Pattern: regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`),
			Type:    "uuid",
			MatchFunc: func(match string, lineNum int, line string) *OOVStringIssue {
				return &OOVStringIssue{
					String:         match,
					StringType:     "uuid",
					LineNumber:     lineNum + 1,
					TokenCount:     estimateUUIDTokenCount(match),
					Context:        line,
					Recommendation: "Replace with <UUID> placeholder",
				}
			},
		},
		{
			Pattern: regexp.MustCompile(`\b[0-9a-f]{32,}\b`),
			Type:    "hash",
			MatchFunc: func(match string, lineNum int, line string) *OOVStringIssue {
				if len(match) >= minHashLength { // At least MD5 length
					return &OOVStringIssue{
						String:         match,
						StringType:     "hash",
						LineNumber:     lineNum + 1,
						TokenCount:     estimateHashTokenCount(match),
						Context:        line,
						Recommendation: "Replace with <HASH> placeholder or semantic name",
					}
				}
				return nil
			},
		},
		{
			Pattern: regexp.MustCompile(`[a-zA-Z0-9_-]{20,}`),
			Type:    "id",
			MatchFunc: func(match string, lineNum int, line string) *OOVStringIssue {
				if len(match) > 20 && !isURL(match) && !isHash(match) {
					return &OOVStringIssue{
						String:         match,
						StringType:     "id",
						LineNumber:     lineNum + 1,
						TokenCount:     len(match) / 2, // Rough estimate
						Context:        line,
						Recommendation: "Use shorter identifier or break into semantic parts",
					}
				}
				return nil
			},
		},
	}

	// Single pass through lines, check all detectors
	for lineNum, line := range ctx.Lines {
		for _, detector := range detectors {
			matches := detector.Pattern.FindAllString(line, -1)
			for _, match := range matches {
				if issue := detector.MatchFunc(match, lineNum, line); issue != nil {
					d.issues = append(d.issues, issue)
				}
			}
		}
	}

	return nil
}

// isURL checks if a string looks like a URL
func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// isHash checks if a string looks like a cryptographic hash
func isHash(s string) bool {
	if len(s) < minHashLength {
		return false
	}
	// Check if it's mostly hex characters
	hexCount := 0
	for _, r := range s {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
			hexCount++
		}
	}
	return float64(hexCount)/float64(len(s)) > hexCharRatioThreshold
}

// estimateURLTokenCount estimates token count for a URL
func estimateURLTokenCount(url string) int {
	return estimateTokenCost("url", url)
}

// estimateUUIDTokenCount estimates token count for a UUID
func estimateUUIDTokenCount(uuid string) int {
	return estimateTokenCost("uuid", uuid)
}

// estimateHashTokenCount estimates token count for a hash
func estimateHashTokenCount(hash string) int {
	return estimateTokenCost("hash", hash)
}
