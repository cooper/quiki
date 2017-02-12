// Copyright (c) 2017, Mitchell Cooper
package wikiclient

// a client is formed by pairing a transport with a session
type Client struct {
	Transport Transport
	Session   *Session
}

func (c Client) sendMessage(msg Message) {

}
