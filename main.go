package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// Repository represents a Github repository
type Repository struct {
	Name    string `json:"name"`
	URL     string `json:"html_url"`
	HasWiki bool   `json:"has_wiki"`
}

// Gets an HTTP client	that doesn't follow redirects
func getClient() *http.Client {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return client
}

// Checks if a repository has a wiki and if it's writable
func checkWiki(repo Repository) {
	if repo.HasWiki {
		url := repo.URL + "/wiki"

		client := getClient()
		resp, err := client.Get(url)
		if err != nil {
			log.Println(err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			// Check if wiki is writable but doesn't have a first page yet
			fmt.Printf("Readable: %s, URL: %s\n", repo.Name, url)

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println("Error reading response body:", err)
				return
			}
			bodyStr := string(body)

			// Check if wiki is writable but doesn't have a first page yet
			if strings.Contains(bodyStr, "Create the first page") {
				fmt.Printf("Writable-Firstpage: %s, URL: %s\n", repo.Name, url)
			} else {

				// If we can go to github.com/<repo>/wiki/<somenewpage> that means we can edit an existing wiki
				url = url + "/notrealpage"

				client := getClient()
				resp, err := client.Get(url)
				if err != nil {
					fmt.Println(err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					fmt.Printf("Writable: %s, URL: %s\n", repo.Name, url)
				}
			}
		}
	}
}

// Gets all repositories for a given organization
func getRepositories(orgName string) ([]Repository, error) {
	// We use /users/ and not /orgs/ because not all Github repositories belong to orgs, but all orgs are users apparently
	url := fmt.Sprintf("https://api.github.com/users/%s/repos", orgName)

	client := getClient()

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch repositories: %s", resp.Status)
	}

	var repos []Repository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	return repos, nil
}

// Scans an organization for repositories with wikis
func scanOrg(orgName string) {
	if orgName == "" {
		fmt.Println("Organization name cannot be empty")
		return
	}
	repos, err := getRepositories(orgName)
	if err != nil {
		log.Fatalln("Error:", err)
	}

	for _, repo := range repos {
		checkWiki(repo)
	}
}

// Main function
func main() {

	if len(os.Args) > 1 {
		orgName := os.Args[1]
		scanOrg(orgName)
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			orgName := strings.TrimSpace(scanner.Text())
			scanOrg(orgName)
		}
		if err := scanner.Err(); err != nil {
			log.Fatalln("Error reading from stdin:", err)
		}
	}
}
