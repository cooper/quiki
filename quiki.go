package main

import (
	"log"
	"os"

	"github.com/cooper/quiki/adminifier"
	"github.com/cooper/quiki/webserver"
)

func main() {
	// find config file
	if len(os.Args) < 2 || os.Args[1] == "" {
		log.Fatal("usage: " + os.Args[0] + " /path/to/quiki.conf")
	}

	// configure webserver using conf file
	webserver.Configure(os.Args[1])

	// configure adminifier using existing server and conf page
	// (it depends on webserver being loaded already)
	adminifier.Configure()

	// listen indefinitely
	webserver.Listen()
}
