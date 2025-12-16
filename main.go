package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

const (
	wikiFirstPageMarker = "Create the first page" // HTML marker indicating wiki has no pages
	wikiTestPagePath    = "/notrealpage"          // Non-existent page used to test write access
	httpTimeout         = 30 * time.Second        // HTTP request timeout
	maxResponseSize     = 10 * 1024 * 1024        // Maximum response body size (10MB)
)

var httpClient = &http.Client{
	Timeout: httpTimeout,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// Repository represents a GitHub repository with wiki information.
type Repository struct {
	Name     string // Repository name
	URL      string // HTML URL of the repository
	HasWiki  bool   // Whether wiki is enabled
	IsPublic bool   // Whether repository is public
}

// Creates a GitHub API client, using authentication if GITHUB_TOKEN is set.
// Authentication provides higher rate limits and access to private resources.
func getGitHubClient(ctx context.Context) *github.Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		return github.NewClient(tc)
	}
	return github.NewClient(nil)
}

// Parses account input supporting "org:" and "user:" prefixes.
// Returns ("org"|"user"|"unknown", accountName).
func parseAccountInput(input string) (accountType, accountName string) {
	if strings.HasPrefix(input, "org:") {
		return "org", strings.TrimPrefix(input, "org:")
	}
	if strings.HasPrefix(input, "user:") {
		return "user", strings.TrimPrefix(input, "user:")
	}
	return "unknown", input
}

// Auto-detects whether a GitHub account is an organization or user.
// It first checks the Organizations API, then falls back to Users API if not found.
func getAccountType(ctx context.Context, client *github.Client, name string) (string, error) {
	org, resp, err := client.Organizations.Get(ctx, name)
	if err == nil && org != nil {
		return "org", nil
	}

	if resp != nil && resp.StatusCode == 404 {
		user, userResp, userErr := client.Users.Get(ctx, name)
		if userErr == nil && user != nil {
			return "user", nil
		}
		if userResp != nil && userResp.StatusCode == 404 {
			return "", fmt.Errorf("account '%s' not found", name)
		}
		return "", userErr
	}

	return "", err
}

// Checks if rate limit has been reached and waits until reset if necessary.
// Outputs waiting message to stderr to avoid polluting stdout with results.
func handleRateLimit(ctx context.Context, client *github.Client, resp *github.Response) {
	if resp != nil && resp.Rate.Remaining == 0 {
		resetTime := resp.Rate.Reset.Time
		waitDuration := time.Until(resetTime)
		if waitDuration > 0 {
			fmt.Fprintf(os.Stderr, "Rate limit reached. Waiting %d seconds...\n", int(waitDuration.Seconds()))
			time.Sleep(waitDuration)
		}
	}
}

// Fetches all public repositories for an organization or user.
// Handles pagination automatically and filters for public repositories only.
// Returns rate limit errors after waiting and retrying once.
func getRepositories(ctx context.Context, accountType string, accountName string) ([]Repository, error) {
	client := getGitHubClient(ctx)
	var allRepos []Repository

	listOpts := github.ListOptions{PerPage: 100}

	for {
		var repos []*github.Repository
		var resp *github.Response
		var err error

		if accountType == "org" {
			orgOpt := &github.RepositoryListByOrgOptions{
				ListOptions: listOpts,
			}
			repos, resp, err = client.Repositories.ListByOrg(ctx, accountName, orgOpt)
		} else {
			userOpt := &github.RepositoryListByUserOptions{
				ListOptions: listOpts,
			}
			repos, resp, err = client.Repositories.ListByUser(ctx, accountName, userOpt)
		}

		if err != nil {
			handleRateLimit(ctx, client, resp)
			if resp != nil && resp.Rate.Remaining == 0 {
				continue
			}
			return nil, err
		}

		for _, repo := range repos {
			if repo.GetPrivate() {
				continue
			}

			allRepos = append(allRepos, Repository{
				Name:     repo.GetName(),
				URL:      repo.GetHTMLURL(),
				HasWiki:  repo.GetHasWiki(),
				IsPublic: !repo.GetPrivate(),
			})
		}

		handleRateLimit(ctx, client, resp)

		if resp.NextPage == 0 {
			break
		}
		listOpts.Page = resp.NextPage
	}

	return allRepos, nil
}

// Determines if a repository's wiki is publicly writable.
// A wiki is considered writable if:
// 1. It has no first page (shows "Create the first page" message), or
// 2. Accessing a non-existent page returns 200 OK instead of redirecting to login
func checkWiki(repo Repository) {
	if !repo.HasWiki {
		return
	}

	if _, err := url.Parse(repo.URL); err != nil {
		log.Printf("Invalid repository URL %s: %v\n", repo.URL, err)
		return
	}

	wikiURL := repo.URL + "/wiki"

	resp, err := httpClient.Get(wikiURL)
	if err != nil {
		log.Printf("Error accessing wiki for %s: %v\n", repo.Name, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	limitedReader := io.LimitReader(resp.Body, maxResponseSize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		log.Printf("Error reading wiki response for %s: %v\n", repo.Name, err)
		return
	}
	bodyStr := string(body)

	if strings.Contains(bodyStr, wikiFirstPageMarker) {
		fmt.Printf("Vulnerable [firstpage]: %s - %s\n", repo.Name, wikiURL)
		return
	}

	testURL := wikiURL + wikiTestPagePath
	testResp, err := httpClient.Get(testURL)
	if err != nil {
		log.Printf("Error testing wiki writeability for %s: %v\n", repo.Name, err)
		return
	}
	defer testResp.Body.Close()

	if testResp.StatusCode == http.StatusOK {
		fmt.Printf("Vulnerable [writeable]: %s - %s\n", repo.Name, testURL)
	}
}

// Scans all public repositories for a GitHub account and checks wikis.
// Supports both organization and user accounts with auto-detection.
func scanAccount(ctx context.Context, accountInput string) {
	if accountInput == "" {
		log.Println("Account name cannot be empty")
		return
	}

	accountType, accountName := parseAccountInput(accountInput)

	if accountType == "unknown" {
		client := getGitHubClient(ctx)
		detectedType, err := getAccountType(ctx, client, accountName)
		if err != nil {
			log.Printf("Error detecting account type: %v\n", err)
			return
		}
		accountType = detectedType
	}

	repos, err := getRepositories(ctx, accountType, accountName)
	if err != nil {
		log.Printf("Error fetching repositories: %v\n", err)
		return
	}

	for _, repo := range repos {
		checkWiki(repo)
	}
}

func main() {
	ctx := context.Background()

	if len(os.Args) > 1 {
		accountInput := os.Args[1]
		scanAccount(ctx, accountInput)
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			accountInput := strings.TrimSpace(scanner.Text())
			scanAccount(ctx, accountInput)
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Error reading from stdin: %v\n", err)
		}
	}
}
