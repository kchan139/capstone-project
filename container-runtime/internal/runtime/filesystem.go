package runtime

import (
    "golang.org/x/sys/unix"
    "unsafe"
    "os"
    "fmt"
    "syscall"
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

func SetupOverlayFS(lowerDir, upperDir, workDir string) error { 
    os.MkdirAll("/tmp/container-overlay/upper", 0755)
	os.MkdirAll("/tmp/container-overlay/work", 0755)
	os.MkdirAll("/tmp/container-overlay/merged", 0755)
	//TODO: need to resolve hardcode
	overlayOptions := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", 
    lowerDir,  // read-only base
    upperDir,               // writable layer
    workDir)  

	err := syscall.Mount("overlay", "/tmp/container-overlay/merged", "overlay", 0, overlayOptions)
    os.MkdirAll("/tmp/container-overlay/merged/put_old", 0755)
	if err != nil {
		fmt.Printf("Overlay mount failed: %v\n", err)
        return err
	}
    return nil
}