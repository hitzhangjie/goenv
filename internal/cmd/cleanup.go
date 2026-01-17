package cmd

import (
	"github.com/hitzhangjie/goenv/internal/installer"
	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up downloaded Go version archives",
	Long:  "Remove all downloaded Go version archives from the downloads directory",
	RunE:  runCleanup,
}

func runCleanup(cmd *cobra.Command, args []string) error {
	if err := installer.CleanupDownloads(); err != nil {
		return err
	}
	return nil
}
