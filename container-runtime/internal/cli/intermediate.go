package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"mrunc/internal/runtime"
	mySpecs "mrunc/pkg/specs"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
	"golang.org/x/sys/unix"
)
func intermediateCommand(ctx *cli.Context) error {
parent := os.NewFile(uintptr(3), "parent-sock")
    defer parent.Close()
	config, err := receiveConfigFrom(parent)
	if err != nil {
		return fmt.Errorf("child: failed to receive config: %v", err)
	}
	// setup cgroup
	runtime.CreateCgroup(config, os.Getpid())

	fmt.Println("Running inside limited cgroup for 10 seconds...")


	// passing the socket to the init process
	passedSocket := os.NewFile(uintptr(4), "init-sock")
	defer passedSocket.Close()
	// passsing the exec.fifo files
	fifo_fd := os.NewFile(uintptr(5), "fifo-file")
	defer fifo_fd.Close()

	cmd := exec.Command("/proc/self/exe", append([]string{"initproc"}, os.Args[2:]...)...)
	// Mark all fds >=3 as CLOEXEC in one go.
	// Kernel will skip stdio and anything already CLOEXEC.
	_ = unix.CloseRange(3, ^uint(0), unix.CLOSE_RANGE_CLOEXEC)
	cmd.ExtraFiles = []*os.File{passedSocket, fifo_fd}
	cmd.SysProcAttr = runtime.CreateNamespaces(config)
	if config.Process.Terminal {
			fmt.Printf("Starting container in interactive mode\n")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
	} else {
		fmt.Printf("Starting container in non-interactive mode\n")
		cmd.Stdin = nil
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	// writing the init PID to the parent process
	if _, err := parent.Write([]byte(fmt.Sprintf("%d", cmd.Process.Pid))); err != nil {
        return fmt.Errorf("write pid: %w", err)
    }

	// if err := cmd.Wait(); err != nil {
	// 	fmt.Printf("Intermediate: Init exited with error: %v\n", err)
	// } else {
	// 	fmt.Println("Intermediate: Init completed successfully")
	// }



	return nil
}



func receiveConfigFrom(r io.Reader) (*mySpecs.ContainerConfig, error) {
    // Use a decoder (stops after one JSON doc; no EOF needed)
    dec := json.NewDecoder(r)
    var cfg mySpecs.ContainerConfig
    if err := dec.Decode(&cfg); err != nil {
        return nil, fmt.Errorf("decode config: %w", err)
    }
    return &cfg, nil
}