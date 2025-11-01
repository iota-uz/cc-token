// Package main implements cc-token, a CLI tool for counting tokens in files and directories
// using Anthropic's Claude API. It supports caching, parallel processing, and multiple output
// formats to help estimate API costs before sending content to Claude.
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"
)

// Version information
const version = "1.0.0"

// API constants
const (
	apiURL         = "https://api.anthropic.com/v1/messages/count_tokens"
	apiVersion     = "2023-06-01"
	defaultModel   = "claude-sonnet-4-5"
	defaultTimeout = 30 * time.Second
)

// Model pricing (USD per 1M tokens - input pricing)
// Source: https://www.anthropic.com/pricing (as of 2025-11-01)
var modelPricing = map[string]float64{
	// Claude 4.x models
	"claude-sonnet-4-5": 3.00,  // Claude Sonnet 4.5
	"claude-sonnet-4.5": 3.00,  // Alternate format
	"claude-haiku-4-5":  1.00,  // Claude Haiku 4.5
	"claude-haiku-4.5":  1.00,  // Alternate format
	"claude-opus-4-1":   15.00, // Claude Opus 4.1
	"claude-opus-4.1":   15.00, // Alternate format
	"claude-sonnet-4":   3.00,  // Claude Sonnet 4
	"claude-4-sonnet":   3.00,  // Alternate format
	"claude-opus-4":     15.00, // Generic Claude Opus 4 (fallback to 4.1 pricing)
	"claude-haiku-4":    1.00,  // Generic Claude Haiku 4 (fallback to 4.5 pricing)

	// Claude 3.x models
	"claude-haiku-3-5":  0.80,  // Claude Haiku 3.5
	"claude-3-5-haiku":  0.80,  // Alternate format
	"claude-haiku-3.5":  0.80,  // Alternate format
	"claude-sonnet-3-7": 3.00,  // Claude Sonnet 3.7 (legacy)
	"claude-3-7-sonnet": 3.00,  // Alternate format
	"claude-sonnet-3.7": 3.00,  // Alternate format
	"claude-3-5-sonnet": 3.00,  // Claude Sonnet 3.5 (legacy, same as 3.7)
	"claude-sonnet-3-5": 3.00,  // Alternate format
	"claude-sonnet-3.5": 3.00,  // Alternate format
	"claude-opus-3":     15.00, // Claude Opus 3 (legacy)
	"claude-3-opus":     15.00, // Alternate format
	"claude-haiku-3":    0.25,  // Claude Haiku 3 (legacy)
	"claude-3-haiku":    0.25,  // Alternate format
	"claude-sonnet-3":   3.00,  // Claude Sonnet 3 (legacy)
	"claude-3-sonnet":   3.00,  // Alternate format
}

// Config holds CLI configuration
type Config struct {
	Model       string
	Extensions  []string
	MaxSize     int64
	Concurrency int
	ShowCost    bool
	JSONOutput  bool
	Verbose     bool
	NoCache     bool
	ClearCache  bool
	Version     bool
}

// TokenResult holds token count result for a file
type TokenResult struct {
	Path     string
	Tokens   int
	Cached   bool
	Error    error
	IsDir    bool
	Children []*TokenResult
}

// CacheEntry represents a cached token count
type CacheEntry struct {
	Tokens   int       `json:"tokens"`
	Hash     string    `json:"hash"`
	Modified time.Time `json:"modified"`
}

// APIRequest represents the token counting API request
type APIRequest struct {
	Model    string         `json:"model"`
	Messages []MessageInput `json:"messages"`
}

// MessageInput represents a message in the API request
type MessageInput struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// APIResponse represents the token counting API response
type APIResponse struct {
	InputTokens int `json:"input_tokens"`
}

// Cache holds the token count cache
type Cache struct {
	mu      sync.RWMutex
	entries map[string]CacheEntry
	path    string
}

var (
	httpClient *http.Client
	cache      *Cache
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}
}

// run is the main entry point for the application. It parses flags, initializes the cache,
// processes input paths, and outputs results in the specified format.
func run() error {
	config, err := parseFlags()
	if err != nil {
		return err
	}

	if config.Version {
		printVersion()
		return nil
	}

	if config.ClearCache {
		return clearCache()
	}

	// Validate API key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY environment variable is not set.\nGet your API key from: https://console.anthropic.com/")
	}

	// Initialize HTTP client
	httpClient = &http.Client{
		Timeout: defaultTimeout,
	}

	// Initialize cache
	if !config.NoCache {
		cache, err = loadCache()
		if err != nil && config.Verbose {
			fmt.Fprintf(os.Stderr, "Warning: Failed to load cache: %v\n", err)
		}
	}

	// Get input path
	args := flag.Args()
	if len(args) == 0 {
		return fmt.Errorf("no input specified.\nUsage: cc-token [flags] <path-to-file-or-directory>")
	}

	// Process each path
	var results []*TokenResult
	for _, path := range args {
		result, err := processPath(path, config, apiKey)
		if err != nil {
			return err
		}
		results = append(results, result)
	}

	// Save cache
	if cache != nil {
		if err := cache.save(); err != nil && config.Verbose {
			fmt.Fprintf(os.Stderr, "Warning: Failed to save cache: %v\n", err)
		}
	}

	// Output results
	return outputResults(results, config)
}

