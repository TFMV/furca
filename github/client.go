package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

// Repository represents a GitHub repository
type Repository struct {
	Owner       string
	Name        string
	FullName    string
	ParentOwner string
	ParentName  string
}

// Client is a wrapper around the GitHub API client
type Client struct {
	client *github.Client
	user   *github.User
}

// NewClient creates a new GitHub client with the provided token
func NewClient(token string) (*Client, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Get authenticated user
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated user: %w", err)
	}

	return &Client{
		client: client,
		user:   user,
	}, nil
}

// GetForkedRepositories returns a list of repositories that are forks
func (c *Client) GetForkedRepositories(ctx context.Context) ([]Repository, error) {
	// First, get all repositories for the authenticated user
	opts := &github.RepositoryListByAuthenticatedUserOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		// Get all repositories regardless of visibility
		Visibility: "all",
		// Include all repositories regardless of affiliation
		Affiliation: "owner,collaborator,organization_member",
	}

	var allRepos []*github.Repository
	for {
		repos, resp, err := c.client.Repositories.ListByAuthenticatedUser(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	fmt.Printf("Found %d total repositories\n", len(allRepos))

	// Now identify which ones are forks
	var forks []Repository
	var forkCount int
	for _, repo := range allRepos {
		// Check if this is a fork
		if repo.GetFork() {
			forkCount++
			fmt.Printf("Processing fork #%d: %s\n", forkCount, repo.GetFullName())

			// For each fork, we need to get the full repository details to access parent info
			fullRepo, _, err := c.client.Repositories.Get(ctx, repo.GetOwner().GetLogin(), repo.GetName())
			if err != nil {
				fmt.Printf("Error getting details for %s: %v\n", repo.GetFullName(), err)
				continue
			}

			// Check if parent information is available
			parent := fullRepo.GetParent()
			if parent != nil {
				forks = append(forks, Repository{
					Owner:       fullRepo.GetOwner().GetLogin(),
					Name:        fullRepo.GetName(),
					FullName:    fullRepo.GetFullName(),
					ParentOwner: parent.GetOwner().GetLogin(),
					ParentName:  parent.GetName(),
				})
				fmt.Printf("Added fork: %s (parent: %s)\n", fullRepo.GetFullName(), parent.GetFullName())
			} else {
				fmt.Printf("Warning: Fork %s has no parent information\n", fullRepo.GetFullName())
			}
		}
	}

	fmt.Printf("Identified %d forks with parent information\n", len(forks))
	return forks, nil
}

// IsRepositoryBehindUpstream checks if a forked repository is behind its upstream
func (c *Client) IsRepositoryBehindUpstream(ctx context.Context, repo Repository) (bool, error) {
	comparison, _, err := c.client.Repositories.CompareCommits(
		ctx,
		repo.Owner,
		repo.Name,
		fmt.Sprintf("%s:%s", repo.ParentOwner, "main"),
		"main",
		&github.ListOptions{},
	)
	if err != nil {
		// Try with master branch if main fails
		comparison, _, err = c.client.Repositories.CompareCommits(
			ctx,
			repo.Owner,
			repo.Name,
			fmt.Sprintf("%s:%s", repo.ParentOwner, "master"),
			"master",
			&github.ListOptions{},
		)
		if err != nil {
			return false, fmt.Errorf("failed to compare commits: %w", err)
		}
	}

	// If AheadBy > 0, the fork has commits that the upstream doesn't
	// If BehindBy > 0, the fork is behind the upstream
	return comparison.GetBehindBy() > 0, nil
}

// SyncRepositoryWithUpstream syncs a forked repository with its upstream
func (c *Client) SyncRepositoryWithUpstream(ctx context.Context, repo Repository) error {
	// Try with main branch first
	err := c.syncBranch(ctx, repo, "main")
	if err != nil {
		// Try with master branch if main fails
		err = c.syncBranch(ctx, repo, "master")
		if err != nil {
			return fmt.Errorf("failed to sync repository: %w", err)
		}
	}
	return nil
}

// syncBranch syncs a specific branch with its upstream
func (c *Client) syncBranch(ctx context.Context, repo Repository, branch string) error {
	// Create a merge commit by merging the upstream branch into the fork
	// This is a direct API call since the MergeUpstream method might not be available in all versions
	url := fmt.Sprintf("repos/%s/%s/merge-upstream", repo.Owner, repo.Name)
	req := struct {
		Branch string `json:"branch"`
	}{
		Branch: branch,
	}

	resp, err := c.client.NewRequest(http.MethodPost, url, req)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = c.client.Do(ctx, resp, nil)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}

	return nil
}
