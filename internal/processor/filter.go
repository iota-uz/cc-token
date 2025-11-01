package processor

import (
	"os"
	"path/filepath"

	"github.com/iota-uz/cc-token/internal/config"
)

// shouldInclude determines whether a file should be included in processing based on
// size and extension filters configured by the user.
func shouldInclude(path string, info os.FileInfo, cfg *config.Config) bool {
	// Check size
	if info.Size() > cfg.MaxSize {
		return false
	}

	// Check extensions
	if len(cfg.Extensions) > 0 {
		ext := filepath.Ext(path)
		found := false
		for _, allowedExt := range cfg.Extensions {
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
