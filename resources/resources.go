package resources

import (
	"embed"
	"io/fs"
	"log"
)

// PUBLIC

// Adminifier provides access to the adminifier resource files.
var Adminifier fs.FS

// Webserver provides access to the webserver resource files.
var Webserver fs.FS

// Shared provides access to shared resource files.
var Shared fs.FS

// Wikis provides embedded base wikis.
var Wikis fs.FS

// PRIVATE

//go:embed adminifier/*
var adminifierFs embed.FS

//go:embed webserver/*
var webserverFs embed.FS

//go:embed shared/*
var sharedFs embed.FS

//go:embed wikis/*
var wikisFs embed.FS

func init() {
	var err error
	var resources = map[string]struct {
		fs       embed.FS
		accessor *fs.FS
	}{
		"adminifier": {adminifierFs, &Adminifier},
		"webserver":  {webserverFs, &Webserver},
		"shared":     {sharedFs, &Shared},
		"wikis":      {wikisFs, &Wikis},
	}
	for name, resource := range resources {
		*resource.accessor, err = fs.Sub(resource.fs, name)
		if err != nil {
			log.Fatalf("failed to access %s resources: %v", name, err)
		}
	}
}
