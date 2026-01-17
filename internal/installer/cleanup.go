package installer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hitzhangjie/goenv/internal/config"
)

// CleanupDownloads removes all files from the downloads directory
func CleanupDownloads() error {
	downloadsDir, err := config.GetDownloadsDir()
	if err != nil {
		return err
	}

	// Check if directory exists
	info, err := os.Stat(downloadsDir)
	if os.IsNotExist(err) {
		fmt.Println("Downloads directory does not exist, nothing to clean.")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to check downloads directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("downloads path exists but is not a directory: %s", downloadsDir)
	}

	// Read directory contents
	entries, err := os.ReadDir(downloadsDir)
	if err != nil {
		return fmt.Errorf("failed to read downloads directory: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("Downloads directory is already empty.")
		return nil
	}

	// Remove all files
	removedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories
		}
		filePath := filepath.Join(downloadsDir, entry.Name())
		if err := os.Remove(filePath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove %s: %v\n", filePath, err)
			continue
		}
		removedCount++
		fmt.Printf("Removed: %s\n", entry.Name())
	}

	fmt.Printf("Cleaned up %d file(s) from downloads directory.\n", removedCount)
	return nil
}
