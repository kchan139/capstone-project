package runtime

import (
    "golang.org/x/sys/unix"
    "unsafe"
    "os"
    "fmt"
    "syscall"
    "path/filepath"
)

func PivotRoot(newRoot, putOld string) error {
    // Convert to C strings (null-terminated)
    newRootPtr, err := unix.BytePtrFromString(newRoot)
    if err != nil {
        return err
    }
    putOldPtr, err := unix.BytePtrFromString(putOld)	
    if err != nil {
        return err
    }

    // Syscall: long pivot_root(const char *new_root, const char *put_old);
    _, _, errno := unix.Syscall(
        unix.SYS_PIVOT_ROOT,
        uintptr(unsafe.Pointer(newRootPtr)),
        uintptr(unsafe.Pointer(putOldPtr)),
        0,
    )
    if errno != 0 {
        return errno
    }
    return nil
}
func SetupOverlayFS(containerID, lowerDir string) error {
    os.MkdirAll("/tmp/container-overlay", 0755)

    containerBase := fmt.Sprintf("/tmp/container-overlay/%s", containerID)
    upperDir := filepath.Join(containerBase, "upper")
    workDir  := filepath.Join(containerBase, "work")
    merged   := filepath.Join(containerBase, "merged")

    os.MkdirAll(upperDir, 0755)
    os.MkdirAll(workDir, 0755)
    os.MkdirAll(merged, 0755)

    overlayOptions := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
        lowerDir, upperDir, workDir)

    if err := syscall.Mount("overlay", merged, "overlay", 0, overlayOptions); err != nil {
        return fmt.Errorf("overlay mount failed: %v", err)
    }

    return nil
}