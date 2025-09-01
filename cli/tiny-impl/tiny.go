package tinyimpl

import (
	"flag"
	"fmt"
	"os"

	"github.com/cooper/quiki/cli"
	"github.com/cooper/quiki/wikifier"
)

type Parser struct{}

func (p *Parser) SetupFlags(c *cli.Config) {
	flag.BoolVar(&c.Interactive, "i", false, "interactive mode, read from stdin")
	flag.BoolVar(&c.JSONOutput, "json", false, "output JSON instead of HTML")
}

func (p *Parser) HandleCommand(c *cli.Config, args []string) error {
	// handle interactive mode
	if c.Interactive {
		cli.RunInteractiveMode(c.JSONOutput)
		return nil
	}

	// must have a page file argument
	if len(args) == 0 {
		Usage()
		os.Exit(1)
	}

	// process standalone page
	page := wikifier.NewPage(args[0])
	cli.RunPageAndExit(page, c.JSONOutput)
	return nil
}

func Usage() {
	fmt.Fprintf(os.Stderr, "usage: wikifier [options] [page file]\n\n")
	fmt.Fprintf(os.Stderr, "minimal wikifier engine for processing standalone page files\n\n")
	fmt.Fprintf(os.Stderr, "common usages:\n")
	fmt.Fprintf(os.Stderr, "  wikifier somepage.page    render page to HTML and output to stdout\n")
	fmt.Fprintf(os.Stderr, "  wikifier -i               read page content from stdin\n\n")
	fmt.Fprintf(os.Stderr, "options:\n")
	flag.PrintDefaults()
}
