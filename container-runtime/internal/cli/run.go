package cli

import (
    "fmt"
    "os"
    "os/exec"
	"my-capstone-project/internal/runtime"
	"my-capstone-project/internal/container"

)

func runCommand() error {
    if len(os.Args) < 3 {
        return fmt.Errorf("usage: mrunc run <config.json>")
    }
	configPath := os.Args[2]
    config, err := container.LoadConfig(configPath)
    if err != nil {
        return err
    }
    cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
    if config.Process.Terminal {
        // Interactive mode - connect to terminal
        fmt.Printf("Starting container in interactive mode\n")
        cmd.Stdin = os.Stdin
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
    } else {
        // Non-interactive mode - disconnect from terminal
        fmt.Printf("Starting container in non-interactive mode\n")
        cmd.Stdin = nil  // No stdin
        cmd.Stdout = os.Stdout  // Still show output (or redirect to logs)
        cmd.Stderr = os.Stderr  // Still show errors (or redirect to logs)
    }

    cmd.SysProcAttr = runtime.CreateNamespaces()
    
    err = cmd.Run()
    if err != nil {
        return fmt.Errorf("failed to run container: %v", err)
    }
    return nil
}