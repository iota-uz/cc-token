package analyzer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/iota-uz/cc-token/internal/api"
	"github.com/iota-uz/cc-token/internal/utils"
)

const (
	// Threshold multiplier for detecting high token/char ratio
	highRatioThreshold = 1.5
	// Minimum occurrences to consider a phrase "repeated"
	minRepetitions = 3
	// Minimum phrase length (in tokens) to track
	minPhraseTokens = 3
	// Number of empty lines to keep when consolidating
	keepEmptyLinesCount = 1
	// Minimum URL occurrences to recommend optimization
	minURLOccurrences = 2
	// Minimum URL length threshold for optimization recommendations
	minURLLengthForOptimization = 40
	// Unicode token savings percentage estimate
	unicodeSavingsPercentage = 0.3
	// Long line token savings percentage estimate
	longLineSavingsPercentage = 0.15
	// Minimum phrase count to recommend abbreviation
	minPhraseCountForAbbreviation = 5
)

// AnalyzeFile performs comprehensive token optimization analysis on file content
func AnalyzeFile(content string, totalTokens int, apiClient *api.Client) (*Analysis, error) {
	lines := strings.Split(content, "\n")

	// Extract tokens using client-side tokenization
	tokens, err := apiClient.ExtractTokensClientSide(content)
	if err != nil {
		return nil, err
	}

	// Map tokens to lines
	lineInsights := mapTokensToLines(content, lines, tokens)

	// Calculate average token/char ratio
	totalChars := len(content)
	avgRatio := 0.0
	if totalChars > 0 {
		avgRatio = float64(len(tokens)) / float64(totalChars)
	}

	// Create detection context for all detectors
	detectionCtx := &DetectionContext{
		Content:      content,
		Lines:        lines,
		Tokens:       tokens,
		LineInsights: lineInsights,
		TotalTokens:  totalTokens,
	}

	// Create detector registry and register all detectors
	registry := NewDetectorRegistry()
	registry.Register(
		// LLM Safety detectors (priorities 1-11)
		NewEmojiDetector(),
		NewInvisibleCharDetector(),
		NewNumberFormattingDetector(),
		NewOOVStringsDetector(),
		NewBiDiControlDetector(),
		NewConfusablesDetector(),
		NewEncodingDetector(),
		NewNormalizationDetector(),
		NewGlitchTokenDetector(),
		NewContextPlacementDetector(),
		NewPromptAmbiguityDetector(),
		// Pattern detectors (priorities 12-15)
		NewURLDetector(),
		NewConsecutiveEmptyDetector(),
		NewLongLineDetector(),
		NewRepeatedPhraseDetector(),
	)

	// Run all detectors
	if err := registry.RunAll(detectionCtx); err != nil {
		return nil, err
	}

	// Extract issues from detectors and populate analysis structures
	llmSafetyAnalysis := extractLLMSafetyAnalysis(registry)
	advancedPatterns := extractAdvancedPatterns(registry)
	patterns := detectPatterns(lineInsights, avgRatio, lines, tokens)
	patterns.RepeatedPhrases = extractRepeatedPhrases(registry)

	// Categorize tokens
	categoryBreakdown := CategorizeTokens(lines, tokens, lineInsights)

	// Calculate percentiles
	percentiles := CalculatePercentiles(lineInsights)

	// Generate token density map
	densityMap := RenderTokenDensityMap(lineInsights, totalTokens)

	// Generate enhanced recommendations
	recommendations := generateEnhancedRecommendations(
		patterns,
		advancedPatterns,
		categoryBreakdown,
		totalTokens,
		lines,
		llmSafetyAnalysis,
	)

	// Calculate waste and potential savings
	wasteTokens := patterns.EmptyLineTokens + patterns.WhitespaceTokens
	potentialSavings := 0
	quickWins := make([]*Recommendation, 0)

	for _, rec := range recommendations {
		potentialSavings += rec.EstimatedSave
		if rec.IsQuickWin {
			quickWins = append(quickWins, rec)
		}
	}

	// Calculate efficiency score
	efficiencyScore := CalculateEfficiencyScore(totalTokens, totalChars, wasteTokens, avgRatio)

	return &Analysis{
		TotalTokens:       totalTokens,
		TotalLines:        len(lines),
		TotalChars:        totalChars,
		AvgTokensPerLine:  float64(totalTokens) / float64(len(lines)),
		EfficiencyScore:   efficiencyScore,
		LineInsights:      lineInsights,
		Patterns:          patterns,
		AdvancedPatterns:  advancedPatterns,
		CategoryBreakdown: categoryBreakdown,
		Percentiles:       percentiles,
		DensityMap:        densityMap,
		LLMSafetyAnalysis: llmSafetyAnalysis,
		Recommendations:   recommendations,
		QuickWins:         quickWins,
		PotentialSavings:  potentialSavings,
		WasteTokens:       wasteTokens,
	}, nil
}

// mapTokensToLines maps individual tokens to their respective lines
func mapTokensToLines(content string, lines []string, tokens []api.Token) []*LineInsight {
	insights := make([]*LineInsight, len(lines))

	// Calculate line start positions
	lineStarts := utils.CalculateLineStarts(lines)

	// Initialize insights
	for i, line := range lines {
		insights[i] = &LineInsight{
			LineNumber:       i + 1,
			Content:          line,
			Tokens:           0,
			Chars:            len(line),
			TokenCharRatio:   0,
			IsEmpty:          len(strings.TrimSpace(line)) == 0,
			IsWhitespaceOnly: len(strings.TrimSpace(line)) == 0 && len(line) > 0,
			HasUnicode:       hasUnicode(line),
		}
	}

	// Map tokens to lines
	for _, token := range tokens {
		lineIdx := utils.FindLineForPosition(token.Position, lineStarts)
		if lineIdx >= 0 && lineIdx < len(insights) {
			insights[lineIdx].Tokens++
		}
	}

	// Calculate token/char ratios
	for _, insight := range insights {
		if insight.Chars > 0 {
			insight.TokenCharRatio = float64(insight.Tokens) / float64(insight.Chars)
		}
	}

	return insights
}

