package cli

import (
	"encoding/json"
	"fmt"
	"mrunc/internal/container"
	"mrunc/internal/runtime"
	"os"
	"os/exec"
	"mrunc/internal/utils"
	"golang.org/x/sys/unix"
	"net"
	"strconv"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

func createCommand(ctx *cli.Context) error {
	var consoleSockPath string
	var fanotifyMonitorFilePath = ctx.String("fanotify-monitor")
	consoleSockPath = ctx.String("console-socket")
	var bundlePath = ctx.String("bundle")
	var configPath, _ = utils.ResolveConfigPath(bundlePath)

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
	var SyncParentSock, SyncChildSock ,_ = utils.SocketPair()
	if config.Process.Terminal {
		parentSock, childSock, err = utils.SocketPair()
		if err != nil {
			return err
		}
		extra = []*os.File{childPipe,fifo_fd, childSock, SyncChildSock}
	} else {
		extra = []*os.File{childPipe,fifo_fd, nil, SyncChildSock}
	}



	cmd := exec.Command("/proc/self/exe", "initproc")
	cmd.ExtraFiles = extra


	cmd.Env = append(os.Environ(), "_MRUNC_PIPE_FD=3","BUNDLE_PATH="+bundlePath)
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
		addr, err := net.ResolveUnixAddr("unix", consoleSockPath)
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
	//0. create /run/mrunc/<container-id> to store state.json, and audit.json
	infoContainerDir := filepath.Join("/run/mrunc", containerId)
	err = os.MkdirAll(infoContainerDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	// ///////// MONITOR PHASE ///////////
	// 1. WAIT FOR CHILD TO FINISH SETTING UP
	buf := make([]byte, 64)
	n, err := SyncParentSock.Read(buf)
	if err != nil {
		fmt.Printf("failed to read from sync socket: %v", err)
	}
	signal := string(buf[:n])
	fmt.Printf("After child ready: %v\n",signal)

	// ////// TODO: Write data to state.json (container pid, other data)
	runtime.UpdateStateFile(config, cmd.Process.Pid, "created")

	if fanotifyMonitorFilePath != "" {
		// 2. child is ready, fork and run the monitor process
		monitorCmd := exec.Command("/proc/self/exe", "monitor")
		monitorCmd.Env = append(os.Environ(),
			"CONTAINER_PID=" + strconv.Itoa(cmd.Process.Pid),
			"CONTAINER_ID=" + containerId,
			"FANOTIFY_FILEPATH=" + fanotifyMonitorFilePath,
			"ROOT_FS="+config.RootFS.Path,
		)
		monitorCmd.Stdin = os.Stdin
		monitorCmd.Stdout = os.Stdout
		monitorCmd.Stderr = os.Stderr
		monitorParentSock, monitorChildSock, _ := utils.SocketPair()
		monitorCmd.ExtraFiles = []*os.File{monitorChildSock}

		err = monitorCmd.Start()
		if err != nil {
			fmt.Printf("failed to start monitor process: %v", err)
		}

		// 3. wait for the monitor to ready
		monitorBuf := make([]byte, 64)
		n, err = monitorParentSock.Read(monitorBuf)
		if err != nil {
			fmt.Printf("failed to read from monitor parent socket: %v", err)
		}
		signal = string(buf[:n])
		fmt.Printf("After monitor ready: %v\n",signal)
		// 4. monitor is ready, send signal to child so child can continue
		SyncParentSock.Write([]byte("OK"))
	} else {
		// if no monitor, then just signal the child process
		SyncParentSock.Write([]byte("OK"))

	}




	// Cleanup veth on exit
	if config.Linux.Network != nil && config.Linux.Network.EnableNetwork {
		runtime.CleanupVeth(config.Linux.Network.VethHost)
	}

	return nil
}

func createExecFifo(containerId string) (*os.File, error) {
	dirPath := "/run/mrunc/" + containerId
	fifoPath := dirPath + "/exec.fifo"

	// Step 1: ensure directory exists
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, err
	}

	// Step 2: create FIFO
	if err := unix.Mkfifo(fifoPath, 0666); err != nil && !os.IsExist(err) {
		return nil, err
	}

	// Step 3: open it (both ends, so it doesn't block yet)
	fifoFile, err := os.OpenFile(fifoPath, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		return nil, err
	}

	return fifoFile, nil
}