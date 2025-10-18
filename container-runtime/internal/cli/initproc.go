package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)


func initprocCommand(ctx *cli.Context) error {
	fmt.Println("Init command run")
	fmt.Printf("Init, my pid are %d", os.Getpid())

	time.Sleep(300 * time.Second)
	return nil
}