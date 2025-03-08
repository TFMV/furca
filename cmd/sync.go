package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/TFMV/furca/github"
	"github.com/TFMV/furca/logger"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// SyncResult represents the result of a sync operation
type SyncResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Behind int    `json:"behind_by,omitempty"`
}

// SyncSummary represents the summary of all sync operations
type SyncSummary struct {
	Synced    []string          `json:"synced"`
	UpToDate  []string          `json:"up_to_date"`
	Errors    map[string]string `json:"errors"`
	Timestamp string            `json:"timestamp"`
}

var (
	dryRun      bool
	jsonOutput  bool
	maxRetries  int
	retryDelay  int
	successIcon = color.GreenString("✅")
	syncIcon    = color.BlueString("🔄")
	errorIcon   = color.RedString("❌")
	dryRunIcon  = color.YellowString("[DRY-RUN]")
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

		// Create a context for all operations
		ctx := context.Background()

		// Get forked repositories
		log.Info("Fetching forked repositories...")
		forks, err := client.GetForkedRepositories(ctx)
		if err != nil {
			log.Fatalf("Failed to fetch forked repositories: %v", err)
		}

		if len(forks) == 0 {
			log.Info("No forked repositories found with parent information.")
			return
		}

		log.Infof("Found %d forked repositories with parent information", len(forks))

		// Process repositories concurrently
		var wg sync.WaitGroup
		results := make(chan SyncResult, len(forks))

		// Initialize summary
		summary := SyncSummary{
			Synced:    []string{},
			UpToDate:  []string{},
			Errors:    make(map[string]string),
			Timestamp: time.Now().Format(time.RFC3339),
		}

		for _, fork := range forks {
			wg.Add(1)
			go func(fork github.Repository) {
				defer wg.Done()

				log.Debugf("Checking repository: %s", fork.Name)

				// Check if fork is behind upstream with retries
				behind, behindBy, err := checkRepositoryWithRetries(ctx, client, fork, maxRetries, retryDelay)
				if err != nil {
					errMsg := fmt.Sprintf("failed to compare commits: %v", err)
					results <- SyncResult{
						Name:   fork.Name,
						Status: "error",
						Error:  errMsg,
					}
					return
				}

				if !behind {
					if dryRun {
						results <- SyncResult{
							Name:   fork.Name,
							Status: "up_to_date",
							Behind: 0,
						}
					} else {
						results <- SyncResult{
							Name:   fork.Name,
							Status: "up_to_date",
							Behind: 0,
						}
					}
					return
				}

				// If dry run, just report what would happen
				if dryRun {
					results <- SyncResult{
						Name:   fork.Name,
						Status: "would_sync",
						Behind: behindBy,
					}
					return
				}

				// Sync fork with upstream with retries
				log.Debugf("Syncing %s with upstream...", fork.Name)
				err = syncRepositoryWithRetries(ctx, client, fork, maxRetries, retryDelay)
				if err != nil {
					errMsg := fmt.Sprintf("failed to sync repository: %v", err)
					results <- SyncResult{
						Name:   fork.Name,
						Status: "error",
						Error:  errMsg,
					}
					return
				}

				results <- SyncResult{
					Name:   fork.Name,
					Status: "synced",
					Behind: behindBy,
				}
			}(fork)
		}

		// Wait for all goroutines to finish
		go func() {
			wg.Wait()
			close(results)
		}()

		// Process results
		for result := range results {
			switch result.Status {
			case "up_to_date":
				summary.UpToDate = append(summary.UpToDate, result.Name)
				if !jsonOutput {
					if dryRun {
						fmt.Printf("%s %s %s is up to date with upstream\n", dryRunIcon, successIcon, result.Name)
					} else {
						fmt.Printf("%s %s is up to date with upstream\n", successIcon, result.Name)
					}
				}
			case "would_sync":
				summary.Synced = append(summary.Synced, result.Name)
				if !jsonOutput {
					fmt.Printf("%s %s Would sync %s (behind by %d commits)\n", dryRunIcon, syncIcon, result.Name, result.Behind)
				}
			case "synced":
				summary.Synced = append(summary.Synced, result.Name)
				if !jsonOutput {
					fmt.Printf("%s Successfully synced %s with upstream (was behind by %d commits)\n", syncIcon, result.Name, result.Behind)
				}
			case "error":
				summary.Errors[result.Name] = result.Error
				if !jsonOutput {
					fmt.Printf("%s Error checking %s: %s\n", errorIcon, result.Name, result.Error)
				}
			}
		}

		// Print summary or JSON output
		if jsonOutput {
			jsonData, err := json.MarshalIndent(summary, "", "  ")
			if err != nil {
				log.Errorf("Failed to generate JSON output: %v", err)
			} else {
				fmt.Println(string(jsonData))
			}
		} else {
			// Print summary
			fmt.Println("\n📊 Summary:")
			if dryRun {
				fmt.Printf("%s Would sync repositories: %d\n", syncIcon, len(summary.Synced))
			} else {
				fmt.Printf("%s Synced repositories: %d\n", syncIcon, len(summary.Synced))
			}
			fmt.Printf("%s Up-to-date repositories: %d\n", successIcon, len(summary.UpToDate))
			fmt.Printf("%s Errors encountered: %d\n", errorIcon, len(summary.Errors))

			if len(summary.Errors) > 0 {
				fmt.Println("\nSee logs for details.")
			}
		}
	},
}

