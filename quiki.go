// Copyright (c) 2017, Mitchell Cooper
package main

import (
	"fmt"
	"github.com/cooper/quiki/config"
	"github.com/cooper/quiki/wikiclient"
	"log"
	"net/http"
	"os"
	"time"
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

	// set up wikis
	if err := initializeWikis(); err != nil {
		log.Fatal(err)
	}

	// setup the transport
	if err := initTransport(); err != nil {
		log.Fatal(err)
	}

	sess := &wikiclient.Session{WikiName: "notroll", WikiPassword: "hi"}
	http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		c := wikiclient.Client{
			Transport: tr,
			Session:   sess,
			Timeout:   3 * time.Second,
		}
		res, err := c.Request(wikiclient.NewMessage("page", map[string]interface{}{
			"name": "hi.page",
		}))
		if err != nil {
			fmt.Fprint(w, "some error happended: ", err)
			return
		}
		fmt.Fprint(w, res)
	})
	http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
	})

	// listen
	log.Fatal(http.ListenAndServe(conf.Get("quiki.http.bind")+":"+port, nil))
}
