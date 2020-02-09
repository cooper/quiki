package adminifier

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/cooper/quiki/webserver"
)

var javascriptTemplates string

func setupWikiHandlers(shortcode string, wi *webserver.WikiInfo) {
	for _, which := range []string{"dashboard", "pages", "categories", "images", "models", "settings", "help"} {
		mux.HandleFunc(host+root+shortcode+"/"+which, func(w http.ResponseWriter, r *http.Request) {
			handleWiki(shortcode, w, r)
		})
	}
	pfx := host + root + shortcode + "/frame/"
	mux.HandleFunc(pfx, func(w http.ResponseWriter, r *http.Request) {
		tmplName := "frame-" + strings.TrimPrefix(r.URL.Path, pfx) + ".tpl"

		// frame template does not exist
		if exists := tmpl.Lookup(tmplName); exists == nil {
			http.NotFound(w, r)
			return
		}

		// execute frame template
		err := tmpl.ExecuteTemplate(w, tmplName, nil)

		// error occurred in template execution
		if err != nil {
			panic(err)
		}
	})
}

func handleWiki(shortcode string, w http.ResponseWriter, r *http.Request) {

	// load javascript templates
	if javascriptTemplates == "" {
		files, _ := filepath.Glob(dirAdminifier + "/template/js-tmpl/*.tpl")
		for _, fileName := range files {
			data, _ := ioutil.ReadFile(fileName)
			javascriptTemplates += string(data)
		}
	}

	// TODO: session verify
	err := tmpl.ExecuteTemplate(w, "wiki.tpl", struct {
		Static      string
		AdminRoot   string
		Root        string
		JSTemplates template.HTML
	}{
		AdminRoot:   root,
		Static:      root + "adminifier-static",
		Root:        root + shortcode,
		JSTemplates: template.HTML(javascriptTemplates),
	})
	if err != nil {
		panic(err)
	}
}
