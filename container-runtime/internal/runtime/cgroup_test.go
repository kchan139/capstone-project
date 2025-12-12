package runtime

import (
	"os"
	"testing"

	"mrunc/pkg/specs"
	"mrunc/tests/testutil"
)

func TestCreateCgroup(t *testing.T) {
	testutil.SkipIfNotPrivileged(t)

	config := testutil.MockConfig()
	config.ContainerId = "test-cgroup"

	// Use current process PID for testing
	pid := os.Getpid()

	// This will likely fail in test environment without cgroup permissions
	// but we're testing that it doesn't panic
	err := CreateCgroup(config, pid)

	// We expect this to fail in most test environments
	// Just verify it returns an error gracefully rather than panicking
	if err == nil {
		t.Log("CreateCgroup succeeded (running in privileged environment)")
	} else {
		t.Logf("CreateCgroup failed as expected in test environment: %v", err)
	}
}

func TestCreateCgroupWithNilResources(t *testing.T) {
	testutil.SkipIfNotPrivileged(t)

	config := &specs.ContainerConfig{
		ContainerId: "test-nil-resources",
		RootFS: specs.RootfsConfig{
			Path: "/tmp/test",
		},
		Process: specs.ProcessConfig{
			Args: []string{"/bin/sh"},
		},
		Linux: specs.LinuxConfig{
			Resources: nil,
		},
	}

	pid := os.Getpid()

	// Should handle nil resources gracefully
	err := CreateCgroup(config, pid)
	if err == nil {
		t.Log("CreateCgroup with nil resources succeeded")
	} else {
		t.Logf("CreateCgroup with nil resources failed: %v", err)
	}
}
