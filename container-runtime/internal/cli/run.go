package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"mrunc/internal/config"
	"mrunc/internal/container"
	"mrunc/internal/runtime"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"
)

func runCommand(ctx *cli.Context) error {
	var configPath string

	if ctx.NArg() < 2 {
		// No container name and config specified → use default path
		baseDir := os.Getenv("MRUNC_BASE")
		if baseDir == "" {
			baseDir = config.BaseImageDir
		}
		configPath = filepath.Join(baseDir, "ubuntu", "ubuntu.json")
		fmt.Printf("No config specified, using default: %s\n", configPath)
	} else {
		// Use provided config
		configPath = ctx.Args().Get(1)
	}
	containerId := ctx.Args().Get(0)

	config, err := container.LoadConfig(configPath)
	if err != nil {
		return err
	}
	config.ContainerId = containerId

	childPipe, parentPipe, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %v", err)
	}

	configData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	var extra []*os.File
	var host *runtime.HostConsole
	var pty *runtime.PtyFiles
	var restoreConsole func()
	var closePty func()
	if config.Process.Terminal {
		fmt.Printf("Starting container in interactive mode\n")
		host, restoreConsole, err = runtime.SetupHostConsole()
		if err != nil {
			return err
		}
		defer restoreConsole()
		// 2. create pty pair
		pty, closePty, err = runtime.SetupPty()
		if err != nil {
			return err
		}
		defer closePty()
		extra = []*os.File{childPipe, pty.SlaveFile}
	} else {
		fmt.Printf("Starting container in non-interactive mode\n")
		extra = []*os.File{childPipe}

	}
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)

	cmd.ExtraFiles = extra

	cmd.Env = append(os.Environ(), "_MRUNC_PIPE_FD=3")
	cmd.SysProcAttr = runtime.CreateNamespaces()

	if err := cmd.Start(); err != nil {
		parentPipe.Close()
		childPipe.Close()
		return err
	}
	// setup cgroup
	fmt.Printf("run process : %d\n", os.Getpid())

	fmt.Printf("child process : %d\n", cmd.Process.Pid)
	if err := runtime.CreateCgroup(config, cmd.Process.Pid); err != nil {
		return fmt.Errorf("failed to create cgroup: %v", err)
	}

	if config.Process.Terminal {
		if err := pty.SlaveFile.Close(); err != nil {
			return fmt.Errorf("failed to close PTY slave file: %v", err)
		}

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGWINCH)
		defer signal.Stop(sigCh)
		stopResize := runtime.StartWinchForwarder(host.Host, pty.MasterConsole, sigCh)
		defer stopResize()
	}
	childPipe.Close()

	// Setup veth pair from parent side if network enabled
	if config.Linux.Network != nil && config.Linux.Network.EnableNetwork {
		// // Give child time to setup network namespace
		// time.Sleep(500 * time.Millisecond)

		// Wait for child to setup network namespace by polling for /proc/<pid>/ns/net
		nsPath := fmt.Sprintf("/proc/%d/ns/net", cmd.Process.Pid)
		const pollInterval = 50 * time.Millisecond
		const pollTimeout = 2 * time.Second
		var waited time.Duration
		for {
			if _, err := os.Stat(nsPath); err == nil {
				break
			}
			if waited >= pollTimeout {
				parentPipe.Close()
				return fmt.Errorf("network namespace for child process (%d) did not appear within timeout", cmd.Process.Pid)
			}
			time.Sleep(pollInterval)
			waited += pollInterval
		}

		netCfg := config.Linux.Network

		// Extract IP without CIDR for firewall script
		containerIP := strings.Split(netCfg.ContainerIP, "/")[0]

		if err := runtime.SetupHostVethPair(
			cmd.Process.Pid,
			netCfg.VethHost,
			netCfg.VethContainer,
			netCfg.ContainerIP,
			netCfg.GatewayCIDR,
		); err != nil {
			fmt.Printf("Warning: Failed to setup veth pair: %v\n", err)
		} else {
			// Apply firewall script if specified
			if netCfg.FirewallScript != "" {
				if err := runtime.ApplyFirewallScript(
					netCfg.FirewallScript,
					netCfg.VethHost,
					containerIP,
				); err != nil {
					fmt.Printf("Warning: Firewall script failed: %v\n", err)
					fmt.Printf("You can apply it manually later\n")
				}
			} else {
				fmt.Printf("\n   No firewall script specified in config\n")
				fmt.Printf("Network created but no NAT/forwarding rules applied\n")
				fmt.Printf("Example script: configs/firewall-setup.sh\n\n")
			}
		}
	}

	_, err = parentPipe.Write(configData)
	parentPipe.Close()
	if err != nil {
		return fmt.Errorf("failed to send config: %v", err)
	}

	if config.Process.Terminal && pty != nil {
		go func() {
			_, _ = io.Copy(os.Stdout, pty.Master)
		}()

		// our terminal → child
		go func() {
			_, _ = io.Copy(pty.Master, os.Stdin)
		}()
	}

	if err := cmd.Wait(); err != nil {
		fmt.Printf("PARENT: Child exited with error: %v\n", err)
	} else {
		fmt.Println("PARENT: Child completed successfully")
	}

	// Cleanup veth on exit
	if config.Linux.Network != nil && config.Linux.Network.EnableNetwork {
		runtime.CleanupVeth(config.Linux.Network.VethHost)
	}

	return nil
}
