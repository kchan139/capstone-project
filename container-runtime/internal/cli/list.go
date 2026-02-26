package cli

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli/v2"
	"mrunc/pkg/specs"
)

const mruncStateDir = "/run/mrunc"

type ContainerStatus string

const (
	StatusCreated ContainerStatus = "created"
	StatusRunning ContainerStatus = "running"
	StatusStopped ContainerStatus = "stopped"
)

// ContainerStateInternal is what we display in mrunc list.
type ContainerStateInternal struct {
	ID      string
	PID     int
	Status  ContainerStatus
	Bundle  string
	Created time.Time
	Owner   string
}

func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil
}

// resolves a UID to a username, falling back to the UID string.
func ownerName(uid int) string {
	u, err := user.LookupId(fmt.Sprintf("%d", uid))
	if err != nil {
		return fmt.Sprintf("%d", uid)
	}
	return u.Username
}

// returns the owner username of the file/directory.
func ownerOfStateFile(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return ownerName(os.Getuid())
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return ownerName(os.Getuid())
	}
	return ownerName(int(stat.Uid))
}

func listCommand(ctx *cli.Context) error {
	entries, err := os.ReadDir(mruncStateDir)
	if err != nil {
		if os.IsNotExist(err) {
			printContainerTable(nil)
			return nil
		}
		return fmt.Errorf("failed to read state directory %s: %v", mruncStateDir, err)
	}

	var containers []ContainerStateInternal

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		stateFile := filepath.Join(mruncStateDir, entry.Name(), "state.json")
		ps, err := specs.LoadContainerState(stateFile)
		if err != nil {
			continue
		}

		cs := ContainerStateInternal{
			ID:      ps.ContainerID,
			Bundle:  ps.BundlePath,
			Created: ps.Created,
			Owner:   ownerOfStateFile(filepath.Join(mruncStateDir, entry.Name())),
		}

		switch ps.Status {
		case "created":
			cs.Status = StatusCreated
			cs.PID = ps.ContainerPID
		case "running":
			if isProcessAlive(ps.ContainerPID) {
				cs.Status = StatusRunning
				cs.PID = ps.ContainerPID
			} else {
				cs.Status = StatusStopped
				cs.PID = 0
			}
		default:
			cs.Status = StatusStopped
			cs.PID = 0
		}

		containers = append(containers, cs)
	}

	printContainerTable(containers)
	return nil
}

// printContainerTable renders the container list in runc-style tabular format.
func printContainerTable(containers []ContainerStateInternal) {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "ID\tPID\tSTATUS\tBUNDLE\tCREATED\tOWNER")
	for _, c := range containers {
		pidStr := fmt.Sprintf("%d", c.PID)
		if c.PID == 0 {
			pidStr = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			c.ID,
			pidStr,
			string(c.Status),
			c.Bundle,
			c.Created.Format(time.RFC3339),
			c.Owner,
		)
	}
}