// Copyright (c) 2017, Mitchell Cooper
package wikiclient

// used outside of transport
type Transport interface {
	Errors() chan error
	ReadMessages() chan Message
	WriteMessage(msg Message) error
	Connect() error // connect to wikiserver
}

// base for all transports
type transport struct {
	errors        chan error
	readMessages  chan Message
	writeMessages chan Message
}

// create transport base
func createTransport() *transport {
	return &transport{
		make(chan error),
		make(chan Message),
		make(chan Message),
	}
}

func (tr *transport) WriteMessage(msg Message) error {
	tr.writeMessages <- msg
	return nil
}

func (tr *transport) Errors() chan error {
	return tr.errors
}

func (tr *transport) ReadMessages() chan Message {
	return tr.readMessages
}