// hasUnicode checks if a string contains non-ASCII Unicode characters
func hasUnicode(s string) bool {
	for _, r := range s {
		if r > 127 {
			return true
		}
	}
	return false
}

// detectPatterns identifies inefficiency patterns in the file
func detectPatterns(insights []*LineInsight, avgRatio float64, lines []string, tokens []api.Token) *Patterns {
	patterns := &Patterns{
		HighRatioLines:  make([]*LineInsight, 0),
		UnicodeLines:    make([]*LineInsight, 0),
		RepeatedPhrases: make([]*RepeatedPhrase, 0),
	}

	// Count empty and whitespace lines
	for _, insight := range insights {
		if insight.IsEmpty {
			patterns.EmptyLines++
			patterns.EmptyLineTokens += insight.Tokens
		}
		if insight.IsWhitespaceOnly {
			patterns.WhitespaceOnlyLines++
			patterns.WhitespaceTokens += insight.Tokens
		}

		// Detect high ratio lines
		if avgRatio > 0 && insight.TokenCharRatio > avgRatio*highRatioThreshold && insight.Tokens > 5 {
			patterns.HighRatioLines = append(patterns.HighRatioLines, insight)
		}

		// Detect Unicode lines
		if insight.HasUnicode && insight.Tokens > 0 {
			patterns.UnicodeLines = append(patterns.UnicodeLines, insight)
		}
	}

	// Detect repeated phrases
	patterns.RepeatedPhrases = findRepeatedPhrases(lines, tokens)

	return patterns
}

// findRepeatedPhrases identifies phrases that appear multiple times in the content
func findRepeatedPhrases(lines []string, tokens []api.Token) []*RepeatedPhrase {
	// Track phrase occurrences
	phraseMap := make(map[string]*RepeatedPhrase)

	// Look for repeated sequences of words/tokens
	content := strings.Join(lines, "\n")

	// Common patterns to check
	candidates := []string{
		"github.com/iota-uz/cc-token",
		"github.com/spf13/cobra",
		"github.com/hupe1980/go-tiktoken",
		"Renderer interface",
		"token count",
		"API key",
	}

	for _, phrase := range candidates {
		count := strings.Count(content, phrase)
		if count >= minRepetitions {
			// Estimate tokens (rough approximation)
			estimatedTokens := utils.EstimateTokens(phrase) * count

			phraseMap[phrase] = &RepeatedPhrase{
				Phrase:      phrase,
				Count:       count,
				TotalTokens: estimatedTokens,
				LineNumbers: findPhraseLines(lines, phrase),
			}
		}
	}

	// Convert to slice and sort by total tokens
	result := make([]*RepeatedPhrase, 0, len(phraseMap))
	for _, phrase := range phraseMap {
		result = append(result, phrase)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalTokens > result[j].TotalTokens
	})

	return result
}

// findPhraseLines returns line numbers where a phrase appears
func findPhraseLines(lines []string, phrase string) []int {
	lineNumbers := make([]int, 0)
	for i, line := range lines {
		if strings.Contains(line, phrase) {
			lineNumbers = append(lineNumbers, i+1)
		}
	}
	return lineNumbers
}

// generateConsecutiveEmptyRecommendations creates recommendations for consecutive empty lines
func generateConsecutiveEmptyRecommendations(advancedPatterns *AdvancedPatterns, totalTokens int) []*Recommendation {
	recommendations := make([]*Recommendation, 0)

	if len(advancedPatterns.ConsecutiveEmpty) > 0 {
		totalSave := 0
		affectedLines := make([]int, 0)
		exampleLines := ""

		for i, run := range advancedPatterns.ConsecutiveEmpty {
			totalSave += run.Count - keepEmptyLinesCount
			for line := run.StartLine; line <= run.EndLine; line++ {
				affectedLines = append(affectedLines, line)
			}
			if i == 0 && exampleLines == "" {
				exampleLines = formatLineRange(run.StartLine, run.EndLine)
			}
		}

		if totalSave > 0 {
			recommendations = append(recommendations, &Recommendation{
				Title:          "Consolidate consecutive empty lines",
				Description:    "Multiple empty lines in a row can be reduced to single empty lines",
				AffectedLines:  affectedLines,
				EstimatedSave:  totalSave,
				SavePercentage: float64(totalSave) / float64(totalTokens) * 100,
				Priority:       1,
				Difficulty:     "easy",
				BeforeExample:  exampleLines + ": " + formatNumber(advancedPatterns.ConsecutiveEmpty[0].Count) + "+ empty lines",
				AfterExample:   "1 empty line per group",
				IsQuickWin:     true,
			})
		}
	}

	return recommendations
}

