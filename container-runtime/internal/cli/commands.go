package cli

import (
	"fmt"
	"github.com/urfave/cli/v2"
)

func NewApp() *cli.App {
	return &cli.App{
		Name:    "mrunc",
		Usage:   "A minimal container runtime",
		Version: GetVersion(),
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "Run a container from an OCI bundle",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "fanotify-monitor",
						Usage: "path to the fanotify monitor configuration",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "bundle",
						Usage: "path to OCI bundle directory containing config.json",
						Value: "",
					},
				},
				ArgsUsage: "<container-id>",
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
					fmt.Fprintln(ctx.App.Writer, GetVersionInfo())
					return nil
				},
			},
			{
				Name:  "create",
				Usage: "Create a container without starting its process",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "console-socket",
						Usage: "path to unix socket for sending console fd",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "fanotify-monitor",
						Usage: "path to the fanotify monitor configuration",
						Value: "",
					},
					&cli.StringFlag{
						Name:  "bundle",
						Usage: "path to OCI bundle directory containing config.json",
						Value: "",
					},
				},
				ArgsUsage: "<container-id>",
				Action:    createCommand,
			},
			{
				Name:   "intermediate",
				Hidden: true,
				Usage:  "The intermediate process between parent process and init process",
				Action: intermediateCommand,
			},
			{
				Name:   "initproc",
				Hidden: true,
				Usage:  "The Init process that is waiting to run",
				Action: initprocCommand,
			},
			{
				Name:   "start",
				Usage:  "Signal the created container to start",
				Action: startCommand,
			},
			{
				Name:   "monitor",
				Usage:  "Signal the created container to start",
				Hidden: true,
				Action: monitorCommand,
			},
			{
				Name:   "list",
				Usage:  "List all the current container in the machine and their metadata",
				Action: listCommand,
			},
			{
				Name:   "delete",
				Usage:  "Delete all the resources which associate with the container",
				Action: deleteCommand,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Usage:   "forcibly deletes the container if it is still running",
					},
				},
			},
			{
				Name:   "kill",
				Usage:  "Kill the container process.",
				Action: killCommand,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "all",
						Aliases: []string{"a"},
						Usage:   "Kill all processes associating with the container",
					},
				},
			},
		},
	}
}
