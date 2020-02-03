// Package webserver is the newest webserver.
package webserver

// Copyright (c) 2020, Mitchell Cooper
// quiki - a standalone web server for wikifier

import (
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cooper/quiki/wikifier"
	"github.com/pkg/errors"
)

var conf *wikifier.Page
var mux *http.ServeMux

// Server represents a quiki webserver.
type Server struct {
	Server *http.Server
	Mux    *http.ServeMux
	Conf   *wikifier.Page
	bind   string
	port   string
}

// New initializes the webserver and returns it.
func New(confFile string) *Server {
	mux = http.NewServeMux()

	// parse configuration
	conf = wikifier.NewPage(confFile)
	conf.VarsOnly = true
	if err := conf.Parse(); err != nil {
		log.Fatal(errors.Wrap(err, "parse config"))
	}

	// extract strings
	var port, bind, dirStatic string
	for key, ptr := range map[string]*string{
		"server.http.port":    &port,
		"server.http.bind":    &bind,
		"server.dir.template": &templateDirs,
		"server.dir.static":   &dirStatic,
	} {
		str, err := conf.GetStr(key)
		if err != nil {
			log.Fatal(err)
		}
		*ptr = str
	}

	// normalize paths
	templateDirs = filepath.FromSlash(templateDirs)
	dirStatic = filepath.FromSlash(dirStatic)

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

	// create webserver
	return &Server{
		Server: server,
		Mux:    mux,
		Conf:   conf,
		port:   port,
		bind:   bind,
	}
}

// Listen runs the webserver indefinitely.
//
// If any errors occur, the program is terminated.
func (s *Server) Listen() {
	if s.port == "unix" {
		listener, err := net.Listen("unix", s.bind)
		if err != nil {
			log.Fatal(errors.Wrap(err, "listen"))
		}
		s.Server.Serve(listener)
	} else {
		s.Server.Addr = s.bind + ":" + s.port
		log.Fatal(errors.Wrap(s.Server.ListenAndServe(), "listen"))
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
