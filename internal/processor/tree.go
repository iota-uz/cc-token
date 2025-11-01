package processor

import (
	"path/filepath"
	"sort"
)

// buildTree constructs a hierarchical tree structure from a flat list of file results,
// grouping files by directory and calculating total token counts.
func buildTree(rootPath string, results []*Result) *Result {
	root := &Result{
		Path:     rootPath,
		IsDir:    true,
		Children: make([]*Result, 0),
	}

	// Group by directory
	filesByDir := make(map[string][]*Result)
	for _, result := range results {
		if result == nil {
			continue
		}
		dir := filepath.Dir(result.Path)
		filesByDir[dir] = append(filesByDir[dir], result)
	}

	// Sort directories
	var dirs []string
	for dir := range filesByDir {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)

	// Add files to tree
	for _, dir := range dirs {
		files := filesByDir[dir]
		sort.Slice(files, func(i, j int) bool {
			return files[i].Path < files[j].Path
		})
		root.Children = append(root.Children, files...)
	}

	// Calculate total tokens
	for _, child := range root.Children {
		if child.Error == nil {
			root.Tokens += child.Tokens
		}
	}

	return root
}
