//go:build !windows

package ipc

import (
	"fmt"
	"net"
	"os"
	"os/user"
	"path"
	"runtime"
)

func init() {
	if runtime.GOOS == "darwin" {
		if home, err := os.UserHomeDir(); err == nil {
			p := path.Join(home, "Library", "Caches", Name)
			os.MkdirAll(p, 0755)
			SocketPath = path.Join(p, SocketName)
		} else if user, err := user.Current(); err == nil {
			SocketPath = fmt.Sprintf("/tmp/%s-%s.sock", Name, user.Uid)
		}
	} else {
		if runtime := os.Getenv("XDG_RUNTIME_DIR"); runtime != "" {
			os.MkdirAll(runtime, 0755)
			SocketPath = path.Join(runtime, Name+".sock")
		} else if user, err := user.Current(); err == nil {
			SocketPath = fmt.Sprintf("/tmp/%s-%s.sock", Name, user.Uid)
		}
	}
}

func Dial() (net.Conn, error) {
	return net.Dial("unix", SocketPath)
}

func Listen() (net.Listener, error) {
	return net.Listen("unix", SocketPath)
}

func DestroyConn() error {
	return os.Remove(SocketPath)
}
