// Package processor handles file and directory processing for token counting.
package processor

// Result holds token count result for a file or directory
type Result struct {
	Path     string
	Tokens   int
	Cached   bool
	Error    error
	IsDir    bool
	Children []*Result
}

// CountFiles recursively counts the number of successfully processed files in this result
func (r *Result) CountFiles() int {
	if !r.IsDir {
		if r.Error == nil {
			return 1
		}
		return 0
	}

	count := 0
	for _, child := range r.Children {
		count += child.CountFiles()
	}
	return count
}
