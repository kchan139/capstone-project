// pkg/specs/container_state.go
package specs

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
)

// ContainerState represents the runtime state of a container
type ContainerState struct {
    ContainerPID int    `json:"container_pid"`
    ContainerID  string `json:"container_id"`
    CgroupPath   string `json:"cgroup_path"`
    Status       string `json:"status"`
}

// Valid status values
const (
    StatusCreated = "created"
    StatusRunning = "running"
    StatusStopped = "stopped"
    StatusPaused  = "paused"
    StatusExited  = "exited"
)

// LoadContainerState reads container state from a JSON file
func LoadContainerState(filepath string) (*ContainerState, error) {
    data, err := os.ReadFile(filepath)
    if err != nil {
        return nil, fmt.Errorf("failed to read state file: %v", err)
    }

    var state ContainerState
    if err := json.Unmarshal(data, &state); err != nil {
        return nil, fmt.Errorf("failed to parse state file: %v", err)
    }

    return &state, nil
}

// SaveContainerState writes container state to a JSON file
func SaveContainerState(filepath string, state *ContainerState) error {
    // Marshal with indentation for readability
    data, err := json.MarshalIndent(state, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal state: %v", err)
    }

    // Write to file with 0644 permissions
    if err := os.WriteFile(filepath, data, 0644); err != nil {
        return fmt.Errorf("failed to write state file: %v", err)
    }

    return nil
}

// Validate checks if the container state is valid
func (s *ContainerState) Validate() error {
    if s.ContainerPID <= 0 {
        return fmt.Errorf("invalid container PID: %d", s.ContainerPID)
    }

    if s.ContainerID == "" {
        return fmt.Errorf("container ID cannot be empty")
    }

    if s.CgroupPath == "" {
        return fmt.Errorf("cgroup path cannot be empty")
    }

    // Validate status
    validStatuses := map[string]bool{
        StatusCreated: true,
        StatusRunning: true,
        StatusStopped: true,
        StatusPaused:  true,
        StatusExited:  true,
    }

    if !validStatuses[s.Status] {
        return fmt.Errorf("invalid status: %s", s.Status)
    }

    return nil
}

// UpdateStatus updates the container status and saves to file
func UpdateContainerStatus(filepath, newStatus string) error {
    state, err := LoadContainerState(filepath)
    if err != nil {
        return err
    }

    state.Status = newStatus
    return SaveContainerState(filepath, state)
}

// GetContainerStateFile returns the standard path for a container's state file
func GetContainerStateFile(containerID string) string {
    return filepath.Join("/run/mrunc", containerID, "state.json")
}

// CreateInitialState creates a new container state with default values
func CreateInitialState(containerID string, pid int, cgroupPath string) *ContainerState {
    return &ContainerState{
        ContainerPID: pid,
        ContainerID:  containerID,
        CgroupPath:   cgroupPath,
        Status:       StatusCreated,
    }
}