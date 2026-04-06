package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// PrepareExec resolves the command path BEFORE seccomp is applied
// This avoids needing stat/access syscalls after seccomp
func PrepareExec(command string, args []string, env []string) (string, []string, error) {
	if filepath.IsAbs(command) {
		return command, args, nil
	}

	resolvedPath, err := resolvePath(command, env)
	if err != nil {
		return "", nil, fmt.Errorf("resolve exec path for %q: %w", command, err)
	}

	resolvedArgs := append([]string(nil), args...)
	if len(resolvedArgs) == 0 {
		resolvedArgs = []string{resolvedPath}
	} else {
		resolvedArgs[0] = resolvedPath
	}

	return resolvedPath, resolvedArgs, nil
}

// ExecuteCommand performs the actual exec - call AFTER seccomp is applied
// This only needs the execve syscall
func ExecuteCommand(execPath string, args []string, env []string) error {
	return syscall.Exec(execPath, args, env)
}

func resolvePath(command string, env []string) (string, error) {
	// Look for PATH in environment
	var pathVar string
	for _, envVar := range env {
		if strings.HasPrefix(envVar, "PATH=") {
			pathVar = strings.TrimPrefix(envVar, "PATH=")
			break
		}
	}

	if pathVar == "" {
		return "", fmt.Errorf("PATH not found in environment")
	}

	// Check each directory in PATH
	for _, dir := range strings.Split(pathVar, ":") {
		if dir == "" {
			continue
		}
		fullPath := filepath.Join(dir, command)
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			// Check if executable
			if info.Mode()&0111 != 0 {
				return fullPath, nil
			}
		}
	}

	return "", fmt.Errorf("command not found in PATH: %s", command)
}
