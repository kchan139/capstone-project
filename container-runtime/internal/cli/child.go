package cli

import (
	"fmt"
	"my-capstone-project/internal/runtime"
	"my-capstone-project/internal/utils"
	"my-capstone-project/pkg/specs"
	"strconv"
	"os"
	"io"
	"encoding/json"
	"syscall"
	"golang.org/x/sys/unix"
)

func childCommand() error {
	
	config, err := receiveConfigFromPipe()
	if err != nil {
        return fmt.Errorf("child: failed to receive config: %v", err)
    }
	// container_id, errstr := utils.RandomHexString(16)
	// if errstr != nil {
	// 	return fmt.Errorf("failed to generate random hex strings for container ID: %v", errstr)
	// }

	if !config.Process.Terminal {
        fmt.Printf("Non-interactive mode: detaching from terminal\n")
        if _, err := syscall.Setsid(); err != nil && err != syscall.EPERM {
            fmt.Printf("Warning: setsid failed: %v\n", err)
        }
    } else {
        fmt.Printf("Interactive mode: keeping terminal connection\n")
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

	// Set environment variables
	if len(config.Process.Env) > 0 {
		os.Clearenv()
		for _, env := range config.Process.Env {
			// env is in format "KEY=VALUE"
			if err := os.Setenv(utils.ParseEnvKey(env), utils.ParseEnvValue(env)); err != nil {
				return fmt.Errorf("failed to set environment: %v", err)
			}
		}
	}

	// if err := runtime.CreateUserNamespacePhase2(); err != nil {
    //     fmt.Printf("Warning: failed to create user namespace: %v\n", err)
    //     // Continue without user namespace for now
    // }

	if err := runtime.SetProcessUser(config.Process.User); err != nil {
        return fmt.Errorf("failed to set process user: %v", err)
    }

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