package runtime



import (
	"mrunc/pkg/specs"
	"fmt"
)

func UpdateStateFile(config *specs.ContainerConfig , containerPid int, status string) error {
    state := &specs.ContainerState{
		ContainerPID: containerPid,
		ContainerID:  config.ContainerId,
		CgroupPath:   config.CgroupPath,
		Status:       status,
	}
    if err := state.Validate(); err != nil {
        return fmt.Errorf("invalid container state: %v", err)
    }

    stateFile := specs.GetContainerStateFile(config.ContainerId)

	if err := specs.SaveContainerState(stateFile, state); err != nil {
        return fmt.Errorf("failed to save state: %v", err)
    }

    return nil

}
