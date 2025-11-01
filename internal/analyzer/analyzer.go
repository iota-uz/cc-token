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

	// Detect basic patterns
	patterns := detectPatterns(lineInsights, avgRatio, lines, tokens)

	// Detect advanced patterns
	advancedPatterns := DetectAdvancedPatterns(lines, lineInsights)

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

// generateEnhancedRecommendations creates detailed actionable optimization advice
func generateEnhancedRecommendations(
	patterns *Patterns,
	advancedPatterns *AdvancedPatterns,
	categoryBreakdown *CategoryBreakdown,
	totalTokens int,
	lines []string,
) []*Recommendation {
	recommendations := make([]*Recommendation, 0)

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
	return strings.TrimSpace(strings.ReplaceAll(fmt.Sprintf("%d", n), " ", ""))
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
