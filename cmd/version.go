package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version information
	version string
	commit  string
	date    string
)

// SetVersionInfo sets the version information for Furca.
// It takes the version string, commit hash, and build date as parameters,
// which are typically set during the build process via ldflags.
func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
}

// versionCmd represents the version command which displays the current version
// information for Furca, including the version number, commit hash, and build date.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the version, commit, and build date information for Furca.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Furca version %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s\n", date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
