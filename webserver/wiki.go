package webserver

// wiki.go - manage the wikis served by this quiki

import (
	"errors"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cooper/quiki/monitor"
	"github.com/cooper/quiki/pregenerate"
	"github.com/cooper/quiki/wiki"
	"github.com/cooper/quiki/wikifier"
)

// WikiInfo represents a wiki hosted on this webserver.
type WikiInfo struct {
	Name               string // wiki shortname
	Title              string // wiki title from @name in the wiki config
	Logo               string
	Host               string
	template           wikiTemplate
	pregenerateManager *pregenerate.Manager
	*wiki.Wiki
}

// Wikis is all wikis served by this webserver.
var Wikis map[string]*WikiInfo

// initialize all the wikis in the configuration
func InitWikis() error {

	// find wikis
	found, err := Conf.Get("server.wiki")
	if err != nil {
		return err
	}
	if found == nil {
		log.Println("no wikis are configured yet")
		return nil
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
	if Wikis == nil {
		Wikis = make(map[string]*WikiInfo)
	}
	for _, wikiName := range wikiNames {

		// already loaded
		if _, ok := Wikis[wikiName]; ok {
			continue
		}

		configPfx := "server.wiki." + wikiName

		// not enabled
		enable, _ := Conf.GetBool(configPfx + ".enable")
		if !enable {
			continue
		}

		// host to accept (optional)
		wikiHost, _ := Conf.GetStr(configPfx + ".host")

		// ceate the wiki instance
		var w *wiki.Wiki

		// first, prefer server.wiki.[name].dir
		dirWiki, _ := Conf.GetStr(configPfx + ".dir")
		if dirWiki != "" {
			w, err = wiki.NewWiki(dirWiki)
			if err != nil {
				return err
			}
		}

		// still no??? use server.dir.wiki/[name]
		if w == nil {

			// if not set, give up because this is last resort
			serverDirWiki, err := Conf.GetStr("server.dir.wiki")
			if err != nil {
				return err
			}

			w, err = wiki.NewWiki(filepath.Join(serverDirWiki, wikiName))
			if err != nil {
				return err
			}
		}

		// if wiki host was found in wiki config, use it ONLY when
		// no host was specified in server config.
		if wikiHost == "" {
			wikiHost = w.Opt.Host.Wiki
		}

		// create wiki info for webserver
		wi := &WikiInfo{Wiki: w, Host: wikiHost, Name: wikiName}

		// initialize git repsitory
		log.Println(w.BranchNames())

		// initialize pregeneration manager - always available in unified architecture
		pregenerationEnabled, _ := Conf.GetBool("server.enable.pregeneration")

		// check for pregeneration mode setting
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

		// allow individual option overrides
		if rateLimit, err := Conf.GetDuration("server.pregen.rate_limit"); err == nil {
			opts.RateLimit = rateLimit
		}
		if progressInterval, err := Conf.GetInt("server.pregen.progress_interval"); err == nil {
			opts.ProgressInterval = progressInterval
		}
		if priorityQueueSize, err := Conf.GetInt("server.pregen.page_priority"); err == nil {
			opts.PriorityQueueSize = priorityQueueSize
		}
		if backgroundQueueSize, err := Conf.GetInt("server.pregen.page_background"); err == nil {
			opts.BackgroundQueueSize = backgroundQueueSize
		}
		if imagePriorityQueueSize, err := Conf.GetInt("server.pregen.img_priority"); err == nil {
			opts.ImagePriorityQueueSize = imagePriorityQueueSize
		}
		if imageBackgroundQueueSize, err := Conf.GetInt("server.pregen.img_background"); err == nil {
			opts.ImageBackgroundQueueSize = imageBackgroundQueueSize
		}
		if priorityWorkers, err := Conf.GetInt("server.pregen.page_workers"); err == nil {
			opts.PriorityWorkers = priorityWorkers
		}
		if backgroundWorkers, err := Conf.GetInt("server.pregen.page_bg_workers"); err == nil {
			opts.BackgroundWorkers = backgroundWorkers
		}
		if imagePriorityWorkers, err := Conf.GetInt("server.pregen.img_workers"); err == nil {
			opts.ImagePriorityWorkers = imagePriorityWorkers
		}
		if imageBackgroundWorkers, err := Conf.GetInt("server.pregen.img_bg_workers"); err == nil {
			opts.ImageBackgroundWorkers = imageBackgroundWorkers
		}
		if requestTimeout, err := Conf.GetDuration("server.pregen.timeout"); err == nil {
			opts.RequestTimeout = requestTimeout
		}
		if forceGen, err := Conf.GetBool("server.pregen.force"); err == nil {
			opts.ForceGen = forceGen
		}
		if logVerbose, err := Conf.GetBool("server.pregen.verbose"); err == nil {
			opts.LogVerbose = logVerbose
		}
		if enableImages, err := Conf.GetBool("server.pregen.images"); err == nil {
			opts.EnableImages = enableImages
		}
		if cleanupInterval, err := Conf.GetDuration("server.pregen.cleanup"); err == nil {
			opts.CleanupInterval = cleanupInterval
		}
		if maxTrackingEntries, err := Conf.GetInt("server.pregen.max_tracking"); err == nil {
			opts.MaxTrackingEntries = maxTrackingEntries
		}

		// always create and start the manager for unified queue system
		if pregenerationEnabled {
			log.Printf("pregenerate: starting background pregeneration for wiki %s", wikiName)
			wi.pregenerateManager = pregenerate.NewWithOptions(w, opts).StartBackground()
		} else {
			log.Printf("pregenerate: DISABLED for wiki %s - content will be generated on-demand only; for production environments, enable pregeneration to improve user performance", wikiName)
			wi.pregenerateManager = pregenerate.NewWithOptions(w, opts).StartWorkers()
		}

		// monitor for changes with pregeneration integration
		if err := monitor.GetManager().AddWikiWithPregeneration(w, wi.pregenerateManager); err != nil {
			return err
		}

		// set up the wiki for webserver
		if err := setupWiki(wi); err != nil {
			return err
		}

		Wikis[wikiName] = wi
	}

	// still no wikis?
	if len(Wikis) == 0 {
		return errors.New("none of the configured wikis are enabled")
	}

	return nil
}

var warnedHostnameUsed bool

// initialize a wiki
func setupWiki(wi *WikiInfo) error {

	// derive root.ext if not configured
	if wi.Opt.Root.Ext == "" {
		host := wi.Opt.Host.Wiki
		if host == "" {
			host = Opts.Host
		}
		if host == "" {
			host, _ = os.Hostname()
			if host != "" && !warnedHostnameUsed {
				log.Printf("[%s] @server.http.host not configured, using system "+
					"hostname to configure external root: %s", wi.Name, host)
				warnedHostnameUsed = true
			}
		}
		if host == "" {
			host = "localhost"
			log.Printf("[%s] warning: no host configured for wiki nor default "+
				"host on server. either set a default host in quiki.conf with "+
				"@server.http.host, or set @host.wiki in wiki.conf", wi.Name)
		}
		port := ""
		if Opts.Port != "80" {
			port = ":" + Opts.Port
		}
		wi.Opt.Root.Ext = "http://" + path.Join(host+port, wi.Opt.Root.Wiki)
	}

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
		template, err = loadTemplateAtPath(templateNameOrPath)
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
			wi.Logo = wi.Opt.Root.Image + "/" + disp.File
		case wiki.DisplayRedirect:
			wi.Logo = wi.Opt.Root.Image + "/" + disp.Redirect
		default:
			log.Printf("[%s] generate logo failed: %+v", wi.Name, res)
		}
	}

	type wikiHandler struct {
		rootType string
		root     string
		handler  func(*WikiInfo, string, http.ResponseWriter, *http.Request)
	}

	wikiRoots := []wikiHandler{
		{
			rootType: "page",
			root:     wi.Opt.Root.Page,
			handler:  handlePage,
		},
		{
			rootType: "image",
			root:     wi.Opt.Root.Image,
			handler:  handleImage,
		},
		{
			rootType: "category",
			root:     wi.Opt.Root.Category,
			handler:  handleCategoryPosts,
		},
	}

	// setup handlers
	wikiRoot := wi.Opt.Root.Wiki
	if !strings.HasPrefix(wikiRoot, "/") {
		wikiRoot = "/" + wikiRoot
	}
	for _, item := range wikiRoots {
		rootType, root, handler := item.rootType, item.root, item.handler

		if !strings.HasPrefix(root, "/") {
			root = "/" + root
		}

		// if this is the page root and it's blank, skip it
		if rootType == "page" && root == "/" {
			log.Printf("[%s] pages will be handled at wiki root: %s", wi.Name, wi.Host+wikiRoot)
			continue
		}

		// if it doesn't already have the wiki root as the prefix, add it
		if !strings.HasPrefix(root, wikiRoot) {
			root = wikiRoot + root
		}

		if !strings.HasSuffix(root, "/") {
			root += "/"
		}

		// add the real handler
		wi := wi // copy pointer so the handler below always refer to this one
		Mux.HandleFunc(wi.Host+root, func(w http.ResponseWriter, r *http.Request) {

			// determine the path relative to the root
			relPath := strings.TrimPrefix(r.URL.Path, root)
			if relPath == "" {
				http.NotFound(w, r)
				return
			}

			handler(wi, relPath, w, r)
		})

		log.Printf("[%s] registered %s root: %s", wi.Name, rootType, wi.Host+root)
	}

	// file server
	rootFile := wi.Opt.Root.File
	dirWiki := wi.Dir()
	if rootFile != "" && dirWiki != "" {
		rootFile += "/"
		fileServer := http.FileServer(http.Dir(dirWiki))
		Mux.Register(wi.Host+rootFile, "site file directory index", http.StripPrefix(rootFile, fileServer))
		log.Printf("[%s] registered file root: %s (%s)", wi.Name, wi.Host+rootFile, dirWiki)
	}

	// store the wiki info
	wi.Title = wi.Opt.Name
	return nil
}

// Copy creates a WikiInfo with all the same options, minus Wiki.
// It is used for working with multiple branches within a wiki.
func (wi *WikiInfo) Copy(w *wiki.Wiki) *WikiInfo {
	return &WikiInfo{
		Name:     wi.Name,
		Title:    wi.Title,
		Logo:     wi.Logo,
		Host:     wi.Host,
		template: wi.template,
		Wiki:     w,
	}
}

// Shutdown gracefully shuts down the wiki and its services
func (wi *WikiInfo) Shutdown() {
	if wi.pregenerateManager != nil {
		wi.pregenerateManager.Stop()
		wi.pregenerateManager = nil
	}
}
