package transport

import (
	"errors"
	"github.com/cooper/quiki/config"
)

var conf *config.Config
var transportFunctions = map[string]func() error{
	"unix": connectUnix,
}

func Connect() error {
	conf = config.Conf

	// find sock type, defaulting to unix
	sockType := conf.Get("server.socket.type")
	if sockType == "" {
		sockType = "unix"
	}

	// we don't know this sock type
	aFunc, ok := transportFunctions[sockType]
	if !ok {
		return errors.New("unsupported @server.socket.type: " + sockType)
	}

	// call the transport function
	return aFunc()
}
