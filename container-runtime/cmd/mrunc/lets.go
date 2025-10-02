package main

import (
	"fmt"
	"mrunc/internal/cli"
	"os"
)

func main() {
	if err := cli.NewApp().Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
