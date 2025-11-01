// Package cache provides a file-based caching system for token counts.
package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// File and directory permissions
	cacheDirPerm = 0755
	// FilePerm is the default file permission for cache and export files
	FilePerm = 0644
)

// Entry represents a cached token count
type Entry struct {
	Tokens   int       `json:"tokens"`
	Hash     string    `json:"hash"`
	Modified time.Time `json:"modified"`
}

// Cache holds the token count cache
type Cache struct {
	mu      sync.RWMutex
	entries map[string]Entry
	path    string
}

// Load loads the token count cache from disk, creating a new cache if one doesn't exist.
// The cache is stored in ~/.cc-token/cache.json.
func Load() (*Cache, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".cc-token")
	cachePath := filepath.Join(cacheDir, "cache.json")

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, cacheDirPerm); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	c := &Cache{
		entries: make(map[string]Entry),
		path:    cachePath,
	}

	// Load existing cache
	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	if err := json.Unmarshal(data, &c.entries); err != nil {
		return nil, fmt.Errorf("failed to parse cache file: %w", err)
	}

	return c, nil
}

// Get retrieves a cache entry for the given path in a thread-safe manner.
func (c *Cache) Get(path string) (Entry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[path]
	return entry, ok
}

// Set stores a cache entry for the given path in a thread-safe manner.
func (c *Cache) Set(path string, entry Entry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[path] = entry
}

// Save persists the cache to disk in JSON format.
func (c *Cache) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := json.MarshalIndent(c.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	if err := os.WriteFile(c.path, data, FilePerm); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Clear removes the cache file from disk and prints a confirmation message.
func Clear() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
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

// ComputeHash calculates the SHA-256 hash of the given content for cache validation.
func ComputeHash(content []byte) string {
	hash := sha256.Sum256(content)
	return fmt.Sprintf("%x", hash)
}
