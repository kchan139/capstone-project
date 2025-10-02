package container

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "path/filepath"
    "mrunc/pkg/specs"
    "mrunc/internal/runtime"

)

func LoadConfig(configPath string) (*specs.ContainerConfig, error) {
    // Read the JSON file
    data, err := ioutil.ReadFile(configPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file %s: %v", configPath, err)
    }

    // Parse JSON
    var config specs.ContainerConfig
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config JSON: %v", err)
    }

    // Validate required fields
    if err := validateConfig(&config); err != nil {
        return nil, fmt.Errorf("invalid configuration: %v", err)
    }

    // Convert relative paths to absolute paths
    if !filepath.IsAbs(config.RootFS.Path) {
        absPath, err := filepath.Abs(config.RootFS.Path)
        if err != nil {
            return nil, fmt.Errorf("failed to resolve rootfs path: %v", err)
        }
        config.RootFS.Path = absPath
    }

    return &config, nil
}

func validateConfig(config *specs.ContainerConfig) error {
    // Validate rootfs
    if config.RootFS.Path == "" {
        return fmt.Errorf("rootfs path is required")
    }
    
    // Validate process
    if len(config.Process.Args) == 0 {
        return fmt.Errorf("process args are required")
    }
    
    // Validate user (if specified)
    if err := runtime.ValidateUser(config.Process.User); err != nil {
        return fmt.Errorf("invalid user configuration: %v", err)
    }
    
    return nil
}