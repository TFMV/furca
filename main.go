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
