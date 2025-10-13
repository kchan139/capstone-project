package cli

import (
	"encoding/json"

	"fmt"
	"os/exec"
	"os"
	"github.com/urfave/cli/v2"
	"mrunc/internal/utils"
	"path/filepath"
	"mrunc/internal/config"
	"mrunc/internal/container"

)
func createCommand(ctx *cli.Context) error {
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

	parentSock, childSock, err := utils.SocketPair()
	if err != nil {
		panic(err)
	}
	defer parentSock.Close()
	defer childSock.Close()

	configData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	cmd := exec.Command("/proc/self/exe", append([]string{"intermediate"}, os.Args[2:]...)...)
	cmd.ExtraFiles = []*os.File{childSock}

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

	// does not create namespace yet
	// cmd.SysProcAttr = runtime.CreateNamespaces()


	if err := cmd.Start(); err != nil {

		return err
	}

	_, err = parentSock.Write(configData)
	if err != nil {
		return fmt.Errorf("failed to send config: %v", err)
	}
	parentSock.Close()
	if err := cmd.Wait(); err != nil {
		fmt.Printf("PARENT: Intermediate exited with error: %v\n", err)
	} else {
		fmt.Println("PARENT: Intermediate completed successfully")
	}

	return nil
}