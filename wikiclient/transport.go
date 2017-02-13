// Copyright (c) 2017, Mitchell Cooper
package wikiclient

// used outside of transport
type Transport interface {
	Errors() chan error             // error channel
	readMessages() chan Message     // messages read channel
	writeMessage(msg Message) error // write a message
	Connect() error                 // connect to wikiserver
	Dead() bool                     // true if not connected
}

// base for all transports
type transport struct {
	errors    chan error
	readChan  chan Message
	writeChan chan Message
	connected bool
}

// create transport base
func createTransport() *transport {
	return &transport{
		make(chan error),
		make(chan Message),
		make(chan Message),
		false,
	}
}

// send an error to the erros chan and mark the transport as dead
func (tr *transport) criticalError(err error) {
	tr.errors <- err
	tr.connected = false
}

func (tr *transport) readMessages() chan Message {
	return tr.readChan
}

func (tr *transport) writeMessage(msg Message) error {
	tr.writeChan <- msg
	return nil
}

func (tr *transport) Errors() chan error {
	return tr.errors
}

func (tr *transport) Dead() bool {
	return !tr.connected
}
