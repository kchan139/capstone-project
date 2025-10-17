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
		panic(err)
	}
	parentSock2, childSock2, err := utils.SocketPair()
	if err != nil {
		panic(err)
	}
	fmt.Print("ignore")
	fmt.Println(parentSock2)
	defer parentSock.Close()
	defer childSock.Close()

	// configData, err := json.Marshal(config)
	// if err != nil {
	// 	return err
	// }

	cmd := exec.Command("/proc/self/exe", append([]string{"intermediate"}, os.Args[2:]...)...)
	cmd.ExtraFiles = []*os.File{childSock, childSock2}

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
	childSock.Close()
	childSock2.Close()
	// _, err = parentSock.Write(configData)
	// if err != nil {
	// 	return fmt.Errorf("failed to send config: %v", err)
	// }
	if err := json.NewEncoder(parentSock).Encode(config); err != nil {
		return fmt.Errorf("send config: %w", err)
	}

	// receive the PID of init process sent from intermediate process
	buf := make([]byte, 32)
	n, err := parentSock.Read(buf)
	if err != nil { return fmt.Errorf("read pid") }


	InitPidStr := string(buf[:n])

	pid, _ := strconv.Atoi(strings.TrimSpace(InitPidStr))
	fmt.Println("Init PID from intermediate:", pid)



	if err := cmd.Wait(); err != nil {
		fmt.Printf("PARENT: Intermediate exited with error: %v\n", err)
	} else {
		fmt.Println("PARENT: Intermediate completed successfully")
	}

	return nil
}