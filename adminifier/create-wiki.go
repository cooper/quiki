package adminifier

import (
	"net/http"
	"path/filepath"
	"slices"

	"github.com/cooper/quiki/webserver"
	"github.com/cooper/quiki/wiki"
	"github.com/cooper/quiki/wikifier"
)

func handleCreateWiki(w http.ResponseWriter, r *http.Request) {
	// ensure user is authenticated
	if redirectIfNotLoggedIn(w, r) {
		return
	}

	// missing parameters or malformed request
	if !parsePost(w, r, "template") {
		return
	}

	// ensure template is valid
	templateName := r.Form.Get("template")
	if !slices.Contains(webserver.TemplateNames(), templateName) {
		http.Error(w, "template not found: "+templateName, http.StatusNotFound)
		return
	}

	// ensure server.dir.wiki is set
	wikisDir, err := webserver.Conf.GetStr("server.dir.wiki")
	if err != nil || wikisDir == "" {
		http.Error(w, "server.dir.wiki is not set, nowhere to create", http.StatusInternalServerError)
		return
	}

	wikiName := r.Form.Get("name")
	normalizedName := wikifier.PageNameLink(wikiName)
	wikiPath := filepath.Join(wikisDir, normalizedName)

	// create the wiki
	err = wiki.CreateWikiFromResource(wikiPath, r.Form.Get("base"), wiki.CreateWikiOpts{
		WikiName:     wikiName,
		TemplateName: templateName,
		MainPage:     "main",
		ErrorPage:    "error",
	})
	if err != nil {
		http.Error(w, "create wiki: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// enable the wiki on server config
	webserver.Conf.Set("server.wiki."+normalizedName+".enable", true)
	if err := webserver.Conf.Write(); err != nil {
		http.Error(w, "write server config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// rehash server after config change
	if err := webserver.Rehash(); err != nil {
		http.Error(w, "rehash server: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// scan wikis on adminifier (webserver wikis are handled by Rehash())
	initWikis()

	// redirect to the root
	http.Redirect(w, r, "../", http.StatusSeeOther)
}
