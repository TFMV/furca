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

// SetVersionInfo sets the version information
func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
}

// versionCmd represents the version command
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
