package wikiimpl

import (
	"flag"
	"fmt"
	"os"

	"github.com/cooper/quiki/cli"
	tiny "github.com/cooper/quiki/cli/tiny-impl"
	"github.com/cooper/quiki/pregenerate"
	"github.com/cooper/quiki/wiki"
	"github.com/cooper/quiki/wikifier"
	"github.com/pkg/errors"
)

type Parser struct {
	tiny.Parser
}

func (p *Parser) SetupFlags(c *cli.Config) {
	p.Parser.SetupFlags(c)

	flag.StringVar(&c.QuikiDir, "dir", "", "path to quiki directory (contains config, wikis, auth, etc.)")
	flag.StringVar(&c.WikiPath, "wiki", "", "path to a wiki for wiki operations")
	flag.BoolVar(&c.ForceGen, "force-gen", false, "regenerate pages even if unchanged")
}

func (p *Parser) HandleCommand(c *cli.Config, args []string) error {
	// handle interactive mode (inherited from tiny)
	if c.Interactive {
		cli.RunInteractiveMode(c.JSONOutput)
		return nil
	}

	// if we have a page file argument, process it with wiki context
	if len(args) > 0 {
		return p.handlePageFile(c, args[0])
	}

	// if wiki path specified but no page, pregenerate the wiki
	if c.WikiPath != "" {
		w, err := wiki.NewWiki(c.WikiPath)
		if err != nil {
			return err
		}

		manager := pregenerate.New(w)
		manager.PregenerateSync()
		return nil
	}

	// no arguments and no wiki path - show usage
	Usage()
	os.Exit(1)
	return nil
}

func (p *Parser) handlePageFile(c *cli.Config, pageFile string) error {
	// if wiki path is specified, use wiki context
	if c.WikiPath != "" {
		w, err := wiki.NewWiki(c.WikiPath)
		if err != nil {
			return err
		}

		// render page within wiki context
		page := w.FindPage(pageFile)
		if page == nil {
			return errors.Errorf("page not found: %s", pageFile)
		}
		cli.RunPageAndExit(page, c.JSONOutput)
		return nil
	}

	// standalone page
	page := wikifier.NewPage(pageFile)
	cli.RunPageAndExit(page, c.JSONOutput)
	return nil
}

func Usage() {
	fmt.Fprintf(os.Stderr, "usage: quiki-wiki [options] [page file]\n\n")
	fmt.Fprintf(os.Stderr, "wikifier engine with wiki context support\n\n")
	fmt.Fprintf(os.Stderr, "common usages:\n")
	fmt.Fprintf(os.Stderr, "  quiki-wiki somepage.page         render standalone page to HTML\n")
	fmt.Fprintf(os.Stderr, "  quiki-wiki -wiki=/path my_page   render page within wiki context\n")
	fmt.Fprintf(os.Stderr, "  quiki-wiki -wiki=/path           pregenerate all pages in wiki\n")
	fmt.Fprintf(os.Stderr, "  quiki-wiki -i                    read page content from stdin\n\n")
	fmt.Fprintf(os.Stderr, "options:\n")
	flag.PrintDefaults()
}
