package webserver

// Copyright (c) 2019, Mitchell Cooper

import (
	"encoding/json"
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cooper/quiki/wiki"
)

var templateDirs string
var templates = make(map[string]wikiTemplate)

var templateFuncs = map[string]interface{}{
	"even": func(i int) bool {
		return i%2 == 0
	},
	"odd": func(i int) bool {
		return i%2 != 0
	},
}

type wikiTemplate struct {
	path       string             // template directory path
	template   *template.Template // master HTML template
	staticPath string             // static file directory path, if any
	staticRoot string             // static file directory HTTP root, if any
	manifest   struct {

		// human-readable template name
		// Name   string

		// template author's name
		// Author string

		// URL to template code on the web, such as GitHub repository
		// Code   string

		// wiki logo info
		Logo struct {

			// ideally one of these dimensions will be specified and the other
			// not. used for the logo specified by the wiki 'logo' directive.
			// usually the height is specified. if both are present, the
			// logo will be generated in those exact dimensions.
			Height int
			Width  int
		}
	}
}

// search all template directories for a template by its name
func findTemplate(name string) (wikiTemplate, error) {

	// template is already cached
	if t, ok := templates[name]; ok {
		return t, nil
	}

	for _, templateDir := range strings.Split(templateDirs, ",") {
		templatePath := templateDir + "/" + name
		t, err := loadTemplate(name, templatePath)

		// an error occurred in loading the template
		if err != nil {
			return t, err
		}

		// no template but no error means try the next directory
		if t.template == nil {
			continue
		}

		return t, nil
	}

	// never found a template
	return wikiTemplate{}, errors.New("unable to find template " + name)
}

// load a template from its known path
func loadTemplate(name, templatePath string) (wikiTemplate, error) {
	var t wikiTemplate
	var tryNextDirectory bool

	// template is already cached
	if t, ok := templates[name]; ok {
		return t, nil
	}

	// parse HTML templates
	tmpl := template.New("")
	err := filepath.Walk(templatePath, func(filePath string, info os.FileInfo, err error) error {

		// walk error, probably missing template
		if err != nil {
			tryNextDirectory = true
			return err
		}

		// found template file
		if strings.HasSuffix(filePath, ".tpl") {

			// error in parsing
			subTmpl, err := tmpl.ParseFiles(filePath)
			if err != nil {
				return err
			}

			// add funcs
			subTmpl.Funcs(templateFuncs)
		}

		// found static content directory
		if info.IsDir() && info.Name() == "static" {
			t.staticPath = filePath
			t.staticRoot = "/tmpl/" + name
			fileServer := http.FileServer(http.Dir(filePath))
			pfx := t.staticRoot + "/"
			mux.Handle(pfx, http.StripPrefix(pfx, fileServer))
			log.Printf("[%s] template registered: %s", name, pfx)
		}

		// found manifest
		if info.Name() == "manifest.json" {

			// couldn't read manifest
			contents, err := ioutil.ReadFile(filePath)
			if err != nil {
				return err
			}

			// couldn't parse manifest
			if err := json.Unmarshal(contents, &t.manifest); err != nil {
				return err
			}
		}

		return err
	})

	// not found
	if tryNextDirectory {
		return t, nil
	}

	// other error
	if err != nil {
		return t, err
	}

	// cache the template
	t.path = templatePath
	t.template = tmpl
	templates[name] = t

	return t, nil
}

type wikiPage struct {
	File       string           // page name, with extension
	Name       string           // page name, without extension
	WholeTitle string           // optional, shown in <title> as-is
	Title      string           // page title
	WikiTitle  string           // wiki titled
	WikiLogo   string           // path to wiki logo image
	WikiRoot   string           // wiki HTTP root
	Res        wiki.DisplayPage // response
	StaticRoot string           // path to static resources
	Pages      []wikiPage       // more pages for category posts
	Message    string           // message for error page
	navigation []interface{}    // slice of nav items [display, url]
}

type navItem struct {
	Display string
	Link    string
}

func (p wikiPage) VisibleTitle() string {
	if p.WholeTitle != "" {
		return p.WholeTitle
	}
	if p.Title == p.WikiTitle || p.Title == "" {
		return p.WikiTitle
	}
	return p.Title + " - " + p.WikiTitle
}

func (p wikiPage) Scripts() []string {
	return []string{
		"/static/mootools.min.js",
		"/static/quiki.js",
		"https://cdn.rawgit.com/google/code-prettify/master/loader/run_prettify.js",
	}
}

func (p wikiPage) PageCSS() template.CSS {
	// FIXME: all_css?
	// css := p.Res.Get("all_css")
	// if css == "" {
	// 	css = p.Res.Get("css")
	// }
	return template.CSS(p.Res.CSS)
}

func (p wikiPage) HTMLContent() template.HTML {
	return template.HTML(p.Res.Content)
}

func (p wikiPage) Navigation() []navItem {

	// no navigation
	if len(p.navigation) != 2 {
		return nil
	}

	// first item is ordered keys, second is values
	displays := p.navigation[0].([]interface{})
	urls := p.navigation[1].([]interface{})
	items := make([]navItem, len(displays))
	for i := 0; i < len(items); i++ {
		items[i] = navItem{
			displays[i].(string),
			urls[i].(string),
		}
	}

	return items
}
