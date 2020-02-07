package adminifier

import (
	"net/http"

	"github.com/cooper/quiki/webserver"
)

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
	// TODO: session verify
	tmpl.ExecuteTemplate(w, "wiki.tpl", struct {
		Static    string
		AdminRoot string
		Root      string
	}{
		AdminRoot: root,
		Static:    root + "adminifier-static/",
		Root:      root + shortcode,
	})
}
