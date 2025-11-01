package processor

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/iota-uz/cc-token/internal/api"
	"github.com/iota-uz/cc-token/internal/cache"
	"github.com/iota-uz/cc-token/internal/config"
)

// Processor handles file and directory processing for token counting
type Processor struct {
	apiClient *api.Client
	cache     *cache.Cache
	config    *config.Config
}

// New creates a new Processor instance
func New(apiClient *api.Client, c *cache.Cache, cfg *config.Config) *Processor {
	return &Processor{
		apiClient: apiClient,
		cache:     c,
		config:    cfg,
	}
}

// ProcessPath handles processing of a single path, which can be a file, directory, or stdin ("-").
// It dispatches to the appropriate handler based on the path type.
func (p *Processor) ProcessPath(path string) (*Result, error) {
	// Handle stdin
	if path == "-" {
		return p.processStdin()
	}

	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to access %s: %w", path, err)
	}

	if info.IsDir() {
		return p.processDirectory(path)
	}

	result, err := p.processFile(path, info)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// processStdin reads content from stdin, validates the size, and counts tokens.
// It returns an error if the content exceeds the maximum file size.
func (p *Processor) processStdin() (*Result, error) {
	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read from stdin: %w", err)
	}

	// Validate size
	if int64(len(content)) > p.config.MaxSize {
		return nil, fmt.Errorf("stdin content too large (%d bytes, max: %d bytes)", len(content), p.config.MaxSize)
	}

	tokens, err := p.apiClient.CountTokens(string(content), p.config.Model)
	if err != nil {
		return nil, err
	}

	return &Result{
		Path:   "<stdin>",
		Tokens: tokens,
		Cached: false,
	}, nil
}

// processDirectory recursively processes all files in a directory, respecting .gitignore patterns
// and configured filters. It uses goroutines for parallel processing with concurrency control.
func (p *Processor) processDirectory(dirPath string) (*Result, error) {
	// Load gitignore patterns
	gitignorePatterns, err := loadGitignore(dirPath)
	if err != nil && p.config.Verbose {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load .gitignore: %v\n", err)
	}

	// Collect all files
	var files []struct {
		path string
		info os.FileInfo
	}

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories in collection
		if info.IsDir() {
			// Check if directory should be ignored
			if shouldIgnore(path, dirPath, gitignorePatterns, true) {
				return filepath.SkipDir
			}
			return nil
		}

		// Apply filters
		if shouldIgnore(path, dirPath, gitignorePatterns, false) {
			return nil
		}

		if !shouldInclude(path, info, p.config) {
			return nil
		}

		files = append(files, struct {
			path string
			info os.FileInfo
		}{path, info})
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	if len(files) == 0 {
		return &Result{
			Path:  dirPath,
			IsDir: true,
		}, nil
	}

	// Process files with concurrency control
	results := make([]*Result, len(files))
	var wg sync.WaitGroup
	sem := make(chan struct{}, p.config.Concurrency)
	errors := make(chan error, len(files))

	for i, file := range files {
		wg.Add(1)
		go func(i int, path string, info os.FileInfo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result, err := p.processFile(path, info)
			if err != nil {
				errors <- fmt.Errorf("%s: %w", path, err)
				return
			}
			results[i] = result
		}(i, file.path, file.info)
	}

	wg.Wait()
	close(errors)

	// Check for errors (collect all and return first one)
	if len(errors) > 0 {
		// Drain all errors from channel
		var errList []error
		for err := range errors {
			errList = append(errList, err)
		}
		// Return first error (could be enhanced to return all)
		return nil, errList[0]
	}

	// Build tree structure
	tree := buildTree(dirPath, results)
	return tree, nil
}

// processFile processes a single file, checking the cache first and counting tokens via the API
// if needed. It updates the cache with new results and respects the maximum file size limit.
func (p *Processor) processFile(filePath string, info os.FileInfo) (*Result, error) {
	// Check file size
	if info.Size() > p.config.MaxSize {
		return &Result{
			Path:  filePath,
			Error: fmt.Errorf("file too large (%d bytes, max: %d bytes)", info.Size(), p.config.MaxSize),
		}, nil
	}

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return &Result{
			Path:  filePath,
			Error: fmt.Errorf("failed to read file: %w", err),
		}, nil
	}

	// Check cache
	var tokens int
	var cached bool

	if p.cache != nil {
		hash := cache.ComputeHash(content)
		if entry, ok := p.cache.Get(filePath); ok {
			if entry.Hash == hash && entry.Modified.Equal(info.ModTime()) {
				tokens = entry.Tokens
				cached = true
			}
		}
	}

	// Count tokens if not cached
	if !cached {
		tokens, err = p.apiClient.CountTokens(string(content), p.config.Model)
		if err != nil {
			return &Result{
				Path:  filePath,
				Error: err,
			}, nil
		}

		// Update cache
		if p.cache != nil {
			hash := cache.ComputeHash(content)
			p.cache.Set(filePath, cache.Entry{
				Tokens:   tokens,
				Hash:     hash,
				Modified: info.ModTime(),
			})
		}
	}

	return &Result{
		Path:   filePath,
		Tokens: tokens,
		Cached: cached,
	}, nil
}
