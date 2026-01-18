package cli

import (
	"encoding/json"
	"fmt"
	"mrunc/internal/config"
	"mrunc/internal/container"
	"mrunc/internal/runtime"
	"os"
	"os/exec"
	"mrunc/internal/utils"
	"golang.org/x/sys/unix"
	"net"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

func createCommand(ctx *cli.Context) error {
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

	// create the exec.fifo files
	fifo_fd, err := createExecFifo(config.ContainerId)
	if err != nil {
		return err
	}
	var extra []*os.File
	var parentSock, childSock *os.File
	if config.Process.Terminal {
		parentSock, childSock, err = utils.SocketPair()
		if err != nil {
			return err
		}
		extra = []*os.File{childPipe,fifo_fd, childSock}
	} else {
		extra = []*os.File{childPipe,fifo_fd}
	}



	cmd := exec.Command("/proc/self/exe", append([]string{"initproc"}, os.Args[2:]...)...)
	cmd.ExtraFiles = extra


	cmd.Env = append(os.Environ(), "_MRUNC_PIPE_FD=3")
	cmd.SysProcAttr = runtime.CreateNamespaces(config)
	if err := cmd.Start(); err != nil {
		parentPipe.Close()
		childPipe.Close()
		return err
	}
	//----------------------------------------- setup cgroup
	fmt.Printf("Child PID: %d",cmd.Process.Pid)
	var cgroupPath string
	if cgroupPath, err = runtime.CreateCgroup(config, cmd.Process.Pid); err != nil {
		return fmt.Errorf("failed to create cgroup: %v", err)
	}
	config.CgroupPath = cgroupPath
	configData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	_, err = parentPipe.Write(configData)
	parentPipe.Close()
	if err != nil {
		return fmt.Errorf("failed to send config: %v", err)
	}

	if config.Process.Terminal {
		childSock.Close()
		buf := make([]byte, 1)
        oob := make([]byte, unix.CmsgSpace(4))
        _, oobn, _, _, err := unix.Recvmsg(int(parentSock.Fd()), buf, oob, 0)
		 if err != nil {
            return fmt.Errorf("recvmsg: %w", err)
        }

        msgs, err := unix.ParseSocketControlMessage(oob[:oobn])
        if err != nil {
            return fmt.Errorf("parse control message: %w", err)
        }

        fds, err := unix.ParseUnixRights(&msgs[0])
        if err != nil {
            return fmt.Errorf("parse unix rights: %w", err)
        }

        masterFd := fds[0]
		fmt.Printf("has master fd: %d",masterFd)
		// ===== connect with outside socket
		addr, err := net.ResolveUnixAddr("unix", "/tmp/fd_broker.sock")
		if err != nil {
			log.Fatal(err)
		}

		conn, err := net.DialUnix("unix", nil, addr)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		msg := []byte(config.ContainerId)

		// Send FD via ancillary data
		oob = unix.UnixRights(masterFd)
		_, _, err = conn.WriteMsgUnix(msg, oob, nil)
		if err != nil {
			log.Fatal("WriteMsgUnix:", err)
		}

		// Wait for acknowledgment
		ack := make([]byte, 16)
		n, err := conn.Read(ack)
		if err != nil {
			log.Fatal("Read ack:", err)
		}

		response := string(ack[:n])
		if response == "OK" {
			fmt.Printf("FD sent successfully with key '%s' (fd=%d)",
				config.ContainerId,  masterFd, )
		} else {
			log.Fatalf("Failed to send FD: %s", response)
		}
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

	// Cleanup veth on exit
	if config.Linux.Network != nil && config.Linux.Network.EnableNetwork {
		runtime.CleanupVeth(config.Linux.Network.VethHost)
	}

	return nil
}
