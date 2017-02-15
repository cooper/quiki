// Copyright (c) 2017, Mitchell Cooper
package wikiclient

import (
	"errors"
	"strconv"
	"time"
)

// a Client is formed by pairing a transport with a session
type Client struct {
	Transport Transport     // wikiclient transport
	Session   *Session      // wikiclient session
	Timeout   time.Duration // how long to waits on requests
}

// create a client and clean the session if necessary
func NewClient(tr Transport, sess *Session, timeout time.Duration) Client {
	sess.Clean(tr)
	return Client{tr, sess, timeout}
}

// display a page
func (c Client) DisplayPage(pageName string) (Message, error) {
	return c.Request("page", map[string]interface{}{"name": pageName})
}

// display an image
func (c Client) DisplayImage(imageName string, width, height int) (Message, error) {
	return c.Request("image", map[string]interface{}{
		"name":   imageName,
		"width":  strconv.Itoa(width),
		"height": strconv.Itoa(height),
	})
}

// display category posts
func (c Client) DisplayCategoryPosts(categoryName string, pageN int) (Message, error) {
	if pageN <= 0 {
		pageN = 1
	}
	return c.Request("cat_posts", map[string]interface{}{
		"name":   categoryName,
		"page_n": string(pageN),
	})
}

func (c Client) Request(command string, args messageArgs) (Message, error) {
	return c.RequestMessage(NewMessage(command, args))
}

// send a message and block until we get its response
func (c Client) RequestMessage(req Message) (res Message, err error) {

	// the transport is not authenticated
	if !c.Session.ReadAccess {
		err = c.sendMessage(NewMessage("wiki", map[string]interface{}{
			"name":     c.Session.WikiName,
			"password": c.Session.WikiPassword,
		}))
		if err != nil {
			return
		}
		c.Session.ReadAccess = true
	}

	// TODO: if the transport is not write authenticated and we have
	// credentials in the session, send them now

	// send
	if err = c.sendMessage(req); err != nil {
		return
	}

	// await the response, or give up after the timeout
	select {
	case res = <-c.Transport.readMessages():

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
func (c Client) sendMessage(msg Message) error {

	// the transport is dead!
	if c.Transport.Dead() {
		return errors.New("transport is dead")
	}

	return c.Transport.writeMessage(msg)
}
