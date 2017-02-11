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
	Errors() chan error
	ReadMessages() chan wikiclient.Message
	WriteMessage(msg wikiclient.Message) error
	Connect() error // connect to wikiserver
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
