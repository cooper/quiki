package main

import (
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
	opts        webserver.Options
)

func main() {
	flag.StringVar(&opts.Config, "config", "", "path to quiki.conf")
	flag.BoolVar(&interactive, "i", false, "interactive mode, read from stdin")
	flag.BoolVar(&wizard, "w", false, "run setup wizard")
	flag.StringVar(&wikiPath, "wiki", "", "path to a wiki to run standalone")
	flag.BoolVar(&forceGen, "force-gen", false, "regenerate pages even if unchanged")
	flag.StringVar(&opts.Bind, "bind", "", "address to bind to")
	flag.StringVar(&opts.Port, "port", "", "port to listen on")
	flag.StringVar(&opts.Host, "host", "", "default HTTP host")
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
	fmt.Println(page.HTML())
	os.Exit(0)
}

// runs a wiki page
func runWikiPageAndExit(w *wiki.Wiki, pagePath string) {
	res := w.DisplayPage(pagePath)
	switch r := res.(type) {
	case wiki.DisplayPage:
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
