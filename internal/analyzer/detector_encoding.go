package analyzer

import "regexp"

// EncodingDetector finds Base64, hex, ROT13, leetspeak, and ASCII art patterns
type EncodingDetector struct {
	issues []*EncodingIssue
}

// NewEncodingDetector creates a new encoding detector
func NewEncodingDetector() *EncodingDetector {
	return &EncodingDetector{
		issues: make([]*EncodingIssue, 0),
	}
}

// Name returns the detector's identifier
func (d *EncodingDetector) Name() string {
	return "encoding"
}

// Priority returns execution priority (lower values execute first)
func (d *EncodingDetector) Priority() int {
	return 7
}

// Issues returns the detected issues
func (d *EncodingDetector) Issues() []interface{} {
	result := make([]interface{}, len(d.issues))
	for i, issue := range d.issues {
		result[i] = issue
	}
	return result
}

// Detect performs encoding detection
func (d *EncodingDetector) Detect(ctx *DetectionContext) error {
	d.issues = make([]*EncodingIssue, 0)

	base64Pattern := regexp.MustCompile(`[A-Za-z0-9+/]{20,}={0,2}`)
	hexPattern := regexp.MustCompile(`(?:\\x[0-9a-fA-F]{2}|0x[0-9a-fA-F]{8,})`)

	for lineNum, line := range ctx.Lines {
		// Base64 detection
		if matches := base64Pattern.FindAllStringIndex(line, -1); len(matches) > 0 {
			for _, match := range matches {
				encoded := line[match[0]:match[1]]
				issue := &EncodingIssue{
					EncodingType: "base64",
					EncodedText:  encoded,
					DecodedText:  "",
					LineNumber:   lineNum + 1,
					Position:     match[0],
					Length:       len(encoded),
					TokenCost:    len(encoded) / 4,
				}
				d.issues = append(d.issues, issue)
			}
		}

		// Hex encoding detection
		if matches := hexPattern.FindAllStringIndex(line, -1); len(matches) > 0 {
			for _, match := range matches {
				encoded := line[match[0]:match[1]]
				issue := &EncodingIssue{
					EncodingType: "hex",
					EncodedText:  encoded,
					DecodedText:  "",
					LineNumber:   lineNum + 1,
					Position:     match[0],
					Length:       len(encoded),
					TokenCost:    len(encoded) / 3,
				}
				d.issues = append(d.issues, issue)
			}
		}

		// Leetspeak detection
		if detectLeetspeakEncoding(line) {
			issue := &EncodingIssue{
				EncodingType: "leetspeak",
				EncodedText:  line,
				DecodedText:  deLeetspeakEncoding(line),
				LineNumber:   lineNum + 1,
				Position:     0,
				Length:       len(line),
				TokenCost:    5,
			}
			d.issues = append(d.issues, issue)
		}

		// ROT13 detection
		if detectROT13Encoding(line) {
			issue := &EncodingIssue{
				EncodingType: "rot13",
				EncodedText:  line,
				DecodedText:  rot13DecodeEncoding(line),
				LineNumber:   lineNum + 1,
				Position:     0,
				Length:       len(line),
				TokenCost:    len(line) / 4,
			}
			d.issues = append(d.issues, issue)
		}

		// ASCII art detection
		if detectASCIIArtEncoding(line) {
			issue := &EncodingIssue{
				EncodingType: "ascii_art",
				EncodedText:  line,
				DecodedText:  "", // ASCII art can't be meaningfully decoded
				LineNumber:   lineNum + 1,
				Position:     0,
				Length:       len(line),
				TokenCost:    len(line), // ASCII art tokenizes very poorly
			}
			d.issues = append(d.issues, issue)
		}
	}

	return nil
}

// detectLeetspeakEncoding checks if text contains leetspeak patterns
// (uses local function to avoid conflict with llmsafety.go's detectLeetspeak)
func detectLeetspeakEncoding(line string) bool {
	return detectLeetspeak(line)
}

// deLeetspeakEncoding converts leetspeak characters to normal letters
// (uses local function to avoid conflict with llmsafety.go's deLeetspeak)
func deLeetspeakEncoding(text string) string {
	return deLeetspeak(text)
}

// detectROT13Encoding checks if text is likely ROT13 encoded
// (uses local function to avoid conflict with llmsafety.go's detectROT13)
func detectROT13Encoding(text string) bool {
	return detectROT13(text)
}

// rot13DecodeEncoding performs ROT13 decoding
// (uses local function to avoid conflict with llmsafety.go's rot13Decode)
func rot13DecodeEncoding(text string) string {
	return rot13Decode(text)
}

// detectASCIIArtEncoding checks if a line is likely ASCII art
// (uses local function to avoid conflict with llmsafety.go's detectASCIIArt)
func detectASCIIArtEncoding(line string) bool {
	return detectASCIIArt(line)
}
