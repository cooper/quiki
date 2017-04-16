// Copyright (c) 2017, Mitchell Cooper
// quiki - a standalone web server for wikifier
package main

import (
	"errors"
	"fmt"
	"github.com/cooper/quiki/config"
	"log"
	"net/http"
	"os"
)

// wikiserver config instance
var conf *config.Config

// wikifier directory path
var wikifierPath string

func main() {

	// find config file
	if len(os.Args) < 2 || os.Args[1] == "" {
		log.Fatal("usage: " + os.Args[0] + " /path/to/quiki.conf")
	}

	// parse configuration
	conf = config.New(os.Args[1])
	if err := conf.Parse(); err != nil {
		log.Fatal(err)
	}

	// these are required
	var port string
	if err := conf.RequireMany(map[string]*string{
		"server.http.port":    &port,
		"server.dir.wikifier": &wikifierPath,
	}); err != nil {
		log.Fatal(err)
	}

	// template search directories may have been specified
	templateDirs = conf.Get("server.dir.template")

	// setup the transport
	if err := initTransport(); err != nil {
		log.Fatal(err)
	}

	// set up wikis
	if err := initWikis(); err != nil {
		log.Fatal(err)
	}

	// setup static files from wikifier
	if err := setupStatic(); err != nil {
		log.Fatal(err)
	}

	log.Println("quiki ready")

	// main handler
	http.HandleFunc("/", handleRoot)

	// listen
	log.Fatal(http.ListenAndServe(conf.Get("server.http.bind")+":"+port, nil))
}

func setupStatic() error {
	if stat, err := os.Stat(wikifierPath); err != nil || !stat.IsDir() {
		if err == nil {
			err = errors.New("not a directory")
		}
		errStr := fmt.Sprintf(
			"@dir.wikifier (%s) error: %v\n",
			wikifierPath,
			err.Error(),
		)
		return errors.New(errStr)
	}
	fileServer := http.FileServer(http.Dir(wikifierPath + "/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))
	return nil
}
