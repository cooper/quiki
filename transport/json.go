// Copyright (c) 2017, Mitchell Cooper
package transport

import (
	"bufio"
	"errors"
	"github.com/cooper/quiki/wikiclient"
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
func (jsonTr *jsonTransport) startLoops() {
	go jsonTr.readLoop()
	go jsonTr.mainLoop()
}

func (jsonTr *jsonTransport) readLoop() {
	for {

		// not ready
		if jsonTr.reader == nil {
			jsonTr.errors <- errors.New("reader is not available")
			return
		}

		// read a full line
		data, err := jsonTr.reader.ReadBytes('\n')

		// some error occurred
		if err != nil {
			jsonTr.errors <- err
			return
		}

		jsonTr.incoming <- data
	}
}

func (jsonTr *jsonTransport) mainLoop() {
	for {
		select {

		// outgoing messages
		case msg := <-jsonTr.writeMessages:
			data := append(msg.ToJson(), '\n')
			if _, err := jsonTr.writer.Write(data); err != nil {
				jsonTr.errors <- err
			}

		// incoming json data
		case json := <-jsonTr.incoming:
			msg, err := wikiclient.MessageFromJson(json)
			if err != nil {
				jsonTr.errors <- errors.New("error creating message: " + err.Error())
				break
			}
			jsonTr.readMessages <- msg

		}
	}
}
