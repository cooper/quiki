// Copyright (c) 2017, Mitchell Cooper
// quiki - a standalone web server for wikifier
package webserver

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/cooper/quiki/config"
)

// wikiserver config instance
var conf *config.Config
var mux *http.ServeMux

// wikifier directory path
var wikifierPath string

func Run() {
	mux = http.NewServeMux()

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

	// create server with main handler
	mux.HandleFunc("/", handleRoot)
	server := &http.Server{Handler: mux}

	// listen
	bind := conf.Get("server.http.bind")
	if port == "unix" {
		listener, err := net.Listen("unix", bind)
		if err != nil {
			log.Fatal(err)
		}
		server.Serve(listener)
	} else {
		server.Addr = bind + ":" + port
		log.Fatal(server.ListenAndServe())
	}
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
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))
	return nil
}
