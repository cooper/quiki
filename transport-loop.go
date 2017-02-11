// Copyright (c) 2017, Mitchell Cooper
package main

import (
	"errors"
	"github.com/cooper/quiki/transport"
	"github.com/cooper/quiki/wikiclient"
	"log"
	"time"
)

var tr transport.Transport

func initTransport() (err error) {

	// setup the transport
	tr, err = transport.New()
	if err != nil {
		return
	}

	// connect
	if err = tr.Connect(); err != nil {
		err = errors.New("can't connect to transport: " + err.Error())
		return
	}

	log.Println("connected to wikifier")
	tr.WriteMessage(wikiclient.NewMessage("wiki", map[string]interface{}{
		"name":     "notroll",
		"password": "hi",
	}))

	// start the loop
	go transportLoop()
	return
}

func transportLoop() {
	for {
		if tr == nil {
			panic("no transport?")
		}
		select {

		// some error occured. let's reinitialize the transport
		case err := <-tr.Errors():
			log.Println("transport error:", err)
			tr = nil
			time.Sleep(5 * time.Second)
			initTransport()

		case msg := <-tr.ReadMessages():
			log.Println(msg)
		}
	}
}
