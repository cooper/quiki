package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/cooper/quiki/adminifier"
	"github.com/cooper/quiki/webserver"
)

func runServer() {

	useDefaultConfigPath := opts.Config == ""
	if useDefaultConfigPath {
		opts.Config = filepath.Join(os.Getenv("HOME"), "quiki", "quiki.conf")
	}

	// if running wizard, create a new config file
	if wizard {
		webserver.CreateWizardConfig(opts)
	}

	// print usage when running with no args and no config in default location
	if useDefaultConfigPath {
		if _, err := os.Stat(opts.Config); err != nil {
			log.Printf("config file not found at default location: %s", opts.Config)
			flag.Usage()
			os.Exit(1)
		}
	}

	// run webserver
	webserver.Configure(opts)
	adminifier.Configure()
	writePIDFile()

	// handle SIGHUP to rehash server config
	go handleSignals()
	webserver.Listen()
}
