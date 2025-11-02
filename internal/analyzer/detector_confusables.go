package analyzer

import (
	"github.com/mtibben/confusables"
)

// ConfusablesDetector finds homoglyphs and visually similar characters
type ConfusablesDetector struct {
	issues []*ConfusableIssue
}

// NewConfusablesDetector creates a new confusables detector
func NewConfusablesDetector() *ConfusablesDetector {
	return &ConfusablesDetector{
		issues: make([]*ConfusableIssue, 0),
	}
}

// Name returns the detector's identifier
func (d *ConfusablesDetector) Name() string {
	return "confusables"
}

// Priority returns execution priority (lower values execute first)
func (d *ConfusablesDetector) Priority() int {
	return 6
}

// Issues returns the detected issues
func (d *ConfusablesDetector) Issues() []interface{} {
	result := make([]interface{}, len(d.issues))
	for i, issue := range d.issues {
		result[i] = issue
	}
	return result
}

// Detect performs confusables detection
func (d *ConfusablesDetector) Detect(ctx *DetectionContext) error {
	d.issues = make([]*ConfusableIssue, 0)

	for lineNum, line := range ctx.Lines {
		runes := []rune(line)
		for pos, r := range runes {
			// Skip ASCII characters and common punctuation
			if r < 128 {
				continue
			}

			// Use UTS #39 skeleton algorithm to detect confusables
			original := string(r)
			skeleton := confusables.Skeleton(original)

			// If skeleton differs from original, it's a confusable character
			if skeleton != original {
				context := extractContext(line, pos)
				isMixedScript := detectMixedScriptConfusable(line, pos)

				// Get the first rune of skeleton as the confusable target
				var confusableRune rune
				if len([]rune(skeleton)) > 0 {
					confusableRune = []rune(skeleton)[0]
				} else {
					confusableRune = r
				}

				// Generate a descriptive name
				charName := getConfusableCharNameHelper(r, confusableRune)

				issue := &ConfusableIssue{
					OriginalChar:   r,
					ConfusableChar: confusableRune,
					CharName:       charName,
					LineNumber:     lineNum + 1,
					Position:       pos,
					Context:        context,
					Count:          1,
					IsMixedScript:  isMixedScript,
				}

				merged := tryMergeIssueByLineAndType(
					d.issues,
					lineNum+1,
					func(e *ConfusableIssue, line int) bool {
						return e.LineNumber == line && e.OriginalChar == r
					},
					func(e *ConfusableIssue) { e.Count++ },
				)

				if !merged {
					d.issues = append(d.issues, issue)
				}
			}
		}
	}

	return nil
}

// getConfusableCharNameHelper generates a descriptive name for confusable characters
// This is a local helper for the detector (llmsafety.go has getConfusableCharName)
func getConfusableCharNameHelper(original, confusable rune) string {
	// Determine script of original character
	scriptName := getUnicodeScriptHelper(original)
	confusableScript := getUnicodeScriptHelper(confusable)

	return scriptName + " '" + string(original) + "' vs " + confusableScript + " '" + string(confusable) + "'"
}

// getUnicodeScriptHelper returns a simple script name for a rune
// This is a local helper for the detector (llmsafety.go has getUnicodeScript)
func getUnicodeScriptHelper(r rune) string {
	switch {
	case r >= 0x0400 && r <= 0x04FF:
		return "Cyrillic"
	case r >= 0x0370 && r <= 0x03FF:
		return "Greek"
	case r >= 0x0041 && r <= 0x007A:
		return "Latin"
	case r >= 0x0600 && r <= 0x06FF:
		return "Arabic"
	case r >= 0x0590 && r <= 0x05FF:
		return "Hebrew"
	case r >= 0x4E00 && r <= 0x9FFF:
		return "CJK"
	case r >= 0x1F00 && r <= 0x1FFF:
		return "Greek Extended"
	case r >= 0x1D400 && r <= 0x1D7FF:
		return "Mathematical"
	case r >= 0xFF00 && r <= 0xFFEF:
		return "Fullwidth"
	default:
		return "Unicode"
	}
}

// detectMixedScriptConfusable checks if there are mixed scripts in the word containing the character
// This is a local helper for the detector (llmsafety.go has detectMixedScript)
func detectMixedScriptConfusable(line string, pos int) bool {
	start := pos
	for start > 0 && (isLetterOrDigit(rune(line[start-1]))) {
		start--
	}
	end := pos
	for end < len(line)-1 && (isLetterOrDigit(rune(line[end+1]))) {
		end++
	}
	if end-start < 2 {
		return false
	}

	word := line[start : end+1]
	hasLatin, hasCyrillic, hasGreek := false, false, false
	for _, r := range word {
		if r >= 0x0041 && r <= 0x007A {
			hasLatin = true
		} else if r >= 0x0400 && r <= 0x04FF {
			hasCyrillic = true
		} else if r >= 0x0370 && r <= 0x03FF {
			hasGreek = true
		}
	}

	mixedCount := 0
	if hasLatin {
		mixedCount++
	}
	if hasCyrillic {
		mixedCount++
	}
	if hasGreek {
		mixedCount++
	}
	return mixedCount > 1
}

// isLetterOrDigit is a helper function for checking if a rune is a letter or digit
func isLetterOrDigit(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		(r >= 0x0400 && r <= 0x04FF) || // Cyrillic
		(r >= 0x0370 && r <= 0x03FF) || // Greek
		(r >= 0x0600 && r <= 0x06FF) || // Arabic
		(r >= 0x0590 && r <= 0x05FF) || // Hebrew
		(r >= 0x4E00 && r <= 0x9FFF) // CJK
}
