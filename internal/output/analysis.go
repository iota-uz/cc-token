package output

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/iota-uz/cc-token/internal/analyzer"
	"github.com/iota-uz/cc-token/internal/config"
)

const (
	maxLinePreview = 80
	topExpensiveN  = 25
)

// AnalysisFormatter formats token optimization analysis
type AnalysisFormatter struct {
	useColor bool
}

// NewAnalysisFormatter creates a new analysis formatter
func NewAnalysisFormatter(useColor bool) *AnalysisFormatter {
	return &AnalysisFormatter{
		useColor: useColor,
	}
}

// FormatAnalysis outputs comprehensive token optimization analysis
func (f *AnalysisFormatter) FormatAnalysis(analysis *analyzer.Analysis, filename string, cfg *config.Config) error {
	// Header
	f.printHeader(filename, analysis)

	// Token density heatmap
	f.printDensityMap(analysis)

	// Category breakdown
	f.printCategoryBreakdown(analysis)

	// Statistical analysis
	f.printStatisticalAnalysis(analysis)

	// LLM Safety Analysis
	if analysis.LLMSafetyAnalysis != nil && analysis.LLMSafetyAnalysis.TotalIssues > 0 {
		f.printLLMSafetyAnalysis(analysis)
	}

	// Top expensive lines
	f.printTopExpensiveLines(analysis)

	// Quick wins
	if len(analysis.QuickWins) > 0 {
		f.printQuickWins(analysis)
	}

	// All recommendations
	f.printRecommendations(analysis)

	// Summary
	f.printSummary(analysis)

	return nil
}

func (f *AnalysisFormatter) printHeader(filename string, analysis *analyzer.Analysis) {
	title := fmt.Sprintf("Token Optimization Analysis: %s", filename)
	subtitle := fmt.Sprintf("Total: %d tokens across %d lines (%.1f tokens/line)",
		analysis.TotalTokens, analysis.TotalLines, analysis.AvgTokensPerLine)
	efficiency := fmt.Sprintf("Efficiency Score: %d/100", analysis.EfficiencyScore)

	if f.useColor {
		color.New(color.Bold, color.FgCyan).Println(title)
		color.New(color.FgWhite).Println(subtitle)

		// Color efficiency score based on value
		if analysis.EfficiencyScore >= 80 {
			color.New(color.FgGreen).Println(efficiency)
		} else if analysis.EfficiencyScore >= 60 {
			color.New(color.FgYellow).Println(efficiency)
		} else {
			color.New(color.FgRed).Println(efficiency)
		}
	} else {
		fmt.Println(title)
		fmt.Println(subtitle)
		fmt.Println(efficiency)
	}
}

func (f *AnalysisFormatter) printDensityMap(analysis *analyzer.Analysis) {
	if analysis.DensityMap == nil {
		return
	}

	if !f.useColor {
		// Skip heatmap in plain mode - too verbose with Unicode bars
		return
	}

	f.printSectionHeader("TOKEN DENSITY HEATMAP")
	heatmap := analysis.DensityMap.FormatHeatmap()
	fmt.Print(heatmap)
}

func (f *AnalysisFormatter) printCategoryBreakdown(analysis *analyzer.Analysis) {
	if analysis.CategoryBreakdown == nil {
		return
	}

	f.printSectionHeader("CATEGORY BREAKDOWN")

	stats := analysis.CategoryBreakdown.GetStats()
	categories := []struct {
		name   string
		tokens int
		pct    float64
	}{
		{"Prose", analysis.CategoryBreakdown.Prose, stats.Prose},
		{"Code Blocks", analysis.CategoryBreakdown.CodeBlocks, stats.CodeBlocks},
		{"URLs", analysis.CategoryBreakdown.URLs, stats.URLs},
		{"Formatting", analysis.CategoryBreakdown.Formatting, stats.Formatting},
		{"Whitespace", analysis.CategoryBreakdown.Whitespace, stats.Whitespace},
	}

	for _, cat := range categories {
		if cat.tokens == 0 {
			continue
		}

		if f.useColor {
			bar := analyzer.RenderCategoryBar(cat.pct, 24)
			line := fmt.Sprintf("%-15s %s %6d tokens (%5.1f%%)",
				cat.name+":", bar, cat.tokens, cat.pct)
			color.New(color.FgWhite).Println(line)
		} else {
			// Plain mode: simple format without bars
			fmt.Printf("%s: %d tokens (%.1f%%)\n", cat.name, cat.tokens, cat.pct)
		}
	}
}