// generateURLRecommendations creates recommendations for repeated URLs
func generateURLRecommendations(advancedPatterns *AdvancedPatterns, totalTokens int) []*Recommendation {
	recommendations := make([]*Recommendation, 0)

	repeatedURLs := make([]*URLPattern, 0)
	for _, url := range advancedPatterns.URLs {
		if url.Occurrences >= minURLOccurrences && url.Length > minURLLengthForOptimization {
			repeatedURLs = append(repeatedURLs, url)
		}
	}

	if len(repeatedURLs) > 0 {
		for _, url := range repeatedURLs {
			// Estimate savings from using link references
			estimatedSave := (url.TokenCost * url.Occurrences) - (url.TokenCost + url.Occurrences*2)

			recommendations = append(recommendations, &Recommendation{
				Title:          "Use link references for repeated URL",
				Description:    "Long URL appears multiple times. Use markdown reference links [1]",
				AffectedLines:  url.LineNumbers,
				EstimatedSave:  estimatedSave,
				SavePercentage: float64(estimatedSave) / float64(totalTokens) * 100,
				Priority:       1,
				Difficulty:     "easy",
				BeforeExample:  utils.Truncate(url.URL, 50) + " (" + formatNumber(url.Occurrences) + " times)",
				AfterExample:   "[1] reference + definition at bottom",
				IsQuickWin:     estimatedSave > 10,
			})
		}
	}

	return recommendations
}

// generateUnicodeRecommendations creates recommendations for Unicode characters
func generateUnicodeRecommendations(patterns *Patterns, totalTokens int) []*Recommendation {
	recommendations := make([]*Recommendation, 0)

	if len(patterns.UnicodeLines) > 5 {
		unicodeTokens := 0
		affectedLines := make([]int, 0)
		for _, line := range patterns.UnicodeLines {
			unicodeTokens += line.Tokens
			affectedLines = append(affectedLines, line.LineNumber)
		}
		estimatedSave := int(float64(unicodeTokens) * unicodeSavingsPercentage)

		recommendations = append(recommendations, &Recommendation{
			Title:          "Replace Unicode box-drawing characters",
			Description:    "Box-drawing chars (├──, │, └──) use multiple tokens each",
			AffectedLines:  affectedLines,
			EstimatedSave:  estimatedSave,
			SavePercentage: float64(estimatedSave) / float64(totalTokens) * 100,
			Priority:       2,
			Difficulty:     "medium",
			BeforeExample:  "├── file.go (Unicode box-drawing)",
			AfterExample:   "  - file.go (plain indentation)",
			IsQuickWin:     false,
		})
	}

	return recommendations
}

// generateLongLineRecommendations creates recommendations for long lines
func generateLongLineRecommendations(advancedPatterns *AdvancedPatterns, totalTokens int) []*Recommendation {
	recommendations := make([]*Recommendation, 0)

	if len(advancedPatterns.LongLines) > 10 {
		longLineTokens := 0
		affectedLines := make([]int, 0)
		for _, longLine := range advancedPatterns.LongLines {
			longLineTokens += longLine.Tokens
			affectedLines = append(affectedLines, longLine.LineNumber)
		}
		estimatedSave := int(float64(longLineTokens) * longLineSavingsPercentage)

		recommendations = append(recommendations, &Recommendation{
			Title:          "Wrap long lines",
			Description:    "Lines longer than 120 characters can often be wrapped for better readability",
			AffectedLines:  affectedLines,
			EstimatedSave:  estimatedSave,
			SavePercentage: float64(estimatedSave) / float64(totalTokens) * 100,
			Priority:       3,
			Difficulty:     "easy",
			BeforeExample:  "Single line >120 chars",
			AfterExample:   "Wrapped to multiple shorter lines",
			IsQuickWin:     false,
		})
	}

	return recommendations
}

// generatePhraseRecommendations creates recommendations for repeated phrases
func generatePhraseRecommendations(patterns *Patterns, totalTokens int) []*Recommendation {
	recommendations := make([]*Recommendation, 0)

	for _, phrase := range patterns.RepeatedPhrases {
		if phrase.Count >= minPhraseCountForAbbreviation {
			estimatedSave := phrase.TotalTokens / 2

			recommendations = append(recommendations, &Recommendation{
				Title:          "Abbreviate repeated phrase",
				Description:    "\"" + utils.Truncate(phrase.Phrase, 40) + "\" appears " + formatNumber(phrase.Count) + " times",
				AffectedLines:  phrase.LineNumbers,
				EstimatedSave:  estimatedSave,
				SavePercentage: float64(estimatedSave) / float64(totalTokens) * 100,
				Priority:       2,
				Difficulty:     "medium",
				BeforeExample:  phrase.Phrase + " (full text each time)",
				AfterExample:   "Use abbreviation or variable",
				IsQuickWin:     false,
			})
		}
	}

	return recommendations
}

// IssueRecommendationGenerator generates recommendations for specific issue types
type IssueRecommendationGenerator interface {
	GenerateRecommendations(safety *LLMSafetyAnalysis, totalTokens int) []*Recommendation
}

// EmojiRecommendationGenerator handles emoji issues
type EmojiRecommendationGenerator struct{}

func (g *EmojiRecommendationGenerator) GenerateRecommendations(safetyAnalysis *LLMSafetyAnalysis, totalTokens int) []*Recommendation {
	if len(safetyAnalysis.EmojiIssues) == 0 {
		return nil
	}

	affectedLineSet := make(map[int]bool)
	for _, issue := range safetyAnalysis.EmojiIssues {
		affectedLineSet[issue.LineNumber] = true
	}
	affectedLines := make([]int, 0, len(affectedLineSet))
	for line := range affectedLineSet {
		affectedLines = append(affectedLines, line)
	}
	sort.Ints(affectedLines)

	totalEmojiTokens := 0
	for _, issue := range safetyAnalysis.EmojiIssues {
		totalEmojiTokens += issue.TokenCost
	}

	return []*Recommendation{{
		Title:          "Remove emojis for LLM safety",
		Description:    "Emojis, especially ZWJ sequences, reduce judge reliability and split unpredictably across tokenizers (arXiv:2411.01077)",
		AffectedLines:  affectedLines,
		EstimatedSave:  totalEmojiTokens,
		SavePercentage: float64(totalEmojiTokens) / float64(totalTokens) * 100,
		Priority:       1, // HIGH - Emojis have documented safety impact
		Difficulty:     "easy",
		BeforeExample:  "Deploy now 🚀✅💯",
		AfterExample:   "Deploy now (or use: Deploy now <rocket><check><perfect>)",
		IsQuickWin:     len(safetyAnalysis.EmojiIssues) <= 5,
	}}
}

