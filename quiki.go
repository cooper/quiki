// Copyright (c) 2017, Mitchell Cooper
package main

import (
	"github.com/cooper/quiki/config"
	"github.com/cooper/quiki/transport"
	"log"
	"net/http"
	"os"
)

func main() {

	// find config file
	if len(os.Args) < 2 || os.Args[1] == "" {
		log.Fatal("usage: " + os.Args[0] +
			" /path/to/wikiserver.conf")
	}

	// parse configuration
	conf := config.New(os.Args[1])
	config.Conf = conf
	if err := conf.Parse(); err != nil {
		log.Fatal(err)
	}

	// port not configured
	port := conf.Get("quiki.http.port")
	if port == "" {
		log.Fatal("@quiki.http.port is required")
	}

	// setup the transport
	if err := transport.Connect(); err != nil {
		log.Fatal("can't connect to transport: " + err.Error())
	}
	log.Println("connected to wikiserver")

	// listen
	log.Fatal(http.ListenAndServe(conf.Get("quiki.http.bind")+":"+port, nil))
}
