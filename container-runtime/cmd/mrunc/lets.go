package main

import (
    "fmt"
    "os"
    "my-capstone-project/internal/cli"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: mrunc <command> [args...]")
        os.Exit(1)
    }
    
    if err := cli.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}