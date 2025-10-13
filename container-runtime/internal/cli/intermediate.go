package cli

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"mrunc/pkg/specs"
	"os"
	"bytes"
	"encoding/json"
	"io"

)
func intermediateCommand(ctx *cli.Context) error {
	config, err := receiveConfigFromSocket()
	if err != nil {
		return fmt.Errorf("child: failed to receive config: %v", err)
	}
	fmt.Println("intermediate called")
	fmt.Println(config.Hostname)
	
	return nil
}


func receiveConfigFromSocket() (*specs.ContainerConfig, error) {
	// hard-code the file descriptor
	f := os.NewFile(uintptr(3), "parent-sock")
	defer f.Close()

	var buf bytes.Buffer

	if _, err := io.Copy(&buf, f); err != nil {
		return nil, err
	}

	var config specs.ContainerConfig
	if err := json.Unmarshal(buf.Bytes(), &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %v", err)
	}
	return &config, nil
}
