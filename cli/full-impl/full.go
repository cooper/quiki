package fullimpl

import (
	"flag"
	"fmt"
	"os"

	"github.com/cooper/quiki/cli"
	impl "github.com/cooper/quiki/cli/wiki-impl"
)

type Parser struct {
	impl.Parser

	AuthHandler   func(*cli.Config)
	ReloadHandler func(*cli.Config)
	ServerHandler func(*cli.Config)
}

func (p *Parser) SetupFlags(c *cli.Config) {
	// start with wiki flags (which includes tiny flags)
	p.Parser.SetupFlags(c)

	// add full-specific flags
	flag.StringVar(&c.Config, "config", "", "(deprecated: use -dir instead) path to quiki.conf")
	flag.BoolVar(&c.Wizard, "w", false, "run setup wizard")
	flag.StringVar(&c.Bind, "bind", "", "address to bind to")
	flag.StringVar(&c.Port, "port", "", "port to listen on")
	flag.StringVar(&c.Host, "host", "", "default HTTP host")
	flag.BoolVar(&c.Reload, "reload", false, "send reload signal to running server")
}

func (p *Parser) HandleCommand(c *cli.Config, args []string) error {
	// handle reload signal
	if c.Reload {
		if p.ReloadHandler != nil {
			p.ReloadHandler(c)
		}
		return nil
	}

	// check for auth command
	if len(args) > 0 && args[0] == "auth" {
		if p.AuthHandler != nil {
			p.AuthHandler(c)
		}
		return nil
	}

	// if we have page arguments or wiki operations, delegate to wiki parser
	if c.Interactive || len(args) > 0 || c.WikiPath != "" {
		return p.Parser.HandleCommand(c, args)
	}

	// no arguments - run server
	if p.ServerHandler != nil {
		p.ServerHandler(c)
	}
	return nil
}

func Usage() {
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
