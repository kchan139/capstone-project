package cli

import (
	"encoding/json"
	"fmt"
	"mrunc/internal/config"
	"mrunc/internal/container"
	"mrunc/internal/runtime"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

func runCommand(ctx *cli.Context) error {
	var configPath string

	if ctx.NArg() < 1 {
		// No config specified → use default path
		baseDir := os.Getenv("MRUNC_BASE")
		if baseDir == "" {
			baseDir = config.BaseImageDir
		}
		configPath = filepath.Join(baseDir, "ubuntu", "ubuntu.json")
		fmt.Printf("No config specified, using default: %s\n", configPath)
	} else {
		// Use provided config
		configPath = ctx.Args().Get(0)
	}

	config, err := container.LoadConfig(configPath)
	if err != nil {
		return err
	}

	childPipe, parentPipe, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %v", err)
	}

	configData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.ExtraFiles = []*os.File{childPipe}

	if config.Process.Terminal {
		fmt.Printf("Starting container in interactive mode\n")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		fmt.Printf("Starting container in non-interactive mode\n")
		cmd.Stdin = nil
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	cmd.Env = append(os.Environ(), "_MRUNC_PIPE_FD=3")
	cmd.SysProcAttr = runtime.CreateNamespaces()

	if err := cmd.Start(); err != nil {
		parentPipe.Close()
		childPipe.Close()
		return err
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
	if err != nil {
		parentPipe.Close()
		return fmt.Errorf("failed to send config: %v", err)
	}
	parentPipe.Close()

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
