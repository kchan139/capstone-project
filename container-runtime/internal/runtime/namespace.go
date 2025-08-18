package runtime

import "syscall"

func CreateNamespaces() *syscall.SysProcAttr {
	// TODO: need to resolve hardcode
    return &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
        Unshareflags: syscall.CLONE_NEWNS,
    }
}