package cli

import (
	"encoding/json"
	"fmt"
	"mrunc/internal/config"
	"mrunc/internal/container"
	"mrunc/internal/runtime"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func runCommand(ctx *cli.Context) error {
	var configPath string

	if ctx.NArg() < 1 {
		// No config specified → use default path
		baseDir := os.Getenv("MRUNC_BASE")
		if baseDir == "" {
			baseDir = config.BaseImageDir
		}
		configPath = filepath.Join(baseDir, "ubuntu", "ubuntu.json")
		fmt.Printf("No config specified, using default: %s\n", configPath)
	} else {
		// Use provided config
		configPath = ctx.Args().Get(0)
	}

	config, err := container.LoadConfig(configPath)
	if err != nil {
		return err
	}

	childPipe, parentPipe, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %v", err)
	}

	configData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.ExtraFiles = []*os.File{childPipe}

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

	cmd.Env = append(os.Environ(), "_MRUNC_PIPE_FD=3")
	cmd.SysProcAttr = runtime.CreateNamespaces()

	if err := cmd.Start(); err != nil {
		parentPipe.Close()
		childPipe.Close()
		return err
	}

	childPipe.Close()
	_, err = parentPipe.Write(configData)
	if err != nil {
		parentPipe.Close()
		return fmt.Errorf("failed to send config: %v", err)
	}
	parentPipe.Close()

	if err := cmd.Wait(); err != nil {
		fmt.Printf("PARENT: Child exited with error: %v\n", err)
	} else {
		fmt.Println("PARENT: Child completed successfully")
	}

	return nil
}
