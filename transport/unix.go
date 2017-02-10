// Copyright (c) 2017, Mitchell Cooper
package transport

import "net"

type unixTransport struct {
	conn net.Conn
}

func connectUnix(path string) (*unixTransport, error) {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return nil, err
	}
	transport := &unixTransport{conn}
	return transport, nil
}
