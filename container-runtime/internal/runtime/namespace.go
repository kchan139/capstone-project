package runtime

import (
	"fmt"
	"os"
	"syscall"
)

func CreateNamespaces() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		// Add CLONE_NEWNET for network namespace
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
		Unshareflags: syscall.CLONE_NEWNS,
	}
}

// This must be called from the parent process after child completes filesystem setup
func SetupUserNamespaceFromParent(childPID int) error {
	fmt.Printf("DEBUG: Setting up user namespace for PID %d\n", childPID)

	// Get current user info
	currentUID := os.Getuid()
	currentGID := os.Getgid()

	fmt.Printf("DEBUG: Current parent UID: %d, GID: %d\n", currentUID, currentGID)

	// Write UID mapping: map container root (0) to current user
	uidMapPath := fmt.Sprintf("/proc/%d/uid_map", childPID)
	uidMapping := fmt.Sprintf("0 %d 1000", currentUID) // Map container UID 0 to host UID

	fmt.Printf("DEBUG: Writing UID mapping: %s to %s\n", uidMapping, uidMapPath)
	if err := os.WriteFile(uidMapPath, []byte(uidMapping), 0644); err != nil {
		return fmt.Errorf("failed to write uid_map: %v", err)
	}

	// Deny setgroups (required before writing gid_map)
	setgroupsPath := fmt.Sprintf("/proc/%d/setgroups", childPID)
	fmt.Printf("DEBUG: Denying setgroups: %s\n", setgroupsPath)
	if err := os.WriteFile(setgroupsPath, []byte("deny"), 0644); err != nil {
		return fmt.Errorf("failed to write setgroups: %v", err)
	}

	// Write GID mapping: map container root (0) to current group
	gidMapPath := fmt.Sprintf("/proc/%d/gid_map", childPID)
	gidMapping := fmt.Sprintf("0 %d 1000", currentGID) // Map container GID 0 to host GID

	fmt.Printf("DEBUG: Writing GID mapping: %s to %s\n", gidMapping, gidMapPath)
	if err := os.WriteFile(gidMapPath, []byte(gidMapping), 0644); err != nil {
		return fmt.Errorf("failed to write gid_map: %v", err)
	}

	fmt.Printf("DEBUG: User namespace mappings completed\n")
	return nil
}
