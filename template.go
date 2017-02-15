// Copyright (c) 2017, Mitchell Cooper
package main

import (
	"github.com/cooper/quiki/wikiclient"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var templateDir string
var templates = make(map[string]wikiTemplate)

type wikiTemplate struct {
	path       string             // template directory path
	template   *template.Template // master HTML template
	staticPath string             // static file directory path, if any
	staticRoot string             // static file directory HTTP root, if any
}

func getTemplate(name string) (wikiTemplate, error) {
	var t wikiTemplate
	path := templateDir + "/" + name

	// template is already cached
	if t, ok := templates[path]; ok {
		return t, nil
	}

	// parse HTML templates
	tmpl := template.New("")
	if err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {

		// a template
		if strings.HasSuffix(filePath, ".tpl") {
			if _, err := tmpl.ParseFiles(filePath); err != nil {
				return err
			}
		}

		// static content directory
		if info.IsDir() && info.Name() == "static" {
			t.staticPath = path
			t.staticRoot = "/tmpl/" + name
			fileServer := http.FileServer(http.Dir(path))
			pfx := t.staticRoot + "/"
			http.Handle(pfx, http.StripPrefix(pfx, fileServer))
		}

		return err
	}); err != nil {
		return t, err
	}

	// cache the template
	t.path = path
	t.template = tmpl
	templates[path] = t
	return t, nil
}

type wikiPage struct {

	// titles -
	// need to provide either WholeTitle (by itself)
	// or Title (the page title) and WikiTitle (the wiki name) together
	WholeTitle string
	Title      string
	WikiTitle  string

	// response
	Res wikiclient.Message

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

func (p wikiPage) HTMLContent() template.HTML {
	return template.HTML(p.Res.String("content"))
}
