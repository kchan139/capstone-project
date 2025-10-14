package cli

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
	"bytes"
	"encoding/json"
	mySpecs "mrunc/pkg/specs"
	"io"
	"time"
	"mrunc/internal/runtime"
)
func intermediateCommand(ctx *cli.Context) error {
	config, err := receiveConfigFromSocket()
	if err != nil {
		return fmt.Errorf("child: failed to receive config: %v", err)
	}
	fmt.Println("intermediate called")
	// setup cgroup
	runtime.CreateScope(config.ContainerId, os.Getpid())

	fmt.Println("Running inside limited cgroup for 10 seconds...")
	time.Sleep(300 * time.Second)
	
	return nil
}


func receiveConfigFromSocket() (*mySpecs.ContainerConfig, error) {
	// hard-code the file descriptor
	f := os.NewFile(uintptr(3), "parent-sock")
	defer f.Close()

	var buf bytes.Buffer

	if _, err := io.Copy(&buf, f); err != nil {
		return nil, err
	}

	var config mySpecs.ContainerConfig
	if err := json.Unmarshal(buf.Bytes(), &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %v", err)
	}
	return &config, nil
}
