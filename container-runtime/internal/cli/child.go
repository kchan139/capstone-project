package cli

import (
	"fmt"
	"my-capstone-project/internal/container"
	"my-capstone-project/internal/runtime"
	"my-capstone-project/internal/utils"
	"os"
	"syscall"
)

func childCommand() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("child: missing config file")
	}
	configPath := os.Args[2]

	config, err := container.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("child: failed to load config: %v", err)
	}

	container_id, errstr := utils.RandomHexString(16)
	if errstr != nil {
		return fmt.Errorf("failed to generate random hex strings for container ID: %v", errstr)
	}

	// Set hostname
	if err := syscall.Sethostname([]byte(container_id)); err != nil {
		return fmt.Errorf("failed to set hostname: %v", err)
	}

	// Setup overlay filesystem
	//TODO: need to get rid of hardcode image
	fmt.Printf("This is the container ID of this containter:%v\n", container_id)
	if err := runtime.SetupOverlayFS(container_id, config.RootFS.Path); err != nil {
		return fmt.Errorf("failed to setup overlay: %v", err)
	}
	merge_path := fmt.Sprintf("/tmp/container-overlay/%s/merged", container_id)
	merge_putold_path := fmt.Sprintf("/tmp/container-overlay/%s/merged/put_old", container_id)
	os.MkdirAll(merge_putold_path, 0755)

	// Pivot root
	if err := runtime.PivotRoot(merge_path, merge_putold_path); err != nil {
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

	// Execute the command
	// Execute the process (replace current process)
	command := config.Process.Args[0]
	args := config.Process.Args[1:]
	env := os.Environ()

	err = syscall.Exec(command, args, env)
	if err != nil {
		return fmt.Errorf("failed to execute command: %v", err)
	}

	return nil
}