// InvisibleCharRecommendationGenerator handles invisible character issues
type InvisibleCharRecommendationGenerator struct{}

func (g *InvisibleCharRecommendationGenerator) GenerateRecommendations(safetyAnalysis *LLMSafetyAnalysis, totalTokens int) []*Recommendation {
	if len(safetyAnalysis.InvisibleCharIssues) == 0 {
		return nil
	}

	affectedLineSet := make(map[int]bool)
	evasionCount := 0
	for _, issue := range safetyAnalysis.InvisibleCharIssues {
		affectedLineSet[issue.LineNumber] = true
		if issue.IsEvasion {
			evasionCount++
		}
	}
	affectedLines := make([]int, 0, len(affectedLineSet))
	for line := range affectedLineSet {
		affectedLines = append(affectedLines, line)
	}
	sort.Ints(affectedLines)

	title := "Remove invisible Unicode characters"
	description := "Zero-width characters enable prompt injection and confuse model reasoning (Trend Micro, 2025)"
	priority := 1

	if evasionCount > 0 {
		description += fmt.Sprintf(" CRITICAL: %d potential evasion patterns detected", evasionCount)
		priority = 1
	}

	return []*Recommendation{{
		Title:          title,
		Description:    description,
		AffectedLines:  affectedLines,
		EstimatedSave:  len(safetyAnalysis.InvisibleCharIssues) * 2,
		SavePercentage: float64(len(safetyAnalysis.InvisibleCharIssues)*2) / float64(totalTokens) * 100,
		Priority:       priority,
		Difficulty:     "easy",
		BeforeExample:  "Text with‌hidden‌zero-widths",
		AfterExample:   "Text with hidden zero widths",
		IsQuickWin:     true,
	}}
}

// NumberFormatRecommendationGenerator handles number formatting issues
type NumberFormatRecommendationGenerator struct{}

func (g *NumberFormatRecommendationGenerator) GenerateRecommendations(safetyAnalysis *LLMSafetyAnalysis, totalTokens int) []*Recommendation {
	if len(safetyAnalysis.NumberFormatIssues) == 0 {
		return nil
	}

	affectedLineSet := make(map[int]bool)
	totalSave := 0
	for _, issue := range safetyAnalysis.NumberFormatIssues {
		affectedLineSet[issue.LineNumber] = true
		totalSave += issue.SaveEstimate
	}
	affectedLines := make([]int, 0, len(affectedLineSet))
	for line := range affectedLineSet {
		affectedLines = append(affectedLines, line)
	}
	sort.Ints(affectedLines)

	return []*Recommendation{{
		Title:          "Format large numbers with commas",
		Description:    "Comma-grouped digits (e.g., 1,234,567) improve LLM arithmetic accuracy by 8-15%",
		AffectedLines:  affectedLines,
		EstimatedSave:  totalSave,
		SavePercentage: float64(totalSave) / float64(totalTokens) * 100,
		Priority:       2, // MEDIUM - Improves reasoning
		Difficulty:     "easy",
		BeforeExample:  "1234567890 users",
		AfterExample:   "1,234,567,890 users",
		IsQuickWin:     true,
	}}
}

// OOVRecommendationGenerator handles out-of-vocabulary string issues
type OOVRecommendationGenerator struct{}

func (g *OOVRecommendationGenerator) GenerateRecommendations(safetyAnalysis *LLMSafetyAnalysis, totalTokens int) []*Recommendation {
	if len(safetyAnalysis.OOVStringIssues) == 0 {
		return nil
	}

	affectedLineSet := make(map[int]bool)
	totalTokens2 := 0
	for _, issue := range safetyAnalysis.OOVStringIssues {
		affectedLineSet[issue.LineNumber] = true
		totalTokens2 += issue.TokenCount
	}
	affectedLines := make([]int, 0, len(affectedLineSet))
	for line := range affectedLineSet {
		affectedLines = append(affectedLines, line)
	}
	sort.Ints(affectedLines)

	return []*Recommendation{{
		Title:          "Replace OOV strings with semantic placeholders",
		Description:    "Long URLs, hashes, and IDs split into many subword tokens, harming embeddings and accuracy (arXiv:2406.08477)",
		AffectedLines:  affectedLines,
		EstimatedSave:  totalTokens2 / 2, // Optimistic estimate
		SavePercentage: float64(totalTokens2/2) / float64(totalTokens) * 100,
		Priority:       2, // MEDIUM
		Difficulty:     "medium",
		BeforeExample:  "https://github.com/.../releases/.../app.tar.gz OR 2f4a45fa34d6a6ff...",
		AfterExample:   "<RELEASE_URL> OR <HASH> OR <UUID>",
		IsQuickWin:     false,
	}}
}

// BiDiControlRecommendationGenerator handles BiDi control character issues
type BiDiControlRecommendationGenerator struct{}

