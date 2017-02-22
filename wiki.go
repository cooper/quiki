// Copyright (c) 2017, Mitchell Cooper
// wikis.go - manage the wikis served by this quiki
package main

import (
	"errors"
	wikiclient "github.com/cooper/go-wikiclient"
	"github.com/cooper/quiki/config"
	"net/http"
	"strings"
	"time"
)

// represents a wiki
type wikiInfo struct {
	name     string            // wiki shortname
	title    string            // wiki title from @name in the wiki config
	password string            // wiki password for read authentication
	confPath string            // path to wiki configuration
	template wikiTemplate      // template
	client   wikiclient.Client // client, only available in handlers
	conf     *config.Config    // wiki config instance
}

// all wikis served by this quiki
var wikis map[string]wikiInfo

// initialize all the wikis in the configuration
func initializeWikis() error {

	// find wikis
	wikiMap := conf.GetMap("server.wiki")
	if len(wikiMap) == 0 {
		return errors.New("no wikis configured")
	}

	// set up each wiki
	wikis = make(map[string]wikiInfo, len(wikiMap))
	for wikiName := range wikiMap {
		configPfx := "server.wiki." + wikiName

		// not enabled
		if !conf.GetBool(configPfx + ".quiki") {
			continue
		}

		// get wiki config path and password
		var wikiConfPath, wikiPassword string
		if err := conf.RequireMany(map[string]*string{
			configPfx + ".config":   &wikiConfPath,
			configPfx + ".password": &wikiPassword,
		}); err != nil {
			return err
		}

		// create wiki info
		wiki := wikiInfo{
			name:     wikiName,
			password: wikiPassword,
			confPath: wikiConfPath,
		}

		// set up the wiki
		if err := setupWiki(wiki); err != nil {
			return err
		}
	}

	// still no wikis?
	if len(wikis) == 0 {
		return errors.New("none of the configured wikis are enabled")
	}

	return nil
}

// wiki roots mapped to handler functions
var wikiRoots = map[string]func(wikiInfo, string, http.ResponseWriter, *http.Request){
	"wiki":  handlePage,
	"page":  handlePage,
	"image": handleImage,
}

// initialize a wiki
func setupWiki(wiki wikiInfo) error {

	// parse the wiki configuration
	wiki.conf = config.New(wiki.confPath)
	if err := wiki.conf.Parse(); err != nil {
		return err
	}

	// maybe we can get the wikifier path from this
	if wikifierPath == "" {
		wikifierPath = wiki.conf.Get("dir.wikifier")
	}

	// find the wiki root
	var wikiRoot = wiki.conf.Get("root.wiki")

	// find the template. if not configured, use default
	templatePath := conf.Get("server.wiki." + wiki.name + ".template")
	if templatePath == "" {
		templatePath = "default"
	}
	template, err := getTemplate(templatePath)
	if err != nil {
		return err
	}
	wiki.template = template

	// make a generic session used for read access for this wiki
	readSess := &wikiclient.Session{
		WikiName:     wiki.name,
		WikiPassword: wiki.password,
	}

	// setup handlers
	for rootType, handler := range wikiRoots {
		root, err := wiki.conf.RequireExists("root." + rootType)
		if err != nil {
			return err
		}

		// if it doesn't already have the wiki root as the prefix, add it
		if !strings.HasPrefix(root, wikiRoot) {
			wiki.conf.Warnf(
				"@root.%s (%s) is configured outside of @root.wiki (%s); assuming %s%s",
				rootType, root, wikiRoot, wikiRoot, root,
			)
			root = wikiRoot + root
		}

		// normally 'something/' handles 'something' as well; this prevents that
		if root != "" {
			http.HandleFunc(root, http.NotFound)
		}
		root += "/"

		// add the real handler
		realHandler := handler
		realRootType := rootType
		http.HandleFunc(root, func(w http.ResponseWriter, r *http.Request) {
			wiki.client = wikiclient.NewClient(tr, readSess, 3*time.Second)

			// the transport is not connected
			if tr.Dead() {
				http.Error(w, "503 service unavailable", http.StatusServiceUnavailable)
				return
			}

			// determine the path relative to the root
			relPath := strings.TrimPrefix(r.URL.Path, root)
			if relPath == "" && realRootType != "wiki" {
				http.NotFound(w, r)
				return
			}

			realHandler(wiki, relPath, w, r)
		})
	}

	// store the wiki info
	wiki.title = wiki.conf.Get("name")
	wikis[wiki.name] = wiki
	return nil
}
