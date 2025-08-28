package runtime

import (
"syscall"
"fmt"
)

func CreateNamespaces() *syscall.SysProcAttr {
	// TODO: need to resolve hardcode
    return &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
        Unshareflags: syscall.CLONE_NEWNS,
    }
}

// CreateUserNamespacePhase2 creates user namespace AFTER filesystem setup
func CreateUserNamespacePhase2() error {
    // Now we can safely create user namespace after pivot_root is done
    if err := syscall.Unshare(syscall.CLONE_NEWUSER); err != nil {
        return fmt.Errorf("failed to create user namespace: %v", err)
    }
    
    // Write UID/GID mappings (this must be done from outside the namespace)
    // This is more complex - typically done by the parent process
    // For now, we'll skip the mappings and assume we're running as root
    
    return nil
}