func (f *AnalysisFormatter) printStatisticalAnalysis(analysis *analyzer.Analysis) {
	if analysis.Percentiles == nil {
		return
	}

	f.printSectionHeader("STATISTICAL ANALYSIS")

	// Distribution
	distLine := "Distribution: " + analysis.Percentiles.FormatPercentiles()
	if f.useColor {
		color.New(color.FgWhite).Println(distLine)
	} else {
		fmt.Println(distLine)
	}

	// Top 10% concentration
	top10Line := fmt.Sprintf("Top 10%% of lines consume %.1f%% of total tokens", analysis.Percentiles.Top10Pct)
	if f.useColor {
		if analysis.Percentiles.Top10Pct > 25 {
			color.New(color.FgYellow).Println(top10Line)
		} else {
			color.New(color.FgWhite).Println(top10Line)
		}
	} else {
		fmt.Println(top10Line)
	}

	// Waste factors
	if analysis.WasteTokens > 0 {
		wastePct := float64(analysis.WasteTokens) / float64(analysis.TotalTokens) * 100
		wasteLine := fmt.Sprintf("Waste factors: %d tokens (%.1f%%) from empty lines and whitespace",
			analysis.WasteTokens, wastePct)
		if f.useColor {
			color.New(color.FgRed).Println(wasteLine)
		} else {
			fmt.Println(wasteLine)
		}
	}
}

func (f *AnalysisFormatter) printQuickWins(analysis *analyzer.Analysis) {
	f.printSectionHeader("QUICK WINS (Easy + High Impact)")

	for _, rec := range analysis.QuickWins {
		if f.useColor {
			title := fmt.Sprintf("• %s", rec.Title)
			savings := fmt.Sprintf("  Savings: ~%d tokens (%.1f%%)", rec.EstimatedSave, rec.SavePercentage)
			color.New(color.Bold, color.FgGreen).Println(title)
			color.New(color.Faint).Println(savings)

			// Before/After if available
			if rec.BeforeExample != "" && rec.AfterExample != "" {
				before := fmt.Sprintf("  Before: %s", rec.BeforeExample)
				after := fmt.Sprintf("  After:  %s", rec.AfterExample)
				color.New(color.Faint).Println(before)
				color.New(color.FgGreen, color.Faint).Println(after)
			}

			if rec.Description != "" {
				desc := fmt.Sprintf("  %s", rec.Description)
				color.New(color.Faint).Println(desc)
			}
		} else {
			// Plain mode: compact single-line format
			fmt.Printf("%s (Impact: %d tokens, %.1f%%)\n",
				rec.Title, rec.EstimatedSave, rec.SavePercentage)
		}
	}
}

func (f *AnalysisFormatter) printTopExpensiveLines(analysis *analyzer.Analysis) {
	topLines := analysis.GetTopExpensiveLines(topExpensiveN)

	// Calculate total tokens in top lines
	totalTopTokens := 0
	for _, line := range topLines {
		totalTopTokens += line.Tokens
	}
	percentage := float64(totalTopTokens) / float64(analysis.TotalTokens) * 100

	// Section header
	f.printSectionHeader(fmt.Sprintf("TOP EXPENSIVE LINES (%d lines, %d tokens, %.1f%%)",
		len(topLines), totalTopTokens, percentage))

	// Print each line
	for _, line := range topLines {
		if line.Tokens < 3 {
			break // Stop showing very low token lines
		}

		lineNumStr := fmt.Sprintf("Line %d:", line.LineNumber)
		tokenStr := fmt.Sprintf("%d tokens", line.Tokens)

		preview := line.Content
		if len(preview) > maxLinePreview {
			preview = preview[:maxLinePreview] + "..."
		}
		preview = strings.TrimSpace(preview)

		if f.useColor {
			color.New(color.FgYellow).Printf("%-12s", lineNumStr)
			color.New(color.FgGreen).Printf("%-12s", tokenStr)
			if line.HasUnicode {
				color.New(color.FgMagenta).Printf(" [Unicode] ")
			}
			fmt.Println()
			color.New(color.FgWhite, color.Faint).Printf("  %s\n", preview)
		} else {
			fmt.Printf("%-12s %-12s", lineNumStr, tokenStr)
			if line.HasUnicode {
				fmt.Printf(" [Unicode]")
			}
			fmt.Println()
			fmt.Printf("  %s\n", preview)
		}

	}
}

