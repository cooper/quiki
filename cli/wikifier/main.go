package main

import (
	"flag"

	"github.com/cooper/quiki/cli"
	tiny "github.com/cooper/quiki/cli/tiny-impl"
)

func main() {
	parser := &tiny.Parser{}
	flag.Usage = tiny.Usage
	config, args := cli.ParseFlags(parser)

	if err := parser.HandleCommand(config, args); err != nil {
		panic(err)
	}
}
