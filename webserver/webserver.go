// Package webserver is the newest webserver.
package webserver

// Copyright (c) 2020, Mitchell Cooper
// quiki - a standalone web server for wikifier

import (
	"context"
	"encoding/gob"
	"fmt"
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
	"github.com/cooper/quiki/pregenerate"
	"github.com/cooper/quiki/resources"
	"github.com/cooper/quiki/router"
	"github.com/cooper/quiki/wiki"
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

	// find all wikis in the new configuration
	found, err := Conf.Get("server.wiki")
	if err != nil {
		return errors.Wrap(err, "rehash: get wiki config")
	}
	if found == nil {
		// no wikis configured, shut down all existing wikis
		for wikiName, wi := range Wikis {
			monitor.GetManager().RemoveWiki(wikiName)
			wi.Shutdown()
			delete(Wikis, wikiName)
		}
		return nil
	}
	wikiMap, ok := found.(*wikifier.Map)
	if !ok {
		return errors.New("rehash: server.wiki is not a map")
	}

	configuredWikis := make(map[string]bool)
	for _, wikiName := range wikiMap.Keys() {
		configuredWikis[wikiName] = true
	}

	// remove wikis that are no longer in the configuration
	for wikiName := range Wikis {
		if !configuredWikis[wikiName] {
			if wi, exists := Wikis[wikiName]; exists {
				monitor.GetManager().RemoveWiki(wikiName)
				wi.Shutdown()
				delete(Wikis, wikiName)
			}
		}
	}

	// rehash each configured wiki individually (config already parsed)
	for wikiName := range configuredWikis {
		if err := rehashWikiWithConfig(wikiName); err != nil {
			return errors.Wrap(err, "rehash: rehash wiki "+wikiName)
		}
	}

	// check if we have any wikis after rehashing
	if len(Wikis) == 0 {
		return errors.New("rehash: none of the configured wikis are enabled")
	}

	return nil
}

// RehashWiki reloads configuration and reinitializes a specific wiki.
// this is more efficient than Rehash() when only one wiki's config has changed.
func RehashWiki(wikiName string) error {
	// re-parse configuration to get latest changes
	if err := Conf.Parse(); err != nil {
		return errors.Wrap(err, "rehash wiki: parse config")
	}

	return rehashWikiWithConfig(wikiName)
}

