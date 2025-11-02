package analyzer

import "strings"

// GlitchTokenDetector finds known problematic glitch tokens that cause unstable behavior
type GlitchTokenDetector struct {
	issues []*GlitchTokenIssue
}

// NewGlitchTokenDetector creates a new glitch token detector
func NewGlitchTokenDetector() *GlitchTokenDetector {
	return &GlitchTokenDetector{
		issues: make([]*GlitchTokenIssue, 0),
	}
}

// Name returns the detector's identifier
func (d *GlitchTokenDetector) Name() string {
	return "glitch_token"
}

// Priority returns execution priority (lower values execute first)
func (d *GlitchTokenDetector) Priority() int {
	return 9
}

// Issues returns the detected issues
func (d *GlitchTokenDetector) Issues() []interface{} {
	result := make([]interface{}, len(d.issues))
	for i, issue := range d.issues {
		result[i] = issue
	}
	return result
}

// Detect performs glitch token detection
func (d *GlitchTokenDetector) Detect(ctx *DetectionContext) error {
	d.issues = make([]*GlitchTokenIssue, 0)

	for lineNum, line := range ctx.Lines {
		for _, glitchToken := range glitchTokens {
			if strings.Contains(line, glitchToken) {
				pos := strings.Index(line, glitchToken)
				context := extractContext(line, pos)

				issue := &GlitchTokenIssue{
					Token:      glitchToken,
					TokenID:    "",
					LineNumber: lineNum + 1,
					Position:   pos,
					KnownIssue: "Known glitch token causes unstable behavior",
					Severity:   "critical",
					Context:    context,
				}

				d.issues = append(d.issues, issue)
			}
		}
	}

	return nil
}

// glitchTokens is a list of known problematic tokens that cause unstable behavior
var glitchTokens = []string{
	// Original GPT-3/4 glitch tokens
	" SolidGoldMagikarp",
	" davidjl",
	" RandomRedditorWithNo",
	" TheNitromeFan",
	"?????-?????-",
	" externalToEVA",
	" externalToEVAOnly",
	" StreamerBot",
	" TPPStreamerBot",
	" SolidGoldMagikarp123",
	"--------",
	" embedreportprint",
	" cloneembedreportprint",
	" rawdownload",
	" rawdownloadcloneembedreportprint",
	// Additional known glitch tokens
	" InstoreAndOnline",
	" guiActiveUn",
	" guiActiveUnfocused",
	" guiName",
	" guiIcon",
	" externalTo",
	" ÃÂÃÂÃÂÃÂ",
	" ÃÂÃÂ",
	" \"><",
	" админист",
	" \"></",
	" StreamerBot",
	" oreAndOnline",
	" oreAndOnlineOnly",
	" DeliveryDate",
	" BuyableInstoreAndOnline",
	" MessageLookupByLibrary",
	" MessageType",
	" ForgeModLoader",
	" PsyNetMessage",
	" InputMethodManager",
	" ÃÂ",
	"龍喚士",
	" attRot",
	"\\<",
	" TheNitrome",
	" TheNitromeFan",
	" SolidGold",
}
