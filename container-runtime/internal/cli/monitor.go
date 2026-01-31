package cli

import (
	"fmt"
	"os"
	"strconv"
	"github.com/urfave/cli/v2"
)

func monitorCommand(ctx *cli.Context) error {
	containerPid, _ := strconv.Atoi(os.Getenv("CONTAINER_PID"))
	fmt.Printf("At monitor process, get container pid: %d\n", containerPid)

	// TODO: Setup & enter the container's namespace
	//////////// paste monitor code here
	//
	//
	//
	// Finish setting up, signal back to parent
	monitorSock := os.NewFile(3, "monitor-sock")
	monitorSock.Write([]byte("OK"))
	// the handle event loops
	return nil
}