// rehashWikiWithConfig reinitializes a specific wiki using the already-parsed config.
// this is used internally by both RehashWiki and Rehash to avoid duplicate config parsing.
func rehashWikiWithConfig(wikiName string) error {
	// check if the wiki exists in current configuration
	configPfx := "server.wiki." + wikiName
	enable, _ := Conf.GetBool(configPfx + ".enable")
	if !enable {
		// wiki is disabled, remove it if it exists
		if wi, exists := Wikis[wikiName]; exists {
			monitor.GetManager().RemoveWiki(wikiName)
			wi.Shutdown()
			delete(Wikis, wikiName)
		}
		return nil
	}

	// store the old wiki for cleanup after successful replacement
	var oldWiki *WikiInfo
	if wi, exists := Wikis[wikiName]; exists {
		oldWiki = wi
	}

	// initialize the new wiki using the same logic as InitWikis()
	// host to accept (optional)
	wikiHost, _ := Conf.GetStr(configPfx + ".host")

	// create the wiki instance
	var w *wiki.Wiki
	var err error

	// first, prefer server.wiki.[name].dir
	dirWiki, _ := Conf.GetStr(configPfx + ".dir")
	if dirWiki != "" {
		w, err = wiki.NewWiki(dirWiki)
		if err != nil {
			return errors.Wrap(err, "rehash wiki: create wiki from dir")
		}
	}

	// still no wiki? use server.dir.wiki/[name]
	if w == nil {
		// if not set, give up because this is last resort
		serverDirWiki, err := Conf.GetStr("server.dir.wiki")
		if err != nil {
			return errors.Wrap(err, "rehash wiki: get server wiki dir")
		}

		w, err = wiki.NewWiki(filepath.Join(serverDirWiki, wikiName))
		if err != nil {
			return errors.Wrap(err, "rehash wiki: create wiki from server dir")
		}
	}

	// if wiki host was found in wiki config, use it ONLY when
	// no host was specified in server config.
	if wikiHost == "" {
		wikiHost = w.Opt.Host.Wiki
	}

	// create wiki info for webserver
	wi := &WikiInfo{Wiki: w, Host: wikiHost, Name: wikiName}

	// initialize pregeneration manager
	pregenerationEnabled, _ := Conf.GetBool("server.enable.pregeneration")

	// get pregeneration options (same logic as InitWikis)
	pregenerationMode, _ := Conf.GetStr("server.pregen.mode")
	var opts pregenerate.Options
	switch pregenerationMode {
	case "fast":
		opts = pregenerate.FastOptions()
	case "slow":
		opts = pregenerate.SlowOptions()
	default:
		opts = pregenerate.DefaultOptions()
	}

	// apply option overrides from config
	if rateLimit, ok, _ := Conf.GetDuration("server.pregen.rate_limit"); ok {
		opts.RateLimit = rateLimit
	}
	if progressInterval, ok, _ := Conf.GetInt("server.pregen.progress_interval"); ok {
		opts.ProgressInterval = progressInterval
	}
	if priorityQueueSize, ok, _ := Conf.GetInt("server.pregen.page_priority"); ok {
		opts.PriorityQueueSize = priorityQueueSize
	}
	if backgroundQueueSize, ok, _ := Conf.GetInt("server.pregen.page_background"); ok {
		opts.BackgroundQueueSize = backgroundQueueSize
	}
	if imagePriorityQueueSize, ok, _ := Conf.GetInt("server.pregen.img_priority"); ok {
		opts.ImagePriorityQueueSize = imagePriorityQueueSize
	}
	if imageBackgroundQueueSize, ok, _ := Conf.GetInt("server.pregen.img_background"); ok {
		opts.ImageBackgroundQueueSize = imageBackgroundQueueSize
	}
	if priorityWorkers, ok, _ := Conf.GetInt("server.pregen.page_workers"); ok {
		opts.PriorityWorkers = priorityWorkers
	}
	if backgroundWorkers, ok, _ := Conf.GetInt("server.pregen.page_bg_workers"); ok {
		opts.BackgroundWorkers = backgroundWorkers
	}
	if imagePriorityWorkers, ok, _ := Conf.GetInt("server.pregen.img_workers"); ok {
		opts.ImagePriorityWorkers = imagePriorityWorkers
	}
	if imageBackgroundWorkers, ok, _ := Conf.GetInt("server.pregen.img_bg_workers"); ok {
		opts.ImageBackgroundWorkers = imageBackgroundWorkers
	}
	if requestTimeout, ok, _ := Conf.GetDuration("server.pregen.timeout"); ok {
		opts.RequestTimeout = requestTimeout
	}
	if val, err := Conf.Get("server.pregen.force"); err == nil {
		if forceGen, ok := val.(bool); ok {
			opts.ForceGen = forceGen
		}
	}
	if val, err := Conf.Get("server.pregen.verbose"); err == nil {
		if logVerbose, ok := val.(bool); ok {
			opts.LogVerbose = logVerbose
		}
	}
	if val, err := Conf.Get("server.pregen.images"); err == nil {
		if enableImages, ok := val.(bool); ok {
			opts.EnableImages = enableImages
		}
	}
	if cleanupInterval, ok, _ := Conf.GetDuration("server.pregen.cleanup"); ok {
		opts.CleanupInterval = cleanupInterval
	}
	if maxTrackingEntries, ok, _ := Conf.GetInt("server.pregen.max_tracking"); ok {
		opts.MaxTrackingEntries = maxTrackingEntries
	}

	// start pregeneration manager
	if pregenerationEnabled {
		wi.pregenerateManager = pregenerate.NewWithOptions(w, opts).StartBackground()
	} else {
		wi.pregenerateManager = pregenerate.NewWithOptions(w, opts).StartWorkers()
	}

	// monitor for changes with pregeneration integration
	if err := monitor.GetManager().AddWikiWithPregeneration(w, wi.pregenerateManager); err != nil {
		return errors.Wrap(err, "rehash wiki: add wiki monitoring")
	}

	// set up the wiki for webserver
	if oldWiki == nil {
		// new wiki, set up everything including routes
		if err := setupWiki(wi); err != nil {
			return errors.Wrap(err, "rehash wiki: setup wiki")
		}
	} else {
		// existing wiki - unregister old routes and set up new ones
		Router.RemoveWiki(wikiName)
		if err := setupWiki(wi); err != nil {
			return errors.Wrap(err, "rehash wiki: setup wiki")
		}
	}

	// store the updated wiki atomically
	Wikis[wikiName] = wi

	// now that the new wiki is in place, safely cleanup the old pregenerate manager only
	if oldWiki != nil {
		oldWiki.Shutdown()
	}

	return nil
}

