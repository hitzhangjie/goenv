package cmd

import (
	"github.com/hitzhangjie/goenv/internal/installer"
	"github.com/spf13/cobra"
)

var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Regenerate wrapper scripts for all installed Go versions",
	Long:  "Regenerate wrapper scripts for all installed Go versions to apply the latest environment variable settings",
	RunE:  runFix,
}

func runFix(cmd *cobra.Command, args []string) error {
	return installer.FixScripts()
}
