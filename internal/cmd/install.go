package cmd

import (
	"fmt"

	"github.com/hitzhangjie/goenv/internal/cache"
	"github.com/hitzhangjie/goenv/internal/installer"
	"github.com/hitzhangjie/goenv/internal/version"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <version>",
	Short: "Install a Go version",
	Long:  "Download and install a specific Go version",
	Args:  cobra.ExactArgs(1),
	RunE:  runInstall,
}

func runInstall(cmd *cobra.Command, args []string) error {
	versionStr := version.NormalizeVersion(args[0])

	// Verify version exists in cache
	cachedData, err := cache.LoadVersions()
	if err == nil && cachedData != nil {
		found := false
		for _, group := range cachedData.Groups {
			for _, v := range group.Versions {
				if v.Tag == versionStr {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			fmt.Printf("Warning: Version %s not found in cached versions list.\n", versionStr)
			fmt.Println("You may want to run 'goenv versions --update' first to refresh the list.")
		}
	}

	// Install
	if err := installer.Install(versionStr); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	return nil
}
