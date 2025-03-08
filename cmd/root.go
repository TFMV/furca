// Package cmd implements the command-line interface for Furca.
//
// It provides commands for synchronizing GitHub forks with their upstream repositories,
// as well as utility functions for configuration management and command execution.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "furca",
	Short: "Furca - Keep your GitHub forks effortlessly fresh",
	Long: `Furca is a command-line tool that automates the synchronization of 
forked GitHub repositories with their upstream sources.

It simplifies the developer experience by automatically fetching repository 
information, determining if forks are behind their upstream repositories, 
and synchronizing them accordingly when executed.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	// Find home directory
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error finding home directory:", err)
		os.Exit(1)
	}

	// Search for config in home directory with name ".furca" (without extension)
	viper.AddConfigPath(home)
	viper.SetConfigName(".furca")
	viper.SetConfigType("env")

	// Also look for .env file in current directory
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")

	// Read in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