func (f *AnalysisFormatter) printRecommendations(analysis *analyzer.Analysis) {
	if len(analysis.Recommendations) == 0 {
		return
	}

	f.printSectionHeader("OPTIMIZATION RECOMMENDATIONS")

	for _, rec := range analysis.Recommendations {
		if f.useColor {
			title := fmt.Sprintf("• %s", rec.Title)
			savings := fmt.Sprintf("  Savings: ~%d tokens (%.1f%%)",
				rec.EstimatedSave, rec.SavePercentage)

			if rec.Priority == 1 {
				color.New(color.Bold, color.FgGreen).Println(title)
			} else if rec.Priority == 2 {
				color.New(color.FgYellow).Println(title)
			} else {
				color.New(color.FgWhite).Println(title)
			}
			color.New(color.Faint).Println(savings)

			if rec.Description != "" {
				desc := fmt.Sprintf("   %s", rec.Description)
				color.New(color.Faint).Println(desc)
			}

			if len(rec.AffectedLines) > 0 && len(rec.AffectedLines) <= 10 {
				lines := "   Lines: "
				for i, lineNum := range rec.AffectedLines {
					if i > 0 {
						lines += ", "
					}
					lines += fmt.Sprintf("%d", lineNum)
					if i >= 9 {
						lines += fmt.Sprintf(", ... (%d more)", len(rec.AffectedLines)-10)
						break
					}
				}
				color.New(color.Faint).Println(lines)
			} else if len(rec.AffectedLines) > 10 {
				lines := fmt.Sprintf("   %d lines affected", len(rec.AffectedLines))
				color.New(color.Faint).Println(lines)
			}
		} else {
			// Plain mode: compact format with priority indicator
			priority := ""
			if rec.Priority == 1 {
				priority = "[HIGH] "
			} else if rec.Priority == 2 {
				priority = "[MED] "
			} else {
				priority = "[LOW] "
			}

			fmt.Printf("%s%s (Impact: %d tokens, %.1f%%",
				priority, rec.Title, rec.EstimatedSave, rec.SavePercentage)

			if len(rec.AffectedLines) > 0 {
				fmt.Printf(", %d lines", len(rec.AffectedLines))
			}
			fmt.Println(")")
		}
	}
}

// IssueSection represents a formatted issue section
type IssueSection struct {
	Title       string
	Count       int
	Impact      string
	Fix         string
	CriticalMsg string // optional critical warning
}

func (f *AnalysisFormatter) printIssueSection(section IssueSection) {
	f.printSubheader(section.Title)
	fmt.Printf("  Found %d %s\n", section.Count, section.Title)

	if section.CriticalMsg != "" {
		if f.useColor {
			color.New(color.FgRed, color.Bold).Printf("  %s\n", section.CriticalMsg)
		} else {
			fmt.Printf("  %s\n", section.CriticalMsg)
		}
	}

	fmt.Printf("  Impact: %s\n", section.Impact)
	fmt.Printf("  Fix: %s\n", section.Fix)
}

