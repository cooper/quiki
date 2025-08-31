package main

import (
	"flag"
	"log"

	"github.com/cooper/quiki/pregenerate"
	"github.com/cooper/quiki/webserver"
	"github.com/cooper/quiki/wiki"
	"github.com/cooper/quiki/wikifier"
	"github.com/pkg/errors"
)

var (
	interactive bool
	wizard      bool
	wikiPath    string
	forceGen    bool
	jsonOutput  bool
	reload      bool
	pidFile     string
	opts        webserver.Options
)

func main() {
	flag.StringVar(&opts.Config, "config", "", "path to quiki.conf")
	flag.BoolVar(&interactive, "i", false, "interactive mode, read from stdin")
	flag.BoolVar(&wizard, "w", false, "run setup wizard")
	flag.StringVar(&wikiPath, "wiki", "", "path to a wiki for wiki operations")
	flag.BoolVar(&forceGen, "force-gen", false, "regenerate pages even if unchanged")
	flag.BoolVar(&jsonOutput, "json", false, "output JSON instead of HTML")
	flag.StringVar(&opts.Bind, "bind", "", "address to bind to")
	flag.StringVar(&opts.Port, "port", "", "port to listen on")
	flag.StringVar(&opts.Host, "host", "", "default HTTP host")
	flag.StringVar(&opts.WikisDir, "wikis-dir", "", "directory to store wikis in")
	flag.BoolVar(&reload, "reload", false, "send reload signal to running server")
	flag.StringVar(&pidFile, "pidfile", "", "path to PID file (default: ~/quiki/server.pid)")
	flag.Parse()

	// handle reload functionality
	if reload {
		handleReload()
		return
	}

	// run interactive mode and exit
	if interactive {
		interactiveMode()
		return
	}

	// load wiki if -wiki used
	var w *wiki.Wiki
	if wikiPath != "" {
		var err error
		w, err = wiki.NewWiki(wikiPath)
		if err != nil {
			log.Fatal(errors.Wrap(err, "load wiki"))
		}
		if forceGen {
			w.Opt.Page.ForceGen = true
		}
	}

	// run standalone
	if len(flag.Args()) > 0 {

		// page in a standalone wiki
		if w != nil {
			runWikiPageAndExit(w, flag.Arg(0))
			return
		}

		// standalone page
		page := wikifier.NewPage(flag.Arg(0))
		runPageAndExit(page)
		return
	} else if w != nil {
		// -wiki with no page means pregenerate the wiki

		log.Println("starting pregeneration...")

		pregen := pregenerate.New(w)
		stats := pregen.PregenerateSync()

		log.Printf("pregeneration complete: %d total, %d generated, %d failed",
			stats.TotalPages, stats.PregenedPages, stats.FailedPages)

		pregen.Stop()
		return
	}

	runServer()
}
