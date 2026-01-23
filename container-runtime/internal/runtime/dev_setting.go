package runtime


import (
	"fmt"
	"os"
	"path/filepath"
	mySpecs "mrunc/pkg/specs"

	"golang.org/x/sys/unix"
)

type Device struct {
	Name  string
	Major uint32
	Minor uint32
	Mode  uint32
}

var devices = []Device{
	{"null",     1, 3, 0666},
	{"zero",     1, 5, 0666},
	{"full",     1, 7, 0666},
	{"random",   1, 8, 0666},
	{"urandom",  1, 9, 0666},
	{"tty",      5, 0, 0666},
}
func mknodChar(path string, major, minor, mode uint32) error {
	dev := unix.Mkdev(major, minor)
	return unix.Mknod(
		path,
		uint32(unix.S_IFCHR)|mode,
		int(dev),
	)
}

func SetupDev(config *mySpecs.ContainerConfig)  error {
	root := config.RootFS.Path + "/dev"

	if err := os.MkdirAll(root, 0755); err != nil {
		return err
	}

	// Avoid umask surprises
	unix.Umask(0)
	// create devices using mknod
	for _, d := range devices {
		path := filepath.Join(root, d.Name)

		// Remove existing node if present
		_ = os.Remove(path)

		if err := mknodChar(path, d.Major, d.Minor, d.Mode); err != nil {
			return fmt.Errorf("mknod %s failed: %v\n", path, err)
			continue
		}

		fmt.Printf("created %s (%d:%d)\n", path, d.Major, d.Minor)
	}


	links := map[string]string{
		root + "/dev/fd":     "/proc/self/fd",
		root + "/dev/stdin":  "/proc/self/fd/0",
		root + "/dev/stdout": "/proc/self/fd/1",
		root + "/dev/stderr": "/proc/self/fd/2",
	}
	for dst, src := range links {
		_ = os.Remove(dst) // ignore error
		if err := os.Symlink(src, dst); err != nil {
			return err
		}
	}
	return nil
}
func LinkPts(config *mySpecs.ContainerConfig) error {
	_ = os.Remove(config.RootFS.Path +"/dev/ptmx")
	return os.Symlink("pts/ptmx", config.RootFS.Path +"/dev/ptmx")
}


func BindConsole(slaveFd int) error {
	consolePath := "/dev/console"
	if err := os.MkdirAll(filepath.Dir(consolePath), 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(consolePath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	f.Close()

	src := fmt.Sprintf("/proc/self/fd/%d", slaveFd)

	// Bind-mount the fd to /dev/console
	if err := unix.Mount(
		src,
		consolePath,
		"",
		unix.MS_BIND,
		"",
	); err != nil {
		return err
	}

	return nil
}