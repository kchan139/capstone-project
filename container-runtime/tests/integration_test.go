//go:build integration
// +build integration

package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"mrunc/internal/container"
	"mrunc/tests/testutil"
)

func TestBinaryBuild(t *testing.T) {
	// Test that the binary can be built
	cmd := exec.Command("make", "build")
	cmd.Dir = filepath.Join("..", "..", "container-runtime")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, output)
	}

	// Verify binary exists
	binaryPath := filepath.Join("..", "..", "container-runtime", "bin", "mrunc")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatal("Binary was not created after build")
	}
}

func TestBinaryVersion(t *testing.T) {
	binaryPath := filepath.Join("..", "..", "container-runtime", "bin", "mrunc")

	// Skip if binary doesn't exist
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skip("Binary not found, run 'make build' first")
	}

	tests := []struct {
		name string
		args []string
	}{
		{"version flag", []string{"--version"}},
		{"version short", []string{"-v"}},
		{"version command", []string{"version"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.CombinedOutput()

			if err != nil {
				// Some version commands might exit with 0, others might not
				// Just check that we get output
				if len(output) == 0 {
					t.Errorf("No output from version command: %v", err)
				}
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, "mrunc") && !strings.Contains(outputStr, "version") {
				t.Errorf("Version output doesn't contain expected text: %s", outputStr)
			}
		})
	}
}

func TestConfigParsing(t *testing.T) {
	// Test that we can parse the example configs
	configFiles := []string{
		filepath.Join("..", "..", "container-runtime", "configs", "examples", "ubuntu.json"),
		filepath.Join("..", "..", "container-runtime", "configs", "examples", "ci-test.json"),
	}

	for _, configFile := range configFiles {
		t.Run(filepath.Base(configFile), func(t *testing.T) {
			if _, err := os.Stat(configFile); os.IsNotExist(err) {
				t.Skipf("Config file not found: %s", configFile)
			}

			_, err := container.LoadConfig(configFile)
			if err != nil {
				t.Errorf("Failed to load config %s: %v", configFile, err)
			}
		})
	}
}

func TestMockConfigCreation(t *testing.T) {
	// Test that our test utilities work
	config := testutil.MockConfig()

	if config == nil {
		t.Fatal("MockConfig returned nil")
	}

	if config.ContainerId == "" {
		t.Error("MockConfig should set container ID")
	}

	if len(config.Process.Args) == 0 {
		t.Error("MockConfig should set process args")
	}
}

func TestMockRootFSCreation(t *testing.T) {
	rootfs := testutil.CreateMockRootFS(t)

	// Verify essential directories exist
	dirs := []string{"bin", "etc", "proc"}
	for _, dir := range dirs {
		path := filepath.Join(rootfs, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected directory %s not found", dir)
		}
	}

	// Verify os-release exists
	osRelease := filepath.Join(rootfs, "etc", "os-release")
	if _, err := os.Stat(osRelease); os.IsNotExist(err) {
		t.Error("Expected /etc/os-release not found")
	}
}

func TestWriteConfigFile(t *testing.T) {
	config := testutil.MockConfig()
	configPath := testutil.WriteConfigFile(t, config)

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not written")
	}

	// Verify we can load it back
	loadedConfig, err := container.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load written config: %v", err)
	}

	if loadedConfig.ContainerId != config.ContainerId {
		t.Errorf("Container ID mismatch: got %s, want %s",
			loadedConfig.ContainerId, config.ContainerId)
	}
}
