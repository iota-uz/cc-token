package analyzer

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

const (
	heatmapBlockSize = 50 // Number of lines per heatmap block
	heatmapMaxWidth  = 30 // Max width of heatmap bar
	chartBarMaxWidth = 24 // Max width for category bars
	barCharFilled    = "█"
	barCharEmpty     = "░"
	percentileCount  = 6 // min, 25%, 50%, 75%, 90%, max
)

// TokenDensityMap represents token distribution across file sections
type TokenDensityMap struct {
	Blocks []DensityBlock
}

// DensityBlock represents a section of the file
type DensityBlock struct {
	StartLine  int
	EndLine    int
	Tokens     int
	Percentage float64
	IsHot      bool // Top 20% most dense
}

// PercentileStats holds statistical distribution of tokens per line
type PercentileStats struct {
	Min          int
	Percentile25 int
	Median       int
	Percentile75 int
	Percentile90 int
	Percentile95 int
	Max          int
	Top10Pct     float64 // Percentage of total tokens in top 10% of lines
}

// RenderTokenDensityMap creates ASCII visualization of token distribution
func RenderTokenDensityMap(insights []*LineInsight, totalTokens int) *TokenDensityMap {
	if len(insights) == 0 {
		return &TokenDensityMap{Blocks: []DensityBlock{}}
	}

	// Create blocks
	numBlocks := (len(insights) + heatmapBlockSize - 1) / heatmapBlockSize
	blocks := make([]DensityBlock, 0, numBlocks)

	for i := 0; i < numBlocks; i++ {
		startIdx := i * heatmapBlockSize
		endIdx := (i + 1) * heatmapBlockSize
		if endIdx > len(insights) {
			endIdx = len(insights)
		}

		// Sum tokens in this block
		blockTokens := 0
		for j := startIdx; j < endIdx; j++ {
			blockTokens += insights[j].Tokens
		}

		percentage := 0.0
		if totalTokens > 0 {
			percentage = float64(blockTokens) / float64(totalTokens) * 100
		}

		blocks = append(blocks, DensityBlock{
			StartLine:  startIdx + 1,
			EndLine:    endIdx,
			Tokens:     blockTokens,
			Percentage: percentage,
		})
	}

	// Mark hot blocks (top 20%)
	if len(blocks) > 0 {
		maxTokens := 0
		for _, block := range blocks {
			if block.Tokens > maxTokens {
				maxTokens = block.Tokens
			}
		}
		threshold := float64(maxTokens) * 0.8
		for i := range blocks {
			if float64(blocks[i].Tokens) >= threshold {
				blocks[i].IsHot = true
			}
		}
	}

	return &TokenDensityMap{Blocks: blocks}
}

// FormatHeatmap returns formatted ASCII heatmap
func (m *TokenDensityMap) FormatHeatmap() string {
	if len(m.Blocks) == 0 {
		return "No data"
	}

	// Find max tokens for scaling
	maxTokens := 0
	for _, block := range m.Blocks {
		if block.Tokens > maxTokens {
			maxTokens = block.Tokens
		}
	}

	var sb strings.Builder
	for _, block := range m.Blocks {
		// Format line range
		lineRange := fmt.Sprintf("Line %3d-%-3d:", block.StartLine, block.EndLine)

		// Calculate bar length
		barLength := 0
		if maxTokens > 0 {
			barLength = int(float64(block.Tokens) / float64(maxTokens) * float64(heatmapMaxWidth))
		}
		filledBars := strings.Repeat(barCharFilled, barLength)
		emptyBars := strings.Repeat(barCharEmpty, heatmapMaxWidth-barLength)

		// Hot indicator
		hotIndicator := "  "
		if block.IsHot {
			hotIndicator = "🔥"
		}

		// Format line
		line := fmt.Sprintf("%s %s%s (%d tokens, %.1f%%) %s\n",
			lineRange, filledBars, emptyBars, block.Tokens, block.Percentage, hotIndicator)
		sb.WriteString(line)
	}

	return sb.String()
}

// CalculatePercentiles computes statistical distribution of tokens per line
func CalculatePercentiles(insights []*LineInsight) *PercentileStats {
	if len(insights) == 0 {
		return &PercentileStats{}
	}

	// Extract and sort token counts
	tokenCounts := make([]int, len(insights))
	totalTokens := 0
	for i, insight := range insights {
		tokenCounts[i] = insight.Tokens
		totalTokens += insight.Tokens
	}
	sortInts(tokenCounts)

	// Calculate percentiles
	stats := &PercentileStats{
		Min:          tokenCounts[0],
		Max:          tokenCounts[len(tokenCounts)-1],
		Median:       percentile(tokenCounts, 0.50),
		Percentile25: percentile(tokenCounts, 0.25),
		Percentile75: percentile(tokenCounts, 0.75),
		Percentile90: percentile(tokenCounts, 0.90),
		Percentile95: percentile(tokenCounts, 0.95),
	}

	// Calculate top 10% concentration
	top10Index := int(float64(len(insights)) * 0.9)
	top10Tokens := 0
	for i := top10Index; i < len(tokenCounts); i++ {
		top10Tokens += tokenCounts[i]
	}
	if totalTokens > 0 {
		stats.Top10Pct = float64(top10Tokens) / float64(totalTokens) * 100
	}

	return stats
}

// FormatPercentiles returns formatted percentile statistics
func (s *PercentileStats) FormatPercentiles() string {
	return fmt.Sprintf("Min: %d | 25%%: %d | 50%%: %d | 75%%: %d | 90%%: %d | 95%%: %d | Max: %d",
		s.Min, s.Percentile25, s.Median, s.Percentile75, s.Percentile90, s.Percentile95, s.Max)
}

// RenderCategoryBar creates ASCII bar chart for a category
func RenderCategoryBar(percentage float64, maxWidth int) string {
	if maxWidth <= 0 {
		maxWidth = chartBarMaxWidth
	}

	barLength := int(math.Round(percentage / 100.0 * float64(maxWidth)))
	if barLength > maxWidth {
		barLength = maxWidth
	}
	if barLength < 0 {
		barLength = 0
	}

	return strings.Repeat(barCharFilled, barLength)
}

// Helper functions

func percentile(sorted []int, p float64) int {
	if len(sorted) == 0 {
		return 0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[len(sorted)-1]
	}

	index := p * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	// Linear interpolation
	fraction := index - float64(lower)
	return int(float64(sorted[lower])*(1-fraction) + float64(sorted[upper])*fraction)
}

func sortInts(arr []int) {
	sort.Ints(arr)
}

// CalculateEfficiencyScore computes overall file efficiency (0-100)
func CalculateEfficiencyScore(totalTokens, totalChars, wasteTokens int, avgRatio float64) int {
	if totalTokens == 0 || totalChars == 0 {
		return 0
	}

	// Base score from token/char ratio (lower is better)
	// Typical ratio is ~0.25-0.35 tokens/char
	idealRatio := 0.25
	ratioScore := 100.0
	if avgRatio > idealRatio {
		ratioScore = 100.0 * (idealRatio / avgRatio)
	}
	if ratioScore > 100 {
		ratioScore = 100
	}

	// Waste penalty
	wastePct := float64(wasteTokens) / float64(totalTokens) * 100
	wasteScore := 100.0 - wastePct*2 // Each 1% waste reduces score by 2 points

	// Combined score
	score := (ratioScore*0.6 + wasteScore*0.4)
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return int(score)
}
