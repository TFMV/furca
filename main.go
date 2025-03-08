// Package main provides the entry point for Furca, a tool to keep GitHub forks effortlessly fresh.
//
// Furca automates the synchronization of forked GitHub repositories with their upstream sources.
// It simplifies the developer experience by automatically fetching repository information,
// determining if forks are behind their upstream repositories, and synchronizing them accordingly.
package main

import (
	"fmt"
	"os"

	"github.com/TFMV/furca/cmd"
)

// Version information set by build flags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Make version info available to commands
	cmd.SetVersionInfo(version, commit, date)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
