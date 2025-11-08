package runtime

// pty_linux.go
import (
	"fmt"
	"os"

	"github.com/containerd/console"
)

type HostConsole struct {
	Host console.Console
}

func SetupHostConsole() (*HostConsole, func(), error) {
	h := console.Current()
	if h == nil {
		return nil, nil, fmt.Errorf("not a tty")
	}
	if err := h.SetRaw(); err != nil {
		return nil, nil, fmt.Errorf("set raw: %w", err)
	}

	// cleanup resets the terminal
	cleanup := func() {
		_ = h.Reset()
	}

	return &HostConsole{Host: h}, cleanup, nil
}

type PtyFiles struct {
	Master        *os.File
	SlavePath     string
	SlaveFile     *os.File
	MasterConsole console.Console
}

func SetupPty() (*PtyFiles, func(), error) {
	master, slavePath, err := console.NewPty()
	if err != nil {
		return nil, nil, fmt.Errorf("new pty: %w", err)
	}

	masterFile := os.NewFile(master.Fd(), "pty-master")

	slaveFile, err := os.OpenFile(slavePath, os.O_RDWR, 0)
	if err != nil {
		master.Close()
		return nil, nil, fmt.Errorf("open slave: %w", err)
	}

	// cleanup closes what we opened
	cleanup := func() {
		_ = master.Close()
		_ = slaveFile.Close()
	}

	return &PtyFiles{
		Master:        masterFile,
		SlavePath:     slavePath,
		SlaveFile:     slaveFile,
		MasterConsole: master,
	}, cleanup, nil
}

// startWinchForwarder wires terminal-resize events from the host console
// to the container's pty. It sends the initial size once, then listens
// on sigCh in a goroutine.
// Call the returned cleanup to stop it (or just let the program exit).
func StartWinchForwarder(host console.Console, master console.Console, sigCh <-chan os.Signal) func() {
	// 1) send initial size
	if sz, err := host.Size(); err == nil {
		_ = master.Resize(sz)
	}

	// 2) forward changes
	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-sigCh:
				if sz, err := host.Size(); err == nil {
					_ = master.Resize(sz)
				}
			case <-done:
				return
			}
		}
	}()

	// return a cleanup
	return func() {
		close(done)
	}
}