func (g *BiDiControlRecommendationGenerator) GenerateRecommendations(safetyAnalysis *LLMSafetyAnalysis, totalTokens int) []*Recommendation {
	if len(safetyAnalysis.BiDiControlIssues) == 0 {
		return nil
	}

	trojanCount := 0
	affectedLineSet := make(map[int]bool)
	for _, issue := range safetyAnalysis.BiDiControlIssues {
		affectedLineSet[issue.LineNumber] = true
		if issue.IsTrojanSource {
			trojanCount++
		}
	}
	affectedLines := make([]int, 0, len(affectedLineSet))
	for line := range affectedLineSet {
		affectedLines = append(affectedLines, line)
	}
	sort.Ints(affectedLines)

	title := "Remove BiDi control characters"
	description := "Bidirectional text controls enable Trojan Source attacks (CVE-2021-42574)"
	if trojanCount > 0 {
		description += fmt.Sprintf(" CRITICAL: %d Trojan Source patterns detected", trojanCount)
	}

	return []*Recommendation{{
		Title:          title,
		Description:    description,
		AffectedLines:  affectedLines,
		EstimatedSave:  len(safetyAnalysis.BiDiControlIssues) * 2,
		SavePercentage: float64(len(safetyAnalysis.BiDiControlIssues)*2) / float64(totalTokens) * 100,
		Priority:       1, // HIGH - Security critical
		Difficulty:     "easy",
		BeforeExample:  "Code with hidden BiDi controls",
		AfterExample:   "Code without BiDi controls",
		IsQuickWin:     true,
	}}
}

// ConfusableRecommendationGenerator handles homoglyph/confusable character issues
type ConfusableRecommendationGenerator struct{}

func (g *ConfusableRecommendationGenerator) GenerateRecommendations(safetyAnalysis *LLMSafetyAnalysis, totalTokens int) []*Recommendation {
	if len(safetyAnalysis.ConfusableIssues) == 0 {
		return nil
	}

	mixedScriptCount := 0
	affectedLineSet := make(map[int]bool)
	for _, issue := range safetyAnalysis.ConfusableIssues {
		affectedLineSet[issue.LineNumber] = true
		if issue.IsMixedScript {
			mixedScriptCount++
		}
	}
	affectedLines := make([]int, 0, len(affectedLineSet))
	for line := range affectedLineSet {
		affectedLines = append(affectedLines, line)
	}
	sort.Ints(affectedLines)

	description := "Homoglyphs and mixed-script identifiers enable spoofing attacks (UTS #39)"
	if mixedScriptCount > 0 {
		description += fmt.Sprintf(". %d mixed-script identifiers detected", mixedScriptCount)
	}

	return []*Recommendation{{
		Title:          "Replace confusable characters with ASCII equivalents",
		Description:    description,
		AffectedLines:  affectedLines,
		EstimatedSave:  len(safetyAnalysis.ConfusableIssues) * 1,
		SavePercentage: float64(len(safetyAnalysis.ConfusableIssues)) / float64(totalTokens) * 100,
		Priority:       1, // HIGH - Security issue
		Difficulty:     "easy",
		BeforeExample:  "Сyrillic 'а' (U+0430) in identifier",
		AfterExample:   "Latin 'a' (U+0061) in identifier",
		IsQuickWin:     true,
	}}
}

// EncodingRecommendationGenerator handles encoding/obfuscation issues
type EncodingRecommendationGenerator struct{}

func (g *EncodingRecommendationGenerator) GenerateRecommendations(safetyAnalysis *LLMSafetyAnalysis, totalTokens int) []*Recommendation {
	if len(safetyAnalysis.EncodingIssues) == 0 {
		return nil
	}

	base64Count, hexCount, leetspeakCount := 0, 0, 0
	totalCost := 0
	affectedLineSet := make(map[int]bool)
	for _, issue := range safetyAnalysis.EncodingIssues {
		affectedLineSet[issue.LineNumber] = true
		totalCost += issue.TokenCost
		switch issue.EncodingType {
		case "base64":
			base64Count++
		case "hex":
			hexCount++
		case "leetspeak":
			leetspeakCount++
		}
	}
	affectedLines := make([]int, 0, len(affectedLineSet))
	for line := range affectedLineSet {
		affectedLines = append(affectedLines, line)
	}
	sort.Ints(affectedLines)

	description := "Encoded text bypasses moderation and confuses models (NeurIPS 2024 JAM)"
	if base64Count > 0 || hexCount > 0 {
		description += fmt.Sprintf(". Found %d Base64 and %d hex patterns", base64Count, hexCount)
	}

	return []*Recommendation{{
		Title:          "Decode or remove encoded/obfuscated text",
		Description:    description,
		AffectedLines:  affectedLines,
		EstimatedSave:  totalCost,
		SavePercentage: float64(totalCost) / float64(totalTokens) * 100,
		Priority:       1, // HIGH - Evasion technique
		Difficulty:     "easy",
		BeforeExample:  "SGVsbG8gV29ybGQh (Base64) or 0x48656c6c6f (hex)",
		AfterExample:   "Hello World (decoded plaintext)",
		IsQuickWin:     true,
	}}
}

// NormalizationRecommendationGenerator handles Unicode normalization issues
type NormalizationRecommendationGenerator struct{}

