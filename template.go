// Copyright (c) 2017, Mitchell Cooper
package main

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var templates = make(map[string]wikiTemplate)

type wikiTemplate struct {
	path     string
	template *template.Template
}

func getTemplate(path string) wikiTemplate {

	// template is already cached
	if t, ok := templates[path]; ok {
		return t
	}

	// parse HTML templates
	tmpl := template.New("")
	if err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if strings.HasSuffix(filePath, ".tmpl") {
			if _, err := tmpl.ParseFiles(filePath); err != nil {
				return err
			}
		}
		return err
	}); err != nil {
		log.Fatal(err)
	}

	// cache the template
	t := wikiTemplate{path, tmpl}
	templates[path] = t
	return t
}

type wikiPage struct {

	// titles -
	// need to provide either WholeTitle (by itself)
	// or Title (the page title) and WikiTitle (the wiki name) together
	WholeTitle string
	Title      string
	WikiTitle  string

	// HTML page content
	Content string

	// path to static/ directory within the template
	StaticRoot string
}

func (p wikiPage) VisibleTitle() string {
	if p.WholeTitle != "" {
		return p.WholeTitle
	}
	return p.Title + " - " + p.WikiTitle
}

func (p wikiPage) PageCSS() string {
	return ""
}
