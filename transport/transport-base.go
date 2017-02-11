// Copyright (c) 2017, Mitchell Cooper
package transport

import (
	"github.com/cooper/quiki/wikiclient"
)

// base for all transports
type transport struct {
	errors        chan error
	readMessages  chan wikiclient.Message
	writeMessages chan wikiclient.Message
}

// create transport base
func createTransport() *transport {
	return &transport{
		make(chan error),
		make(chan wikiclient.Message),
		make(chan wikiclient.Message),
	}
}

func (tr *transport) WriteMessage(msg wikiclient.Message) error {
	tr.writeMessages <- msg
	return nil
}

func (tr *transport) Errors() chan error {
	return tr.errors
}

func (tr *transport) ReadMessages() chan wikiclient.Message {
	return tr.readMessages
}
