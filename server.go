package main

import (
	"path/filepath"

	"github.com/cooper/quiki/adminifier"
	"github.com/cooper/quiki/cli"
	"github.com/cooper/quiki/webserver"
)

func runServer(c *cli.Config) {
	// setup server options from config
	opts := webserver.Options{
		Config:   c.Config,
		WikisDir: filepath.Join(c.QuikiDir, "wikis"),
		Bind:     c.Bind,
		Port:     c.Port,
		Host:     c.Host,
	}

	// if running wizard, create a new config file
	if c.Wizard {
		webserver.CreateWizardConfig(opts)
	}

	// run webserver
	webserver.Configure(opts)
	adminifier.Configure()
	writePIDFile(c)

	// handle SIGHUP to rehash server config
	go handleSignals()
	webserver.Listen()
}
