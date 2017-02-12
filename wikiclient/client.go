// Copyright (c) 2017, Mitchell Cooper
package wikiclient

import (
	"errors"
	"time"
)

// a client is formed by pairing a transport with a session
type Client struct {
	Transport   Transport     // wikiclient transport
	Session     Session       // wikiclient session
	Timeout     time.Duration // how long to waits on requests
	ReadAccess  bool          // true if authenticated for reading
	WriteAccess bool          // true if authenticated for writing
}

// send a message and block until we get its response
func (c *Client) Request(req Message) (res Message, err error) {

	// the transport is not authenticated
	if !c.ReadAccess {
		err = c.sendMessage(NewMessage("wiki", map[string]interface{}{
			"name":     c.Session.WikiName,
			"password": c.Session.WikiPassword,
		}))
		if err != nil {
			return
		}
		c.ReadAccess = true
	}

	// TODO: if the transport is not write authenticated and we have
	// credentials in the session, send them now

	// send
	if err = c.sendMessage(req); err != nil {
		return
	}

	select {
	case res = <-c.Transport.ReadMessages():

		// this is the correct ID
		if res.ID == req.ID {
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

// send a message to the transport, but do not await reply
func (c *Client) sendMessage(msg Message) error {

	// the transport is dead!
	if c.Transport.Dead() {
		return errors.New("transport is dead")
	}

	return c.Transport.WriteMessage(msg)
}
