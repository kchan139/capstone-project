package runtime

import (
	"syscall"
	"testing"
)

func TestCreateNamespaces(t *testing.T) {
	t.Run("returns valid SysProcAttr", func(t *testing.T) {
		attr := CreateNamespaces()

		if attr == nil {
			t.Fatal("CreateNamespaces returned nil")
		}

		// Verify expected clone flags are set
		expectedFlags := uintptr(syscall.CLONE_NEWUTS | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWPID | syscall.CLONE_NEWNET)

		if attr.Cloneflags != expectedFlags {
			t.Errorf("Cloneflags = %v, want %v", attr.Cloneflags, expectedFlags)
		}

		if attr.Unshareflags != uintptr(syscall.CLONE_NEWNS) {
			t.Errorf("Unshareflags = %v, want %v", attr.Unshareflags, syscall.CLONE_NEWNS)
		}
	})
}
