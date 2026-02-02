package cli

import (
	"fmt"
	"strconv"
	"github.com/urfave/cli/v2"
	"mrunc/pkg/specs"
	"log"
	"mrunc/internal/monitorfanotify"
	"runtime"
	"golang.org/x/sys/unix"
	"os"
	"time"
	"unsafe"
	"syscall"
	"path/filepath"
)
// mountInfo tracks mount information for cleanup
type mountInfo struct {
	path    string
	mounted bool
}


func monitorCommand(ctx *cli.Context) error {
	containerPid, _ := strconv.Atoi(os.Getenv("CONTAINER_PID"))
	containerName := os.Getenv("CONTAINER_ID")
	ownPid := os.Getpid()
	log.Printf("At monitor process, get container pid: %d\n", containerPid)
	// open the fd of container and host mount namespace
	containerMountNsPath := fmt.Sprintf("/proc/%d/ns/mnt", containerPid)
    containerMountNsFd, err := unix.Open(containerMountNsPath, unix.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open container mount namespace: %v", err)
	}

	ownMountNsPath := fmt.Sprintf("/proc/%d/ns/mnt", ownPid)
    ownMountNsFd, err := unix.Open(ownMountNsPath, unix.O_RDONLY, 0)
	// TODO: Setup & enter the container's namespace
	//////////// paste monitor code here
	watchPID := containerPid

	conf, err := specs.LoadConfigFanotify(os.Getenv("FANOTIFY_FILEPATH"))
	if err != nil {
		log.Printf("Failed to load config: %v", err)
	}

	// Validate configuration
	if err := conf.Validate(); err != nil {
		log.Printf("Invalid config: %v\n", err)
	}

	if len(conf.WatchRules) == 0 {
		log.Printf("No watch rules defined in configuration")
	}

	// trying to setns into the container's mount namespace
	runtime.LockOSThread()
	if err := unix.Unshare(unix.CLONE_FS); err != nil {
		fmt.Printf("unshare fs failed: %v\n", err)
	}

	if err := unix.Setns(containerMountNsFd, unix.CLONE_NEWNS); err != nil {
		fmt.Printf("set namespace to container failed: %v\n", err)
	}
	// successfully enter the container's mount ns, now continue setting up
	// Validate all paths
	var validRules []specs.WatchRule
	for i, rule := range conf.WatchRules {
		absPath, err := filepath.Abs(rule.Path)
		if err != nil {
			log.Printf("Warning: Rule %d - Failed to get absolute path for %s: %v", i+1, rule.Path, err)
			continue
		}

		// Check if directory exists
		if info, err := os.Stat(absPath); err != nil {
			log.Printf("Warning: Rule %d - Cannot access %s: %v", i+1, absPath, err)
			continue
		} else if !info.IsDir() {
			log.Printf("Warning: Rule %d - %s is not a directory", i+1, absPath)
			continue
		}

		// Update rule with absolute path
		rule.Path = absPath
		validRules = append(validRules, rule)
	}

	if len(validRules) == 0 {
		log.Fatal("No valid watch rules after validation")
	}

	fmt.Printf("Will monitor %d paths:\n", len(validRules))
	for i, rule := range validRules {
		fmt.Printf("  %d. %s (events: %v, action: %s)\n", i+1, rule.Path, rule.Events, rule.Action)
	}
	fmt.Println()

	// Check if any rule requires blocking
	hasBlockAction := false
	for _, rule := range validRules {
		if rule.Action == "block" {
			hasBlockAction = true
			break
		}
	}

	// Initialize fanotify with appropriate class
	var initFlags uint
	if hasBlockAction {
		initFlags = unix.FAN_CLASS_CONTENT | unix.FAN_CLOEXEC
		log.Println("Initializing with PERMISSION mode (blocking enabled)")
	} else {
		initFlags = unix.FAN_CLASS_NOTIF | unix.FAN_CLOEXEC
		log.Println("Initializing with NOTIFICATION mode (audit only)")
	}

	fd, err := unix.FanotifyInit(initFlags, unix.O_RDONLY|unix.O_LARGEFILE)
	if err != nil {
		log.Printf("FanotifyInit failed: %v", err)
	}
	// Track mounted directories
	mounts := make([]*mountInfo, 0, len(validRules))
	// Setup cleanup function
	cleanup := func() {
		fmt.Println("\nCleaning up...")

		// 1. Close fanotify fd first (VERY important)
		fmt.Println("Closing fanotify fd...")
		if fd >= 0 {
			unix.Close(fd)
		}

		// 2. Small delay to let kernel release references
		time.Sleep(200 * time.Millisecond)
		// 3. Unmount all directories in reverse order
		for i := len(mounts) - 1; i >= 0; i-- {
			mount := mounts[i]
			if !mount.mounted {
				continue
			}
			fmt.Printf("Unmounting: %s\n", mount.path)

			// Try multiple times
			for attempt := 0; attempt < 3; attempt++ {
				err := unix.Unmount(mount.path, 0)
				if err == nil {
					fmt.Printf("  ✓ Successfully unmounted layer %d: %s\n", attempt+1, mount.path)
					continue
				}

				if err == syscall.EBUSY {
					fmt.Printf("  Layer %d busy, retrying...\n", attempt+1)
					time.Sleep(100 * time.Millisecond)
					continue
				}

				if err == syscall.EINVAL || err == syscall.ENOENT {
					fmt.Printf("  Already unmounted: %s\n", mount.path)
					break
				}

				fmt.Printf("  Unmount error: %v\n", err)
				break
			}

			// Final fallback: lazy unmount
			err := unix.Unmount(mount.path, unix.MNT_DETACH)
			if err == nil {
				fmt.Printf("  ✓ Lazy unmount successful: %s\n", mount.path)
			} else if err != syscall.EINVAL && err != syscall.ENOENT {
				fmt.Printf("  ✗ Lazy unmount failed: %v\n", err)
			}
		}

		fmt.Println("Cleanup complete.")
	}
	// defer cleanup()

	if watchPID > 0 {
		// Verify the process exists first
		err := syscall.Kill(watchPID, 0)
		if err != nil {
			log.Fatalf("Process %d does not exist: %v", watchPID, err)
		}

		// Start watcher in goroutine
		go watchProcess(watchPID, cleanup)
	}

	// Mount and mark each directory
	for _, rule := range validRules {
		fmt.Printf("Setting up monitoring for: %s\n", rule.Path)

		// Create bind mount
		err = unix.Mount(rule.Path, rule.Path, "", unix.MS_BIND, "")
		if err != nil {
			log.Printf("  ✗ Bind mount failed for %s: %v (skipping)", rule.Path, err)
			continue
		}
		fmt.Printf("  ✓ Bind mount created\n")

		// Make it private (recommended for isolation)
		err = unix.Mount("", rule.Path, "", unix.MS_PRIVATE, "")
		if err != nil {
			log.Printf("  Warning: Failed to make mount private: %v", err)
		}

		// Track this mount
		mounts = append(mounts, &mountInfo{
			path:    rule.Path,
			mounted: true,
		})

		// Build event mask based on events and action
		eventMask, err := monitorfanotify.BuildEventMask(rule)
		if err != nil {
			log.Printf("  ✗ Failed to build event mask: %v (skipping)", err)
			continue
		}

		// Mark the mount point with dynamic events
		err = unix.FanotifyMark(
			fd,
			unix.FAN_MARK_ADD|unix.FAN_MARK_MOUNT,
			eventMask,
			unix.AT_FDCWD,
			rule.Path,
		)
		if err != nil {
			log.Printf("  ✗ FanotifyMark failed for %s: %v (skipping)", rule.Path, err)
			continue
		}
		fmt.Printf("  ✓ Fanotify mark added (mask: 0x%x, action: %s)\n", eventMask, rule.Action)
	}

	if len(mounts) == 0 {
		log.Fatal("Failed to setup monitoring for any directory")
	}

	if hasBlockAction {
		fmt.Printf("\n✓ Successfully monitoring %d/%d paths (PERMISSION mode - blocking enabled)\n",
			len(mounts), len(validRules))
		fmt.Println("Events will be BLOCKED according to rules... (Ctrl+C to stop)\n")
	} else {
		fmt.Printf("\n✓ Successfully monitoring %d/%d paths (NOTIFICATION mode - audit only)\n",
			len(mounts), len(validRules))
		fmt.Println("Events are logged but NOT blocked... (Ctrl+C to stop)\n")
	}
	monitorSock := os.NewFile(3, "monitor-sock")
	monitorSock.Write([]byte("OK"))

	// Finish setting up, signal back to parent
	buf := make([]byte, 4096)
	firstIterate := 0
	// Event processing loop
	for {
		// if this is not the first iteration, then not enter mount namespace
		if firstIterate != 0 {
			if err := unix.Setns(containerMountNsFd, unix.CLONE_NEWNS); err != nil {
				fmt.Printf("set namespace to container failed: %v\n", err)
			}
		}
		n, err := unix.Read(fd, buf)
		if err != nil {
			log.Printf("Read error: %v", err)
			continue
		}

		var offset int
		for offset < n {
			event := (*unix.FanotifyEventMetadata)(unsafe.Pointer(&buf[offset]))

			if event.Fd >= 0 {
				// Check if this is a permission event that needs response
				if event.Mask&(unix.FAN_ALL_PERM_EVENTS|unix.FAN_OPEN_EXEC_PERM) != 0 {
					response := unix.FanotifyResponse{
						Fd: event.Fd,
					}
					//  permissive event => block
					response.Response = unix.FAN_DENY
					// Send response
					written, err := unix.Write(fd, (*[unsafe.Sizeof(response)]byte)(unsafe.Pointer(&response))[:])
					if err != nil {
						log.Printf("Failed to send response: %v", err)
					}
					fmt.Printf("  DEBUG: response Write n=%d, err=%v, sizeof=%d\n", written, err, unsafe.Sizeof(response))
				}
				// go back to host mount namespace
				if err := unix.Setns(ownMountNsFd, unix.CLONE_NEWNS); err != nil {
					fmt.Printf("set namespace to container failed: %v\n", err)
				}

				// Get the file path
				procPath := fmt.Sprintf("/proc/self/fd/%d", event.Fd)
				path, _ := os.Readlink(procPath)

				// Log the event
				eventType := monitorfanotify.EventMaskToString(uint64(event.Mask))
				timestamp := time.Now().Format("15:04:05.000")

				eventLog, err := os.OpenFile("/run/mrunc/" + containerName + "/audit.json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
				if err != nil {
					return fmt.Errorf("failed to open log file: %v", err)
				}
				logEntry := fmt.Sprintf("[%s] [%s] %s (PID: %d)\n", timestamp, eventType, path, event.Pid)
				fmt.Printf("[%s] [%s] %s (PID: %d)\n", timestamp, eventType, path, event.Pid)
				eventLog.WriteString(logEntry)
				eventLog.Close()
				// Close the file descriptor
				unix.Close(int(event.Fd))
			}

			offset += int(event.Event_len)
		}
		firstIterate += 1
	}
	return nil
}


func watchProcess(pid int, cleanup func()) {
	// Open a pidfd for the process
	pidfd, err := unix.PidfdOpen(pid, 0)
	if err != nil {
		log.Fatalf("PidfdOpen failed (requires Linux 5.3+): %v", err)
	}
	defer unix.Close(pidfd)

	fmt.Printf("Watching process %d using pidfd (blocking wait)...\n", pid)

	// Use poll to wait for process exit
	// When process exits, pidfd becomes readable
	fds := []unix.PollFd{
		{
			Fd:     int32(pidfd),
			Events: unix.POLLIN,
		},
	}

	// Block until process exits (infinite timeout)
	_, err = unix.Poll(fds, -1)
	if err != nil {
		log.Printf("Poll error: %v\n", err)
		cleanup()
		os.Exit(1)
	}

	fmt.Printf("\nProcess %d has exited, triggering cleanup...\n", pid)
	cleanup()
	os.Exit(0)
}