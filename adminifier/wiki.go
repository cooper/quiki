package adminifier

import (
	"net/http"

	"github.com/cooper/quiki/webserver"
)

func setupWikiHandlers(shortcode string, wi *webserver.WikiInfo) {
	for _, which := range []string{"dashboard", "pages", "categories", "images", "models", "settings", "help"} {
		mux.HandleFunc(host+root+shortcode+"/"+which, handleWiki)
	}
}

func handleWiki(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "wiki.tpl", struct {
		Static string
	}{
		Static: root + "adminifier-static/",
	})
}
