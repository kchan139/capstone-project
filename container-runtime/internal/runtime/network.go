package runtime

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// SetupContainerNetwork configures network inside container
func SetupContainerNetwork(vethName, containerIP, gatewayIP string, dnsServers []string) error {
	fmt.Printf("Setting up container network: %s with IP %s\n", vethName, containerIP)

	// Bring up the veth interface
	cmd := exec.Command("ip", "link", "set", vethName, "up")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring up %s: %v", vethName, err)
	}

	// Assign IP address
	cmd = exec.Command("ip", "addr", "add", containerIP, "dev", vethName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to assign IP %s: %v", containerIP, err)
	}

	// Add default route
	cmd = exec.Command("ip", "route", "add", "default", "via", gatewayIP)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add default route: %v", err)
	}

	// Setup DNS
	if len(dnsServers) > 0 {
		if err := setupDNS(dnsServers); err != nil {
			return fmt.Errorf("failed to setup DNS: %v", err)
		}
	}

	fmt.Println("Container network configured successfully")
	return nil
}

// creates /etc/resolv.conf with nameservers
func setupDNS(dnsServers []string) error {
	resolvConf := "/etc/resolv.conf"
	content := ""
	for _, dns := range dnsServers {
		content += fmt.Sprintf("nameserver %s\n", dns)
	}

	_ = os.Remove(resolvConf)

	if err := os.WriteFile(resolvConf, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write resolv.conf: %v", err)
	}

	fmt.Printf("DNS configured: %v\n", dnsServers)
	return nil
}

// creates veth pair from parent (host) side
func SetupHostVethPair(containerPID int, vethHost, vethContainer, containerIP, gatewayIP string) error {
	fmt.Printf("Creating veth pair: %s <-> %s\n", vethHost, vethContainer)

	// Create veth pair
	cmd := exec.Command("ip", "link", "add", vethHost, "type", "veth", "peer", "name", vethContainer)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create veth pair: %v\nOutput: %s", err, string(output))
	}

	// Move container end into container's network namespace
	netnsPath := fmt.Sprintf("/proc/%d/ns/net", containerPID)
	cmd = exec.Command("ip", "link", "set", vethContainer, "netns", netnsPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to move veth to container: %v", err)
	}

	// Configure host end
	cmd = exec.Command("ip", "addr", "add", gatewayIP+"/24", "dev", vethHost)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to assign IP to host veth: %v", err)
	}

	cmd = exec.Command("ip", "link", "set", vethHost, "up")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring up host veth: %v", err)
	}

	fmt.Printf("Host veth %s configured with IP %s\n", vethHost, gatewayIP)
	return nil
}

// executes an external firewall script
func ApplyFirewallScript(scriptPath, vethHost, containerIP string) error {
	// Check if script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("firewall script not found: %s", scriptPath)
	}

	fmt.Printf("\n  Applying firewall rules from: %s\n", scriptPath)
	fmt.Printf("    veth: %s, container IP: %s\n\n", vethHost, containerIP)

	// Execute script with parameters
	cmd := exec.Command(scriptPath, vethHost, containerIP)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("firewall script failed: %v", err)
	}

	fmt.Println("Firewall rules applied successfully")
	return nil
}

// removes veth interface (host side cleanup)
func CleanupVeth(vethHost string) error {
	// Check if interface exists
	if _, err := os.Stat(filepath.Join("/sys/class/net", vethHost)); os.IsNotExist(err) {
		return nil // Already cleaned up
	}

	cmd := exec.Command("ip", "link", "delete", vethHost)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete veth %s: %v", vethHost, err)
	}

	fmt.Printf("Cleaned up veth interface: %s\n", vethHost)
	return nil
}
