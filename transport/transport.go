// Copyright (c) 2017, Mitchell Cooper
package transport

import (
	"errors"
	"github.com/cooper/quiki/config"
)

var conf *config.Config

type Transport interface {
}

func Connect() (Transport, error) {
	conf = config.Conf
	sockType := conf.Get("server.socket.type")
	switch sockType {

	// unix socket, this is default
	case "unix", "":
		path, err := conf.Require("server.socket.path")
		if err != nil {
			return nil, err
		}
		return connectUnix(path)
	}

	// not sure
	return nil, errors.New("unsupported @server.socket.type: " + sockType)
}
