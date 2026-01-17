package system

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// GetGOOS returns the current GOOS value
func GetGOOS() (string, error) {
	// Try to get from go env first
	if goos, err := getFromGoEnv("GOOS"); err == nil && goos != "" {
		return goos, nil
	}
	// Fallback to runtime
	return runtime.GOOS, nil
}

// GetGOARCH returns the current GOARCH value
func GetGOARCH() (string, error) {
	// Try to get from go env first
	if goarch, err := getFromGoEnv("GOARCH"); err == nil && goarch != "" {
		return goarch, nil
	}
	// Fallback to runtime
	return runtime.GOARCH, nil
}

func getFromGoEnv(key string) (string, error) {
	cmd := exec.Command("go", "env", key)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetDownloadURL constructs the download URL for a Go version
func GetDownloadURL(version, goos, goarch string) string {
	// Remove "go" prefix if present
	ver := version
	if strings.HasPrefix(ver, "go") {
		ver = ver[2:]
	}
	return fmt.Sprintf("https://go.dev/dl/%s.%s-%s.tar.gz", version, goos, goarch)
}
