package socket

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

// GetListener returns an active net.Listener.
// If running under systemd socket activation (or WASI pre-opened fd 3),
// it reconstructs the listener from file descriptor 3.
// Otherwise, it binds directly to the specified TCP port.
func GetListener(port string) (net.Listener, error) {
	// Guard: only inspect fd 3 if systemd explicitly passed LISTEN_FDS
	// to avoid hijacking internal runtime kqueue/epoll descriptors on host OS.
	if listenFds := os.Getenv("LISTEN_FDS"); listenFds != "" {
		if count, err := strconv.Atoi(listenFds); err == nil && count > 0 {
			file := os.NewFile(3, "systemd-socket")
			if file != nil {
				if l, err := net.FileListener(file); err == nil {
					return l, nil
				}
			}
		}
	}

	return net.Listen("tcp", fmt.Sprintf(":%s", port))
}
