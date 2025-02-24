package webserver

// Copyright (c) 2020, Mitchell Cooper

import (
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cooper/quiki/wikifier"
)

var templateFses []fs.FS
var templates = make(map[string]wikiTemplate)

var templateFuncs = map[string]any{
	"even": func(i int) bool {
		return i%2 == 0
	},
	"odd": func(i int) bool {
		return i%2 != 0
	},
}

type wikiTemplate struct {
	// path       string             // template directory path
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

// Returns the names of all available templates.
func TemplateNames() []string {
	names := make([]string, 0)
	for _, templateFs := range templateFses {
		dirs, err := fs.ReadDir(templateFs, ".")
		if err != nil {
			log.Printf("error reading template directory: %s", err)
			continue
		}
		for _, dir := range dirs {
			if dir.IsDir() {
				names = append(names, dir.Name())
			}
		}
	}
	return names
}

// search all template directories for a template by its name
func findTemplate(name string) (wikiTemplate, error) {

	// template is already cached
	if t, ok := templates[name]; ok {
		return t, nil
	}

	t, err := loadTemplate(name, name, templateFses)
	if err != nil {
		return t, err
	}

	// never found a template
	if t.template == nil {
		return wikiTemplate{}, fmt.Errorf("unable to find template '%s' in any of the provided directories", name)
	}

	return t, nil
}

// load a template from its known path
func loadTemplateAtPath(path string) (wikiTemplate, error) {
	basename := filepath.Base(path)
	return loadTemplate(".", basename, []fs.FS{os.DirFS(path)})
}

// load template given fses
func loadTemplate(walkPath, name string, fses []fs.FS) (wikiTemplate, error) {
	var t wikiTemplate

	// template is already cached
	if t, ok := templates[name]; ok {
		return t, nil
	}

	// parse HTML templates
	tmpl := template.New("")
	for _, templateFs := range fses {
		var tryNextDirectory bool
		err := fs.WalkDir(templateFs, walkPath, func(filePath string, d fs.DirEntry, err error) error {

			// walk error, probably missing template
			if err != nil {
				tryNextDirectory = true
				return err
			}

			// found template file
			if strings.HasSuffix(filePath, ".tpl") {

				// error in parsing
				subTmpl, err := tmpl.ParseFS(templateFs, filePath)
				if err != nil {
					return err
				}

				// add funcs
				subTmpl.Funcs(templateFuncs)
			}

			// found static content directory
			if d.IsDir() && d.Name() == "static" {
				t.staticPath = filePath
				t.staticRoot = "/tmpl/" + name
				if subFs, err := fs.Sub(templateFs, filePath); err == nil {
					fileServer := http.FileServer(http.FS(subFs))
					pfx := t.staticRoot + "/"
					Mux.Handle(pfx, http.StripPrefix(pfx, fileServer))
					log.Printf("[%s] template registered: %s", name, pfx)
				} else {
					log.Printf("[%s] error registering static content: %s", name, err)
				}
			}

			// found manifest
			if d.Name() == "manifest.json" {

				// couldn't read manifest
				contents, err := fs.ReadFile(templateFs, filePath)
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
			continue
		}

		// other error
		if err != nil {
			return t, err
		}

		// cache the template
		// t.path = name
		t.template = tmpl
		templates[name] = t

		return t, nil
	}

	return t, nil
}

type wikiPage struct {
	File        string                       // page name, with extension
	Name        string                       // page name, without extension
	WholeTitle  string                       // optional, shown in <title> as-is
	Title       string                       // page title
	Description string                       // page description
	Keywords    []string                     // page keywords
	Author      string                       // page author
	WikiTitle   string                       // wiki titled
	WikiLogo    string                       // path to wiki logo image (deprecated, use Logo)
	WikiRoot    string                       // wiki HTTP root (deprecated, use Root.Wiki)
	Root        wikifier.PageOptRoot         // all roots
	StaticRoot  string                       // path to static resources
	Pages       []wikiPage                   // more pages for category posts
	Message     string                       // message for error page
	Navigation  []wikifier.PageOptNavigation // slice of nav items
	PageN       int                          // for category posts, the page number (first page = 1)
	NumPages    int                          // for category posts, the number of pages
	PageCSS     template.CSS                 // css
	HTMLContent template.HTML                // html
	retina      []int                        // retina scales for logo
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
		"/static/ext/mootools.min.js",
		"/static/quiki.js",
	}
}

// for category posts, the page numbers available.
// if there is only one page, this is nothing
func (p wikiPage) PageNumbers() []int {
	if p.NumPages == 1 {
		return nil
	}
	numbers := make([]int, p.NumPages)
	for i := 1; i <= p.NumPages; i++ {
		numbers[i-1] = i
	}
	return numbers
}

func (p wikiPage) Logo() template.HTML {
	if p.WikiLogo == "" {
		return template.HTML("")
	}
	h := `<img alt="` + html.EscapeString(p.WikiTitle) + `" src="` + p.WikiLogo + `"`

	// retina
	if len(p.retina) != 0 {
		h += ` srcset="` + wikifier.ScaleString(p.WikiLogo, p.retina) + `"`
	}

	h += ` />`
	return template.HTML(h)
}

func (p wikiPage) KeywordString() string {
	return strings.Join(p.Keywords, ", ")
}
