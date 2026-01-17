package version

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Version represents a Go version
type Version struct {
	Tag         string    `json:"tag"`
	Major       int       `json:"major"`
	Minor       int       `json:"minor"`
	Patch       int       `json:"patch"`
	RC          int       `json:"rc"` // 0 means not an RC version
	FullVersion string    `json:"full_version"`
	IsRC        bool      `json:"is_rc"`
	FetchedAt   time.Time `json:"fetched_at"`
}

// VersionGroup represents a group of versions under the same major.minor
type VersionGroup struct {
	MajorMinor string     `json:"major_minor"` // e.g., "1.22"
	Versions   []*Version `json:"versions"`
}

// VersionsData represents the cached versions data
type VersionsData struct {
	FetchedAt time.Time      `json:"fetched_at"`
	Groups    []VersionGroup `json:"groups"`
}

var versionRegex = regexp.MustCompile(`^go(\d+)\.(\d+)(?:\.(\d+))?(?:rc(\d+))?$`)

// ParseVersion parses a version tag string (e.g., "go1.22.1" or "go1.22rc1")
func ParseVersion(tag string) (*Version, error) {
	matches := versionRegex.FindStringSubmatch(tag)
	if matches == nil {
		return nil, fmt.Errorf("invalid version tag format: %s", tag)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])

	var patch, rc int
	var isRC bool

	if matches[3] != "" {
		patch, _ = strconv.Atoi(matches[3])
	}

	if matches[4] != "" {
		rc, _ = strconv.Atoi(matches[4])
		isRC = true
	}

	fullVersion := fmt.Sprintf("%d.%d", major, minor)
	if patch > 0 {
		fullVersion += fmt.Sprintf(".%d", patch)
	}
	if isRC {
		fullVersion += fmt.Sprintf("rc%d", rc)
	}

	return &Version{
		Tag:         tag,
		Major:       major,
		Minor:       minor,
		Patch:       patch,
		RC:          rc,
		FullVersion: fullVersion,
		IsRC:        isRC,
	}, nil
}

// GetMajorMinor returns the major.minor version string (e.g., "1.22")
func (v *Version) GetMajorMinor() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

// Compare compares two versions, returns -1 if v < other, 0 if equal, 1 if v > other
func (v *Version) Compare(other *Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}
	// RC versions are less than non-RC versions
	if v.IsRC && !other.IsRC {
		return -1
	}
	if !v.IsRC && other.IsRC {
		return 1
	}
	if v.IsRC && other.IsRC {
		if v.RC < other.RC {
			return -1
		}
		if v.RC > other.RC {
			return 1
		}
	}
	return 0
}

// GroupVersions groups versions by major.minor
func GroupVersions(versions []*Version) []VersionGroup {
	groups := make(map[string][]*Version)

	for _, v := range versions {
		key := v.GetMajorMinor()
		groups[key] = append(groups[key], v)
	}

	var result []VersionGroup
	for majorMinor, vers := range groups {
		// Sort versions within group
		for i := 0; i < len(vers)-1; i++ {
			for j := i + 1; j < len(vers); j++ {
				if vers[i].Compare(vers[j]) > 0 {
					vers[i], vers[j] = vers[j], vers[i]
				}
			}
		}
		result = append(result, VersionGroup{
			MajorMinor: majorMinor,
			Versions:   vers,
		})
	}

	// Sort groups by version
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			v1, _ := ParseVersion("go" + result[i].MajorMinor + ".0")
			v2, _ := ParseVersion("go" + result[j].MajorMinor + ".0")
			if v1.Compare(v2) > 0 {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// NormalizeVersion normalizes a version string to ensure it starts with "go"
func NormalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	if !strings.HasPrefix(version, "go") {
		version = "go" + version
	}
	return version
}
