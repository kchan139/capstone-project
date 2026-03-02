package cli

import (
	"github.com/urfave/cli/v2"
	"fmt"
	"path/filepath"
	"mrunc/pkg/specs"
	"syscall"
	"strconv"
	"strings"
	"time"
	"os"
)

func parseSignal(sigStr string) (syscall.Signal, error) {
	if sigStr == "" {
		return syscall.SIGTERM, nil
	}

	// numeric signal
	if n, err := strconv.Atoi(sigStr); err == nil {
		return syscall.Signal(n), nil
	}

	s := strings.ToUpper(sigStr)
	s = strings.TrimPrefix(s, "SIG")

	signalMap := map[string]syscall.Signal{
		"TERM": syscall.SIGTERM,
		"KILL": syscall.SIGKILL,
		"INT":  syscall.SIGINT,
		"HUP":  syscall.SIGHUP,
		"QUIT": syscall.SIGQUIT,
		"STOP": syscall.SIGSTOP,
		"CONT": syscall.SIGCONT,
	}

	if sig, ok := signalMap[s]; ok {
		return sig, nil
	}

	return -1, fmt.Errorf("unknown signal: %s", sigStr)
}


func killCommand(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		return fmt.Errorf("Missing container name")
	}
	containerId := ctx.Args().Get(0)
	// second arg = signal (optional)
	sigStr := ""
	if ctx.NArg() >= 2 {
		sigStr = ctx.Args().Get(1)
	}
	sig, err := parseSignal(sigStr)
	if err != nil {
		return err
	}

	// 1. Get the information from /run/mrunc/<containerId>
	stateFile := filepath.Join(mruncStateDir, containerId, "state.json")
	ps, err := specs.LoadContainerState(stateFile)
	if err != nil {
		return fmt.Errorf("Error loading state.json: %v\n",err)
	}
	if (ps.Status != "created" && ps.Status != "running") {
		return fmt.Errorf("Container is not created or running")
	}
	cs := ContainerStateInternal{
		ID:      ps.ContainerID,
		Bundle:  ps.BundlePath,
		Created: ps.Created,
		PID: ps.ContainerPID,
	}
	all := ctx.Bool("all")
	if (all) {
		killFile := filepath.Join(ps.CgroupPath, "cgroup.kill")
		if err := os.WriteFile(killFile, []byte("1"), 0); err != nil {
			return fmt.Errorf("failed to kill cgroup processes: %v", err)
		}
		// Give the kernel a moment to kill all processes
		time.Sleep(100 * time.Millisecond)
	} else {
		if err := syscall.Kill(cs.PID, sig); err != nil {
			return fmt.Errorf("failed to send signal %d to process %d: %v", sig, cs.PID, err)
		}
		// only wait if force kill
		if sig == syscall.SIGKILL {
			syscall.Wait4(cs.PID, nil, 0, nil)
		}
	}
	// update status
	specs.UpdateContainerStatus(specs.GetContainerStateFile(ps.ContainerID), "stopped")
	return nil
}
