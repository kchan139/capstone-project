package cli

import (
	"encoding/json"

	"fmt"
	"mrunc/internal/config"
	"mrunc/internal/container"
	"mrunc/internal/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/sys/unix"
)

func createCommand(ctx *cli.Context) error {
	var configPath string

	if ctx.NArg() < 2 {
		// No config specified → use default path
		baseDir := os.Getenv("MRUNC_BASE")
		if baseDir == "" {
			baseDir = config.BaseImageDir
		}
		configPath = filepath.Join(baseDir, "ubuntu", "ubuntu.json")
		fmt.Printf("No config specified, using default: %s\n", configPath)
	} else {
		// Use provided config
		configPath = ctx.Args().Get(1)
	}
	containerId := ctx.Args().Get(0)
	config, err := container.LoadConfig(configPath)
	config.ContainerId = containerId
	if err != nil {
		return err
	}
	// create 2 unix socket, one for intermediate and one for init process
	parentSock, childSock, err := utils.SocketPair()
	if err != nil {
		return err
	}
	parentSock2, childSock2, err := utils.SocketPair()
	if err != nil {
		return err
	}
	fmt.Print("ignore")
	fmt.Println(parentSock2)
	defer parentSock.Close()
	defer childSock.Close()

	// create the exec.fifo files
	fifo_fd, err := createExecFifo(config.ContainerId)
	if err != nil {
		return err
	}
	cmd := exec.Command("/proc/self/exe", append([]string{"intermediate"}, os.Args[2:]...)...)
	// Mark all fds >=3 as CLOEXEC in one go.
	// Kernel will skip stdio and anything already CLOEXEC.
	_ = unix.CloseRange(3, ^uint(0), unix.CLOSE_RANGE_CLOEXEC)
	cmd.ExtraFiles = []*os.File{childSock, childSock2, fifo_fd}

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
	fifo_fd.Close()
	childSock.Close()
	childSock2.Close()

	if err := json.NewEncoder(parentSock).Encode(config); err != nil {
		return fmt.Errorf("send config: %w", err)
	}
	// receive the PID of init process sent from intermediate process
	buf := make([]byte, 32)
	n, err := parentSock.Read(buf)
	if err != nil {
		return fmt.Errorf("read pid: %w", err)
	}

	InitPidStr := string(buf[:n])

	pid, _ := strconv.Atoi(strings.TrimSpace(InitPidStr))
	fmt.Println("Init PID from intermediate:", pid)

	// send config to init proc
	if err := json.NewEncoder(parentSock2).Encode(config); err != nil {
		return fmt.Errorf("send config: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		fmt.Printf("PARENT: Intermediate exited with error: %v\n", err)
	} else {
		fmt.Println("PARENT: Intermediate completed successfully")
	}

	return nil
}

func createExecFifo(containerId string) (*os.File, error) {
	dirPath := "/run/mrunc/" + containerId
	fifoPath := dirPath + "/exec.fifo"

	// Step 1: ensure directory exists
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, err
	}

	// Step 2: create FIFO
	if err := unix.Mkfifo(fifoPath, 0666); err != nil && !os.IsExist(err) {
		return nil, err
	}

	// Step 3: open it (both ends, so it doesn't block yet)
	fifoFile, err := os.OpenFile(fifoPath, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		return nil, err
	}

	return fifoFile, nil
}
