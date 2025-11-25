package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"mrunc/internal/runtime"
	"mrunc/internal/utils"
	"mrunc/pkg/specs"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"
	"golang.org/x/sys/unix"
)

func childCommand(ctx *cli.Context) error {
	config, err := receiveConfigFromPipe()
	if err != nil {
		return fmt.Errorf("child: failed to receive config: %v", err)
	}

	if !config.Process.Terminal {
		fmt.Printf("Non-interactive mode: detaching from terminal\n")
		if _, err := syscall.Setsid(); err != nil && err != syscall.EPERM {
			fmt.Printf("Warning: setsid failed: %v\n", err)
		}
	} else {
		if _, err := unix.Setsid(); err != nil {
			return fmt.Errorf("setsid: %w", err)
		}

		// receive slave from parent
		slave := os.NewFile(uintptr(4), "pty-slave")
		if slave == nil {
			return fmt.Errorf("no slave fd")
		}
		defer slave.Close()
		// dup slave → 0,1,2
		for _, fd := range []int{0, 1, 2} {
			if err := unix.Dup2(int(slave.Fd()), fd); err != nil {
				return fmt.Errorf("dup2 %d: %w", fd, err)
			}
		}
		// now fd 0 is the pty slave, make it controlling tty
		if err := unix.IoctlSetInt(0, unix.TIOCSCTTY, 0); err != nil {
			return fmt.Errorf("TIOCSCTTY: %w", err)
		}
		// fmt.Println("hahahah") haha cai dit con me may
	}

	// Set hostname
	if err := syscall.Sethostname([]byte(config.Hostname)); err != nil {
		return fmt.Errorf("failed to set hostname: %v", err)
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

	// Setup network namespace
	if err := runtime.SetupLoopback(); err != nil {
		fmt.Printf("Warning: Failed to setup loopback: %v\n", err)
	}

	// Setup veth network if enabled
	if config.Linux.Network != nil && config.Linux.Network.EnableNetwork {
		netCfg := config.Linux.Network
		if err := runtime.SetupContainerNetwork(
			netCfg.VethContainer,
			netCfg.ContainerIP,
			netCfg.GatewayCIDR,
			netCfg.DNS,
		); err != nil {
			fmt.Printf("Warning: Failed to setup container network: %v\n", err)
		}
	}

	// Optional: Verify network setup
	runtime.VerifyNetwork()

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

	// apply capabilities
	if err := runtime.SetupCaps(config); err != nil {
		return fmt.Errorf("failed to set the capabilities : %v", err)
	}

	time.Sleep(120 *time.Second)
	// Execute the process (replace current process)
	command := config.Process.Args[0]
	args := config.Process.Args
	env := os.Environ()

	return runtime.ExecuteCommand(command, args, env)
}

// receiveConfigFromPipe reads configuration from pipe passed by parent
func receiveConfigFromPipe() (*specs.ContainerConfig, error) {
	// Get pipe FD from environment variable
	pipeFdStr := os.Getenv("_MRUNC_PIPE_FD")
	if pipeFdStr == "" {
		return nil, fmt.Errorf("_MRUNC_PIPE_FD environment variable not set")
	}

	pipeFd, err := strconv.Atoi(pipeFdStr)
	if err != nil {
		return nil, fmt.Errorf("invalid pipe FD: %v", err)
	}

	// Create file from FD
	pipe := os.NewFile(uintptr(pipeFd), "config-pipe")
	defer pipe.Close()

	// Read all data from pipe
	configData, err := io.ReadAll(pipe)
	if err != nil {
		return nil, fmt.Errorf("failed to read config data: %v", err)
	}

	// Deserialize config
	var config specs.ContainerConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %v", err)
	}

	return &config, nil
}
