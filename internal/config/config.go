package config

import (
	"os"
	"path/filepath"
)

const (
	GoenvDir     = ".goenv"
	VersionsFile = "versions.json"
	DownloadsDir = "downloads"
	SDKDir       = "sdk"
	BinDir       = "bin"
)

// GetGoenvRoot returns the root directory for goenv (~/.goenv)
func GetGoenvRoot() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, GoenvDir), nil
}

// GetVersionsFile returns the path to versions.json
func GetVersionsFile() (string, error) {
	root, err := GetGoenvRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, VersionsFile), nil
}

// GetDownloadsDir returns the downloads directory path
func GetDownloadsDir() (string, error) {
	root, err := GetGoenvRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, DownloadsDir), nil
}

// GetSDKDir returns the SDK directory path
func GetSDKDir() (string, error) {
	root, err := GetGoenvRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, SDKDir), nil
}

// GetBinDir returns the bin directory path
func GetBinDir() (string, error) {
	root, err := GetGoenvRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, BinDir), nil
}

// EnsureDir ensures that a directory exists, creating it if necessary
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}
