// package main
// import (
//     "os"
//     "os/exec"
// 	"fmt"
// 	"syscall"
// 	"unsafe"
// 	"golang.org/x/sys/unix"
// )

// func main() {
// 	switch os.Args[1] {
// 	case "run": run()
// 	case "child": child()
// 	default: print("bad command")
// 	}
// }

// func pivotRoot(newRoot, putOld string) error {
//     // Convert to C strings (null-terminated)
//     newRootPtr, err := unix.BytePtrFromString(newRoot)
//     if err != nil {
//         return err
//     }
//     putOldPtr, err := unix.BytePtrFromString(putOld)
//     if err != nil {
//         return err
//     }

//     // Syscall: long pivot_root(const char *new_root, const char *put_old);
//     _, _, errno := unix.Syscall(
//         unix.SYS_PIVOT_ROOT,
//         uintptr(unsafe.Pointer(newRootPtr)),
//         uintptr(unsafe.Pointer(putOldPtr)),
//         0,
//     )
//     if errno != 0 {
//         return errno
//     }
//     return nil
// }


// func run() {
// 	fmt.Printf("Running %v as %d\n",os.Args[2:],os.Getpid())

// 	cmd := exec.Command("/proc/self/exe",append([]string{"child"},os.Args[2:]...)...)
// 	cmd.Stdin = os.Stdin
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr
// 	cmd.SysProcAttr = &syscall.SysProcAttr{
// 		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS ,
// 		Unshareflags: syscall.CLONE_NEWNS, // does that mean other ns can share ? like the pid, uts, ?
// 	}
// 	err := cmd.Run()
//     if err != nil {
//         fmt.Printf("Error: %v\n", err)
//     }

// }

// func child() {
// 	fmt.Printf("Running %v as %d\n",os.Args[2:],os.Getpid())
// 	syscall.Sethostname([]byte("container"))
// 	// create overlay fs (isolation on write)
// 	os.MkdirAll("/tmp/container-overlay/upper", 0755)
// 	os.MkdirAll("/tmp/container-overlay/work", 0755)
// 	os.MkdirAll("/tmp/container-overlay/merged", 0755)

// 	overlayOptions := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", 
//     "/home/phiung/container_image/fake_ubuntu",  // read-only base
//     "/tmp/container-overlay/upper",               // writable layer
//     "/tmp/container-overlay/work")  

// 	err := syscall.Mount("overlay", "/tmp/container-overlay/merged", "overlay", 0, overlayOptions)
// 	if err != nil {
// 		fmt.Printf("Overlay mount failed: %v\n", err)
// 	}

// 	// Now pivot root to the overlay
// 	syscall.Mount("/tmp/container-overlay/merged", "/tmp/container-overlay/merged", "", syscall.MS_BIND, "")
// 	os.MkdirAll("/tmp/container-overlay/merged/put_old", 0755)
// 	err = pivotRoot("/tmp/container-overlay/merged", "/tmp/container-overlay/merged/put_old")
// 	syscall.Chdir("/")

// 	syscall.Unmount("/put_old", syscall.MNT_DETACH)
// 	os.RemoveAll("/put_old")

// 	syscall.Mount("proc", "proc", "proc", 0, "")
	
// 	cmd := exec.Command(os.Args[2],os.Args[3:]...)
// 	cmd.Stdin = os.Stdin
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr

// 	err = cmd.Run()
//     if err != nil {
//         fmt.Printf("Error: %v\n", err)
//     }

// }