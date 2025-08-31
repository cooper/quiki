package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

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
	QuikiDir    string
	opts        webserver.Options
)

func main() {
	flag.Usage = printUsage

	// quiki directory and config options
	flag.StringVar(&QuikiDir, "dir", "", "path to quiki directory (contains config, wikis, auth, etc.)")
	flag.StringVar(&opts.Config, "config", "", "(deprecated: use -dir instead) path to quiki.conf")

	// operation modes
	flag.BoolVar(&interactive, "i", false, "interactive mode, read from stdin")
	flag.BoolVar(&wizard, "w", false, "run setup wizard")
	flag.StringVar(&wikiPath, "wiki", "", "path to a wiki for wiki operations")
	flag.BoolVar(&forceGen, "force-gen", false, "regenerate pages even if unchanged")
	flag.BoolVar(&jsonOutput, "json", false, "output JSON instead of HTML")

	// server options
	flag.StringVar(&opts.Bind, "bind", "", "address to bind to")
	flag.StringVar(&opts.Port, "port", "", "port to listen on")
	flag.StringVar(&opts.Host, "host", "", "default HTTP host")
	flag.BoolVar(&reload, "reload", false, "send reload signal to running server")
	flag.Parse()

	initQuikiDir()

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

	if len(flag.Args()) > 0 {

		// subcommands
		switch flag.Arg(0) {
		case "auth":
			handleAuthCommand()
			return
		}

		// otherwise, running a page in standalone mode

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

// printUsage prints custom usage information
func printUsage() {
	fmt.Fprintf(os.Stderr, "usage: quiki [options] [page file]\n\n")
	fmt.Fprintf(os.Stderr, "quiki directory:\n")
	fmt.Fprintf(os.Stderr, "  by default, quiki uses ~/quiki to store its data\n")
	fmt.Fprintf(os.Stderr, "  use -dir to specify a different place\n\n")
	fmt.Fprintf(os.Stderr, "common usages:\n")
	fmt.Fprintf(os.Stderr, "  quiki -w                    run webserver with setup wizard\n")
	fmt.Fprintf(os.Stderr, "  quiki                       run webserver\n")
	fmt.Fprintf(os.Stderr, "  quiki -dir=/var/lib/quiki   run webserver w/ different config/data dir\n")
	fmt.Fprintf(os.Stderr, "  quiki somepage.page         render standalone page to stdout\n")
	fmt.Fprintf(os.Stderr, "  quiki -wiki=/path my_page   render page within wiki context\n\n")
	fmt.Fprintf(os.Stderr, "options:\n")
	flag.PrintDefaults()
}

func initQuikiDir() {
	if QuikiDir == "" && opts.Config == "" {
		QuikiDir = filepath.Join(os.Getenv("HOME"), "quiki")
	} else if QuikiDir == "" && opts.Config != "" {
		log.Printf("warning: -config flag is deprecated, use -dir instead")
		QuikiDir = filepath.Dir(opts.Config)
	}

	opts.Config = filepath.Join(QuikiDir, "quiki.conf")
	opts.WikisDir = filepath.Join(QuikiDir, "wikis")
}
