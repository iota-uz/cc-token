// Package main is the entry point for the cc-token CLI tool.
package main

import (
	"os"

	"github.com/iota-uz/cc-token/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
