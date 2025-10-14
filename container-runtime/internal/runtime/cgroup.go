package runtime

import (
    "fmt"
    "os"
    "path/filepath"
    "strconv"
)

func CreateCgroup(containerID string, pid int) error {
    uid := os.Getenv("SUDO_UID")
    if uid == "" {
        uid = fmt.Sprint(os.Getuid())
    }

    cgroupBase := filepath.Join("/sys/fs/cgroup/user.slice", fmt.Sprintf("user-%s.slice", uid))
    cgroupPath := filepath.Join(cgroupBase, containerID)

    if err := os.MkdirAll(cgroupPath, 0755); err != nil {
        return fmt.Errorf("create cgroup dir: %w", err)
    }

    // Attach process
    procs := filepath.Join(cgroupPath, "cgroup.procs")
    if err := os.WriteFile(procs, []byte(strconv.Itoa(pid)), 0644); err != nil {
        return fmt.Errorf("add pid to cgroup: %w", err)
    }

    // Optionally, set resource limits:
    // _ = os.WriteFile(filepath.Join(cgroupPath, "memory.max"), []byte("500M"), 0644)

    return nil
}
