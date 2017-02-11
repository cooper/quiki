// Copyright (c) 2017, Mitchell Cooper
package transport

import (
	"errors"
	"github.com/cooper/quiki/config"
	"github.com/cooper/quiki/wikiclient"
)

var conf *config.Config

// used outside of transport
type Transport interface {
	WriteMessage(msg wikiclient.Message) error
	Connect() error // connect to wikiserver
}

// base for all transports
type transport struct {
	readMessages  chan wikiclient.Message
	writeMessages chan wikiclient.Message
}

// create transport base
func createTransport() *transport {
	return &transport{
		make(chan wikiclient.Message),
		make(chan wikiclient.Message),
	}
}

func (tr *transport) WriteMessage(msg wikiclient.Message) error {
	tr.writeMessages <- msg
	return nil
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
