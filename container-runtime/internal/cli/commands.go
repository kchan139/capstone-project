package cli

import (
    "fmt"
    "os"
)

func Execute() error {
     if len(os.Args) < 2 {
        return showUsage()
    }
    switch os.Args[1] {
    case "run":
        return runCommand()
    case "child":
        return childCommand()
    // case "create":
    //     return createCommand()
    // case "start":
    //     return startCommand()
    // case "stop":
    //     return stopCommand()
    // case "delete":
    //     return deleteCommand()
    default:
        return fmt.Errorf("unknown command: %s", os.Args[1])
    }
}

func showUsage() error {
    fmt.Println("Usage:")
    fmt.Println("  mrunc run <config.json>")
    fmt.Println("")
    fmt.Println("Examples:")
    fmt.Println("  mrunc run configs/examples/ubuntu-container.json")
    return nil
}