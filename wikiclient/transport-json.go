// Copyright (c) 2017, Mitchell Cooper
package wikiclient

// jsonTransport is a base type for all JSON stream transports.
// with this type of transport, each message is represented by a JSON
// array and is terminated with a newline. the message format is
//
// [ command, arguments, ID ]
//
// where command is the string message type, arguments is a JSON object of
// message parameters, and ID is an optional integer message identifier which
// will reappear in the message response.
//

import (
	"bufio"
	"errors"
	"io"
)

type jsonTransport struct {
	*transport

	// a channel of incoming line data waiting to be parsed
	incoming chan []byte

	// messages written to the transport with WriteMessage() or 'write' channel
	// will be translated to JSON by the jsonTransport main loop. the resultant
	// JSON will be written to this writer.
	writer io.Writer

	// the jsonTransport will read data from this buffer line-by-line, parsing
	// it as JSON and creating wikiclient messages. the created messages will
	// be sent to the 'read' channel.
	reader *bufio.Reader
}

// create json transport base
func createJson() *jsonTransport {
	return &jsonTransport{
		createTransport(),
		make(chan []byte),
		nil,
		nil,
	}
}

// start the loop
func (tr *jsonTransport) startLoops() {
	go tr.readLoop()
	go tr.mainLoop()
}

// read data loop
func (tr *jsonTransport) readLoop() {
	for {

		// not ready
		if tr.reader == nil {
			tr.errors <- errors.New("reader is not available")
			return
		}

		// read a full line
		data, err := tr.reader.ReadBytes('\n')

		// some error occurred
		if err != nil {
			tr.errors <- err
			return
		}

		tr.incoming <- data
	}
}

// main loop
func (tr *jsonTransport) mainLoop() {
	for {
		select {

		// outgoing messages
		case msg := <-tr.write:
			data := append(msg.ToJson(), '\n')
			if _, err := tr.writer.Write(data); err != nil {
				tr.errors <- err
			}

		// incoming json data
		case json := <-tr.incoming:
			msg, err := MessageFromJson(json)
			if err != nil {
				tr.errors <- errors.New("error creating message from JSON: " + err.Error())
				break
			}
			tr.read <- msg
		}
	}
}
