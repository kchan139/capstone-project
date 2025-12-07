package runtime

import (
	"fmt"
	"log"
	"strings"

	mySpecs "mrunc/pkg/specs"
	"os/exec"
	"github.com/moby/sys/capability"
	"golang.org/x/sys/unix"
)
// to test
func PrintCaps() error {
	cmd := exec.Command("bash", "-c",
		"grep Cap /proc/self/status | while read line; do echo $line; capsh --decode=$(echo $line | awk '{print $2}'); done")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get capabilities: %v", err)
	}

	fmt.Println(string(output))
	return nil
}
func SetupCaps(config *mySpecs.ContainerConfig) error {

	if config.Process.Capabilities == nil {
		log.Println("WARNING: No capabilities specified - process will have full root capabilities")
		return nil
	}

	// Get capability handle for the target process
	caps, err := capability.NewPid2(0)
	if err != nil {
		return fmt.Errorf("failed to get capabilities for child: %v", err)
	}

	// Load current capabilities
	if err := caps.Load(); err != nil {
		return fmt.Errorf("failed to load capabilities: %v", err)
	}

	// Clear all capabilities first
	caps.Clear(capability.CAPS | capability.BOUNDS | capability.AMBS)

	capConfig := config.Process.Capabilities

	// Set bounding capabilities
	if capConfig.Bounding != nil {
		for _, capStr := range capConfig.Bounding {
			cap, err := parseCapability(capStr)
			if err != nil {
				return fmt.Errorf("invalid bounding capability %s: %v", capStr, err)
			}
			caps.Set(capability.BOUNDING, cap)
		}
	}

	// Set permitted capabilities
	if capConfig.Permitted != nil {
		for _, capStr := range capConfig.Permitted {
			cap, err := parseCapability(capStr)
			if err != nil {
				return fmt.Errorf("invalid permitted capability %s: %v", capStr, err)
			}
			caps.Set(capability.PERMITTED, cap)
		}
	}

	// Set effective capabilities
	if capConfig.Effective != nil {
		for _, capStr := range capConfig.Effective {
			cap, err := parseCapability(capStr)
			if err != nil {
				return fmt.Errorf("invalid effective capability %s: %v", capStr, err)
			}
			caps.Set(capability.EFFECTIVE, cap)
		}
	}

	// Set inheritable capabilities
	if capConfig.Inheritable != nil {
		for _, capStr := range capConfig.Inheritable {
			cap, err := parseCapability(capStr)
			if err != nil {
				return fmt.Errorf("invalid inheritable capability %s: %v", capStr, err)
			}
			caps.Set(capability.INHERITABLE, cap)
		}
	}

	// Set ambient capabilities (will be applied with Apply())
	if capConfig.Ambient != nil {
		for _, capStr := range capConfig.Ambient {
			cap, err := parseCapability(capStr)
			if err != nil {
				return fmt.Errorf("invalid ambient capability %s: %v", capStr, err)
			}
			caps.Set(capability.AMBIENT, cap)
		}
	}

	// Apply all capability changes at once
	if err := caps.Apply(capability.CAPS | capability.BOUNDS | capability.AMBS); err != nil {
		return fmt.Errorf("failed to apply capabilities: %v", err)
	}

	// TODO:: needing to implement the nonewprivilege field
	// must specify no new privilege, otherwise it may become like root
	if config.Process.NoNewPrivileges {
		if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
			return fmt.Errorf("failed to set no_new_privs: %v", err)
		}
	}

	return nil
}

func parseCapability(capStr string) (capability.Cap, error) {
	// Remove "CAP_" prefix if present
	capName := strings.TrimPrefix(capStr, "CAP_")

	// Convert to the format expected by the capability package
	// The capability package uses lowercase with underscores
	capName = strings.ToLower(capName)

	// Try to find the capability by name
	for _, cap := range capability.ListKnown() {
		if cap.String() == capName {
			return cap, nil
		}
	}

	return capability.Cap(0), fmt.Errorf("unknown capability: %s (looking for '%s')", capStr, capName)
}