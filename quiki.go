package main

import (
	"flag"

	"github.com/cooper/quiki/cli"
	impl "github.com/cooper/quiki/cli/full-impl"
)

func main() {
	parser := &impl.Parser{
		AuthHandler:   func(c *cli.Config) { handleAuthCommand(c) },
		ReloadHandler: func(c *cli.Config) { handleReload(c) },
		ServerHandler: func(c *cli.Config) { runServer(c) },
	}

	flag.Usage = impl.Usage
	config, args := cli.ParseFlags(parser)

	if err := parser.HandleCommand(config, args); err != nil {
		panic(err)
	}
}
