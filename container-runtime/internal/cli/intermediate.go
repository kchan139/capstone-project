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

	// unshare the PID namespace

	// if err := unix.Unshare(unix.CLONE_NEWPID); err != nil {
	// 	panic(fmt.Errorf("unshare: %w", err))
	// }

	// passing the socket to the init process
	passedSocket := os.NewFile(uintptr(4), "init-sock")
	defer passedSocket.Close()

	cmd := exec.Command("/proc/self/exe", append([]string{"initproc"}, os.Args[2:]...)...)
	cmd.ExtraFiles = []*os.File{passedSocket}
	cmd.SysProcAttr = runtime.CreateNamespaces()
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

	if err := cmd.Wait(); err != nil {
		fmt.Printf("Intermediate: Init exited with error: %v\n", err)
	} else {
		fmt.Println("Intermediate: Init completed successfully")
	}

	return nil
}


// func receiveConfigFromSocket() (*mySpecs.ContainerConfig, error) {
// 	// hard-code the file descriptor
// 	f := os.NewFile(uintptr(3), "parent-sock")

// 	buf := make([]byte, 4096)
// 	n, err := f.Read(buf)
// 	if err != nil {
//     	return nil, fmt.Errorf("read error: %v", err)
// 	}
// 	 f.Close()


// 	var config mySpecs.ContainerConfig
// 	if err := json.Unmarshal(buf[:n], &config); err != nil {
// 		return nil, fmt.Errorf("failed to parse config JSON: %v", err)
// 	}
// 	return &config, nil
// }

func receiveConfigFrom(r io.Reader) (*mySpecs.ContainerConfig, error) {
    // Use a decoder (stops after one JSON doc; no EOF needed)
    dec := json.NewDecoder(r)
    var cfg mySpecs.ContainerConfig
    if err := dec.Decode(&cfg); err != nil {
        return nil, fmt.Errorf("decode config: %w", err)
    }
    return &cfg, nil
}