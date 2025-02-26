package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/cooper/quiki/adminifier"
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
	flag.Parse()

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
		// -wiki with no page means pregnerate the wiki
		w.Pregenerate()
		log.Println("wiki pregenerated")
		return
	}

	useDefaultConfigPath := opts.Config == ""
	if useDefaultConfigPath {
		opts.Config = filepath.Join(os.Getenv("HOME"), "quiki", "quiki.conf")
	}

	// if running wizard, create a new config file
	if wizard {
		webserver.CreateWizardConfig(opts)
	}

	// print usage when running with no args and no config in default location
	if useDefaultConfigPath {
		if _, err := os.Stat(opts.Config); err != nil {
			log.Printf("config file not found at default location: %s", opts.Config)
			flag.Usage()
			os.Exit(1)
		}
	}

	// run webserver
	webserver.Configure(opts)
	adminifier.Configure()
	webserver.Listen()
}

// reads from stdin
func interactiveMode() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	page := wikifier.NewPageSource(string(input))
	runPageAndExit(page)
}

// runs a page
func runPageAndExit(page *wikifier.Page) {
	err := page.Parse()
	if err != nil {
		log.Fatal(err)
	}
	if jsonOutput {
		json.NewEncoder(os.Stdout).Encode(page)
		os.Exit(0)
	}
	fmt.Println(page.HTMLAndCSS())
	os.Exit(0)
}

// runs a wiki page
func runWikiPageAndExit(w *wiki.Wiki, pagePath string) {
	res := w.DisplayPage(pagePath)
	if jsonOutput {
		json.NewEncoder(os.Stdout).Encode(res)
		os.Exit(0)
	}
	switch r := res.(type) {
	case wiki.DisplayPage:
		if r.CSS != "" {
			fmt.Println(`<style type="text/css">`)
			fmt.Println(r.CSS)
			fmt.Println(`</style>`)
		}
		fmt.Println(r.Content)
	case wiki.DisplayError:
		log.Fatal(r.Error)
	case wiki.DisplayRedirect:
		log.Fatal("Redirect: " + r.Redirect)
	default:
		log.Fatal("unsupported response type from wiki.DisplayPage()")
	}
	os.Exit(0)
}
