package cli

import (
	"fmt"
	"runtime/debug"
)

// These variables are set at build time via ldflags
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// GetVersion returns the full version string
func GetVersion() string {
	if Version == "dev" {
		// Try to get version from build info (works with go install)
		if info, ok := debug.ReadBuildInfo(); ok {
			if info.Main.Version != "" && info.Main.Version != "(devel)" {
				return info.Main.Version
			}
		}
	}
	return Version
}

// GetVersionInfo returns detailed version information
func GetVersionInfo() string {
	version := GetVersion()
	return fmt.Sprintf("mrunc version %s\nCommit: %s\nBuilt: %s",
		version, GitCommit, BuildDate)
}
