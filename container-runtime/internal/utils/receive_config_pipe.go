package utils
import (
	"os"
	"fmt"
	"encoding/json"
	"io"
	"strconv"
	"mrunc/pkg/specs"


)
func ReceiveConfigFromPipe() (*specs.ContainerConfig, error) {
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