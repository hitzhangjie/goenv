package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hitzhangjie/goenv/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed Go versions",
	Long:  "Display all Go versions that have been installed",
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	sdkDir, err := config.GetSDKDir()
	if err != nil {
		return err
	}

	// Check if SDK directory exists
	if _, err := os.Stat(sdkDir); os.IsNotExist(err) {
		fmt.Println("No Go versions installed.")
		return nil
	}

	// Read all directories in SDK
	entries, err := os.ReadDir(sdkDir)
	if err != nil {
		return fmt.Errorf("failed to read SDK directory: %w", err)
	}

	var installedVersions []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "go") {
			// Verify it's a valid installation by checking for bin/go
			goBin := filepath.Join(sdkDir, entry.Name(), "bin", "go")
			if _, err := os.Stat(goBin); err == nil {
				installedVersions = append(installedVersions, entry.Name())
			}
		}
	}

	if len(installedVersions) == 0 {
		fmt.Println("No Go versions installed.")
		return nil
	}

	fmt.Println("Installed Go versions:")
	for _, v := range installedVersions {
		fmt.Printf("  - %s\n", v)
	}

	return nil
}