// parseFlags parses command-line flags and returns a Config struct with validated settings.
// It also resolves model aliases to full model names and validates flag values.
func parseFlags() (*Config, error) {
	config := &Config{}

	modelFlag := flag.String("model", defaultModel, "Model to use for token counting (supports aliases: sonnet, haiku, opus)")
	extFlag := flag.String("ext", "", "Comma-separated list of file extensions to include (e.g., .go,.txt,.md)")
	flag.Int64Var(&config.MaxSize, "max-size", 10*1024*1024, "Maximum file size in bytes (default: 10MB)")
	flag.IntVar(&config.Concurrency, "concurrency", 5, "Number of concurrent API requests for directories")
	flag.BoolVar(&config.ShowCost, "show-cost", true, "Show estimated API cost")
	flag.BoolVar(&config.JSONOutput, "json", false, "Output results in JSON format")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&config.NoCache, "no-cache", false, "Disable caching")
	flag.BoolVar(&config.ClearCache, "clear-cache", false, "Clear the cache and exit")
	flag.BoolVar(&config.Version, "version", false, "Print version information")

	flag.Usage = printUsage
	flag.Parse()

	// Resolve model aliases
	config.Model = resolveModelAlias(*modelFlag)

	// Parse extensions
	if *extFlag != "" {
		config.Extensions = strings.Split(*extFlag, ",")
		for i, ext := range config.Extensions {
			ext = strings.TrimSpace(ext)
			if ext != "" && !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			config.Extensions[i] = ext
		}
	}

	// Validate flags
	if config.Concurrency <= 0 {
		return nil, fmt.Errorf("concurrency must be greater than 0")
	}
	if config.MaxSize <= 0 {
		return nil, fmt.Errorf("max-size must be greater than 0")
	}

	return config, nil
}

// resolveModelAlias converts short model aliases (haiku, sonnet, opus) to their full
// model names. It performs case-insensitive matching and returns the original model
// name if no alias is found.
func resolveModelAlias(model string) string {
	// Map of short aliases to full model names (latest versions)
	aliases := map[string]string{
		"sonnet": "claude-sonnet-4-5", // Latest Sonnet (Claude 4.5)
		"haiku":  "claude-haiku-4-5",  // Latest Haiku (Claude 4.5)
		"opus":   "claude-opus-4-1",   // Latest Opus (Claude 4.1)
	}

	// Convert to lowercase for case-insensitive matching
	modelLower := strings.ToLower(strings.TrimSpace(model))

	if resolved, ok := aliases[modelLower]; ok {
		return resolved
	}

	// Return original if no alias found
	return model
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `cc-token - Claude token counting tool

Usage: cc-token [flags] <path-to-file-or-directory>

Flags:
`)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
Examples:
  cc-token document.txt                    # Count tokens in a single file
  cc-token .                               # Count tokens in current directory
  cc-token -ext .go,.md src/               # Count only .go and .md files
  cc-token -model haiku file.txt           # Use Haiku 4.5 (fast & cheap)
  cc-token -model opus document.txt        # Use Opus 4.1 (most capable)
  cc-token -json . > tokens.json           # Output JSON format
  cat file.txt | cc-token -                # Read from stdin

Model Aliases (Latest Claude 4.x):
  haiku                                    # Claude Haiku 4.5 (fastest/cheapest)
  sonnet                                   # Claude Sonnet 4.5 (best balance, default)
  opus                                     # Claude Opus 4.1 (most capable)

Environment Variables:
  ANTHROPIC_API_KEY    Your Anthropic API key (required)
