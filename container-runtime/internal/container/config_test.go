package container

import (
	"os"
	"path/filepath"
	"testing"

	"mrunc/pkg/specs"
)

func TestLoadConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		// Create temp config file
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.json")

		configJSON := `{
			"root": {
				"path": "/tmp/rootfs",
				"readonly": false
			},
			"process": {
				"args": ["/bin/sh"],
				"env": ["PATH=/usr/bin"],
				"cwd": "/",
				"terminal": false,
				"user": {"uid": 0, "gid": 0}
			},
			"hostname": "test"
		}`

		if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}

		config, err := LoadConfig(configPath)
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		if config.RootFS.Path == "" {
			t.Error("Expected rootfs path to be set")
		}
		if len(config.Process.Args) == 0 {
			t.Error("Expected process args to be set")
		}
	})

	t.Run("missing config file", func(t *testing.T) {
		_, err := LoadConfig("/nonexistent/config.json")
		if err == nil {
			t.Error("Expected error for missing config file")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "bad.json")

		if err := os.WriteFile(configPath, []byte("not json"), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		_, err := LoadConfig(configPath)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}

func TestValidateConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := &specs.ContainerConfig{
			RootFS: specs.RootfsConfig{
				Path: "/tmp/rootfs",
			},
			Process: specs.ProcessConfig{
				Args: []string{"/bin/sh"},
				User: &specs.User{UID: 0, GID: 0},
			},
		}

		if err := validateConfig(config); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("missing rootfs path", func(t *testing.T) {
		config := &specs.ContainerConfig{
			RootFS: specs.RootfsConfig{
				Path: "",
			},
			Process: specs.ProcessConfig{
				Args: []string{"/bin/sh"},
				User: &specs.User{UID: 0, GID: 0},
			},
		}

		err := validateConfig(config)
		if err == nil {
			t.Error("Expected error for missing rootfs path")
		}
	})

	t.Run("missing process args", func(t *testing.T) {
		config := &specs.ContainerConfig{
			RootFS: specs.RootfsConfig{
				Path: "/tmp/rootfs",
			},
			Process: specs.ProcessConfig{
				Args: []string{},
				User: &specs.User{UID: 0, GID: 0},
			},
		}

		err := validateConfig(config)
		if err == nil {
			t.Error("Expected error for missing process args")
		}
	})
}
