// Copyright (c) 2017, Mitchell Cooper
package wikiclient

var transportID uint

// used outside of transport
type Transport interface {
	Errors() <-chan error           // error channel
	readMessages() <-chan Message   // messages read channel
	writeMessage(msg Message) error // write a message
	Connect() error                 // connect to wikiserver
	Dead() bool                     // true if not connected
	ID() uint                       // transport identifier
}

// base for all transports
type transport struct {
	errors    chan error   // transport errors
	read      chan Message // read messages waiting to be processed
	write     chan Message // messages waiting to be written
	connected bool         // transport is active
	id        uint
}

// create transport base
func createTransport() *transport {
	transportID++
	return &transport{
		make(chan error),
		make(chan Message),
		make(chan Message),
		false,
		transportID,
	}
}

// send an error to the erros chan and mark the transport as dead
func (tr *transport) criticalError(err error) {
	tr.errors <- err
	tr.connected = false
}

// returns a channel of read messages waiting to be handled
func (tr *transport) readMessages() <-chan Message {
	return tr.read
}

// adds a message to the write buffer
func (tr *transport) writeMessage(msg Message) error {
	tr.write <- msg
	return nil
}

// returns a channel of read/write errors
func (tr *transport) Errors() <-chan error {
	return tr.errors
}

// true if the transport is not connected
func (tr *transport) Dead() bool {
	return !tr.connected
}

// transport identifier
func (tr *transport) ID() uint {
	return tr.id
}
