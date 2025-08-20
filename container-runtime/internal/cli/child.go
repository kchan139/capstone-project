package cli

import (
    "fmt"
    "os"
    "os/exec"
    "syscall"
    "my-capstone-project/internal/runtime"
)

func childCommand() error {
    fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())
    
    // Set hostname
    if err := syscall.Sethostname([]byte("container")); err != nil {
        return fmt.Errorf("failed to set hostname: %v", err)
    }
    
    // Setup overlay filesystem
    //TODO: need to get rid of hardcode image
    if err := runtime.SetupOverlayFS("/home/phiung/container_image/fake_ubuntu","/tmp/container-overlay/upper","/tmp/container-overlay/work"); err != nil {
        return fmt.Errorf("failed to setup overlay: %v", err)
    }
    
    // Pivot root
    if err := runtime.PivotRoot("/tmp/container-overlay/merged", "/tmp/container-overlay/merged/put_old"); err != nil {
        return fmt.Errorf("failed to pivot root: %v", err)
    }
    
    // Change to root directory
    if err := syscall.Chdir("/"); err != nil {
        return fmt.Errorf("failed to chdir: %v", err)
    }
    
    // Cleanup old root
    syscall.Unmount("/put_old", syscall.MNT_DETACH)
    os.RemoveAll("/put_old")
    
    // Mount proc
    if err := syscall.Mount("proc", "proc", "proc", 0, ""); err != nil {
        return fmt.Errorf("failed to mount proc: %v", err)
    }
    
    // Execute the command
    cmd := exec.Command(os.Args[2], os.Args[3:]...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    
    err := cmd.Run()
    if err != nil {
        return fmt.Errorf("failed to execute command: %v", err)
    }
    return nil
}
