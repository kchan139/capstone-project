package cli

import (
    "fmt"
    "os"
    "syscall"
    "my-capstone-project/internal/runtime"
    "my-capstone-project/internal/utils"

)

func childCommand() error {
    fmt.Printf("Running %v as %d\n", os.Args[2:], os.Getpid())
    container_id,errstr := utils.RandomHexString(16)
    if errstr != nil {
        return fmt.Errorf("failed to generate random hex strings for container ID: %v", errstr)
    }
    
    // Set hostname
    if err := syscall.Sethostname([]byte(container_id)); err != nil {
        return fmt.Errorf("failed to set hostname: %v", err)
    }
    
    // Setup overlay filesystem
    //TODO: need to get rid of hardcode image
    fmt.Printf("This is the container ID of this containter:%v\n",container_id)
    if err := runtime.SetupOverlayFS(container_id,"/home/phiung/container_image/fake_ubuntu"); err != nil {
        return fmt.Errorf("failed to setup overlay: %v", err)
    }
    merge_path:= fmt.Sprintf("/tmp/container-overlay/%s/merged",container_id)
    merge_putold_path:= fmt.Sprintf("/tmp/container-overlay/%s/merged/put_old",container_id)
    os.MkdirAll(merge_putold_path, 0755) 
    
    // Pivot root
    if err := runtime.PivotRoot(merge_path, merge_putold_path); err != nil {
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
    err := syscall.Exec(os.Args[2], os.Args[2:], os.Environ())
    if err != nil {
        return fmt.Errorf("failed to execute command: %v", err)
    }
    return nil
}
