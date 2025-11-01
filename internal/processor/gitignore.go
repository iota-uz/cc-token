package processor

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// loadGitignore loads and parses .gitignore patterns from the specified directory.
// It returns an empty list if no .gitignore file exists.
func loadGitignore(dirPath string) ([]string, error) {
	gitignorePath := filepath.Join(dirPath, ".gitignore")
	file, err := os.Open(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}

	return patterns, scanner.Err()
}

// shouldIgnore checks if a file or directory should be ignored based on .gitignore patterns.
// It always ignores the .git directory and matches against provided gitignore patterns.
func shouldIgnore(path, basePath string, patterns []string, isDir bool) bool {
	relPath, err := filepath.Rel(basePath, path)
	if err != nil {
		return false
	}

	// Always ignore .git directory
	if strings.Contains(relPath, ".git"+string(filepath.Separator)) || relPath == ".git" {
		return true
	}

	// Check gitignore patterns
	for _, pattern := range patterns {
		matched, _ := filepath.Match(pattern, filepath.Base(relPath))
		if matched {
			return true
		}

		// Check directory patterns
		if isDir {
			matched, _ = filepath.Match(pattern, filepath.Base(relPath)+"/")
			if matched {
				return true
			}
		}
	}

	return false
}
