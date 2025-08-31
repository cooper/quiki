package main

import (
	"github.com/cooper/quiki/adminifier"
	"github.com/cooper/quiki/webserver"
)

func runServer() {
	// if running wizard, create a new config file
	if wizard {
		webserver.CreateWizardConfig(opts)
	}

	// run webserver
	webserver.Configure(opts)
	adminifier.Configure()
	writePIDFile()

	// handle SIGHUP to rehash server config
	go handleSignals()
	webserver.Listen()
}
