package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/TFMV/furca/github"
	"github.com/TFMV/furca/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize your forked repositories with their upstream sources",
	Long: `The sync command fetches all your forked repositories and synchronizes them
with their upstream sources if they are behind.

It requires a GitHub token with appropriate permissions, which can be provided
via the GITHUB_TOKEN environment variable or in a .env file.`,
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.GetLogger()

		// Get GitHub token from environment
		token := viper.GetString("GITHUB_TOKEN")
		if token == "" {
			fmt.Println("\n❌ ERROR: GitHub token not found")
			fmt.Println("\nTo use Furca, you need to provide a GitHub personal access token with 'repo' scope.")
			fmt.Println("\nYou can set it in one of these ways:")
			fmt.Println("  1. Create a .env file in the current directory with:")
			fmt.Println("     GITHUB_TOKEN=your_github_token_here")
			fmt.Println("  2. Set an environment variable:")
			fmt.Println("     export GITHUB_TOKEN=your_github_token_here")
			fmt.Println("\nTo create a token, visit: https://github.com/settings/tokens")
			os.Exit(1)
		}

		// Create GitHub client
		client, err := github.NewClient(token)
		if err != nil {
			log.Fatalf("Failed to create GitHub client: %v", err)
		}

		// Get forked repositories
		log.Info("Fetching forked repositories...")
		forks, err := client.GetForkedRepositories(context.Background())
		if err != nil {
			log.Fatalf("Failed to fetch forked repositories: %v", err)
		}

		if len(forks) == 0 {
			log.Info("No forked repositories found.")
			return
		}

		log.Infof("Found %d forked repositories", len(forks))

		// Process repositories concurrently
		var wg sync.WaitGroup
		results := make(chan string, len(forks))

		for _, fork := range forks {
			wg.Add(1)
			go func(fork github.Repository) {
				defer wg.Done()

				log.Infof("Checking repository: %s", fork.Name)

				// Check if fork is behind upstream
				behind, err := client.IsRepositoryBehindUpstream(context.Background(), fork)
				if err != nil {
					results <- fmt.Sprintf("❌ Error checking %s: %v", fork.Name, err)
					return
				}

				if !behind {
					results <- fmt.Sprintf("✅ %s is up to date with upstream", fork.Name)
					return
				}

				// Sync fork with upstream
				log.Infof("Syncing %s with upstream...", fork.Name)
				if err := client.SyncRepositoryWithUpstream(context.Background(), fork); err != nil {
					results <- fmt.Sprintf("❌ Failed to sync %s: %v", fork.Name, err)
					return
				}

				results <- fmt.Sprintf("🔄 Successfully synced %s with upstream", fork.Name)
			}(fork)
		}

		// Wait for all goroutines to finish
		go func() {
			wg.Wait()
			close(results)
		}()

		// Print results
		for result := range results {
			fmt.Println(result)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
