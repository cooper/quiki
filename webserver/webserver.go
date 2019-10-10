// Package webserver is the newest webserver.
package webserver

// Copyright (c) 2019, Mitchell Cooper
// quiki - a standalone web server for wikifier

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/cooper/quiki/wikifier"
	"github.com/pkg/errors"
)

var conf *wikifier.Page
var mux *http.ServeMux

// Run runs the webserver.
func Run() {
	mux = http.NewServeMux()

	// find config file
	if len(os.Args) < 2 || os.Args[1] == "" {
		log.Fatal("usage: " + os.Args[0] + " /path/to/quiki.conf")
	}

	// parse configuration
	conf = wikifier.NewPage(os.Args[1])
	conf.VarsOnly = true
	if err := conf.Parse(); err != nil {
		log.Fatal(errors.Wrap(err, "parse config"))
	}

	var port, dirStatic string
	for key, ptr := range map[string]*string{
		"server.http.port":    &port,
		"server.dir.template": &templateDirs,
		"server.dir.static":   &dirStatic,
	} {
		str, err := conf.GetStr(key)
		if err != nil {
			log.Fatal(err)
		}
		*ptr = str
	}

	// set up wikis
	if err := initWikis(); err != nil {
		log.Fatal(errors.Wrap(err, "init wikis"))
	}

	// setup static files from wikifier
	if err := setupStatic(dirStatic); err != nil {
		log.Fatal(errors.Wrap(err, "setup static"))
	}

	log.Println("quiki ready")

	// create server with main handler
	mux.HandleFunc("/", handleRoot)
	server := &http.Server{Handler: mux}

	// listen
	bind, err := conf.GetStr("server.http.bind")
	if err != nil {
		log.Fatal(err)
	}
	if port == "unix" {
		listener, err := net.Listen("unix", bind)
		if err != nil {
			log.Fatal(errors.Wrap(err, "listen"))
		}
		server.Serve(listener)
	} else {
		server.Addr = bind + ":" + port
		log.Fatal(errors.Wrap(server.ListenAndServe(), "listen"))
	}
}

func setupStatic(staticPath string) error {
	if stat, err := os.Stat(staticPath); err != nil || !stat.IsDir() {
		if err == nil {
			err = errors.New("not a directory")
		}
		return err
	}
	fileServer := http.FileServer(http.Dir(staticPath))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))
	return nil
}
