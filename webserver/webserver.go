// Package webserver is the newest webserver.
package webserver

// Copyright (c) 2020, Mitchell Cooper
// quiki - a standalone web server for wikifier

import (
	"context"
	"encoding/gob"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/cooper/quiki/authenticator"
	"github.com/cooper/quiki/monitor"
	"github.com/cooper/quiki/resources"
	"github.com/cooper/quiki/wikifier"
	"github.com/pkg/errors"
)

// Options is the webserver command line options.
type Options struct {
	Config   string
	Bind     string
	Port     string
	Host     string
	WikisDir string
	Pregen   bool
}

// Rehash re-parses the server configuration and updates runtime settings.
// It re-reads the config file, extracts HTTP and template settings, and gracefully updates wikis.
func Rehash() error {
	// re-parse configuration
	if err := Conf.Parse(); err != nil {
		return errors.Wrap(err, "rehash: parse config")
	}

	// extract updated HTTP and template settings (if not overridden)
	var templateDirs string
	for key, ptr := range map[string]*string{
		"server.http.port":    &Opts.Port,
		"server.http.bind":    &Opts.Bind,
		"server.http.host":    &Opts.Host,
		"server.dir.template": &templateDirs,
	} {
		if *ptr != "" {
			continue
		}
		str, err := Conf.GetStr(key)
		if err != nil {
			return err
		}
		*ptr = str
	}

	// Get monitor manager for coordination
	monitorMgr := monitor.GetManager()

	// temporarily store current wikis for fallback
	oldWikis := Wikis

	// Stop monitoring and pregeneration for wikis that will be removed/replaced
	for wikiName := range oldWikis {
		if wi, exists := oldWikis[wikiName]; exists {
			monitor.GetManager().RemoveWiki(wikiName)
			wi.Shutdown() // Gracefully stop pregeneration
		}
	}

	// reset wikis map to allow InitWikis to rebuild from config
	Wikis = make(map[string]*WikiInfo) // re-initialize wikis based on updated config
	if err := InitWikis(); err != nil {
		// restore old wikis and their monitors on failure
		Wikis = oldWikis
		for _, wi := range oldWikis {
			monitorMgr.AddWikiWithPregeneration(wi.Wiki, wi.pregenerateManager)
		}
		return errors.Wrap(err, "rehash: init wikis")
	}

	// old wikis are no longer needed, garbage collection will handle cleanup

	return nil
}

// Conf is the webserver configuration page.
//
// It is available only after Configure is called.
var Conf *wikifier.Page

// Mux is the *http.ServeMux.
//
// It is available only after Configure is called.
var Mux *ServeMux

// Server is the *http.Server.
//
// It is available only after Configure is called.
var Server *http.Server

// Opts is the webserver options.
var Opts Options

// Auth is the server authentication service.
var Auth *authenticator.Authenticator

// SessMgr is the session manager service.
var SessMgr *scs.SessionManager

// Configure parses a configuration file and initializes webserver.
//
// If any errors occur, the program is terminated.
func Configure(_initial_options Options) {
	var err error
	Opts = _initial_options
	Mux = NewServeMux()
	gob.Register(&authenticator.User{})

	// parse configuration
	Conf = wikifier.NewPage(Opts.Config)
	Conf.VarsOnly = true
	if err = Conf.Parse(); err != nil {
		log.Fatal(errors.Wrap(err, "parse config"))
	}

	// extract strings
	var templateDirs string
	for key, ptr := range map[string]*string{
		"server.http.port":    &Opts.Port,
		"server.http.bind":    &Opts.Bind,
		"server.http.host":    &Opts.Host,
		"server.dir.template": &templateDirs,
	} {
		if *ptr != "" {
			// already set by opts
			continue
		}
		str, err := Conf.GetStr(key)
		if err != nil {
			log.Fatal(err)
		}
		*ptr = str
	}

	// convert templateDirs to fs.FSes on the os filesystem
	for _, dir := range strings.Split(templateDirs, ",") {
		if dir == "" {
			continue
		}
		templateFs := os.DirFS(dir)
		templateFses = append(templateFses, templateFs)
	}

	// add embedded templates
	if sub, err := fs.Sub(resources.Webserver, "templates"); err == nil {
		templateFses = append(templateFses, sub)
	} else {
		log.Fatal(errors.Wrap(err, "loading embedded templates"))
	}

	// set up wikis
	if err = InitWikis(); err != nil {
		log.Fatal(errors.Wrap(err, "init wikis"))
	}

	// setup static files from wikifier
	if err = setupStatic(); err != nil {
		log.Fatal(errors.Wrap(err, "setup static"))
	}

	// create session manager
	SessMgr = scs.New()

	// create server with main handler
	Mux.RegisterFunc("/", "webserver root", handleRoot)
	Server = &http.Server{Handler: SessMgr.LoadAndSave(Mux)}

	// create authenticator
	Auth, err = authenticator.Open(filepath.Join(filepath.Dir(Opts.Config), "quiki-auth.json"))
	if err != nil {
		log.Fatal(errors.Wrap(err, "init server authenticator"))
	}
}

// Listen runs the webserver indefinitely.
// Stop gracefully shuts down the server
func Stop() error {
	if Server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return Server.Shutdown(ctx)
	}
	return nil
}

// start listening.
// Configure must be called first.
// If any errors occur, the program is terminated.
func Listen() {
	if Opts.Port == "unix" {
		listener, err := net.Listen("unix", Opts.Bind)
		log.Println("quiki ready: " + Opts.Bind)
		if err != nil {
			log.Fatal(errors.Wrap(err, "listen"))
		}
		Server.Serve(listener)
	} else {
		Server.Addr = Opts.Bind + ":" + Opts.Port
		log.Println("quiki ready on port " + Opts.Port)
		log.Fatal(errors.Wrap(Server.ListenAndServe(), "listen"))
	}
}

// Shutdown gracefully shuts down the webserver and all wikis
func Shutdown() {
	// Stop all pregeneration managers
	for _, wi := range Wikis {
		if wi != nil {
			wi.Shutdown()
		}
	}

	// Stop the monitor manager
	monitor.GetManager().Stop()

	// Shutdown the HTTP server with a timeout
	if Server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := Server.Shutdown(ctx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}
	}
}

func setupStatic() error {
	subFS, err := fs.Sub(resources.Webserver, "static")
	if err != nil {
		return errors.Wrap(err, "creating static sub filesystem")
	}
	fileServer := http.FileServer(http.FS(subFS))
	Mux.Register("/static/", "webserver static files", http.StripPrefix("/static/", fileServer))
	return nil
}
