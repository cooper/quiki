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

type wikiInfo struct {
	name     string         // wiki name
	password string         // wiki password for read authentication
	confPath string         // path to wiki configuration
	conf     *config.Config // wiki config instance
}

var conf *config.Config
var wikis map[string]wikiInfo

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

	// parse configuration for each wiki
	wikis := conf.GetMap("server.wiki")
	for wikiName := range wikis {

		// wiki configuration path
		wikiConfPath, err := conf.Require("server.wiki." + wikiName + ".config")
		if err != nil {
			log.Fatal(err)
		}

		// wiki password for read authentication
		wikiPassword, err := conf.Require("server.wiki." + wikiName + ".password")
		if err != nil {
			log.Fatal(err)
		}

		// parse the wiki configuration
		wikiConf := config.New(wikiConfPath)
		if err := wikiConf.Parse(); err != nil {
			log.Fatal(err)
		}

		// store the wiki info
		wikis[wikiName] = wikiInfo{
			name:     wikiName,
			password: wikiPassword,
			confPath: wikiConfPath,
			conf:     wikiConf,
		}
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

	// listen
	log.Fatal(http.ListenAndServe(conf.Get("quiki.http.bind")+":"+port, nil))
}
