package adminifier

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/cooper/quiki/webserver"
)

var javascriptTemplates string

func setupWikiHandlers(shortcode string, wi *webserver.WikiInfo) {
	for _, which := range []string{"dashboard", "pages", "categories", "images", "models", "settings", "help"} {
		mux.HandleFunc(host+root+shortcode+"/"+which, func(w http.ResponseWriter, r *http.Request) {
			handleWiki(shortcode, w, r)
		})
	}
	mux.HandleFunc(host+root+shortcode+"/frame/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Frame content " + r.URL.Path))
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
