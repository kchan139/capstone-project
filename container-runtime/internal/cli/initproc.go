package cli

import (
	"fmt"
	"mrunc/internal/runtime"
	"mrunc/internal/utils"
	"os"
	"syscall"
	"strings"
	"path/filepath"
	"github.com/urfave/cli/v2"
	"golang.org/x/sys/unix"
)

func initprocCommand(ctx *cli.Context) error {

	config, err := utils.ReceiveConfigFromPipe()
	fmt.Println("I am still in initproc")

	if err != nil {
		return fmt.Errorf("child: failed to receive config: %v", err)
	}


	// Set hostname
	if err := syscall.Sethostname([]byte(config.Hostname)); err != nil {
		return fmt.Errorf("failed to set hostname: %v", err)
	}




	root_fs := config.RootFS.Path
	root_fs_putold := config.RootFS.Path + "/put_old"
	os.MkdirAll(root_fs_putold, 0755)

	if err := unix.Mount(root_fs, root_fs, "", unix.MS_BIND, ""); err != nil {
		panic(fmt.Errorf("bind mount failed: %w", err))
	}
	// mount directories
	for _,mount := range config.Mounts {
		destination := filepath.Join(root_fs, mount.Destination)
		if err := os.MkdirAll(destination, 0755); err != nil {
			return fmt.Errorf("failed to create mount point %s: %v", destination, err)
		}
		var flags uintptr = 0
		var dataOpts []string
		// handle slightly different for cgroup mounts
		if mount.Type == "cgroup" || mount.Type == "cgroup2" {
			if err := syscall.Mount(config.CgroupPath , root_fs + mount.Destination, "", syscall.MS_BIND | syscall.MS_REC,""); err != nil {
				return fmt.Errorf("failed to mount %s at %s: %v", mount.Source, destination, err)
			}
			if err := syscall.Mount("", root_fs + mount.Destination, "", syscall.MS_REMOUNT | syscall.MS_RDONLY | syscall.MS_BIND, ""); err != nil {
				return fmt.Errorf("failed to mount %s at %s: %v", mount.Source, destination, err)
			}
			continue
		}
		for _, opt := range mount.Options {
			switch opt {
			case "nosuid":
				flags |= syscall.MS_NOSUID
			case "noexec":
				flags |= syscall.MS_NOEXEC
			case "nodev":
				flags |= syscall.MS_NODEV
			case "ro", "readonly":
				flags |= syscall.MS_RDONLY
			case "bind":
				flags |= syscall.MS_BIND
			case "rbind":
				flags |= syscall.MS_BIND | syscall.MS_REC
			 case "relatime":
            	flags |= syscall.MS_RELATIME
			case "noatime":
				flags |= syscall.MS_NOATIME
			case "strictatime":
				flags |= syscall.MS_STRICTATIME
			default:
				dataOpts = append(dataOpts, opt)
			}
		}
		dataStr := strings.Join(dataOpts, ",")
		if err := syscall.Mount(mount.Source, destination, mount.Type, flags, dataStr); err != nil {
			return fmt.Errorf("failed to mount %s at %s: %v", mount.Source, destination, err)
		}
	}


	// Pivot root
	if err := runtime.PivotRoot(root_fs, root_fs_putold); err != nil {
		return fmt.Errorf("failed to pivot root: %v", err)
	}

	workDir := config.Process.Cwd
	if workDir == "" {
		workDir = "/"
	}

	// Change to root directory
	if err := syscall.Chdir(workDir); err != nil {
		return fmt.Errorf("failed to chdir: %v", err)
	}

	// Cleanup old root
	syscall.Unmount("/put_old", syscall.MNT_DETACH)
	os.RemoveAll("/put_old")

	// TODO: need to re-implement this, the master fd will be sent to a console socket
	if !config.Process.Terminal {
		fmt.Printf("Non-interactive mode: detaching from terminal\n")
		if _, err := syscall.Setsid(); err != nil && err != syscall.EPERM {
			fmt.Printf("Warning: setsid failed: %v\n", err)
		}
	} else {
		if _, err := unix.Setsid(); err != nil {
			return fmt.Errorf("setsid: %w", err)
		}
		consoleSock := os.NewFile(uintptr(5), "console-socket")
        if consoleSock == nil {
            return fmt.Errorf("no console socket")
        }
        defer consoleSock.Close()
		fmt.Printf("DEBUG CHILD: Got console socket, fd=%d\n", consoleSock.Fd())
		// ======= create pty pair
		pty, closePty, err := runtime.SetupPty()
        if err != nil {
            return fmt.Errorf("setup pty: %w", err)
        }
        defer closePty()
		fmt.Printf("DEBUG CHILD: Created PTY, master fd=%d\n", pty.Master.Fd())
        // ======= Send master FD back to parent via socket
        rights := unix.UnixRights(int(pty.Master.Fd()))
        dummy := []byte{0}
        if err := unix.Sendmsg(int(consoleSock.Fd()), dummy, rights, nil, 0); err != nil {
            return fmt.Errorf("sendmsg: %w", err)
        }
		 fmt.Printf("DEBUG CHILD: Sent master FD successfully\n")
        // Close master in child - parent owns it now
        pty.Master.Close()


		// dup slave → 0,1,2
		for _, fd := range []int{0, 1, 2} {
			if err := unix.Dup2(int(pty.SlaveFile.Fd()), fd); err != nil {
				return fmt.Errorf("dup2 %d: %w", fd, err)
			}
		}
		// now fd 0 is the pty slave, make it controlling tty
		if err := unix.IoctlSetInt(0, unix.TIOCSCTTY, 0); err != nil {
			return fmt.Errorf("TIOCSCTTY: %w", err)
		}
	}


	// Setup network namespace
	if err := runtime.SetupLoopback(); err != nil {
		fmt.Printf("Warning: Failed to setup loopback: %v\n", err)
	}

	// Setup veth network if enabled
	if config.Linux.Network != nil && config.Linux.Network.EnableNetwork {
		netCfg := config.Linux.Network
		if err := runtime.SetupContainerNetwork(
			netCfg.VethContainer,
			netCfg.ContainerIP,
			netCfg.GatewayCIDR,
			netCfg.DNS,
		); err != nil {
			fmt.Printf("Warning: Failed to setup container network: %v\n", err)
		}
	}

	// Optional: Verify network setup
	runtime.VerifyNetwork()

	// Set environment variables
	if len(config.Process.Env) > 0 {
		os.Clearenv()
		for _, env := range config.Process.Env {
			if err := os.Setenv(utils.ParseEnvKey(env), utils.ParseEnvValue(env)); err != nil {
				return fmt.Errorf("failed to set environment: %v", err)
			}
		}
	}

	if err := runtime.SetProcessUser(config.Process.User); err != nil {
		return fmt.Errorf("failed to set process user: %v", err)
	}




	// Execute the process (replace current process)
	command := config.Process.Args[0]
	args := config.Process.Args
	env := os.Environ()

	execPath, execArgs, err := runtime.PrepareExec(command, args, env)
	if err != nil {
		return fmt.Errorf("failed to prepare exec: %v", err)
	}

	// apply capabilities
	if err := runtime.SetupCaps(config); err != nil {
		return fmt.Errorf("failed to set the capabilities : %v", err)
	}


	// apply seccomp
	if err := runtime.SetupSeccomp(config.Linux.SeccompConfig); err != nil {
		return fmt.Errorf("failed to set the seccomp : %v", err)
	}

	fd := uintptr(4)
	fifo_fd := os.NewFile(fd, "inherited-fifo")
	sync_buf := make([]byte, 100)
	fmt.Println("before calling fifo read")
	n, _ := fifo_fd.Read(sync_buf)
	fmt.Printf("after calling fifo read %d\n", n)

	return  runtime.ExecuteCommand(execPath, execArgs, env)
}