// checkRepositoryWithRetries checks if a repository is behind its upstream with retries
func checkRepositoryWithRetries(ctx context.Context, client *github.Client, repo github.Repository, maxRetries, retryDelay int) (bool, int, error) {
	var err error
	var behind bool
	var behindBy int

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(retryDelay) * time.Second)
		}

		behind, behindBy, err = client.IsRepositoryBehindUpstream(ctx, repo)
		if err == nil {
			return behind, behindBy, nil
		}

		// Log retry attempt
		if attempt < maxRetries {
			logger.GetLogger().Debugf("Retry %d/%d: checking if %s is behind upstream", attempt+1, maxRetries, repo.Name)
		}
	}

	return false, 0, err
}

// syncRepositoryWithRetries syncs a repository with its upstream with retries
func syncRepositoryWithRetries(ctx context.Context, client *github.Client, repo github.Repository, maxRetries, retryDelay int) error {
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(retryDelay) * time.Second)
		}

		err = client.SyncRepositoryWithUpstream(ctx, repo)
		if err == nil {
			return nil
		}

		// Log retry attempt
		if attempt < maxRetries {
			logger.GetLogger().Debugf("Retry %d/%d: syncing %s with upstream", attempt+1, maxRetries, repo.Name)
		}
	}

	return err
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// Add flags with default values from environment variables
	defaultDryRun := viper.GetBool("DRY_RUN")
	syncCmd.Flags().BoolVar(&dryRun, "dry-run", defaultDryRun, "Preview which repositories would be synced without making changes")

	// JSON output flag with default from environment
	defaultJsonOutput := viper.GetBool("JSON_OUTPUT")
	syncCmd.Flags().BoolVar(&jsonOutput, "json", defaultJsonOutput, "Output results in JSON format")

	// Retry configuration with defaults from environment
	defaultMaxRetries := viper.GetInt("MAX_RETRIES")
	if defaultMaxRetries == 0 {
		defaultMaxRetries = 2 // Default if not set in environment
	}
	syncCmd.Flags().IntVar(&maxRetries, "max-retries", defaultMaxRetries, "Maximum number of retry attempts for API operations")

	defaultRetryDelay := viper.GetInt("RETRY_DELAY")
	if defaultRetryDelay == 0 {
		defaultRetryDelay = 3 // Default if not set in environment
	}
	syncCmd.Flags().IntVar(&retryDelay, "retry-delay", defaultRetryDelay, "Delay in seconds between retry attempts")
}
