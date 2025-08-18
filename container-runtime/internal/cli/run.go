package cli

import (
    "fmt"
    "os"
    "os/exec"
	"my-capstone-project/internal/runtime"
)

func runCommand() error {
    fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())

    cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.SysProcAttr = runtime.CreateNamespaces()
    
    err := cmd.Run()
    if err != nil {
        return fmt.Errorf("failed to run container: %v", err)
    }
    return nil
}