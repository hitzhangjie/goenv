package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hitzhangjie/goenv/internal/cache"
	"github.com/hitzhangjie/goenv/internal/github"
	"github.com/hitzhangjie/goenv/internal/version"
	"github.com/spf13/cobra"
)

var versionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "List all available Go versions",
	Long:  "Fetch and display all available Go versions from GitHub, grouped by major.minor version",
	RunE:  runVersions,
}

func init() {
	versionsCmd.Flags().Bool("update", false, "Force update from GitHub")
	versionsCmd.Flags().String("min-version", "", "Minimum version to fetch (e.g., go1.22)")
	versionsCmd.Flags().Int("min-year", 0, "Minimum year to fetch versions from (e.g., 2020)")
	versionsCmd.Flags().Bool("all", false, "Fetch all versions (ignore filters)")
}

func runVersions(cmd *cobra.Command, args []string) error {
	update, _ := cmd.Flags().GetBool("update")

	// Try to load cached versions
	cachedData, err := cache.LoadVersions()
	if err != nil {
		return fmt.Errorf("failed to load cached versions: %w", err)
	}

	// Check if we should update
	shouldUpdate := update || cachedData == nil
	if !shouldUpdate {
		shouldUpdate = askForUpdate()
	}

	var versionsData *version.VersionsData

	if shouldUpdate {
		// Build existing tags map for early stop
		existingTags := make(map[string]bool)
		if cachedData != nil {
			for _, group := range cachedData.Groups {
				for _, v := range group.Versions {
					existingTags[v.Tag] = true
				}
			}
		}

		// Build fetch options
		var fetchOptions []github.Option
		allVersions, _ := cmd.Flags().GetBool("all")
		if allVersions {
			fetchOptions = append(fetchOptions, github.WithAllVersions())
		} else {
			minVersion, _ := cmd.Flags().GetString("min-version")
			if minVersion != "" {
				fetchOptions = append(fetchOptions, github.WithMinVersion(minVersion))
			}
			minYear, _ := cmd.Flags().GetInt("min-year")
			if minYear > 0 {
				fetchOptions = append(fetchOptions, github.WithMinYear(minYear))
			}
		}

		fmt.Println("Fetching versions from GitHub...")
		tags, err := github.FetchTags(existingTags, fetchOptions...)
		if err != nil {
			// Log error but continue with tags that were successfully fetched
			fmt.Fprintf(os.Stderr, "Warning: Error occurred while fetching tags: %v\n", err)
			fmt.Fprintf(os.Stderr, "Continuing with %d tags that were successfully fetched.\n", len(tags))
		}

		// Parse new tags
		var newVersions []*version.Version
		for _, tag := range tags {
			v, err := version.ParseVersion(tag)
			if err != nil {
				// Skip invalid tags
				continue
			}
			v.FetchedAt = time.Now()
			newVersions = append(newVersions, v)
		}

		// Merge with existing versions
		var allVersionsList []*version.Version
		if cachedData != nil {
			for _, group := range cachedData.Groups {
				for _, v := range group.Versions {
					allVersionsList = append(allVersionsList, v)
				}
			}
		}
		allVersionsList = append(allVersionsList, newVersions...)

		groups := version.GroupVersions(allVersionsList)
		versionsData = &version.VersionsData{
			FetchedAt: time.Now(),
			Groups:    groups,
		}

		// Save to cache
		if err := cache.SaveVersions(versionsData); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save versions cache: %v\n", err)
		}
	} else {
		versionsData = cachedData
	}

	// Display versions
	displayVersions(versionsData)

	return nil
}

func askForUpdate() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Found cached versions. Check for updates? (y/N): ")
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func displayVersions(data *version.VersionsData) {
	if data == nil || len(data.Groups) == 0 {
		fmt.Println("No versions found.")
		return
	}

	fmt.Printf("\nAvailable Go versions (fetched at %s):\n\n", data.FetchedAt.Format("2006-01-02 15:04:05"))

	for _, group := range data.Groups {
		fmt.Printf("Go %s:\n", group.MajorMinor)
		for _, v := range group.Versions {
			rcStr := ""
			if v.IsRC {
				rcStr = fmt.Sprintf(" (RC%d)", v.RC)
			}
			fmt.Printf("  - %s%s\n", v.Tag, rcStr)
		}
		fmt.Println()
	}
}
