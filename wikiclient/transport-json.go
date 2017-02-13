// Copyright (c) 2017, Mitchell Cooper
package wikiclient

import (
	"bufio"
	"errors"
	"io"
)

type jsonTransport struct {
	*transport
	incoming chan []byte
	writer   io.Writer
	reader   *bufio.Reader
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

func (tr *jsonTransport) mainLoop() {
	for {
		select {

		// outgoing messages
		case msg := <-tr.writeChan:
			data := append(msg.ToJson(), '\n')
			if _, err := tr.writer.Write(data); err != nil {
				tr.errors <- err
			}

		// incoming json data
		case json := <-tr.incoming:
			msg, err := MessageFromJson(json)
			if err != nil {
				tr.errors <- errors.New("error creating message: " + err.Error())
				break
			}
			tr.readChan <- msg
		}
	}
}
