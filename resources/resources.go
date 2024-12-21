package resources

import (
	"embed"
	"io/fs"
)

//go:embed adminifier/*
var adminifierFS embed.FS

//go:embed webserver/*
var webserverFS embed.FS

// Adminifier provides access to the adminifier resource files.
var Adminifier, _ = fs.Sub(adminifierFS, "adminifier")

// Webserver provides access to the webserver resource files.
var Webserver, _ = fs.Sub(webserverFS, "webserver")
