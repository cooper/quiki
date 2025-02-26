package resources

import (
	"embed"
	"io/fs"
	"log"
)

//go:embed adminifier/*
var adminifierFS embed.FS

//go:embed webserver/*
var webserverFS embed.FS

// Adminifier provides access to the adminifier resource files.
var Adminifier fs.FS

// Webserver provides access to the webserver resource files.
var Webserver fs.FS

// Wikis provides embedded base wikis.
//
//go:embed wikis/*
var Wikis embed.FS

func init() {
	var err error
	Adminifier, err = fs.Sub(adminifierFS, "adminifier")
	if err != nil {
		log.Fatal(err)
	}

	Webserver, err = fs.Sub(webserverFS, "webserver")
	if err != nil {
		log.Fatal(err)
	}
}
