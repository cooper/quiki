// Copyright (c) 2017, Mitchell Cooper
package main

import (
	"github.com/cooper/quiki/config"
	"log"
	"net/http"
	"os"
)

var conf *config.Config

func main() {

	// find config file
	if len(os.Args) < 2 || os.Args[1] == "" {
		log.Fatal("usage: " + os.Args[0] +
			" /path/to/wikiserver.conf")
	}

	// parse configuration
	conf = config.New(os.Args[1])
	if err := conf.Parse(); err != nil {
		log.Fatal(err)
	}

	// port is required
	port, err := conf.Require("quiki.http.port")
	if err != nil {
		log.Fatal(err)
	}

	// setup the transport
	if err := initTransport(); err != nil {
		log.Fatal("can't initialize transport: " + err.Error())
	}

	// listen
	log.Fatal(http.ListenAndServe(conf.Get("quiki.http.bind")+":"+port, nil))
}
