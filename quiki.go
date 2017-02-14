// Copyright (c) 2017, Mitchell Cooper
// quiki - a standalone web server for wikifier
package main

import (
	"github.com/cooper/quiki/config"
	"log"
	"net/http"
	"os"
)

// wikiserver config instance
var conf *config.Config

func main() {

	// find config file
	if len(os.Args) < 2 || os.Args[1] == "" {
		log.Fatal("usage: " + os.Args[0] + " /path/to/wikiserver.conf")
	}

	// parse configuration
	conf = config.New(os.Args[1])
	if err := conf.Parse(); err != nil {
		log.Fatal(err)
	}

	// port is required
	port, err := conf.Require("server.http.port")
	if err != nil {
		log.Fatal(err)
	}

	// set up wikis
	if err := initializeWikis(); err != nil {
		log.Fatal(err)
	}

	// setup the transport
	if err := initTransport(); err != nil {
		log.Fatal(err)
	}

	// listen
	log.Fatal(http.ListenAndServe(conf.Get("server.http.bind")+":"+port, nil))
}
