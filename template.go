// Copyright (c) 2017, Mitchell Cooper
package main

import (
	wikiclient "github.com/cooper/go-wikiclient"
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
	logo       string             // path for logo file, if any
}

func getTemplate(name string) (wikiTemplate, error) {
	var t wikiTemplate
	templatePath := templateDir + "/" + name

	// template is already cached
	if t, ok := templates[templatePath]; ok {
		return t, nil
	}

	// parse HTML templates
	tmpl := template.New("")
	if err := filepath.Walk(templatePath, func(filePath string, info os.FileInfo, err error) error {

		// a template
		if strings.HasSuffix(filePath, ".tpl") {
			if _, err := tmpl.ParseFiles(filePath); err != nil {
				return err
			}
		}

		// static content directory
		if info.IsDir() && info.Name() == "static" {
			t.staticPath = filePath
			t.staticRoot = "/tmpl/" + name
			fileServer := http.FileServer(http.Dir(filePath))
			pfx := t.staticRoot + "/"
			http.Handle(pfx, http.StripPrefix(pfx, fileServer))
		}

		// found a logo
		if t.staticRoot != "" && strings.HasPrefix(filePath, t.staticPath+"/logo.") {
			t.logo = t.staticRoot + "/" + info.Name()
		}

		return err
	}); err != nil {
		return t, err
	}

	// cache the template
	t.path = templatePath
	t.template = tmpl
	templates[templatePath] = t
	return t, nil
}

type wikiPage struct {
	WholeTitle string             // optional, shown in <title> as-is
	Title      string             // page title
	WikiTitle  string             // wiki title
	WikiLogo   string             // path to wiki logo image
	Res        wikiclient.Message // response
	StaticRoot string             // path to static resources
}

func (p wikiPage) VisibleTitle() string {
	if p.WholeTitle != "" {
		return p.WholeTitle
	}
	return p.Title + " - " + p.WikiTitle
}

func (p wikiPage) PageCSS() template.CSS {
	return template.CSS(p.Res.String("css"))
}

func (p wikiPage) HTMLContent() template.HTML {
	return template.HTML(p.Res.String("content"))
}
