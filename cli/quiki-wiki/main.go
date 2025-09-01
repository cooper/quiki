package main

import (
	"flag"

	"github.com/cooper/quiki/cli"
	impl "github.com/cooper/quiki/cli/wiki-impl"
)

func main() {
	parser := &impl.Parser{}
	flag.Usage = impl.Usage
	config, args := cli.ParseFlags(parser)

	if err := parser.HandleCommand(config, args); err != nil {
		panic(err)
	}
}
