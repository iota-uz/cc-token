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
