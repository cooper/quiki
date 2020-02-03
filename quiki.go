package main

import (
	"log"
	"os"

	"github.com/cooper/quiki/webserver"
)

func main() {
	// find config file
	if len(os.Args) < 2 || os.Args[1] == "" {
		log.Fatal("usage: " + os.Args[0] + " /path/to/quiki.conf")
	}

	webserver.New(os.Args[1]).Listen()
}
