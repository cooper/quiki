// Copyright (c) 2017, Mitchell Cooper
package wikiclient

// UnixTransport is a unix socket transport.
// it is based on the JSON stream transport.

import (
	"bufio"
	"net"
)

type UnixTransport struct {
	*jsonTransport
	path string
	conn net.Conn
}

// create
func NewUnixTransport(path string) *UnixTransport {
	return &UnixTransport{
		createJson(),
		path,
		nil,
	}
}

// connect
func (tr *UnixTransport) Connect() error {
	// TODO: check if already connected
	conn, err := net.Dial("unix", tr.path)
	if err != nil {
		return err
	}
	tr.connected = true
	tr.conn = conn
	tr.reader = bufio.NewReader(conn)
	tr.writer = conn
	tr.startLoops()
	return nil
}
