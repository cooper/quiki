// Copyright (c) 2017, Mitchell Cooper
package transport

import "bufio"
import "time"
import "log"

type jsonTransport struct {
	*transport
	incoming chan []byte
	err      chan error
	rw       *bufio.ReadWriter
}

// create json transport base
func createJson() *jsonTransport {
	return &jsonTransport{
		createTransport(),
		make(chan []byte),
		make(chan error),
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
		if jsonTr.rw == nil {
			log.Println("not ready")
			time.Sleep(5 * time.Second)
			continue
		}

		// read a full line
		data, err := jsonTr.rw.ReadBytes('\n')

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
		case msg := <-jsonTr.writeMessages:
			log.Println("found a message to write:", msg)
			data := append(wikiclientMessageToJson(msg), '\n')
            jsonTr.rw.WriteString("does this even work?\n")
			if _, err := jsonTr.rw.Write(data); err != nil {
				log.Println("error writing! ", err)
			}
		case json := <-jsonTr.incoming:
			log.Println("found some data to handle:", string(json))
			msg := jsonToWikiclientMessage(json)
			jsonTr.readMessages <- msg
		}
	}
}

func wikiclientMessageToJson(msg wikiclientMessage) []byte {
	return []byte{byte(msg)}
}

func jsonToWikiclientMessage(json []byte) wikiclientMessage {
	return 0
}
