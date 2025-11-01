package utils

import "strings"

// CalculateLineStarts computes the starting byte position of each line
// Returns a slice where index i contains the starting position of line i
func CalculateLineStarts(lines []string) []int {
	starts := make([]int, len(lines))
	pos := 0
	for i, line := range lines {
		starts[i] = pos
		pos += len(line) + 1 // +1 for newline character
	}
	return starts
}

// FindLineForPosition returns the line index for a given byte position
// Uses binary search for efficiency with large files
func FindLineForPosition(pos int, lineStarts []int) int {
	if len(lineStarts) == 0 {
		return -1
	}

	// Binary search for the line
	left, right := 0, len(lineStarts)-1
	for left <= right {
		mid := (left + right) / 2
		if lineStarts[mid] == pos {
			return mid
		} else if lineStarts[mid] < pos {
			// Check if position falls within this line
			if mid == len(lineStarts)-1 || pos < lineStarts[mid+1] {
				return mid
			}
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return right
}

// CalculateLineMetrics computes line count and average tokens per line
// Returns (lineCount, avgTokensPerLine)
func CalculateLineMetrics(content string, tokens int) (int, float64) {
	lineCount := len(strings.Split(content, "\n"))
	avgTokensPerLine := 0.0
	if lineCount > 0 {
		avgTokensPerLine = float64(tokens) / float64(lineCount)
	}
	return lineCount, avgTokensPerLine
}
