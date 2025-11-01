package output

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/iota-uz/cc-token/internal/analyzer"
	"github.com/iota-uz/cc-token/internal/config"
)

const (
	separatorLine  = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
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
	fmt.Println()
}

func (f *AnalysisFormatter) printDensityMap(analysis *analyzer.Analysis) {
	if analysis.DensityMap == nil {
		return
	}

	f.printSectionHeader("TOKEN DENSITY HEATMAP")
	heatmap := analysis.DensityMap.FormatHeatmap()
	fmt.Print(heatmap)
	fmt.Println()
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
		icon   string
	}{
		{"Prose", analysis.CategoryBreakdown.Prose, stats.Prose, "📝"},
		{"Code Blocks", analysis.CategoryBreakdown.CodeBlocks, stats.CodeBlocks, "💻"},
		{"URLs", analysis.CategoryBreakdown.URLs, stats.URLs, "🔗"},
		{"Formatting", analysis.CategoryBreakdown.Formatting, stats.Formatting, "✨"},
		{"Whitespace", analysis.CategoryBreakdown.Whitespace, stats.Whitespace, "⬜"},
	}

	for _, cat := range categories {
		if cat.tokens == 0 {
			continue
		}

		icon := cat.icon
		if !f.useColor {
			icon = "" // No emojis in plain mode
		}

		bar := analyzer.RenderCategoryBar(cat.pct, 24)
		line := fmt.Sprintf("%-15s %s %6d tokens (%5.1f%%)",
			icon+" "+cat.name+":", bar, cat.tokens, cat.pct)

		if f.useColor {
			color.New(color.FgWhite).Println(line)
		} else {
			fmt.Println(line)
		}
	}

	fmt.Println()
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

	fmt.Println()
}

func (f *AnalysisFormatter) printQuickWins(analysis *analyzer.Analysis) {
	f.printSectionHeader("QUICK WINS (Easy + High Impact)")

	for _, rec := range analysis.QuickWins {
		icon := "⚡"
		if !f.useColor {
			icon = "•"
		}

		title := fmt.Sprintf("%s %s", icon, rec.Title)
		savings := fmt.Sprintf("   Savings: ~%d tokens (%.1f%%)", rec.EstimatedSave, rec.SavePercentage)

		if f.useColor {
			color.New(color.Bold, color.FgGreen).Println(title)
			color.New(color.Faint).Println(savings)
		} else {
			fmt.Println(title)
			fmt.Println(savings)
		}

		// Before/After if available
		if rec.BeforeExample != "" && rec.AfterExample != "" {
			before := fmt.Sprintf("   Before: %s", rec.BeforeExample)
			after := fmt.Sprintf("   After:  %s", rec.AfterExample)
			if f.useColor {
				color.New(color.Faint).Println(before)
				color.New(color.FgGreen, color.Faint).Println(after)
			} else {
				fmt.Println(before)
				fmt.Println(after)
			}
		}

		if rec.Description != "" {
			desc := fmt.Sprintf("   %s", rec.Description)
			if f.useColor {
				color.New(color.Faint).Println(desc)
			} else {
				fmt.Println(desc)
			}
		}

		fmt.Println()
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
	for i, line := range topLines {
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

		if i < len(topLines)-1 {
			fmt.Println()
		}
	}

	fmt.Println()
}

func (f *AnalysisFormatter) printRecommendations(analysis *analyzer.Analysis) {
	if len(analysis.Recommendations) == 0 {
		return
	}

	f.printSectionHeader("OPTIMIZATION RECOMMENDATIONS")

	for _, rec := range analysis.Recommendations {
		icon := "🔹"
		if !f.useColor {
			icon = "•"
		}

		title := fmt.Sprintf("%s %s", icon, rec.Title)
		savings := fmt.Sprintf("   Savings: ~%d tokens (%.1f%%)",
			rec.EstimatedSave, rec.SavePercentage)

		if f.useColor {
			if rec.Priority == 1 {
				color.New(color.Bold, color.FgGreen).Println(title)
			} else if rec.Priority == 2 {
				color.New(color.FgYellow).Println(title)
			} else {
				color.New(color.FgWhite).Println(title)
			}
			color.New(color.Faint).Println(savings)
		} else {
			fmt.Println(title)
			fmt.Println(savings)
		}

		if rec.Description != "" {
			desc := fmt.Sprintf("   %s", rec.Description)
			if f.useColor {
				color.New(color.Faint).Println(desc)
			} else {
				fmt.Println(desc)
			}
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
			if f.useColor {
				color.New(color.Faint).Println(lines)
			} else {
				fmt.Println(lines)
			}
		} else if len(rec.AffectedLines) > 10 {
			lines := fmt.Sprintf("   %d lines affected", len(rec.AffectedLines))
			if f.useColor {
				color.New(color.Faint).Println(lines)
			} else {
				fmt.Println(lines)
			}
		}

		fmt.Println()
	}
}

func (f *AnalysisFormatter) printSummary(analysis *analyzer.Analysis) {
	f.printSectionHeader(fmt.Sprintf("TOTAL POTENTIAL SAVINGS: ~%d tokens (%.1f%% reduction)",
		analysis.PotentialSavings,
		float64(analysis.PotentialSavings)/float64(analysis.TotalTokens)*100))
}

func (f *AnalysisFormatter) printSectionHeader(title string) {
	if f.useColor {
		color.New(color.FgCyan).Println(separatorLine)
		color.New(color.Bold, color.FgWhite).Println(title)
		color.New(color.FgCyan).Println(separatorLine)
	} else {
		fmt.Println(separatorLine)
		fmt.Println(title)
		fmt.Println(separatorLine)
	}
	fmt.Println()
}
