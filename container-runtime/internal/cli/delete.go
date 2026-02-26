package cli

import (
	"github.com/urfave/cli/v2"
	"fmt"
	"path/filepath"
	"mrunc/pkg/specs"
	"syscall"
	"time"
	"os"
)

func deleteCommand(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		return fmt.Errorf("Missing container name")
	}
	containerId := ctx.Args().Get(0)
	// 1. Get the information from /run/mrunc/<containerId>
	stateFile := filepath.Join(mruncStateDir, containerId, "state.json")
	ps, err := specs.LoadContainerState(stateFile)
	if err != nil {
		return fmt.Errorf("Error loading state.json: %v\n",err)
	}
	cs := ContainerStateInternal{
		ID:      ps.ContainerID,
		Bundle:  ps.BundlePath,
		Created: ps.Created,
		PID: ps.ContainerPID,
	}

	if (IsProcessAlive(cs.PID)) {
		force := ctx.Bool("force")
		if !force {
			return fmt.Errorf("The container is not stopped yet")
		}
		// kill the container
		if err := syscall.Kill(cs.PID, syscall.SIGKILL); err != nil {
			return fmt.Errorf("failed to kill container process %d: %v", cs.PID, err)
		}
		// Wait for it to be reaped
		syscall.Wait4(cs.PID, nil, 0, nil)
	}
	// container is stopped. Proceed to delete things
	// 1. delete the state directory
	if err := os.RemoveAll(filepath.Join(mruncStateDir, cs.ID)); err != nil {
		return fmt.Errorf("failed to delete state directory for container %s: %v", cs.ID, err)
	}
	// 2. remove the cgroup path
	killFile := filepath.Join(ps.CgroupPath, "cgroup.kill")
	if err := os.WriteFile(killFile, []byte("1"), 0); err != nil {
		return fmt.Errorf("failed to kill cgroup processes: %v", err)
	}

	// Give the kernel a moment to kill all processes
	time.Sleep(100 * time.Millisecond)
	if err := os.Remove(ps.CgroupPath); err != nil {
		return fmt.Errorf("failed to remove cgroup %s: %v", ps.CgroupPath, err)
	}
	return nil
}
