package main
import (
    "os"
    "os/exec"
	"fmt"
	"syscall"
)

func main() {
	switch os.Args[1] {
	case "run": run()
	case "child": child()
	default: print("bad command")
	}
}


func run() {
	fmt.Printf("Running %v as %d\n",os.Args[2:],os.Getpid())

	cmd := exec.Command("/proc/self/exe",append([]string{"child"},os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS ,
		Unshareflags: syscall.CLONE_NEWNS, // does that mean other ns can share ? like the pid, uts, ?
	}
	err := cmd.Run()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    }

}

func child() {
	fmt.Printf("Running %v as %d\n",os.Args[2:],os.Getpid())
	syscall.Sethostname([]byte("container"))
	syscall.Chroot("/home/phiung/container_image/fake_ubuntu") // need to generalize this
	syscall.Chdir("/")
	syscall.Mount("proc","proc","proc",0,"")

	cmd := exec.Command(os.Args[2],os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    }

}