package runtime
import (
    "path/filepath"
	"syscall"
	"strings"
	"fmt"
	"os"

)

func ExecuteCommand (command string, args []string, env []string,) error {
	// Try direct execution first (like real runtimes)
    if err := tryDirectExec(command, args, env); err == nil {
        return nil
    }

	// Fall back to PATH resolution
    if resolvedPath, err := resolvePath(command, env); err == nil {
        args[0] = resolvedPath
        return syscall.Exec(resolvedPath, args, env)
    }

	// Last resort: shell (but log a warning)
    fmt.Printf("Warning: Using shell execution for '%s' - potential security risk\n", command)
    shellCmd := strings.Join(args, " ")
    return syscall.Exec("/bin/sh", []string{"/bin/sh", "-c", shellCmd}, env)
}

func tryDirectExec(command string, args []string, env []string) error {
    if filepath.IsAbs(command) {
        return syscall.Exec(command, args, env)
    }
    return fmt.Errorf("not absolute path")
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