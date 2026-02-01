package cli

import (
	"fmt"
	"strconv"
	"github.com/urfave/cli/v2"
	"mrunc/pkg/specs"
	"log"
	"runtime"
	"golang.org/x/sys/unix"
	"os"
)

func monitorCommand(ctx *cli.Context) error {
	containerPid, _ := strconv.Atoi(os.Getenv("CONTAINER_PID"))
	ownPid := os.Getpid()
	log.Printf("At monitor process, get container pid: %d\n", containerPid)
	// open the fd of container and host mount namespace
	containerMountNsPath := fmt.Sprintf("/proc/%d/ns/mnt", containerPid)
    containerMountNsFd, err := unix.Open(containerMountNsPath, unix.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open container mount namespace: %v", err)
	}

	ownMountNsPath := fmt.Sprintf("/proc/%d/ns/mnt", ownPid)
    _, err = unix.Open(ownMountNsPath, unix.O_RDONLY, 0)
	// TODO: Setup & enter the container's namespace
	//////////// paste monitor code here
	watchPID := containerPid

	conf, err := specs.LoadConfigFanotify("configs/examples/monitor_config.json")
	if err != nil {
		log.Printf("Failed to load config: %v", err)
	}
	log.Printf("Successfully load config: %v\n",conf.WatchRules[0].Path)
	log.Printf("Container pid: %d\n",watchPID)

	// Validate configuration
	if err := conf.Validate(); err != nil {
		log.Printf("Invalid config: %v\n", err)
	}

	if len(conf.WatchRules) == 0 {
		log.Printf("No watch rules defined in configuration")
	}

	// trying to setns into the container's mount namespace
	runtime.LockOSThread()
	// VERY IMPORTANT
	if err := unix.Unshare(unix.CLONE_FS); err != nil {
		fmt.Printf("unshare fs failed: %v\n", err)
	}

	if err := unix.Setns(containerMountNsFd, unix.CLONE_NEWNS); err != nil {
		fmt.Printf("set namespace failed: %v\n", err)
	}

	// Finish setting up, signal back to parent
	monitorSock := os.NewFile(3, "monitor-sock")
	monitorSock.Write([]byte("OK"))

	return nil
}
