package analyzer

// EmojiDetector finds emoji and ZWJ sequences that can harm tokenization
type EmojiDetector struct {
	issues []*EmojiIssue
}

// NewEmojiDetector creates a new emoji detector
func NewEmojiDetector() *EmojiDetector {
	return &EmojiDetector{
		issues: make([]*EmojiIssue, 0),
	}
}

// Name returns the detector's identifier
func (d *EmojiDetector) Name() string {
	return "emoji"
}

// Priority returns execution priority (lower values execute first)
func (d *EmojiDetector) Priority() int {
	return 1
}

// Issues returns the detected issues
func (d *EmojiDetector) Issues() []interface{} {
	result := make([]interface{}, len(d.issues))
	for i, issue := range d.issues {
		result[i] = issue
	}
	return result
}

// Detect performs emoji detection
func (d *EmojiDetector) Detect(ctx *DetectionContext) error {
	d.issues = make([]*EmojiIssue, 0)

	for lineNum, line := range ctx.Lines {
		runes := []rune(line)
		for i, r := range runes {
			if isEmoji(r) {
				// Determine emoji type
				emojiType := "standard"
				if i+1 < len(runes) {
					nextRune := runes[i+1]
					if nextRune == 0x200D { // ZWJ
						emojiType = "zwj_sequence"
					} else if r >= 0x1F3FB && r <= 0x1F3FF { // Skin tone
						emojiType = "skin_tone"
					} else if r >= 0x1F1E0 && r <= 0x1F1FF { // Flag
						emojiType = "flag"
					}
				}

				issue := &EmojiIssue{
					Emoji:       string(r),
					EmojiType:   emojiType,
					LineNumber:  lineNum + 1,
					Count:       1,
					LineContent: line,
					TokenCost:   estimateEmojiTokenCost(emojiType),
				}

				// Check for existing emoji issue on same line to merge
				merged := tryMergeEmojiIssue(d.issues, lineNum+1, emojiType)
				if merged == nil {
					d.issues = append(d.issues, issue)
				} else {
					merged.Count++
					merged.TokenCost += estimateEmojiTokenCost(emojiType)
				}
			}
		}
	}

	return nil
}

// tryMergeEmojiIssue attempts to find an existing issue to merge with
func tryMergeEmojiIssue(issues []*EmojiIssue, lineNum int, emojiType string) *EmojiIssue {
	for _, existing := range issues {
		if existing.LineNumber == lineNum && existing.EmojiType == emojiType {
			return existing
		}
	}
	return nil
}

// isEmoji checks if a rune is an emoji
func isEmoji(r rune) bool {
	for _, rang := range emojiRanges {
		if r >= rang[0] && r <= rang[1] {
			return true
		}
	}
	return false
}

// estimateEmojiTokenCost returns estimated token cost for different emoji types
func estimateEmojiTokenCost(emojiType string) int {
	switch emojiType {
	case "zwj_sequence":
		return 3 // ZWJ sequences are more expensive
	case "skin_tone":
		return 2 // Skin tone modifiers add tokens
	case "flag":
		return 2 // Flags use regional indicators
	default:
		return 1 // Standard emoji
	}
}

// emojiRanges defines Unicode ranges for emoji detection
var emojiRanges = [][2]rune{
	{0x1F600, 0x1F64F}, // Emoticons
	{0x1F300, 0x1F5FF}, // Misc Symbols and Pictographs
	{0x1F680, 0x1F6FF}, // Transport and Map
	{0x1F1E0, 0x1F1FF}, // Flags
	{0x1F900, 0x1F9FF}, // Supplemental Symbols and Pictographs
	{0x2600, 0x26FF},   // Misc symbols
	{0x2700, 0x27BF},   // Dingbats
}
