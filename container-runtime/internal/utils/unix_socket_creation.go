package utils

import (
	"os"
	"golang.org/x/sys/unix"
)
func SocketPair() (*os.File, *os.File, error) {
	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_STREAM, 0)
	if err != nil {
		return nil, nil, err
	}

	// Wrap raw FDs into *os.File
	parentFile := os.NewFile(uintptr(fds[0]), "parentSock")
	childFile := os.NewFile(uintptr(fds[1]), "childSock")

	// Convert parent side to *net.UnixConn
	// parentConn, err := net.FileConn(parentFile)
	if err != nil {
		return nil, nil, err
	}

	// Important: net.FileConn() duplicates the FD internally,
	// so you can close the original *os.File safely.
	// parentFile.Close()

	return parentFile, childFile, nil
}