// Conf is the webserver configuration page.
//
// It is available only after Configure is called.
var Conf *wikifier.Page

// Router is the HTTP router.
//
// It is available only after Configure is called.
var Router *router.Router

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

// GlobalPermissionChecker is the global permission checker for webserver handlers
var GlobalPermissionChecker *PermissionChecker

// Configure parses a configuration file and initializes webserver.
//
// If any errors occur, the program is terminated.
func Configure(_initial_options Options) {
	var err error
	Opts = _initial_options
	Router = router.New()
	gob.Register(&Session{})

	// parse configuration
	if _, err := os.Stat(Opts.Config); err != nil {
		fmt.Printf("no quiki config found at: %s\n", Opts.Config)
		fmt.Printf("use -dir=/path to specify a different quiki directory, or -w to run setup wizard\n")
		os.Exit(1)
	}

	Conf = wikifier.NewPage(Opts.Config)
	Conf.VarsOnly = true
	if err = Conf.Parse(); err != nil {
		log.Fatal(errors.Wrap(err, "parse config"))
	}

	// extract strings
	var templateDirs, domain string
	for key, ptr := range map[string]*string{
		"server.http.port":    &Opts.Port,
		"server.http.bind":    &Opts.Bind,
		"server.http.host":    &Opts.Host,
		"server.dir.template": &templateDirs,
		"server.domain":       &domain,
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

	// add shared templates
	if sub, err := fs.Sub(resources.Shared, "template"); err == nil {
		templateFses = append(templateFses, sub)
	} else {
		log.Fatal(errors.Wrap(err, "loading shared templates"))
	}

	// set up wikis
	if err = InitWikis(); err != nil {
		log.Fatal(errors.Wrap(err, "init wikis"))
	}

	// setup static files from wikifier
	if err = setupStatic(); err != nil {
		log.Fatal(errors.Wrap(err, "setup static"))
	}

	// create session manager with security hardening
	SessMgr = scs.New()

	// security hardening for production
	SessMgr.Lifetime = 24 * time.Hour                 // maximum session lifetime
	SessMgr.IdleTimeout = 4 * time.Hour               // session expires after inactivity
	SessMgr.Cookie.Secure = true                      // require https in production
	SessMgr.Cookie.HttpOnly = true                    // prevent js access to session cookies
	SessMgr.Cookie.SameSite = http.SameSiteStrictMode // prevent csrf
	SessMgr.Cookie.Name = "__quiki_session"           // use custom session cookie name for security

	// configure cookie domain for cross-subdomain sharing if specified
	if domain != "" {
		SessMgr.Cookie.Domain = domain
	}

	// create global permission checker
	GlobalPermissionChecker = NewPermissionChecker(SessMgr)

	// create server with main handler
	Router.HandleFunc("/", "webserver root", handleRoot)
	Server = &http.Server{Handler: SessMgr.LoadAndSave(Router)}

	// create authenticator
	var authPath string
	if dir := filepath.Dir(Opts.Config); dir != "" {
		authPath = filepath.Join(dir, "quiki-auth.json")
	} else {
		authPath = "quiki-auth.json"
	}

	Auth, err = authenticator.OpenServer(authPath)
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
	Router.Handle("/static/", "webserver static files", http.StripPrefix("/static/", fileServer))

	// setup shared static files
	sharedFS, err := fs.Sub(resources.Shared, "static")
	if err != nil {
		return errors.Wrap(err, "creating shared static sub filesystem")
	}
	sharedFileServer := http.FileServer(http.FS(sharedFS))
	Router.Handle("/shared/", "shared static files", http.StripPrefix("/shared/", sharedFileServer))

	return nil
}

type WikiSetupHook func(string, *WikiInfo) error

var wikiSetupHooks []WikiSetupHook

// RegisterWikiSetupHook registers a callback to be called when a wiki is set up or rehashed
func RegisterWikiSetupHook(hook WikiSetupHook) {
	wikiSetupHooks = append(wikiSetupHooks, hook)
}

func callWikiSetupHooks(shortcode string, wi *WikiInfo) error {
	for _, hook := range wikiSetupHooks {
		if err := hook(shortcode, wi); err != nil {
			return errors.Wrapf(err, "wiki setup hook failed for %s", shortcode)
		}
	}
	return nil
}
