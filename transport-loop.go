// Copyright (c) 2017, Mitchell Cooper
package main

import (
	"errors"
	"github.com/cooper/quiki/wikiclient"
	"log"
	"time"
)

var tr wikiclient.Transport
var transportDead bool

func newTransport() (wikiclient.Transport, error) {
	sockType := conf.Get("server.socket.type")
	switch sockType {

	// unix socket, this is default
	case "unix", "":
		path, err := conf.Require("server.socket.path")
		if err != nil {
			return nil, err
		}
		return wikiclient.NewUnixTransport(path), nil
	}

	// not sure
	return nil, errors.New("unsupported @server.socket.type: " + sockType)
}

func initTransport() (err error) {

	// setup the transport
	tr, err = newTransport()
	if err != nil {
		return
	}

	// connect
	if err = tr.Connect(); err != nil {
		err = errors.New("can't connect transport: " + err.Error())
		return
	}

	log.Println("connected to wikifier")

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
			for err != nil {
				transportDead = true
				log.Println("transport error:", err)
				time.Sleep(5 * time.Second)
				err = initTransport()
			}
			transportDead = false
			return

		case msg := <-tr.ReadMessages():
			log.Println(msg)
		}
	}
}
