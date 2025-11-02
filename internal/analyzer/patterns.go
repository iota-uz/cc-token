package analyzer

// NOTE: This file now contains only type definitions and constants.
// All detection logic has been refactored into detector_*.go files
// using the Detector interface and DetectorRegistry pattern.

const (
	// Minimum consecutive empty lines to detect
	minConsecutiveEmptyLines = 2
	// Default long line threshold in characters
	defaultLongLineThreshold = 120
)

// AdvancedPatterns holds detected advanced patterns
type AdvancedPatterns struct {
	URLs             []*URLPattern
	ConsecutiveEmpty []*ConsecutiveEmptyLines
	LongLines        []*LongLine
}

// URLPattern represents a detected URL
type URLPattern struct {
	URL         string
	Length      int
	Occurrences int
	LineNumbers []int
	TokenCost   int
}

// ConsecutiveEmptyLines represents a run of empty lines
type ConsecutiveEmptyLines struct {
	StartLine int
	EndLine   int
	Count     int
}

// LongLine represents a line that's unusually long
type LongLine struct {
	LineNumber int
	Length     int
	Tokens     int
	Content    string
}
