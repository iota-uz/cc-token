package cmd

import (
	"github.com/iota-uz/cc-token/internal/cache"
	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage token count cache",
	Long: `Manage the local cache of token counts.

The cache is stored in ~/.cc-token/cache.json and helps avoid redundant API calls
by storing previously counted token values along with file hashes and modification times.`,
}

var clearCacheCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the token count cache",
	Long:  `Remove all cached token counts from ~/.cc-token/cache.json`,
	Example: `  # Clear the cache
  cc-token cache clear`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cache.Clear()
	},
}

func init() {
	cacheCmd.AddCommand(clearCacheCmd)
	rootCmd.AddCommand(cacheCmd)
}
