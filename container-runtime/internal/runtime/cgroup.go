package runtime

import (
    "fmt"
    "os"
    "path/filepath"
    "bufio"
    "strconv"
    "strings"
)

func CreateCgroup(containerID string, pid int) error {
    parent_cgroup_path, err := getParentCgroupPath()
	if err != nil {
		panic(err)
	}
	fmt.Println("Current cgroup path:", parent_cgroup_path)
    uid := os.Getenv("SUDO_UID")
    if uid == "" {
        uid = fmt.Sprint(os.Getuid())
    }
	cgroup_path := parent_cgroup_path + "/" + containerID;
    if err := os.MkdirAll(cgroup_path, 0755); err != nil {
        return fmt.Errorf("create cgroup dir: %w", err)
    }

    // Attach process
    procs := filepath.Join(cgroup_path, "cgroup.procs")
    if err := os.WriteFile(procs, []byte(strconv.Itoa(pid)), 0644); err != nil {
        return fmt.Errorf("add pid to cgroup: %w", err)
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