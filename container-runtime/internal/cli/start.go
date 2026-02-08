package cli

import (
	"fmt"
	"os"
	"mrunc/pkg/specs"
	"github.com/urfave/cli/v2"
)

func startCommand(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		fmt.Println("Error: must have the container name")
		return fmt.Errorf("missing container name")
	}

	containerId := ctx.Args().Get(0)
	fifoPath := "/run/mrunc/" + containerId + "/exec.fifo"
	fifoFile, err := os.OpenFile(fifoPath, os.O_WRONLY, os.ModeNamedPipe)

	if err != nil {
		return err
	}

	_, err = fifoFile.Write([]byte{'1'})
	if err != nil {
		return err
	}

	defer fifoFile.Close()


	// update state.json, from created to running
	specs.UpdateContainerStatus(specs.GetContainerStateFile(containerId), "running")

	return nil
}
