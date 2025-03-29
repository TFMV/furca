// Package cmd implements the command-line interface for Furca.
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

// CICheckResult represents the result of a CI check operation
type CICheckResult struct {
	BehindRepos    []string          `json:"behind_repos"`
	UpToDateRepos  []string          `json:"up_to_date_repos"`
	Errors         map[string]string `json:"errors"`
	Timestamp      string            `json:"timestamp"`
	TotalBehind    int               `json:"total_behind"`
	TotalUpToDate  int               `json:"total_up_to_date"`
	TotalErrors    int               `json:"total_errors"`
	TotalRepos     int               `json:"total_repos"`
	OutdatedStatus bool              `json:"outdated_status"`
}

var (
	failOnOutdated bool
	ciJsonOutput   bool
)

// ciCheckCmd represents the ci-check command
var ciCheckCmd = &cobra.Command{
	Use:   "ci-check",
	Short: "Check if any forks are behind upstream (for CI/CD integration)",
	Long: `The ci-check command checks all your forked repositories to see if any
are behind their upstream sources, without performing any synchronization.

This command is designed for integration with CI/CD pipelines and can be
configured to exit with a non-zero status code if any forks are behind
their upstream repositories.

Example usage in CI/CD pipelines:

  GitHub Actions:
    - name: Check if forks are up to date
      run: furca ci-check --fail-on-outdated

  CircleCI:
    steps:
      - checkout
      - run:
          name: Check if forks are up to date
          command: furca ci-check --fail-on-outdated

  GitLab CI/CD:
    check_forks:
      script:
        - furca ci-check --fail-on-outdated

  Argo or Tekton workflows:
    - name: check-forks
      container:
        image: your-image-with-furca
        command: [furca, ci-check, --fail-on-outdated]`,
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
		type repoStatus struct {
			Name     string
			IsBehind bool
			BehindBy int
			Error    string
		}
		results := make(chan repoStatus, len(forks))

		for _, fork := range forks {
			wg.Add(1)
			go func(fork github.Repository) {
				defer wg.Done()

				log.Debugf("Checking repository: %s", fork.Name)

				// Check if fork is behind upstream
				behind, behindBy, err := client.IsRepositoryBehindUpstream(ctx, fork)
				if err != nil {
					results <- repoStatus{
						Name:  fork.Name,
						Error: fmt.Sprintf("failed to compare commits: %v", err),
					}
					return
				}

				results <- repoStatus{
					Name:     fork.Name,
					IsBehind: behind,
					BehindBy: behindBy,
				}
			}(fork)
		}

		// Wait for all goroutines to finish
		go func() {
			wg.Wait()
			close(results)
		}()

		// Process results
		ciResult := CICheckResult{
			BehindRepos:   []string{},
			UpToDateRepos: []string{},
			Errors:        make(map[string]string),
			Timestamp:     time.Now().Format(time.RFC3339),
		}

		for result := range results {
			if result.Error != "" {
				ciResult.Errors[result.Name] = result.Error
				if !ciJsonOutput {
					fmt.Printf("%s Error checking %s: %s\n", errorIcon, result.Name, result.Error)
				}
			} else if result.IsBehind {
				ciResult.BehindRepos = append(ciResult.BehindRepos, result.Name)
				if !ciJsonOutput {
					fmt.Printf("%s %s is behind upstream by %d commits\n", syncIcon, result.Name, result.BehindBy)
				}
			} else {
				ciResult.UpToDateRepos = append(ciResult.UpToDateRepos, result.Name)
				if !ciJsonOutput {
					fmt.Printf("%s %s is up to date with upstream\n", successIcon, result.Name)
				}
			}
		}

		// Set count fields
		ciResult.TotalBehind = len(ciResult.BehindRepos)
		ciResult.TotalUpToDate = len(ciResult.UpToDateRepos)
		ciResult.TotalErrors = len(ciResult.Errors)
		ciResult.TotalRepos = ciResult.TotalBehind + ciResult.TotalUpToDate + ciResult.TotalErrors
		ciResult.OutdatedStatus = ciResult.TotalBehind > 0

		// Print JSON output if requested
		if ciJsonOutput {
			jsonData, err := json.MarshalIndent(ciResult, "", "  ")
			if err != nil {
				log.Errorf("Failed to generate JSON output: %v", err)
			} else {
				fmt.Println(string(jsonData))
			}
		} else {
			// Print summary
			fmt.Println("\n📊 Summary:")
			fmt.Printf("%s Repositories behind upstream: %d\n", syncIcon, ciResult.TotalBehind)
			fmt.Printf("%s Repositories up to date: %d\n", successIcon, ciResult.TotalUpToDate)
			fmt.Printf("%s Errors encountered: %d\n", errorIcon, ciResult.TotalErrors)
			fmt.Printf("%s Total repositories checked: %d\n", color.CyanString("ℹ️"), ciResult.TotalRepos)

			if ciResult.TotalBehind > 0 {
				fmt.Println("\n" + color.YellowString("⚠️  Some repositories are behind their upstream sources"))
				if failOnOutdated {
					fmt.Println(color.RedString("❌ Exiting with non-zero status code due to --fail-on-outdated flag"))
				}
			}
		}

		// Exit with non-zero status code if any forks are behind and --fail-on-outdated is specified
		if failOnOutdated && ciResult.TotalBehind > 0 {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(ciCheckCmd)

	// Add flags with default values from environment variables
	defaultFailOnOutdated := viper.GetBool("CI_FAIL_ON_OUTDATED")
	ciCheckCmd.Flags().BoolVar(&failOnOutdated, "fail-on-outdated", defaultFailOnOutdated, "Exit with non-zero status code if any repositories are behind upstream")

	// JSON output flag with default from environment
	defaultJsonOutput := viper.GetBool("JSON_OUTPUT")
	ciCheckCmd.Flags().BoolVar(&ciJsonOutput, "json", defaultJsonOutput, "Output results in JSON format")
}
