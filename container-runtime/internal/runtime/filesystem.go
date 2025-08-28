package runtime

import (
    "golang.org/x/sys/unix"
    "unsafe"
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
