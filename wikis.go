// Copyright (c) 2017, Mitchell Cooper
package main

import (
	"errors"
	"github.com/cooper/quiki/config"
	"net/http"
)

type wikiInfo struct {
	name     string         // wiki name
	password string         // wiki password for read authentication
	confPath string         // path to wiki configuration
	conf     *config.Config // wiki config instance
}

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

var wikiRoots = [...]string{"wiki", "page", "image"}

// initialize a wiki
func setupWiki(wiki wikiInfo) error {

	// parse the wiki configuration
	wiki.conf = config.New(wiki.confPath)
	if err := wiki.conf.Parse(); err != nil {
		return err
	}

	// setup the handlers
	for _, rootType := range wikiRoots {
		root, err := wiki.conf.Require("root." + rootType)
		if err != nil {
			return err
		}
		http.HandleFunc(root, func(w http.ResponseWriter, r *http.Request) {
			handler(rootType, root, w, r)
		})
	}

	// store the wiki info
	wikis[wiki.name] = wiki
	return nil
}
