// Copyright (c) 2017, Mitchell Cooper
// wiki.go - manage the wikis served by this quiki
package webserver

import (
	"errors"
	wikiclient "github.com/cooper/go-wikiclient"
	"github.com/cooper/quiki/config"
	"log"
	"net/http"
	"path"
	"strings"
	"time"
)

// represents a wiki
type wikiInfo struct {
	name        string              // wiki shortname
	title       string              // wiki title from @name in the wiki config
	host        string              // wiki hostname
	password    string              // wiki password for read authentication
	confPath    string              // path to wiki configuration
	logo        string              // set internally to properly-sized logo filename
	template    wikiTemplate        // template
	client      wikiclient.Client   // client, only available in handlers
	conf        *config.Config      // wiki config instance
	defaultSess *wikiclient.Session // default session
}

// all wikis served by this quiki
var wikis map[string]wikiInfo

// initialize all the wikis in the configuration
func initWikis() error {

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
		if !conf.GetBool(configPfx + ".enable") {
			continue
		}

		// host to accept (optional)
		wikiHost := conf.Get(configPfx + ".host")

		// get wiki config path and password
		wikiPassword := conf.Get(configPfx + ".password")
		wikiConfPath := conf.Get(configPfx + ".config")
		if wikiConfPath == "" {
			// config not specified, so use server.dir.wiki and wiki.conf
			dirWiki, err := conf.Require("server.dir.wiki")
			if err != nil {
				return err
			}
			wikiConfPath = dirWiki + "/" + wikiName + "/wiki.conf"
		}

		// create wiki info
		wiki := wikiInfo{
			host:     wikiHost,
			name:     wikiName,
			password: wikiPassword,
			confPath: wikiConfPath,
		}

		// set up the wiki
		if err := wiki.setup(); err != nil {
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
	"page":     handlePage,
	"image":    handleImage,
	"category": handleCategoryPosts,
}

// initialize a wiki
func (wiki wikiInfo) setup() error {

	// make a generic session and client used for read access for this wiki
	wiki.defaultSess = &wikiclient.Session{
		WikiName:     wiki.name,
		WikiPassword: wiki.password,
	}
	defaultClient := wikiclient.NewClient(tr, wiki.defaultSess, 60*time.Second)

	// connect the client, so that we can get config info
	if err := defaultClient.Connect(); err != nil {
		return err
	}

	// Safe point - we are authenticated for read access

	// create a configuration from the response
	wiki.conf = config.NewFromMap("("+wiki.name+")", wiki.defaultSess.Config)

	// maybe we can get the wikifier path from this
	if wikifierPath == "" {
		wikifierPath = wiki.conf.Get("dir.wikifier")
	}

	// find the wiki root
	wikiRoot := wiki.conf.Get("root.wiki")

	// if not configured, use default template
	templateNameOrPath := wiki.conf.Get("template")
	if templateNameOrPath == "" {
		templateNameOrPath = "default"
	}

	// find the template
	var template wikiTemplate
	var err error
	if strings.Contains(templateNameOrPath, "/") {
		// if a path is given, try to load the template at this exact path
		template, err = loadTemplate(path.Base(templateNameOrPath), templateNameOrPath)
	} else {
		// otherwise, search template directories
		template, err = findTemplate(templateNameOrPath)
	}

	// couldn't find it, or an error occured in loading it
	if err != nil {
		return err
	}
	wiki.template = template

	// generate logo according to template
	logoInfo := wiki.template.manifest.Logo
	logoName := wiki.conf.Get("logo")
	if logoName != "" && (logoInfo.Width != 0 || logoInfo.Height != 0) {
		log.Printf("[%s] generating logo %s; %dx%d\n",
			wiki.name, logoName, logoInfo.Width, logoInfo.Height)
		wiki.client = wikiclient.NewClient(tr, wiki.defaultSess, 10*time.Second)
		res, _ := wiki.client.DisplayImageOverride(logoName, logoInfo.Width, logoInfo.Height)
		if file, ok := res.Args["file"].(string); ok && file != "" {
			wiki.logo = file
		}
	}

	// setup handlers
	for rootType, handler := range wikiRoots {
		root, err := wiki.conf.Require("root." + rootType)

		// can't be empty
		if err != nil {
			continue
		}

		// if it doesn't already have the wiki root as the prefix, add it
		if !strings.HasPrefix(root, wikiRoot) {
			wiki.conf.Warnf(
				"@root.%s (%s) is configured outside of @root.wiki (%s); assuming %s%s",
				rootType, root, wikiRoot, wikiRoot, root,
			)
			root = wikiRoot + root
		}

		root += "/"

		// add the real handler
		rootType, handler := rootType, handler
		mux.HandleFunc(wiki.host+root, func(w http.ResponseWriter, r *http.Request) {
			wiki.client = wikiclient.NewClient(tr, wiki.defaultSess, 60*time.Second)
			wiki.conf.Vars = wiki.defaultSess.Config

			// the transport is not connected
			if tr.Dead() {
				http.Error(w, "503 service unavailable", http.StatusServiceUnavailable)
				return
			}

			// determine the path relative to the root
			relPath := strings.TrimPrefix(r.URL.Path, root)
			if relPath == "" && rootType != "wiki" {
				http.NotFound(w, r)
				return
			}

			handler(wiki, relPath, w, r)
		})

		log.Printf("[%s] registered %s root: %s", wiki.name, rootType, wiki.host+root)
	}

	// file server
	rootFile := wiki.conf.Get("root.file")
	dirWiki := wiki.conf.Get("dir.wiki")
	if rootFile != "" && dirWiki != "" {
		rootFile += "/"
		fileServer := http.FileServer(http.Dir(dirWiki))
		mux.Handle(wiki.host+rootFile, http.StripPrefix(rootFile, fileServer))
		log.Printf("[%s] registered file root: %s (%s)", wiki.name, wiki.host+rootFile, dirWiki)
	}

	// store the wiki info
	wiki.title = wiki.conf.Get("name")
	wikis[wiki.name] = wiki
	return nil
}

// logo path passed to templates
func (wiki wikiInfo) getLogo() string {
	if logo := wiki.logo; logo != "" {
		return wiki.conf.Get("root.image") + "/" + logo
	}
	if logo := wiki.conf.Get("logo"); logo != "" {
		return wiki.conf.Get("root.image") + "/" + logo
	}
	return ""
}