// Copyright (c) 2017, Mitchell Cooper
// wikis.go - manage the wikis served by this quiki
package main

import (
	"errors"
	"github.com/cooper/quiki/config"
	"github.com/cooper/quiki/wikiclient"
	"net/http"
	"strings"
	"time"
)

// represents a wiki
type wikiInfo struct {
	name     string         // wiki name
	password string         // wiki password for read authentication
	confPath string         // path to wiki configuration
	template wikiTemplate   // template
	conf     *config.Config // wiki config instance
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

		// get wiki config path and password
		var wikiConfPath, wikiPassword string
		if err := conf.RequireMany(map[string]*string{
			"server.wiki." + wikiName + ".config":   &wikiConfPath,
			"server.wiki." + wikiName + ".password": &wikiPassword,
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

	return nil
}

// wiki roots mapped to handler functions
var wikiRoots = map[string]func(wikiclient.Client, string, http.ResponseWriter, *http.Request){
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

	// find the template. if not configured, use default
	templateName := conf.Get("server.wiki." + wiki.name + ".template")
	if templateName == "" {
		templateName = "default"
	}
	template, err := getTemplate(templateName)
	if err != nil {
		return err
	}
	wiki.template = template

	// find the wiki root. if not configured, use the wiki name
	var wikiRoot = wiki.conf.Get("root.wiki")
	if wikiRoot == "" {
		wikiRoot = "/" + wiki.name
		wiki.conf.Warn("@root.wiki not configured; using wiki name: " + wikiRoot)
	}

	// make a generic session used for read access for this wiki
	readSess := &wikiclient.Session{
		WikiName:     wiki.name,
		WikiPassword: wiki.password,
	}

	// setup handlers
	for rootType, handler := range wikiRoots {
		root, err := wiki.conf.Require("root." + rootType)
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
		http.HandleFunc(root, http.NotFound)

		// add the real handler
		root += "/"
		realHandler := handler
		http.HandleFunc(root, func(w http.ResponseWriter, r *http.Request) {
			c := wikiclient.Client{tr, readSess, 3 * time.Second}

			// the transport is not connected
			if tr.Dead() {
				http.Error(w, "503 service unavailable", http.StatusServiceUnavailable)
				return
			}

			// determine the path relative to the root
			relPath := strings.TrimPrefix(r.URL.Path, root)
			if relPath == "" {
				http.NotFound(w, r)
				return
			}

			realHandler(c, relPath, w, r)
		})
	}

	// store the wiki info
	wikis[wiki.name] = wiki
	return nil
}