func (f *AnalysisFormatter) printLLMSafetyAnalysis(analysis *analyzer.Analysis) {
	safetyAnalysis := analysis.LLMSafetyAnalysis
	if safetyAnalysis == nil {
		return
	}

	f.printSectionHeader(fmt.Sprintf("LLM SAFETY ANALYSIS (Reliability Score: %d/100)", safetyAnalysis.ReliabilityScore))

	// Count evasion patterns
	evasionCount := 0
	for _, issue := range safetyAnalysis.InvisibleCharIssues {
		if issue.IsEvasion {
			evasionCount++
		}
	}

	criticalMsg := ""
	if evasionCount > 0 {
		criticalMsg = fmt.Sprintf("CRITICAL: %d potential prompt injection/evasion patterns detected!", evasionCount)
	}

	// Define issue sections
	// Count Trojan Source patterns for BiDi controls
	trojanSourceMsg := ""
	trojanSourceCount := 0
	for _, issue := range safetyAnalysis.BiDiControlIssues {
		if issue.IsTrojanSource {
			trojanSourceCount++
		}
	}
	if trojanSourceCount > 0 {
		trojanSourceMsg = fmt.Sprintf("CRITICAL: %d Trojan Source attack patterns detected!", trojanSourceCount)
	}

	sections := []IssueSection{
		{
			Title:  "emoji issues (tokenization cost)",
			Count:  len(safetyAnalysis.EmojiIssues),
			Impact: "Reduce judge reliability by 23-47% (arXiv:2411.01077)",
			Fix:    "Replace emojis with text tags (:smile:, :rocket:, etc.)",
		},
		{
			Title:       "invisible character issues (zero-width, control chars)",
			Count:       len(safetyAnalysis.InvisibleCharIssues),
			Impact:      "Enable prompt injection, confuse model reasoning (Trend Micro 2025)",
			Fix:         "Remove all zero-width and invisible characters",
			CriticalMsg: criticalMsg,
		},
		{
			Title:       "BiDi control characters (Trojan Source)",
			Count:       len(safetyAnalysis.BiDiControlIssues),
			Impact:      "Enable code injection attacks (CVE-2021-42574)",
			Fix:         "Remove all bidirectional text control characters",
			CriticalMsg: trojanSourceMsg,
		},
		{
			Title:  "homoglyphs/confusable characters",
			Count:  len(safetyAnalysis.ConfusableIssues),
			Impact: "Enable spoofing and phishing attacks (UTS #39)",
			Fix:    "Replace with ASCII equivalents or flag mixed-script identifiers",
		},
		{
			Title:  "encoded/obfuscated text (Base64, hex, leetspeak)",
			Count:  len(safetyAnalysis.EncodingIssues),
			Impact: "Bypass moderation and confuse models (NeurIPS 2024 JAM)",
			Fix:    "Decode or remove encoded text before processing",
		},
		{
			Title:  "Unicode normalization issues",
			Count:  len(safetyAnalysis.NormalizationIssues),
			Impact: "Cause tokenization inconsistencies (UAX #15)",
			Fix:    "Normalize all text to NFC form",
		},
		{
			Title:  "glitch tokens",
			Count:  len(safetyAnalysis.GlitchTokenIssues),
			Impact: "Cause unstable model behavior (arXiv:2404.09894)",
			Fix:    "Remove or space-separate known glitch tokens",
		},
		{
			Title:  "long context placement issues",
			Count:  len(safetyAnalysis.ContextIssues),
			Impact: "Reduce attention to middle sections (arXiv:2307.03172)",
			Fix:    "Move key facts to start/end; add TL;DR and recap",
		},
		{
			Title:  "prompt ambiguity patterns",
			Count:  len(safetyAnalysis.AmbiguityIssues),
			Impact: "Reduce truthfulness and accuracy (PLOS ONE 2025)",
			Fix:    "Clarify instructions; remove sycophantic framing",
		},
		{
			Title:  "unformatted large numbers",
			Count:  len(safetyAnalysis.NumberFormatIssues),
			Impact: "Reduces arithmetic accuracy by 8-15%",
			Fix:    "Format with commas (1,234,567 instead of 1234567)",
		},
		{
			Title:  "OOV strings (URLs, hashes, IDs, tokens)",
			Count:  len(safetyAnalysis.OOVStringIssues),
			Impact: "Split into many subword tokens, harming embeddings (arXiv:2406.08477)",
			Fix:    "Use semantic placeholders (<URL>, <HASH>, <UUID>, <TOKEN>)",
		},
	}

	for _, section := range sections {
		if section.Count > 0 {
			f.printIssueSection(section)
		}
	}
}

func (f *AnalysisFormatter) printSubheader(title string) {
	if f.useColor {
		color.New(color.FgYellow, color.Bold).Printf("  • %s\n", title)
	} else {
		fmt.Printf("  • %s\n", title)
	}
}

func (f *AnalysisFormatter) printSummary(analysis *analyzer.Analysis) {
	f.printSectionHeader(fmt.Sprintf("TOTAL POTENTIAL SAVINGS: ~%d tokens (%.1f%% reduction)",
		analysis.PotentialSavings,
		float64(analysis.PotentialSavings)/float64(analysis.TotalTokens)*100))
}

func (f *AnalysisFormatter) printSectionHeader(title string) {
	if f.useColor {
		color.New(color.Bold, color.FgCyan).Printf("## %s\n", title)
	} else {
		fmt.Printf("## %s\n", title)
	}
}
