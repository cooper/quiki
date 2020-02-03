package webserver

// wiki.go - manage the wikis served by this quiki

import (
	"errors"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/cooper/quiki/monitor"
	"github.com/cooper/quiki/wiki"
	"github.com/cooper/quiki/wikifier"
)

type wikiInfo struct {
	name     string // wiki shortname
	title    string // wiki title from @name in the wiki config
	logo     string
	host     string
	template wikiTemplate
	*wiki.Wiki
}

// all wikis served by this quiki
var wikis map[string]*wikiInfo

// initialize all the wikis in the configuration
func initWikis() error {

	// find wikis
	found, err := Conf.Get("server.wiki")
	if err != nil {
		return err
	}
	wikiMap, ok := found.(*wikifier.Map)
	if !ok {
		return errors.New("server.wiki is not a map")
	}

	wikiNames := wikiMap.Keys()
	if len(wikiNames) == 0 {
		return errors.New("no wikis configured")
	}

	// set up each wiki
	wikis = make(map[string]*wikiInfo, len(wikiNames))
	for _, wikiName := range wikiNames {
		configPfx := "server.wiki." + wikiName

		// not enabled
		enable, _ := Conf.GetBool(configPfx + ".enable")
		if !enable {
			continue
		}

		// host to accept (optional)
		wikiHost, _ := Conf.GetStr(configPfx + ".host")

		// get wiki config path and password
		wikiConfPath, _ := Conf.GetStr(configPfx + ".config")
		privConfPath, _ := Conf.GetStr(configPfx + ".private")

		if wikiConfPath == "" {
			// config not specified, so use server.dir.wiki and wiki.conf
			dirWiki, err := Conf.GetStr("server.dir.wiki")
			if err != nil {
				return err
			}
			wikiConfPath = filepath.Join(dirWiki, wikiName, "wiki.conf")
		}

		// create wiki
		w, err := wiki.NewWiki(wikiConfPath, privConfPath)
		if err != nil {
			return err
		}

		// if wiki host was found in wiki config, use it ONLY when
		// no host was specified in server config.
		if wikiHost == "" {
			wikiHost = w.Opt.Host.Wiki
		}

		// create wiki info for webserver
		wi := &wikiInfo{Wiki: w, host: wikiHost, name: wikiName}

		// pregenerate
		w.Pregenerate()

		// monitor for changes
		go monitor.WatchWiki(w)

		// set up the wiki for webserver
		if err := setupWiki(wi); err != nil {
			return err
		}

		wikis[wikiName] = wi
	}

	// still no wikis?
	if len(wikis) == 0 {
		return errors.New("none of the configured wikis are enabled")
	}

	return nil
}

// initialize a wiki
func setupWiki(wi *wikiInfo) error {

	// if not configured, use default template
	templateNameOrPath := wi.Opt.Template
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

	// couldn't find it, or an error occurred in loading it
	if err != nil {
		return err
	}
	wi.template = template

	// generate logo according to template
	logoInfo := wi.template.manifest.Logo
	logoName := wi.Opt.Logo
	if logoName != "" && (logoInfo.Width != 0 || logoInfo.Height != 0) {
		si := wiki.SizedImageFromName(logoName)
		si.Width = logoInfo.Width
		si.Height = logoInfo.Height
		res := wi.DisplaySizedImageGenerate(si, true)
		switch disp := res.(type) {
		case wiki.DisplayImage:
			wi.logo = wi.Opt.Root.Image + "/" + disp.File
		case wiki.DisplayRedirect:
			wi.logo = disp.Redirect
		default:
			log.Printf("[%s] generate logo failed: %+v", wi.name, res)
		}
	}

	type wikiHandler struct {
		rootType string
		root     string
		handler  func(*wikiInfo, string, http.ResponseWriter, *http.Request)
	}

	wikiRoots := []wikiHandler{
		wikiHandler{
			rootType: "page",
			root:     wi.Opt.Root.Page,
			handler:  handlePage,
		},
		wikiHandler{
			rootType: "image",
			root:     wi.Opt.Root.Image,
			handler:  handleImage,
		},
		wikiHandler{
			rootType: "category",
			root:     wi.Opt.Root.Category,
			handler:  handleCategoryPosts,
		},
	}

	// setup handlers
	wikiRoot := wi.Opt.Root.Wiki
	for _, item := range wikiRoots {
		rootType, root, handler := item.rootType, item.root, item.handler

		// if it doesn't already have the wiki root as the prefix, add it
		if !strings.HasPrefix(root, wikiRoot) {
			log.Printf(
				"@root.%s (%s) is configured outside of @root.wiki (%s); assuming %s%s",
				rootType, root, wikiRoot, wikiRoot, root,
			)
			root = wikiRoot + root
		}

		root += "/"

		// add the real handler
		wi := wi // copy pointer so the handler below always refer to this one
		Mux.HandleFunc(wi.host+root, func(w http.ResponseWriter, r *http.Request) {

			// determine the path relative to the root
			relPath := strings.TrimPrefix(r.URL.Path, root)
			if relPath == "" && rootType != "wiki" {
				http.NotFound(w, r)
				return
			}

			handler(wi, relPath, w, r)
		})

		log.Printf("[%s] registered %s root: %s", wi.name, rootType, wi.host+root)
	}

	// file server
	rootFile := wi.Opt.Root.File
	dirWiki := wi.Opt.Dir.Wiki
	if rootFile != "" && dirWiki != "" {
		rootFile += "/"
		fileServer := http.FileServer(http.Dir(dirWiki))
		Mux.Handle(wi.host+rootFile, http.StripPrefix(rootFile, fileServer))
		log.Printf("[%s] registered file root: %s (%s)", wi.name, wi.host+rootFile, dirWiki)
	}

	// store the wiki info
	wi.title = wi.Opt.Name
	return nil
}
