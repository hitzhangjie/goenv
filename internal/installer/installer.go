package installer

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hitzhangjie/goenv/internal/config"
	"github.com/hitzhangjie/goenv/internal/system"
)

// Install installs a Go version
func Install(version string) error {
	version = strings.TrimSpace(version)
	if !strings.HasPrefix(version, "go") {
		version = "go" + version
	}

	// Get system info
	goos, err := system.GetGOOS()
	if err != nil {
		return fmt.Errorf("failed to get GOOS: %w", err)
	}

	goarch, err := system.GetGOARCH()
	if err != nil {
		return fmt.Errorf("failed to get GOARCH: %w", err)
	}

	// Construct download URL
	url := system.GetDownloadURL(version, goos, goarch)
	fmt.Printf("Installing %s for %s/%s...\n", version, goos, goarch)
	fmt.Printf("Download URL: %s\n", url)

	// Ensure directories exist
	downloadsDir, err := config.GetDownloadsDir()
	if err != nil {
		return err
	}
	if err := config.EnsureDir(downloadsDir); err != nil {
		return fmt.Errorf("failed to create downloads directory: %w", err)
	}

	sdkDir, err := config.GetSDKDir()
	if err != nil {
		return err
	}
	if err := config.EnsureDir(sdkDir); err != nil {
		return fmt.Errorf("failed to create SDK directory: %w", err)
	}

	binDir, err := config.GetBinDir()
	if err != nil {
		return err
	}
	if err := config.EnsureDir(binDir); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Download (check if already exists first)
	tarballPath := filepath.Join(downloadsDir, filepath.Base(url))
	if _, err := os.Stat(tarballPath); err == nil {
		fmt.Printf("Found existing download: %s, skipping download.\n", tarballPath)
	} else {
		if err := downloadFile(url, tarballPath); err != nil {
			return fmt.Errorf("failed to download: %w", err)
		}
	}

	// Extract
	installDir := filepath.Join(sdkDir, version)
	if err := extractTarball(tarballPath, installDir); err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	// Create wrapper scripts
	if err := createGoScript(version, installDir, binDir); err != nil {
		return fmt.Errorf("failed to create go script: %w", err)
	}

	if err := createGofmtScript(version, installDir, binDir); err != nil {
		return fmt.Errorf("failed to create gofmt script: %w", err)
	}

	fmt.Printf("Successfully installed %s\n", version)
	fmt.Printf("Use '%s' to run this version of Go\n", version)

	return nil
}

func downloadFile(url, dest string) error {
	fmt.Printf("Downloading %s...\n", url)

	client := &http.Client{
		Timeout: 10 * time.Minute,
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status code %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// Show progress
	total := resp.ContentLength
	var downloaded int64
	buf := make([]byte, 32*1024)

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			written, writeErr := out.Write(buf[:n])
			if writeErr != nil {
				return writeErr
			}
			downloaded += int64(written)
			if total > 0 {
				percent := float64(downloaded) / float64(total) * 100
				fmt.Printf("\rProgress: %.1f%%", percent)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	fmt.Println() // New line after progress

	return nil
}

func extractTarball(tarballPath, destDir string) error {
	fmt.Printf("Extracting to %s...\n", destDir)

	file, err := os.Open(tarballPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	// Detect and strip the top-level directory prefix (usually "go/")
	var stripPrefix string
	firstEntry := true

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Detect the top-level directory on first entry
		if firstEntry {
			firstEntry = false
			// Extract the top-level directory name
			parts := strings.Split(strings.TrimPrefix(header.Name, "./"), "/")
			if len(parts) > 0 && parts[0] != "" {
				stripPrefix = parts[0] + "/"
			}
		}

		// Skip if not a regular file or directory
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeDir {
			continue
		}

		// Strip the top-level directory prefix
		name := header.Name
		if stripPrefix != "" && strings.HasPrefix(name, stripPrefix) {
			name = strings.TrimPrefix(name, stripPrefix)
		}

		// Skip if name is empty after stripping (this is the top-level directory itself)
		if name == "" || name == "./" {
			continue
		}

		target := filepath.Join(destDir, name)

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
			continue
		}

		// Create file
		outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
		if err != nil {
			return err
		}

		if _, err := io.Copy(outFile, tr); err != nil {
			outFile.Close()
			return err
		}
		outFile.Close()
	}

	return nil
}

func createGoScript(version, installDir, binDir string) error {
	scriptPath := filepath.Join(binDir, version)

	root, err := config.GetGoenvRoot()
	if err != nil {
		return err
	}

	goroot := installDir
	gopath := filepath.Join(root, version)
	gobin := filepath.Join(binDir, version)
	goBin := filepath.Join(installDir, "bin", "go")

	script := fmt.Sprintf(`#!/bin/bash
GOROOT="%s"
GOPATH="%s"
GOBIN="%s"
exec "%s" "$@"
`, goroot, gopath, gobin, goBin)

	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return err
	}

	return nil
}

func createGofmtScript(version, installDir, binDir string) error {
	suffix := strings.TrimPrefix(version, "go")
	scriptPath := filepath.Join(binDir, "gofmt"+suffix)
	gofmtBin := filepath.Join(installDir, "bin", "gofmt")

	// Check if gofmt script already exists
	if _, err := os.Stat(scriptPath); err == nil {
		// Script exists, check if it points to the same version
		data, err := os.ReadFile(scriptPath)
		if err == nil && strings.Contains(string(data), gofmtBin) {
			// Already points to this version, skip
			return nil
		}
	}

	script := fmt.Sprintf(`#!/bin/bash
exec "%s" "$@"
`, gofmtBin)

	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return err
	}

	return nil
}
