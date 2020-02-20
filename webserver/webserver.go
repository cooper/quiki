// Package webserver is the newest webserver.
package webserver

// Copyright (c) 2020, Mitchell Cooper
// quiki - a standalone web server for wikifier

import (
	"encoding/gob"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/alexedwards/scs/v2"
	"github.com/cooper/quiki/authenticator"
	"github.com/cooper/quiki/wikifier"
	"github.com/pkg/errors"
)

// Conf is the webserver configuration page.
//
// It is available only after Configure is called.
var Conf *wikifier.Page

// Mux is the *http.ServeMux.
//
// It is available only after Configure is called.
var Mux *http.ServeMux

// Server is the *http.Server.
//
// It is available only after Configure is called.
var Server *http.Server

// Bind is the string to bind to, as extracted from the configuration file.
//
// It is available only after Configure is called.
var Bind string

// Port is the port to bind to or "unix" for a UNIX socket, as extracted from the configuration file.
//
// It is available only after Configure is called.
var Port string

// Auth is the server authentication service.
var Auth *authenticator.Authenticator

// SessMgr is the session manager service.
var SessMgr *scs.SessionManager

// Configure parses a configuration file and initializes webserver.
//
// If any errors occur, the program is terminated.
func Configure(confFile string) {
	var err error
	Mux = http.NewServeMux()
	gob.Register(authenticator.User{})

	// parse configuration
	Conf = wikifier.NewPage(confFile)
	Conf.VarsOnly = true
	if err = Conf.Parse(); err != nil {
		log.Fatal(errors.Wrap(err, "parse config"))
	}

	// extract strings
	var dirStatic string
	for key, ptr := range map[string]*string{
		"server.http.port":    &Port,
		"server.http.bind":    &Bind,
		"server.dir.template": &templateDirs,
		"server.dir.static":   &dirStatic,
	} {
		str, err := Conf.GetStr(key)
		if err != nil {
			log.Fatal(err)
		}
		*ptr = str
	}

	// normalize paths
	templateDirs = filepath.FromSlash(templateDirs)
	dirStatic = filepath.FromSlash(dirStatic)

	// set up wikis
	if err = initWikis(); err != nil {
		log.Fatal(errors.Wrap(err, "init wikis"))
	}

	// setup static files from wikifier
	if err = setupStatic(dirStatic); err != nil {
		log.Fatal(errors.Wrap(err, "setup static"))
	}

	// create session manager
	SessMgr = scs.New()

	// create server with main handler
	Mux.HandleFunc("/", handleRoot)
	Server = &http.Server{Handler: SessMgr.LoadAndSave(Mux)}

	// create authenticator
	Auth, err = authenticator.Open(filepath.Join(filepath.Dir(confFile), "quiki-auth.json"))
	if err != nil {
		log.Fatal(errors.Wrap(err, "init server authenticator"))
	}
}

// Listen runs the webserver indefinitely.
//
// Configure must be called first.
// If any errors occur, the program is terminated.
func Listen() {
	if Port == "unix" {
		listener, err := net.Listen("unix", Bind)
		log.Println("quiki ready: " + Bind)
		if err != nil {
			log.Fatal(errors.Wrap(err, "listen"))
		}
		Server.Serve(listener)
	} else {
		Server.Addr = Bind + ":" + Port
		log.Println("quiki ready on port " + Port)
		log.Fatal(errors.Wrap(Server.ListenAndServe(), "listen"))
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
	Mux.Handle("/static/", http.StripPrefix("/static/", fileServer))
	return nil
}