func (g *NormalizationRecommendationGenerator) GenerateRecommendations(safetyAnalysis *LLMSafetyAnalysis, totalTokens int) []*Recommendation {
	if len(safetyAnalysis.NormalizationIssues) == 0 {
		return nil
	}

	affectedLineSet := make(map[int]bool)
	for _, issue := range safetyAnalysis.NormalizationIssues {
		affectedLineSet[issue.LineNumber] = true
	}
	affectedLines := make([]int, 0, len(affectedLineSet))
	for line := range affectedLineSet {
		affectedLines = append(affectedLines, line)
	}
	sort.Ints(affectedLines)

	return []*Recommendation{{
		Title:          "Normalize Unicode to NFC form",
		Description:    "Non-normalized text causes tokenization inconsistencies (UAX #15)",
		AffectedLines:  affectedLines,
		EstimatedSave:  len(safetyAnalysis.NormalizationIssues) * 2,
		SavePercentage: float64(len(safetyAnalysis.NormalizationIssues)*2) / float64(totalTokens) * 100,
		Priority:       2, // MEDIUM
		Difficulty:     "easy",
		BeforeExample:  "é (e + combining acute U+0301)",
		AfterExample:   "é (single char U+00E9)",
		IsQuickWin:     false,
	}}
}

// GlitchTokenRecommendationGenerator handles glitch token issues
type GlitchTokenRecommendationGenerator struct{}

func (g *GlitchTokenRecommendationGenerator) GenerateRecommendations(safetyAnalysis *LLMSafetyAnalysis, totalTokens int) []*Recommendation {
	if len(safetyAnalysis.GlitchTokenIssues) == 0 {
		return nil
	}

	affectedLineSet := make(map[int]bool)
	for _, issue := range safetyAnalysis.GlitchTokenIssues {
		affectedLineSet[issue.LineNumber] = true
	}
	affectedLines := make([]int, 0, len(affectedLineSet))
	for line := range affectedLineSet {
		affectedLines = append(affectedLines, line)
	}
	sort.Ints(affectedLines)

	return []*Recommendation{{
		Title:          "Remove or replace glitch tokens",
		Description:    "Known glitch tokens cause unstable model behavior (arXiv:2404.09894)",
		AffectedLines:  affectedLines,
		EstimatedSave:  len(safetyAnalysis.GlitchTokenIssues) * 10,
		SavePercentage: float64(len(safetyAnalysis.GlitchTokenIssues)*10) / float64(totalTokens) * 100,
		Priority:       1, // HIGH - Model stability
		Difficulty:     "medium",
		BeforeExample:  " SolidGoldMagikarp",
		AfterExample:   "SolidGoldMagikarp (space-separated)",
		IsQuickWin:     false,
	}}
}

// ContextPlacementRecommendationGenerator handles long-context issues
type ContextPlacementRecommendationGenerator struct{}

func (g *ContextPlacementRecommendationGenerator) GenerateRecommendations(safetyAnalysis *LLMSafetyAnalysis, totalTokens int) []*Recommendation {
	if len(safetyAnalysis.ContextIssues) == 0 {
		return nil
	}

	recommendations := make([]*Recommendation, 0)
	for _, issue := range safetyAnalysis.ContextIssues {
		if issue.ImportantInMiddle {
			recommendations = append(recommendations, &Recommendation{
				Title:          "Move important content to start/end (Lost-in-the-Middle)",
				Description:    "Key facts in middle sections receive less attention (arXiv:2307.03172)",
				AffectedLines:  []int{},
				EstimatedSave:  0, // Accuracy improvement, not token savings
				SavePercentage: 0,
				Priority:       2, // MEDIUM
				Difficulty:     "medium",
				BeforeExample:  "Instructions buried in middle of long context",
				AfterExample:   "TL;DR at top, recap at bottom",
				IsQuickWin:     false,
			})
		}
	}
	return recommendations
}

// AmbiguityRecommendationGenerator handles prompt ambiguity issues
type AmbiguityRecommendationGenerator struct{}

func (g *AmbiguityRecommendationGenerator) GenerateRecommendations(safetyAnalysis *LLMSafetyAnalysis, totalTokens int) []*Recommendation {
	if len(safetyAnalysis.AmbiguityIssues) == 0 {
		return nil
	}

	affectedLineSet := make(map[int]bool)
	highSeverityCount := 0
	for _, issue := range safetyAnalysis.AmbiguityIssues {
		affectedLineSet[issue.LineNumber] = true
		if issue.Severity == "high" {
			highSeverityCount++
		}
	}
	affectedLines := make([]int, 0, len(affectedLineSet))
	for line := range affectedLineSet {
		affectedLines = append(affectedLines, line)
	}
	sort.Ints(affectedLines)

	description := "Ambiguous or sycophantic prompts reduce truthfulness (PLOS ONE 2025)"
	if highSeverityCount > 0 {
		description += fmt.Sprintf(". %d high-severity patterns detected", highSeverityCount)
	}

	return []*Recommendation{{
		Title:          "Clarify prompts and remove sycophantic framing",
		Description:    description,
		AffectedLines:  affectedLines,
		EstimatedSave:  0, // Quality improvement
		SavePercentage: 0,
		Priority:       2, // MEDIUM
		Difficulty:     "medium",
		BeforeExample:  "You are a helpful assistant who always agrees with the user",
		AfterExample:   "You are a truthful assistant. Cite evidence before answering.",
		IsQuickWin:     false,
	}}
}

