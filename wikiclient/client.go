// Copyright (c) 2017, Mitchell Cooper
package wikiclient

import (
	"errors"
	"time"
)

// a client is formed by pairing a transport with a session
type Client struct {
	Transport Transport
	Session   Session
	Timeout   time.Duration
}

func (c Client) SendMessage(msg Message) error {

	// the transport is dead!
	if c.Transport.Dead() {
		return errors.New("transport is dead")
	}

	return c.Transport.WriteMessage(msg)
}

// send a message and block until we get its response
func (c Client) Request(req Message) (msg Message, err error) {

	// send
	if err = c.SendMessage(msg); err != nil {
		return
	}

	select {
	case res := <-c.Transport.ReadMessages():

		// this is the correct ID
		if res.ID == req.ID {
			msg = res
			return
		}

		// some other message
		err = errors.New("Got response with incorrect message ID")
		return

	case <-time.After(c.Timeout):
		err = errors.New("Timed out")
		return
	}
}
