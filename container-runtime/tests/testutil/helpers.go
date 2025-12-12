package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"mrunc/pkg/specs"
)

// SkipIfNotRoot skips test if not running as root
func SkipIfNotRoot(t *testing.T) {
	t.Helper()
	if os.Geteuid() != 0 {
		t.Skip("Test requires root privileges")
	}
}

// SkipIfNotPrivileged skips test if CI or non-privileged environment
func SkipIfNotPrivileged(t *testing.T) {
	t.Helper()
	if os.Getenv("CI") != "" {
		t.Skip("Skipping privileged test in CI")
	}
	SkipIfNotRoot(t)
}

// TempDir creates a temporary directory for testing
func TempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "mrunc-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// MockConfig returns a minimal valid container config
func MockConfig() *specs.ContainerConfig {
	return &specs.ContainerConfig{
		ContainerId: "test-container",
		RootFS: specs.RootfsConfig{
			Path:     "/tmp/test-rootfs",
			Readonly: false,
		},
		Process: specs.ProcessConfig{
			Args: []string{"/bin/sh"},
			Env:  []string{"PATH=/usr/bin:/bin"},
			Cwd:  "/",
			User: &specs.User{
				UID: 0,
				GID: 0,
			},
			Terminal: false,
		},
		Hostname: "test-hostname",
		Linux: specs.LinuxConfig{
			Resources: &specs.LinuxResources{
				CPU: &specs.CPUConfig{
					Shares: 1024,
					Quota:  50000,
					Period: 100000,
				},
				Memory: &specs.MemoryConfig{
					Limit:       268435456,
					Reservation: 268435456,
					Swap:        268435456,
				},
				Pids: &specs.PidsConfig{
					Limit: 50,
				},
			},
		},
	}
}

// WriteConfigFile writes a config to a temporary JSON file
func WriteConfigFile(t *testing.T, config *specs.ContainerConfig) string {
	t.Helper()

	tmpDir := TempDir(t)
	configPath := filepath.Join(tmpDir, "config.json")

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	return configPath
}

// CreateMockRootFS creates a minimal rootfs structure for testing
func CreateMockRootFS(t *testing.T) string {
	t.Helper()

	rootfs := TempDir(t)

	// Create essential directories
	dirs := []string{"bin", "etc", "proc", "dev", "sys"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(rootfs, dir), 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}
	}

	// Create a minimal /etc/os-release
	osRelease := filepath.Join(rootfs, "etc", "os-release")
	content := `NAME="Mock Linux"
VERSION="1.0"
ID=mock
`
	if err := os.WriteFile(osRelease, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create os-release: %v", err)
	}

	return rootfs
}

// AssertNoError fails the test if err is not nil
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// AssertError fails the test if err is nil
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected error but got nil")
	}
}
