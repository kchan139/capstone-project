package cli

import (
    "fmt"
    "os"
)

func Execute() error {
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