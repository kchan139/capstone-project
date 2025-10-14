package runtime

import (
    "fmt"
    "github.com/coreos/go-systemd/v22/dbus"
   godbus "github.com/godbus/dbus/v5"
)

// CreateScope creates a transient systemd scope for a process
func CreateScope(containerID string, pid int) error {
    conn, err := dbus.NewSystemConnection()
    if err != nil {
        return fmt.Errorf("connect to systemd: %w", err)
    }
    defer conn.Close()

    // Use direct dbus.Property{ Name, Value } instead of dbus.PropBool, etc.
    props := []dbus.Property{
		 dbus.PropDescription("mrunc container " + containerID),
        dbus.PropSlice("user.slice"),
        dbus.Property{
            Name:  "Delegate",
            Value: godbus.MakeVariant(true),
        },
        dbus.Property{
            Name:  "CPUAccounting",
            Value: godbus.MakeVariant(true),
        },
        dbus.Property{
            Name:  "MemoryAccounting",
            Value: godbus.MakeVariant(true),
        },
        dbus.Property{
            Name:  "PIDs",
            Value: godbus.MakeVariant([]uint32{uint32(pid)}),
        },
    }

    ch := make(chan string)
    unitName := fmt.Sprintf("mrunc-%s.scope", containerID)

    _, err = conn.StartTransientUnit(unitName, "replace", props, ch)
    if err != nil {
        return fmt.Errorf("create scope: %w", err)
    }

    // wait for the job result
    <-ch
    return nil
}
