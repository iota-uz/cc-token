package analyzer

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
