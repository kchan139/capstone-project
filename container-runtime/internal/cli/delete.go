package cli

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"mrunc/pkg/specs"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

func deleteCommand(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		return fmt.Errorf("missing container name")
	}
	containerId := ctx.Args().Get(0)
	// 1. Get the information from /run/mrunc/<containerId>
	stateFile := filepath.Join(mruncStateDir, containerId, "state.json")
	ps, err := specs.LoadContainerState(stateFile)
	if err != nil {
		return fmt.Errorf("error loading state.json: %v", err)
	}
	cs := ContainerStateInternal{
		ID:      ps.ContainerID,
		Bundle:  ps.BundlePath,
		Created: ps.Created,
		PID:     ps.ContainerPID,
	}

	if IsProcessAlive(cs.PID) {
		force := ctx.Bool("force")
		if !force {
			return fmt.Errorf("container is not stopped yet")
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
	if _, err := os.Stat(ps.CgroupPath); err == nil {
		os.WriteFile(filepath.Join(ps.CgroupPath, "cgroup.kill"),
			[]byte("1"), 0)
		time.Sleep(100 * time.Millisecond)
		if err := os.Remove(ps.CgroupPath); err != nil {
			return fmt.Errorf("failed to remove cgroup: %v", err)
		}
	}
	return nil
}
