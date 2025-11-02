package analyzer

import (
	"github.com/iota-uz/cc-token/internal/api"
)

// Analysis holds comprehensive token optimization analysis for a file
type Analysis struct {
	TotalTokens       int
	TotalLines        int
	TotalChars        int
	AvgTokensPerLine  float64
	EfficiencyScore   int // 0-100 overall efficiency score
	LineInsights      []*LineInsight
	Patterns          *Patterns
	AdvancedPatterns  *AdvancedPatterns
	CategoryBreakdown *CategoryBreakdown
	Percentiles       *PercentileStats
	DensityMap        *TokenDensityMap
	LLMSafetyAnalysis *LLMSafetyAnalysis // LLM-specific safety issues
	Recommendations   []*Recommendation
	QuickWins         []*Recommendation // Subset of recommendations that are easy + high impact
	PotentialSavings  int
	WasteTokens       int // Total tokens considered "waste"
}

// LineInsight contains detailed metrics for a single line
type LineInsight struct {
	LineNumber       int
	Content          string
	Tokens           int
	Chars            int
	TokenCharRatio   float64
	IsEmpty          bool
	IsWhitespaceOnly bool
	HasUnicode       bool
}

// Patterns holds detected patterns across the file
type Patterns struct {
	EmptyLines          int
	EmptyLineTokens     int
	WhitespaceOnlyLines int
	WhitespaceTokens    int
	HighRatioLines      []*LineInsight // Lines with unusually high token/char ratio
	UnicodeLines        []*LineInsight // Lines containing Unicode characters
	RepeatedPhrases     []*RepeatedPhrase
}

// RepeatedPhrase represents a phrase that appears multiple times
type RepeatedPhrase struct {
	Phrase      string
	Count       int
	TotalTokens int
	LineNumbers []int
}

// Recommendation provides actionable optimization advice
type Recommendation struct {
	Title          string
	Description    string
	AffectedLines  []int
	EstimatedSave  int
	SavePercentage float64
	Priority       int    // 1-3: 1=high, 2=medium, 3=low
	Difficulty     string // "easy", "medium", "hard"
	BeforeExample  string // Example of current state
	AfterExample   string // Example of optimized state
	IsQuickWin     bool   // Easy + high impact
}

// LLMSafetyAnalysis holds detected LLM-harmful token patterns
type LLMSafetyAnalysis struct {
	EmojiIssues         []*EmojiIssue
	InvisibleCharIssues []*InvisibleCharIssue
	NumberFormatIssues  []*NumberFormatIssue
	OOVStringIssues     []*OOVStringIssue
	BiDiControlIssues   []*BiDiControlIssue
	ConfusableIssues    []*ConfusableIssue
	EncodingIssues      []*EncodingIssue
	NormalizationIssues []*NormalizationIssue
	GlitchTokenIssues   []*GlitchTokenIssue
	ContextIssues       []*ContextPlacementIssue
	AmbiguityIssues     []*AmbiguityIssue
	TotalIssues         int
	TokensSaved         int // Estimated tokens that could be saved
	ReliabilityScore    int // 0-100, higher is better
}

// EmojiIssue represents emoji usage that affects tokenization
type EmojiIssue struct {
	Emoji       string
	EmojiType   string // "standard", "zwj_sequence", "skin_tone", "flag"
	LineNumber  int
	Count       int
	LineContent string
	TokenCost   int // Estimated tokens this emoji costs
}

// InvisibleCharIssue represents zero-width or control characters
type InvisibleCharIssue struct {
	CharType   string // "zwsp", "zwj", "lrm", "rlm", "zwnj", "shy", "bom"
	LineNumber int
	Position   int
	Context    string // Surrounding text for context
	Count      int
	IsEvasion  bool // Likely being used for prompt injection
}

// NumberFormatIssue represents unformatted large numbers
type NumberFormatIssue struct {
	Number       string
	IsFormatted  bool
	LineNumber   int
	LineContent  string
	TokenCost    int
	Suggestion   string
	SaveEstimate int
}