`)
}

func printVersion() {
	fmt.Printf("cc-token version %s\n", version)
	if info, ok := debug.ReadBuildInfo(); ok {
		fmt.Printf("Go version: %s\n", info.GoVersion)
	}
}

// processPath handles processing of a single path, which can be a file, directory, or stdin ("-").
// It dispatches to the appropriate handler based on the path type.
func processPath(path string, config *Config, apiKey string) (*TokenResult, error) {
	// Handle stdin
	if path == "-" {
		return processStdin(config, apiKey)
	}

	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to access %s: %w", path, err)
	}

	if info.IsDir() {
		return processDirectory(path, config, apiKey)
	}

	result, err := processFile(path, info, config, apiKey)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// processStdin reads content from stdin, validates the size, and counts tokens.
// It returns an error if the content exceeds the maximum file size.
func processStdin(config *Config, apiKey string) (*TokenResult, error) {
	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read from stdin: %w", err)
	}

	// Validate size
	if int64(len(content)) > config.MaxSize {
		return nil, fmt.Errorf("stdin content too large (%d bytes, max: %d bytes)", len(content), config.MaxSize)
	}

	tokens, err := countTokens(string(content), config.Model, apiKey)
	if err != nil {
		return nil, err
	}

	return &TokenResult{
		Path:   "<stdin>",
		Tokens: tokens,
		Cached: false,
	}, nil
}

// processDirectory recursively processes all files in a directory, respecting .gitignore patterns
// and configured filters. It uses goroutines for parallel processing with concurrency control.
func processDirectory(dirPath string, config *Config, apiKey string) (*TokenResult, error) {
	// Load gitignore patterns
	gitignorePatterns, err := loadGitignore(dirPath)
	if err != nil && config.Verbose {
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

		if !shouldInclude(path, info, config) {
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
		return &TokenResult{
			Path:  dirPath,
			IsDir: true,
		}, nil
	}

	// Process files with concurrency control
	results := make([]*TokenResult, len(files))
	var wg sync.WaitGroup
	sem := make(chan struct{}, config.Concurrency)
	errors := make(chan error, len(files))

	for i, file := range files {
		wg.Add(1)
		go func(i int, path string, info os.FileInfo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result, err := processFile(path, info, config, apiKey)
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
func processFile(filePath string, info os.FileInfo, config *Config, apiKey string) (*TokenResult, error) {
	// Check file size
	if info.Size() > config.MaxSize {
		return &TokenResult{
			Path:  filePath,
			Error: fmt.Errorf("file too large (%d bytes, max: %d bytes)", info.Size(), config.MaxSize),
		}, nil
	}

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return &TokenResult{
			Path:  filePath,
			Error: fmt.Errorf("failed to read file: %w", err),
		}, nil
	}

	// Check cache
	var tokens int
	var cached bool

	if cache != nil {
		hash := computeHash(content)
		if entry, ok := cache.get(filePath); ok {
			if entry.Hash == hash && entry.Modified.Equal(info.ModTime()) {
				tokens = entry.Tokens
				cached = true
			}
		}
	}

	// Count tokens if not cached
	if !cached {
		tokens, err = countTokens(string(content), config.Model, apiKey)
		if err != nil {
			return &TokenResult{
				Path:  filePath,
				Error: err,
			}, nil
		}

		// Update cache
		if cache != nil {
			hash := computeHash(content)
			cache.set(filePath, CacheEntry{
				Tokens:   tokens,
				Hash:     hash,
				Modified: info.ModTime(),
			})
		}
	}

	return &TokenResult{
		Path:   filePath,
		Tokens: tokens,
		Cached: cached,
	}, nil
}

// countTokens calls the Anthropic API to count tokens in the given content using the specified model.
// It returns the number of input tokens or an error if the API request fails.
func countTokens(content, model, apiKey string) (int, error) {
	reqBody := APIRequest{
		Model: model,
		Messages: []MessageInput{
			{Role: "user", Content: content},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(jsonData))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", apiVersion)
	req.Header.Set("x-api-key", apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return 0, fmt.Errorf("API returned status %d (failed to read response body: %w)", resp.StatusCode, readErr)
		}
		return 0, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return apiResp.InputTokens, nil
}

// shouldInclude determines whether a file should be included in processing based on
// size and extension filters configured by the user.
func shouldInclude(path string, info os.FileInfo, config *Config) bool {
	// Check size
	if info.Size() > config.MaxSize {
		return false
	}

	// Check extensions
	if len(config.Extensions) > 0 {
		ext := filepath.Ext(path)
		found := false
		for _, allowedExt := range config.Extensions {
			if ext == allowedExt {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
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

// buildTree constructs a hierarchical tree structure from a flat list of file results,
// grouping files by directory and calculating total token counts.
func buildTree(rootPath string, results []*TokenResult) *TokenResult {
	root := &TokenResult{
		Path:     rootPath,
		IsDir:    true,
		Children: make([]*TokenResult, 0),
	}

	// Group by directory
	filesByDir := make(map[string][]*TokenResult)
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

// outputResults formats and outputs the token counting results in either tree or JSON format
// based on the configuration.
func outputResults(results []*TokenResult, config *Config) error {
	if config.JSONOutput {
		return outputJSON(results, config)
	}

	return outputTree(results, config)
}

func outputTree(results []*TokenResult, config *Config) error {
	totalTokens := 0
	totalFiles := 0

	for _, result := range results {
		if result.IsDir {
			printTreeNode(result, "", config.Verbose)
			totalTokens += result.Tokens
			totalFiles += countFiles(result)
		} else {
			if result.Error != nil {
				fmt.Fprintf(os.Stderr, "%s: ERROR - %v\n", result.Path, result.Error)
			} else {
				cachedMark := ""
				if config.Verbose && result.Cached {
					cachedMark = " (cached)"
				}
				fmt.Printf("%s: %d tokens%s\n", result.Path, result.Tokens, cachedMark)
				totalTokens += result.Tokens
				totalFiles++
			}
		}
	}

	// Print summary
	if len(results) > 1 || (len(results) == 1 && results[0].IsDir) {
		fmt.Println(strings.Repeat("-", 50))
		fmt.Printf("Total: %d tokens across %d files\n", totalTokens, totalFiles)

		if config.ShowCost {
			cost := calculateCost(totalTokens, config.Model)
			fmt.Printf("Estimated cost: $%.6f\n", cost)
		}
	} else if config.ShowCost && totalTokens > 0 {
		cost := calculateCost(totalTokens, config.Model)
		fmt.Printf("Estimated cost: $%.6f\n", cost)
	}

	return nil
}

func printTreeNode(node *TokenResult, prefix string, verbose bool) {
	basePath := filepath.Base(node.Path)
	if node.IsDir && len(node.Children) > 0 {
		fmt.Printf("%s%s/\n", prefix, basePath)

		for i, child := range node.Children {
			isLast := i == len(node.Children)-1
			childPrefix := prefix + "  "

			if child.Error != nil {
				fmt.Fprintf(os.Stderr, "%s%s: ERROR - %v\n", childPrefix, filepath.Base(child.Path), child.Error)
			} else {
				cachedMark := ""
				if verbose && child.Cached {
					cachedMark = " (cached)"
				}

				connector := "├─"
				if isLast {
					connector = "└─"
				}

				fmt.Printf("%s%s %s: %d tokens%s\n", prefix, connector, filepath.Base(child.Path), child.Tokens, cachedMark)
			}
		}
	}
}

func countFiles(result *TokenResult) int {
	if !result.IsDir {
		if result.Error == nil {
			return 1
		}
		return 0
	}

	count := 0
	for _, child := range result.Children {
		count += countFiles(child)
	}
	return count
}

func outputJSON(results []*TokenResult, config *Config) error {
	output := make([]map[string]interface{}, 0, len(results))

	for _, result := range results {
		item := map[string]interface{}{
			"path":   result.Path,
			"tokens": result.Tokens,
		}

		if result.Error != nil {
			item["error"] = result.Error.Error()
		}

		if result.Cached {
			item["cached"] = true
		}

		if result.IsDir {
			item["type"] = "directory"
			item["files"] = countFiles(result)
		} else {
			item["type"] = "file"
		}

		if config.ShowCost {
			item["estimated_cost"] = calculateCost(result.Tokens, config.Model)
		}

		output = append(output, item)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// calculateCost estimates the API cost for the given number of tokens using the specified model.
// It returns the cost in USD based on the model's pricing per million input tokens.
func calculateCost(tokens int, model string) float64 {
	pricePerMillion, ok := modelPricing[model]
	if !ok {
		pricePerMillion = 3.00 // Default to Sonnet pricing
	}
	return float64(tokens) * pricePerMillion / 1_000_000
}

// computeHash calculates the SHA-256 hash of the given content for cache validation.
func computeHash(content []byte) string {
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash)
}

// Cache methods

// loadCache loads the token count cache from disk, creating a new cache if one doesn't exist.
// The cache is stored in ~/.cc-token/cache.json.
func loadCache() (*Cache, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	cacheDir := filepath.Join(homeDir, ".cc-token")
	cachePath := filepath.Join(cacheDir, "cache.json")

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}

	c := &Cache{
		entries: make(map[string]CacheEntry),
		path:    cachePath,
	}

	// Load existing cache
	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, &c.entries); err != nil {
		return nil, err
	}

	return c, nil
}

// get retrieves a cache entry for the given path in a thread-safe manner.
func (c *Cache) get(path string) (CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[path]
	return entry, ok
}

// set stores a cache entry for the given path in a thread-safe manner.
func (c *Cache) set(path string, entry CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[path] = entry
}

// save persists the cache to disk in JSON format.
func (c *Cache) save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := json.MarshalIndent(c.entries, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.path, data, 0644)
}

// clearCache removes the cache file from disk and prints a confirmation message.
func clearCache() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	cachePath := filepath.Join(homeDir, ".cc-token", "cache.json")

	if err := os.Remove(cachePath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Cache is already empty")
			return nil
		}
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	fmt.Println("Cache cleared successfully")
	return nil
}
