package cli

import (
    "fmt"
    "os"
    "os/exec"
	"my-capstone-project/internal/runtime"
	"my-capstone-project/internal/container"
    "encoding/json"
    "golang.org/x/sys/unix"
)

func runCommand() error {
    if len(os.Args) < 3 {
        return fmt.Errorf("usage: mrunc run <config.json>")
    }
	configPath := os.Args[2]
    config, err := container.LoadConfig(configPath)
    if err != nil {
        return err
    }
    childPipe,parentPipe ,err := os.Pipe()
    if err != nil {
        return fmt.Errorf("failed to create pipe: %v", err)
    }
    configData, err := json.Marshal(config)
    if err != nil {
        return err
    }
    // socket pair use for sync parent and child when creating user namespace
    fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_STREAM, 0)
	if err != nil {
		panic(err)
	}
    parentSock := os.NewFile(uintptr(fds[0]), "parent")
	childSock := os.NewFile(uintptr(fds[1]), "child")

    cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
    cmd.ExtraFiles = []*os.File{childPipe, childSock}
    if config.Process.Terminal {
        // Interactive mode - connect to terminal
        fmt.Printf("Starting container in interactive mode\n")
        cmd.Stdin = os.Stdin
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
    } else {
        // Non-interactive mode - disconnect from terminal
        fmt.Printf("Starting container in non-interactive mode\n")
        cmd.Stdin = nil  // No stdin
        cmd.Stdout = os.Stdout  // Still show output (or redirect to logs)
        cmd.Stderr = os.Stderr  // Still show errors (or redirect to logs)
    }
    cmd.Env = append(os.Environ(), "_MRUNC_PIPE_FD=3", "_MRUNC_SYNC_FD=4")
    cmd.SysProcAttr = runtime.CreateNamespaces()
    
    if err := cmd.Start(); err != nil {
        // Clean up on error
        parentPipe.Close()
        childPipe.Close()
        return err
    }
    childPipe.Close()
    childSock.Close()
    _, err = parentPipe.Write(configData)

    childPID := cmd.Process.Pid
    if err != nil {
        parentPipe.Close()
        return fmt.Errorf("failed to send config: %v", err)
    }
    parentPipe.Close()
    // ✅ Parent waits for signal from child
	buf := make([]byte, 1)
	_, err = parentSock.Read(buf) // blocks until child writes
	if err != nil {
		panic(err)
	}
	fmt.Println("Parent got signal from child, continuing...")

    
    if err := runtime.SetupUserNamespaceFromParent(childPID); err != nil {
        // Clean up on error
        return err
    }
    _, err = parentSock.Write([]byte{1})
	if err != nil {
		panic(err)
	}
	fmt.Println("Parent: signaled back to child!")
    return cmd.Wait()
}