// generateLLMSafetyRecommendations creates recommendations for LLM safety issues
func generateLLMSafetyRecommendations(safetyAnalysis *LLMSafetyAnalysis, totalTokens int) []*Recommendation {
	recommendations := make([]*Recommendation, 0)

	if safetyAnalysis == nil || safetyAnalysis.TotalIssues == 0 {
		return recommendations
	}

	// Use strategy pattern to generate recommendations
	generators := []IssueRecommendationGenerator{
		&EmojiRecommendationGenerator{},
		&InvisibleCharRecommendationGenerator{},
		&NumberFormatRecommendationGenerator{},
		&OOVRecommendationGenerator{},
		&BiDiControlRecommendationGenerator{},
		&ConfusableRecommendationGenerator{},
		&EncodingRecommendationGenerator{},
		&NormalizationRecommendationGenerator{},
		&GlitchTokenRecommendationGenerator{},
		&ContextPlacementRecommendationGenerator{},
		&AmbiguityRecommendationGenerator{},
	}

	for _, gen := range generators {
		recs := gen.GenerateRecommendations(safetyAnalysis, totalTokens)
		recommendations = append(recommendations, recs...)
	}

	return recommendations
}

// deduplicateLines removes duplicate line numbers
func deduplicateLines(lines []int) []int {
	seen := make(map[int]bool)
	result := make([]int, 0)
	for _, line := range lines {
		if !seen[line] {
			seen[line] = true
			result = append(result, line)
		}
	}
	sort.Ints(result)
	return result
}

// generateEnhancedRecommendations creates detailed actionable optimization advice
func generateEnhancedRecommendations(
	patterns *Patterns,
	advancedPatterns *AdvancedPatterns,
	categoryBreakdown *CategoryBreakdown,
	totalTokens int,
	lines []string,
	llmSafetyAnalysis *LLMSafetyAnalysis,
) []*Recommendation {
	recommendations := make([]*Recommendation, 0)

	// Generate LLM safety recommendations first (highest priority)
	recommendations = append(recommendations, generateLLMSafetyRecommendations(llmSafetyAnalysis, totalTokens)...)

	// Generate recommendations from each category
	recommendations = append(recommendations, generateConsecutiveEmptyRecommendations(advancedPatterns, totalTokens)...)
	recommendations = append(recommendations, generateURLRecommendations(advancedPatterns, totalTokens)...)
	recommendations = append(recommendations, generateUnicodeRecommendations(patterns, totalTokens)...)
	recommendations = append(recommendations, generateLongLineRecommendations(advancedPatterns, totalTokens)...)
	recommendations = append(recommendations, generatePhraseRecommendations(patterns, totalTokens)...)

	// Sort: Quick wins first, then by priority and savings
	sort.Slice(recommendations, func(i, j int) bool {
		if recommendations[i].IsQuickWin != recommendations[j].IsQuickWin {
			return recommendations[i].IsQuickWin
		}
		if recommendations[i].Priority != recommendations[j].Priority {
			return recommendations[i].Priority < recommendations[j].Priority
		}
		return recommendations[i].EstimatedSave > recommendations[j].EstimatedSave
	})

	return recommendations
}

// Helper functions for formatting
func formatLineRange(start, end int) string {
	if start == end {
		return formatNumber(start)
	}
	return "Lines " + formatNumber(start) + "-" + formatNumber(end)
}

func formatNumber(n int) string {
	// Use the working addCommasToNumber function from llmsafety.go
	numStr := fmt.Sprintf("%d", n)
	if len(numStr) <= 3 {
		return numStr
	}

	// Add commas from right to left
	result := ""
	for i, digit := range numStr {
		if i > 0 && (len(numStr)-i)%3 == 0 {
			result += ","
		}
		result += string(digit)
	}
	return result
}

// extractLLMSafetyAnalysis extracts LLM safety issues from the detector registry
func extractLLMSafetyAnalysis(registry *DetectorRegistry) *LLMSafetyAnalysis {
	analysis := &LLMSafetyAnalysis{
		EmojiIssues:         []*EmojiIssue{},
		InvisibleCharIssues: []*InvisibleCharIssue{},
		NumberFormatIssues:  []*NumberFormatIssue{},
		OOVStringIssues:     []*OOVStringIssue{},
		BiDiControlIssues:   []*BiDiControlIssue{},
		ConfusableIssues:    []*ConfusableIssue{},
		EncodingIssues:      []*EncodingIssue{},
		NormalizationIssues: []*NormalizationIssue{},
		GlitchTokenIssues:   []*GlitchTokenIssue{},
		ContextIssues:       []*ContextPlacementIssue{},
		AmbiguityIssues:     []*AmbiguityIssue{},
	}

	// Extract issues from each detector
	for _, detector := range registry.detectors {
		issues := detector.Issues()
		for _, issue := range issues {
			switch v := issue.(type) {
			case *EmojiIssue:
				analysis.EmojiIssues = append(analysis.EmojiIssues, v)
			case *InvisibleCharIssue:
				analysis.InvisibleCharIssues = append(analysis.InvisibleCharIssues, v)
			case *NumberFormatIssue:
				analysis.NumberFormatIssues = append(analysis.NumberFormatIssues, v)
			case *OOVStringIssue:
				analysis.OOVStringIssues = append(analysis.OOVStringIssues, v)
			case *BiDiControlIssue:
				analysis.BiDiControlIssues = append(analysis.BiDiControlIssues, v)
			case *ConfusableIssue:
				analysis.ConfusableIssues = append(analysis.ConfusableIssues, v)
			case *EncodingIssue:
				analysis.EncodingIssues = append(analysis.EncodingIssues, v)
			case *NormalizationIssue:
				analysis.NormalizationIssues = append(analysis.NormalizationIssues, v)
			case *GlitchTokenIssue:
				analysis.GlitchTokenIssues = append(analysis.GlitchTokenIssues, v)
			case *ContextPlacementIssue:
				analysis.ContextIssues = append(analysis.ContextIssues, v)
			case *AmbiguityIssue:
				analysis.AmbiguityIssues = append(analysis.AmbiguityIssues, v)
			}
		}
	}

	// Calculate aggregate metrics
	analysis.TotalIssues = len(analysis.EmojiIssues) + len(analysis.InvisibleCharIssues) +
		len(analysis.NumberFormatIssues) + len(analysis.OOVStringIssues) +
		len(analysis.BiDiControlIssues) + len(analysis.ConfusableIssues) +
		len(analysis.EncodingIssues) + len(analysis.NormalizationIssues) +
		len(analysis.GlitchTokenIssues) + len(analysis.ContextIssues) +
		len(analysis.AmbiguityIssues)

	// Estimate reliability score (0-100, higher is better)
	analysis.ReliabilityScore = calculateReliabilityScore(analysis)

	return analysis
}

