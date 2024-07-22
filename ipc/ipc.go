package ipc

import "github.com/gavintan/autoTyper/config"

var (
	Name       = config.Name
	PipeName   = `\\.\pipe\` + Name
	SocketName = Name + ".sock"
	SocketPath = "/tmp/" + SocketName

	PingPath = "/ping"
	ShowPath = "/window/show"
	QuitPath = "/window/quit"
)

type Response struct {
	Error string `json:"error"`
}
