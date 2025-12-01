package cli

import (
	"fmt"
	"mrunc/internal/runtime"
	"mrunc/internal/utils"
	"os"
	"syscall"

	"github.com/urfave/cli/v2"
	"golang.org/x/sys/unix"
)

func initprocCommand(ctx *cli.Context) error {
	fmt.Println("Init command run")
	fmt.Printf("Init, my pid are %d\n", os.Getpid())

	parent := os.NewFile(uintptr(3), "parent-sock")
	config, err := receiveConfigFrom(parent)
	if err != nil {
		return fmt.Errorf("child: failed to receive config: %v", err)
	}

	root_fs := config.RootFS.Path
	root_fs_putold := config.RootFS.Path + "/put_old"
	os.MkdirAll(root_fs_putold, 0755)

	if err := unix.Mount(root_fs, root_fs, "", unix.MS_BIND, ""); err != nil {
		panic(fmt.Errorf("bind mount failed: %w", err))
	}

	// Pivot root
	if err := runtime.PivotRoot(root_fs, root_fs_putold); err != nil {
		return fmt.Errorf("failed to pivot root: %v", err)
	}

	workDir := config.Process.Cwd
	if workDir == "" {
		workDir = "/"
	}

	// Change to root directory
	if err := syscall.Chdir(workDir); err != nil {
		return fmt.Errorf("failed to chdir: %v", err)
	}

	// Cleanup old root
	syscall.Unmount("/put_old", syscall.MNT_DETACH)
	os.RemoveAll("/put_old")

	// Mount proc
	if err := syscall.Mount("proc", "proc", "proc", 0, ""); err != nil {
		return fmt.Errorf("failed to mount proc: %v", err)
	}

	// Set environment variables
	if len(config.Process.Env) > 0 {
		os.Clearenv()
		for _, env := range config.Process.Env {
			if err := os.Setenv(utils.ParseEnvKey(env), utils.ParseEnvValue(env)); err != nil {
				return fmt.Errorf("failed to set environment: %v", err)
			}
		}
	}

	if err := runtime.SetProcessUser(config.Process.User); err != nil {
		return fmt.Errorf("failed to set process user: %v", err)
	}
	command := config.Process.Args[0]
	args := config.Process.Args
	env := os.Environ()

	// reading from the fifo - waiting for the start signal
	fd := uintptr(4)
	fifo_fd := os.NewFile(fd, "inherited-fifo")
	sync_buf := make([]byte, 100)
	fmt.Println("before calling fifo read")
	n, _ := fifo_fd.Read(sync_buf)
	fmt.Printf("after calling fifo read %d\n", n)

	return runtime.ExecuteCommand(command, args, env)
}