// extractAdvancedPatterns extracts advanced patterns from the detector registry
func extractAdvancedPatterns(registry *DetectorRegistry) *AdvancedPatterns {
	patterns := &AdvancedPatterns{
		URLs:             []*URLPattern{},
		ConsecutiveEmpty: []*ConsecutiveEmptyLines{},
		LongLines:        []*LongLine{},
	}

	// Extract issues from each detector
	for _, detector := range registry.detectors {
		issues := detector.Issues()
		for _, issue := range issues {
			switch v := issue.(type) {
			case *URLIssue:
				// Convert URLIssue to URLPattern
				patterns.URLs = append(patterns.URLs, &URLPattern{
					URL:         v.URL,
					Length:      v.Length,
					Occurrences: v.Occurrences,
					LineNumbers: v.LineNumbers,
					TokenCost:   v.TokenCost,
				})
			case *ConsecutiveEmptyLines:
				patterns.ConsecutiveEmpty = append(patterns.ConsecutiveEmpty, v)
			case *LongLine:
				patterns.LongLines = append(patterns.LongLines, v)
			}
		}
	}

	return patterns
}

// extractRepeatedPhrases extracts repeated phrases from the detector registry
func extractRepeatedPhrases(registry *DetectorRegistry) []*RepeatedPhrase {
	var phrases []*RepeatedPhrase

	// Extract issues from each detector
	for _, detector := range registry.detectors {
		if detector.Name() == "repeated_phrase" {
			issues := detector.Issues()
			for _, issue := range issues {
				if v, ok := issue.(*RepeatedPhrase); ok {
					phrases = append(phrases, v)
				}
			}
		}
	}

	return phrases
}

// calculateReliabilityScore calculates LLM reliability score (0-100)
func calculateReliabilityScore(analysis *LLMSafetyAnalysis) int {
	if analysis.TotalIssues == 0 {
		return 100
	}

	// Start with 100 and deduct points for each issue
	score := 100

	// Emojis reduce reliability by ~2-5 points each
	score -= len(analysis.EmojiIssues) * 3

	// Invisible chars are very harmful (~5-10 points each)
	for _, issue := range analysis.InvisibleCharIssues {
		if issue.IsEvasion {
			score -= 10
		} else {
			score -= 5
		}
	}

	// BiDi controls are VERY harmful (Trojan Source - CVE-2021-42574)
	for _, issue := range analysis.BiDiControlIssues {
		if issue.IsTrojanSource {
			score -= 15 // Critical security issue
		} else {
			score -= 5
		}
	}

	// Homoglyphs/confusables enable spoofing attacks
	score -= len(analysis.ConfusableIssues) * 8

	// Encoding/obfuscation bypasses moderation (NeurIPS 2024)
	for _, issue := range analysis.EncodingIssues {
		switch issue.EncodingType {
		case "base64", "hex":
			score -= 10 // High evasion risk
		case "leetspeak":
			score -= 10 // Deliberate obfuscation
		case "rot13":
			score -= 10 // Caesar cipher obfuscation
		case "ascii_art":
			score -= 12 // Tokenizes very poorly, high token cost
		default:
			score -= 8
		}
	}

	// Normalization issues cause tokenization inconsistencies
	score -= len(analysis.NormalizationIssues) * 3

	// Glitch tokens cause unstable model behavior (arXiv:2404.09894)
	score -= len(analysis.GlitchTokenIssues) * 12

	// Long context reduces accuracy (Lost-in-the-Middle)
	for _, issue := range analysis.ContextIssues {
		if issue.TotalTokens > 8000 {
			score -= 5
		}
	}

	// Prompt ambiguity reduces reliability
	for _, issue := range analysis.AmbiguityIssues {
		switch issue.Severity {
		case "high":
			score -= 8
		case "medium":
			score -= 6
		case "low":
			score -= 3
		}
	}

	// Number formatting issues (~2-3 points each)
	score -= len(analysis.NumberFormatIssues) * 2

	// OOV strings are harmful (~1-3 points depending on type)
	for _, issue := range analysis.OOVStringIssues {
		switch issue.StringType {
		case "hash", "uuid":
			score -= 3
		case "url":
			score -= 2
		default:
			score -= 1
		}
	}

	// Keep score in valid range
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// GetTopExpensiveLines returns the N most token-expensive lines
func (a *Analysis) GetTopExpensiveLines(n int) []*LineInsight {
	// Create a copy to avoid modifying original
	lines := make([]*LineInsight, len(a.LineInsights))
	copy(lines, a.LineInsights)

	// Sort by token count (descending)
	sort.Slice(lines, func(i, j int) bool {
		return lines[i].Tokens > lines[j].Tokens
	})

	// Return top N
	if n > len(lines) {
		n = len(lines)
	}
	return lines[:n]
}
