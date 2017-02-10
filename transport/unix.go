// Copyright (c) 2017, Mitchell Cooper
package transport

import "net"

func connectUnix() error {

	// get sock path
	path, err := conf.Require("server.socket.path")
	if err != nil {
		return err
	}

	// connect
	conn, err := net.Dial("unix", path)
	if err != nil {
		return err
	}

	conn.Close()
	return nil
}
