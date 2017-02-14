// Copyright (c) 2017, Mitchell Cooper
// transport-loop.go - maintain a connection to the wikiserver
package main

import (
	"errors"
	"github.com/cooper/quiki/wikiclient"
	"log"
	"time"
)

// the transport instance
var tr wikiclient.Transport

// create the transport based on the configuration
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

// create and connect the transport
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

// loop for errors
func transportLoop() {
	if tr == nil {
		panic("no transport?")
	}
	err := <-tr.Errors()
	for err != nil {
		log.Println("transport error:", err)
		time.Sleep(5 * time.Second)
		err = initTransport()
	}
	return
}
