// Copyright (c) 2017, Mitchell Cooper
package transport

import (
	"errors"
	"github.com/cooper/quiki/config"
	// "log"
)

var conf *config.Config

type wikiclientMessage int // will change

// used outside of transport
type Transport interface {
	StartLoop()
	Connect() error // connect to wikiserver
}

// base for all transports
type transport struct {
	readMessages  chan wikiclientMessage
	writeMessages chan wikiclientMessage
}

// create transport base
func createTransport() *transport {
	return &transport{
		make(chan wikiclientMessage),
		make(chan wikiclientMessage),
	}
}

func New() (Transport, error) {
	conf = config.Conf
	sockType := conf.Get("server.socket.type")
	switch sockType {

	// unix socket, this is default
	case "unix", "":
		return createUnix()
	}

	// not sure
	return nil, errors.New("unsupported @server.socket.type: " + sockType)
}
