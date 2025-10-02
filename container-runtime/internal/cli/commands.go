package cli

import (
	"github.com/urfave/cli/v2"
)

const Version = "0.1.1"

func NewApp() *cli.App {
	return &cli.App{
		Name:    "mrunc",
		Usage:   "A minimal container runtime",
		Version: Version,
		Commands: []*cli.Command{
			{
				Name:      "run",
				Usage:     "Run a container from a config file",
				ArgsUsage: "<config.json>",
				Action:    runCommand,
			},
			{
				Name:   "child",
				Usage:  "Internal command for child process (do not call directly)",
				Hidden: true,
				Action: childCommand,
			},
			{
				Name:   "init",
				Usage:  "Initialize base rootfs images for mrunc",
				Action: initCommand,
			},
		},
	}
}