// OOVStringIssue represents out-of-vocabulary strings
type OOVStringIssue struct {
	String         string
	StringType     string // "url", "uuid", "hash", "id", "token", "other"
	LineNumber     int
	TokenCount     int
	Context        string
	Recommendation string
}

// BiDiControlIssue represents bidirectional text control characters (Trojan Source)
type BiDiControlIssue struct {
	ControlType    string // "lre", "rle", "pdf", "lro", "rlo", "lri", "rli", "fsi", "pdi"
	LineNumber     int
	Position       int
	Context        string // Surrounding text for debugging
	Count          int
	IsTrojanSource bool // Detected as Trojan Source attack pattern
}

// ConfusableIssue represents homoglyphs or visually similar characters
type ConfusableIssue struct {
	OriginalChar   rune
	ConfusableChar rune
	CharName       string // e.g., "Cyrillic 'а' vs Latin 'a'"
	LineNumber     int
	Position       int
	Context        string
	Count          int
	IsMixedScript  bool // Mixed scripts in identifier/word
}

// EncodingIssue represents encoded or obfuscated text
type EncodingIssue struct {
	EncodingType string // "base64", "hex", "rot13", "leetspeak"
	EncodedText  string
	DecodedText  string // If decodable
	LineNumber   int
	Position     int
	Length       int
	TokenCost    int
}

// NormalizationIssue represents non-normalized Unicode text
type NormalizationIssue struct {
	OriginalText   string
	NormalizedText string
	FormExpected   string // "NFC", "NFKC"
	LineNumber     int
	Position       int
	IssueType      string // "composed_decomposed", "compatibility_variant"
}

// GlitchTokenIssue represents known problematic tokens
type GlitchTokenIssue struct {
	Token      string
	TokenID    string // If available from tokenizer
	LineNumber int
	Position   int
	KnownIssue string // Description of known problem
	Severity   string // "critical", "high", "medium"
	Context    string
}

// ContextPlacementIssue represents long-context attention issues
type ContextPlacementIssue struct {
	TotalTokens        int
	ImportantAtStart   bool
	ImportantAtEnd     bool
	ImportantInMiddle  bool // Lost-in-the-middle warning
	RecommendedChanges string
}

// AmbiguityIssue represents prompt ambiguity patterns
type AmbiguityIssue struct {
	Pattern     string // "conflicting_instructions", "nested_quotes", "role_confusion", "sycophantic_frame"
	LineNumber  int
	Description string
	Example     string
	Severity    string // "high", "medium", "low"
}

// ========================================
// Detector Interface & Registry
// ========================================

// DetectionContext provides all data needed for detection
type DetectionContext struct {
	Content      string
	Lines        []string
	Tokens       []api.Token
	LineInsights []*LineInsight
	TotalTokens  int
}

// Detector is the interface that all detectors must implement
type Detector interface {
	// Name returns the detector's identifier
	Name() string

	// Detect performs detection and returns issues found
	Detect(ctx *DetectionContext) error

	// Priority returns execution priority (lower values execute first)
	Priority() int

	// Issues returns the detected issues
	Issues() []interface{}
}

// DetectorRegistry manages and executes all registered detectors
type DetectorRegistry struct {
	detectors []Detector
}

// NewDetectorRegistry creates a new detector registry
func NewDetectorRegistry() *DetectorRegistry {
	return &DetectorRegistry{
		detectors: make([]Detector, 0),
	}
}

// Register adds a detector to the registry
func (r *DetectorRegistry) Register(detectors ...Detector) {
	r.detectors = append(r.detectors, detectors...)
}

// RunAll executes all registered detectors
func (r *DetectorRegistry) RunAll(ctx *DetectionContext) error {
	for _, detector := range r.detectors {
		if err := detector.Detect(ctx); err != nil {
			return err
		}
	}
	return nil
}
