package runtime
import (
    "fmt"
    "syscall"
    "mrunc/pkg/specs"
)

// SetProcessUser sets the user and group for the container process
func SetProcessUser(user *specs.User) error {
    if user == nil {
        // No user specified - keep current user (usually root in namespace)
        fmt.Printf("DEBUG: No user specified, keeping current user (UID: %d, GID: %d)\n", 
            syscall.Getuid(), syscall.Getgid())
        return nil
    }

    fmt.Printf("DEBUG: Setting process user to UID: %d, GID: %d\n", user.UID, user.GID)

    // Set additional groups first (if any)
    if len(user.AdditionalGids) > 0 {
        fmt.Printf("DEBUG: Setting additional groups: %v\n", user.AdditionalGids)
        if err := syscall.Setgroups(user.AdditionalGids); err != nil {
            return fmt.Errorf("failed to set additional groups: %v", err)
        }
    }

    // Set primary group BEFORE setting user (this is important!)
    if err := syscall.Setgid(user.GID); err != nil {
        return fmt.Errorf("failed to set GID %d: %v", user.GID, err)
    }
    fmt.Printf("DEBUG: GID set to %d\n", user.GID)

    // Set user ID
    if err := syscall.Setuid(user.UID); err != nil {
        return fmt.Errorf("failed to set UID %d: %v", user.UID, err)
    }
    fmt.Printf("DEBUG: UID set to %d\n", user.UID)

    // Verify the change
    currentUID := syscall.Getuid()
    currentGID := syscall.Getgid()
    fmt.Printf("DEBUG: Process now running as UID: %d, GID: %d\n", currentUID, currentGID)

    return nil
}

// ValidateUser checks if the user configuration is valid
func ValidateUser(user *specs.User) error {
    if user == nil {
        return nil // User field is optional
    }

    // Check for valid UID range (0-65535 is common)
    if user.UID < 0 || user.UID > 65535 {
        return fmt.Errorf("invalid UID: %d (must be 0-65535)", user.UID)
    }

    // Check for valid GID range
    if user.GID < 0 || user.GID > 65535 {
        return fmt.Errorf("invalid GID: %d (must be 0-65535)", user.GID)
    }

    // Check additional GIDs
    for i, gid := range user.AdditionalGids {
        if gid < 0 || gid > 65535 {
            return fmt.Errorf("invalid additional GID at index %d: %d (must be 0-65535)", i, gid)
        }
    }

    return nil
}