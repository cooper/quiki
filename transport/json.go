// Copyright (c) 2017, Mitchell Cooper
package transport

import (
	"bufio"
	"github.com/cooper/quiki/wikiclient"
	"io"
	"log"
	"time"
)

type jsonTransport struct {
	*transport
	incoming chan []byte
	err      chan error
	writer   io.Writer
	reader   *bufio.Reader
}

// create json transport base
func createJson() *jsonTransport {
	return &jsonTransport{
		createTransport(),
		make(chan []byte),
		make(chan error),
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
		log.Println("readLoop")

		// not ready
		if jsonTr.reader == nil {
			log.Println("reader not ready")
			time.Sleep(5 * time.Second)
			continue
		}

		// read a full line
		data, err := jsonTr.reader.ReadBytes('\n')

		// some error occurred
		if err != nil {
			jsonTr.err <- err
			break
		}

		jsonTr.incoming <- data
	}
}

func (jsonTr *jsonTransport) mainLoop() {
	for {
		select {

        // read error
        case err := <-jsonTr.err:
            log.Println("error reading!", err)
            go jsonTr.readLoop()

        // outgoing messages
		case msg := <-jsonTr.writeMessages:
			log.Println("found a message to write:", msg)
			data := append(msg.ToJson(), '\n')
			if _, err := jsonTr.writer.Write(data); err != nil {
				log.Println("error writing!", err)
			}

        // incoming json data
		case json := <-jsonTr.incoming:
			log.Println("found some data to handle:", string(json))
			msg, err := wikiclient.MessageFromJson(json)
			if err != nil {
				log.Println("error creating message:", err)
				continue
			}
			jsonTr.readMessages <- msg
		}
	}
}
