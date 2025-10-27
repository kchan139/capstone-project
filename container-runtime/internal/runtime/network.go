package runtime

import (
	"fmt"
	"os/exec"
)

// brings up the loopback interface inside the container
func SetupLoopback() error {
	fmt.Println("Setting up loopback interface...")

	// Bring up lo interface
	cmd := exec.Command("ip", "link", "set", "lo", "up")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to bring up loopback: %v, output: %s", err, string(output))
	}

	fmt.Println("Loopback interface is up")
	return nil
}

// checks if basic networking is working
func VerifyNetwork() {
	// Check available interfaces
	cmd := exec.Command("ip", "link", "show")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Network verification warning: %v\n", err)
		return
	}
	fmt.Printf("Network interfaces:\n%s\n", string(output))
}
