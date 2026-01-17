package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/hitzhangjie/goenv/internal/config"
	"github.com/hitzhangjie/goenv/internal/version"
)

// LoadVersions loads cached versions from disk
func LoadVersions() (*version.VersionsData, error) {
	filePath, err := config.GetVersionsFile()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // File doesn't exist, return nil
		}
		return nil, fmt.Errorf("failed to read versions file: %w", err)
	}

	var versionsData version.VersionsData
	if err := json.Unmarshal(data, &versionsData); err != nil {
		return nil, fmt.Errorf("failed to parse versions file: %w", err)
	}

	return &versionsData, nil
}

// SaveVersions saves versions data to disk
func SaveVersions(data *version.VersionsData) error {
	filePath, err := config.GetVersionsFile()
	if err != nil {
		return err
	}

	// Ensure directory exists
	root, err := config.GetGoenvRoot()
	if err != nil {
		return err
	}
	if err := config.EnsureDir(root); err != nil {
		return fmt.Errorf("failed to create goenv directory: %w", err)
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal versions data: %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write versions file: %w", err)
	}

	return nil
}

// ShouldUpdate checks if cached data is old enough to warrant an update
func ShouldUpdate(data *version.VersionsData, maxAge time.Duration) bool {
	if data == nil {
		return true
	}
	return time.Since(data.FetchedAt) > maxAge
}
