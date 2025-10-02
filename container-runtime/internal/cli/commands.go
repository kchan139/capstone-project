package cli

import (
	"github.com/urfave/cli/v2"
)

func NewApp() *cli.App {
	return &cli.App{
		Name:    "mrunc",
		Usage:   "A minimal container runtime",
		Version: GetVersion(),
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
			{
				Name:  "version",
				Usage: "Show detailed version information",
				Action: func(ctx *cli.Context) error {
					println(GetVersionInfo())
					return nil
				},
			},
		},
	}
}
