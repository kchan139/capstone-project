package runtime

import (
    "fmt"
    "os"
    "path/filepath"
    "bufio"
    "strconv"
	mySpecs "mrunc/pkg/specs"

    "strings"
)

func CreateCgroup(config *mySpecs.ContainerConfig, pid int) error {
    parent_cgroup_path, err := getParentCgroupPath()
	if err != nil {
		panic(err)
	}
	fmt.Println("Current cgroup path:", parent_cgroup_path)
	cgroup_path := parent_cgroup_path + "/" + config.ContainerId;
    if err := os.MkdirAll(cgroup_path, 0755); err != nil {
        return fmt.Errorf("create cgroup dir: %w", err)
    }

    // Attach process
    procs := filepath.Join(cgroup_path, "cgroup.procs")
    if err := os.WriteFile(procs, []byte(strconv.Itoa(pid)), 0644); err != nil {
        return fmt.Errorf("add pid to cgroup: %w", err)
    }
	if config.Linux.Resources != nil {
		// cpu controllers
		if config.Linux.Resources.CPU != nil {
			cpuCfg := config.Linux.Resources.CPU

			// cpu.shares -> cpu.weight
			if cpuCfg.Shares > 0 {
				weight := 1 + ((cpuCfg.Shares-2)*9999)/262142
				weightStr := strconv.FormatInt(weight, 10)
				err := os.WriteFile(filepath.Join(cgroup_path, "cpu.weight"), []byte(weightStr), 0644)
				if err != nil {
					return fmt.Errorf("set cpu.weight: %w", err)
				}
			}

			// cpu.quota + cpu.period -> cpu.max
			if cpuCfg.Quota > 0 && cpuCfg.Period > 0 {
				value := fmt.Sprintf("%d %d", cpuCfg.Quota, cpuCfg.Period)
				err := os.WriteFile(filepath.Join(cgroup_path, "cpu.max"), []byte(value), 0644)
				if err != nil {
					return fmt.Errorf("set cpu.max: %w", err)
				}
			}
		}
		// memory controller
		if config.Linux.Resources.Memory != nil {
			mem := config.Linux.Resources.Memory
			if mem.Limit > 0 {
				_ = os.WriteFile(filepath.Join(cgroup_path, "memory.max"), []byte(strconv.FormatInt(mem.Limit, 10)), 0644)
			}
			if mem.Reservation > 0 {
				_ = os.WriteFile(filepath.Join(cgroup_path, "memory.low"), []byte(strconv.FormatInt(mem.Reservation, 10)), 0644)
			}
			if mem.Swap > 0 {
				_ = os.WriteFile(filepath.Join(cgroup_path, "memory.swap.max"), []byte(strconv.FormatInt(mem.Swap, 10)), 0644)
			}
		}
		// Pid controller
		if config.Linux.Resources.Pids != nil {
			pidsCfg := config.Linux.Resources.Pids
			if pidsCfg.Limit > 0 {
				err := os.WriteFile(filepath.Join(cgroup_path, "pids.max"),
					[]byte(strconv.FormatInt(pidsCfg.Limit, 10)), 0644)
				if err != nil {
					return fmt.Errorf("set pids.max: %w", err)
				}
			}
		}
	}
	
    // Optionally, set resource limits:
    // _ = os.WriteFile(filepath.Join(cgroupPath, "memory.max"), []byte("500M"), 0644)

    return nil
}

func getParentCgroupPath() (string, error) {
	// 1. Read /proc/self/cgroup
	f, err := os.Open("/proc/self/cgroup")
	if err != nil {
		return "", err
	}
	defer f.Close()

	var relPath string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		// Format: 0::/user.slice/user-1001.slice/session-3467.scope
		parts := strings.SplitN(line, ":", 3)
		if len(parts) == 3 {
			relPath = parts[2]
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if relPath == "" {
		return "", fmt.Errorf("no cgroup v2 entry found")
	}

	// 2. Construct absolute path
	absPath := filepath.Join("/sys/fs/cgroup", relPath)

	// 3. Move one level up
	parent := filepath.Dir(absPath)

	return parent, nil
}