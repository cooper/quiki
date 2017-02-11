// Copyright (c) 2017, Mitchell Cooper
package transport

import "net"
import "bufio"

type unixTransport struct {
	*jsonTransport
	path string
	conn net.Conn
}

// create
func createUnix() (*unixTransport, error) {
	path, err := conf.Require("server.socket.path")
	if err != nil {
		return nil, err
	}
	return &unixTransport{
		createJson(),
		path,
		nil,
	}, nil
}

// connect
func (unixTr *unixTransport) Connect() error {
	// TODO: check if already connected
	conn, err := net.Dial("unix", unixTr.path)
	if err != nil {
		return err
	}
	unixTr.conn = conn
	unixTr.rw = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	return nil
}
