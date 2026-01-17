package github

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	gh "github.com/google/go-github/v81/github"
	"github.com/hitzhangjie/goenv/internal/version"
	"golang.org/x/oauth2"
)

const (
	OverallTimeout = 5 * time.Minute // 整体操作超时
)

// FetchOptions contains options for fetching tags
type FetchOptions struct {
	MinVersion  string // Minimum version (e.g., "go1.22"), default is "go1.10"
	MinYear     int    // Minimum year (e.g., 2020)
	AllVersions bool   // Fetch all versions (default: false, respects MinVersion/MinYear)
}

const (
	DefaultMinVersion = "go1.10" // Default minimum version
)

// Option is a function that modifies FetchOptions
type Option func(*FetchOptions)

// WithMinVersion sets the minimum version to fetch
func WithMinVersion(v string) Option {
	return func(opts *FetchOptions) {
		opts.MinVersion = v
	}
}

// WithMinYear sets the minimum year to fetch versions from
func WithMinYear(year int) Option {
	return func(opts *FetchOptions) {
		opts.MinYear = year
	}
}

// WithAllVersions fetches all versions regardless of filters
func WithAllVersions() Option {
	return func(opts *FetchOptions) {
		opts.AllVersions = true
	}
}

// createGitHubClient creates a GitHub client with authentication if token is available
func createGitHubClient(ctx context.Context) *gh.Client {
	var httpClient *http.Client

	// Check for GitHub token
	if token := os.Getenv("GITHUB_ACCESS_TOKEN"); token != "" {
		token = strings.TrimSpace(token)
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		httpClient = oauth2.NewClient(ctx, ts)
	}

	return gh.NewClient(httpClient)
}

// FetchTags fetches tags from golang/go repository with options and early stop support
// existingTags is a map of tag names that already exist locally, used for early stopping
// Returns fetched tags and any error encountered. Even if an error occurs, all successfully fetched tags are returned.
func FetchTags(existingTags map[string]bool, options ...Option) ([]string, error) {
	// Apply options
	opts := &FetchOptions{}
	for _, opt := range options {
		opt(opts)
	}

	// Set default min version if not specified and not fetching all versions
	if opts.MinVersion == "" && !opts.AllVersions {
		opts.MinVersion = DefaultMinVersion
	}

	// Parse min version if provided
	var minVersion *version.Version
	var parseErr error
	if opts.MinVersion != "" && !opts.AllVersions {
		minVersion, parseErr = version.ParseVersion(version.NormalizeVersion(opts.MinVersion))
		if parseErr != nil {
			// Return error immediately if min version is invalid
			return nil, fmt.Errorf("invalid min version %s: %w", opts.MinVersion, parseErr)
		}
	}

	// Create context with overall timeout
	ctx, cancel := context.WithTimeout(context.Background(), OverallTimeout)
	defer cancel()

	// Create GitHub client
	client := createGitHubClient(ctx)

	var allTags []string
	page := 1
	perPage := 100
	allExistingFound := false
	var fetchErr error // Store error but continue processing

	// Build filter description
	filterDesc := "all versions"
	if opts.AllVersions {
		filterDesc = "all versions (no filter)"
	} else if minVersion != nil {
		filterDesc = fmt.Sprintf("versions >= %s", opts.MinVersion)
	} else if opts.MinYear > 0 {
		filterDesc = fmt.Sprintf("versions from year >= %d", opts.MinYear)
	}

	fmt.Printf("Starting to fetch tags from GitHub (filter: %s, timeout: %v overall)...\n",
		filterDesc, OverallTimeout)

	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			fetchErr = fmt.Errorf("fetch tags timeout after %v: %w", OverallTimeout, ctx.Err())
			fmt.Printf("Warning: %v. Returning %d tags fetched so far.\n", fetchErr, len(allTags))
			return allTags, fetchErr
		default:
		}

		fmt.Printf("Fetching page %d... ", page)
		startTime := time.Now()

		// Use go-github library to fetch tags
		tags, resp, err := client.Repositories.ListTags(ctx, "golang", "go", &gh.ListOptions{
			Page:    page,
			PerPage: perPage,
		})

		requestDuration := time.Since(startTime)

		if err != nil {
			// Store error but continue with what we have
			if _, ok := err.(*gh.RateLimitError); ok {
				fetchErr = fmt.Errorf("GitHub API rate limit exceeded. Please wait or use GITHUB_ACCESS_TOKEN for higher limits")
			} else if _, ok := err.(*gh.AbuseRateLimitError); ok {
				fetchErr = fmt.Errorf("GitHub API abuse rate limit exceeded. Please wait before retrying")
			} else {
				fetchErr = fmt.Errorf("failed to fetch tags at page %d: %w", page, err)
			}
			fmt.Printf("Error: %v. Returning %d tags fetched so far.\n", fetchErr, len(allTags))
			return allTags, fetchErr
		}

		// Check response status
		if resp != nil && resp.StatusCode != http.StatusOK {
			fetchErr = fmt.Errorf("GitHub API returned status %d at page %d", resp.StatusCode, page)
			fmt.Printf("Error: %v. Returning %d tags fetched so far.\n", fetchErr, len(allTags))
			return allTags, fetchErr
		}

		if len(tags) == 0 {
			fmt.Printf("No more tags found.\n")
			break
		}

		// Process tags with filtering and early stop check
		newTagsCount := 0
		existingCount := 0
		skippedCount := 0

		for _, tag := range tags {
			// Get tag name directly from the tag object
			tagName := tag.GetName()

			// Check if tag already exists locally (early stop mechanism)
			if existingTags != nil && existingTags[tagName] {
				existingCount++
				// If all tags in this page exist, we can stop fetching
				if existingCount == len(tags) {
					allExistingFound = true
					fmt.Printf("All tags in this page already exist locally, stopping fetch.\n")
					break
				}
				continue
			}

			// Parse version for filtering
			v, err := version.ParseVersion(tagName)
			if err != nil {
				// Skip invalid version tags
				skippedCount++
				continue
			}

			// Apply filters
			shouldInclude := true
			if !opts.AllVersions {
				// Filter by min version
				if minVersion != nil {
					if v.Compare(minVersion) < 0 {
						shouldInclude = false
					}
				}
				// Filter by min year (use heuristic based on version numbers)
				if opts.MinYear > 0 && shouldInclude {
					// Rough heuristic: Go 1.0 was released in 2012, so we can estimate
					// This is approximate, but better than nothing
					estimatedYear := 2012 + (v.Major-1)*1 + (v.Minor/6)*1
					if estimatedYear < opts.MinYear {
						shouldInclude = false
					}
				}
			}

			if shouldInclude {
				allTags = append(allTags, tagName)
				newTagsCount++
			} else {
				skippedCount++
			}
		}

		// Log the number of tags fetched in this request
		fmt.Printf("Fetched %d tags (new: %d, existing: %d, skipped: %d, took %v, total new: %d)\n",
			len(tags), newTagsCount, existingCount, skippedCount, requestDuration, len(allTags))

		// Early stop if all tags in this page exist locally
		if allExistingFound {
			fmt.Printf("Stopped early: all remaining tags already exist locally.\n")
			break
		}

		// Check if there are more pages
		if resp == nil || resp.NextPage == 0 {
			fmt.Printf("Reached last page.\n")
			break
		}

		page = resp.NextPage
	}

	fmt.Printf("Successfully fetched %d new tags in total.\n", len(allTags))
	return allTags, nil
}
