package runtime

import (
	"encoding/json"
	"syscall"
	"testing"

	"mrunc/pkg/specs"
)

func TestCreateNamespaces(t *testing.T) {
	t.Run("returns empty SysProcAttr for nil config", func(t *testing.T) {
		attr := CreateNamespaces(nil)

		if attr == nil {
			t.Fatal("CreateNamespaces returned nil")
		}

		if attr.Cloneflags != 0 {
			t.Errorf("Cloneflags = %v, want 0", attr.Cloneflags)
		}

		if attr.Unshareflags != 0 {
			t.Errorf("Unshareflags = %v, want 0", attr.Unshareflags)
		}
	})

	t.Run("sets clone and unshare flags from configured namespaces", func(t *testing.T) {
		var config specs.ContainerConfig
		configJSON := []byte(`{
			"linux": {
				"namespaces": [
					{"type": "uts"},
					{"type": "mount"},
					{"type": "pid"},
					{"type": "network"},
					{"type": "ipc"},
					{"type": "cgroup"},
					{"type": "ignored"}
				]
			}
		}`)

		if err := json.Unmarshal(configJSON, &config); err != nil {
			t.Fatalf("failed to build namespace test config: %v", err)
		}

		attr := CreateNamespaces(&config)
		if attr == nil {
			t.Fatal("CreateNamespaces returned nil")
		}

		expectedCloneFlags := uintptr(
			syscall.CLONE_NEWUTS |
				syscall.CLONE_NEWNS |
				syscall.CLONE_NEWPID |
				syscall.CLONE_NEWNET |
				syscall.CLONE_NEWIPC |
				syscall.CLONE_NEWCGROUP,
		)

		if attr.Cloneflags != expectedCloneFlags {
			t.Errorf("Cloneflags = %v, want %v", attr.Cloneflags, expectedCloneFlags)
		}

		expectedUnshareFlags := uintptr(syscall.CLONE_NEWNS)
		if attr.Unshareflags != expectedUnshareFlags {
			t.Errorf("Unshareflags = %v, want %v", attr.Unshareflags, expectedUnshareFlags)
		}
	})
}
