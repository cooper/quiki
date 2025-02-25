package adminifier

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/cooper/quiki/resources"
	"github.com/cooper/quiki/webserver"
	"github.com/cooper/quiki/wikifier"
)

func handleCreateWiki(w http.ResponseWriter, r *http.Request) {
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

	wikiDir, err := webserver.Conf.GetStr("server.dir.wiki")
	if err != nil || wikiDir == "" {
		http.Error(w, "server.dir.wiki is not set, nowhere to create", http.StatusInternalServerError)
		return
	}

	normalizedName := wikifier.PageNameLink(r.Form.Get("name"))
	wikiPath := filepath.Join(wikiDir, normalizedName)

	// copy blank wiki from resources
	newWikiFs, err := fs.Sub(resources.Adminifier, "new-wiki")
	if err != nil {
		http.Error(w, "get new wiki fs: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := os.CopyFS(wikiPath, newWikiFs); err != nil {
		http.Error(w, "copy new wiki: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// create config page
	conf := wikifier.NewPage(filepath.Join(wikiPath, "wiki.conf"))
	vars := map[string]any{
		"name":       r.Form.Get("name"),
		"root.wiki":  normalizedName,
		"template":   templateName,
		"main_page":  "main",
		"error_page": "error",
	}
	for k, v := range vars {
		conf.Set(k, v)
	}

	// write wiki config
	conf.VarsOnly = true
	if err := conf.Write(); err != nil {
		http.Error(w, "write wiki config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// enable the wiki on server config
	webserver.Conf.Set("server.wiki."+normalizedName+".enable", true)
	if err := webserver.Conf.Write(); err != nil {
		http.Error(w, "write server config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// scan wikis on webserver
	err = webserver.InitWikis()
	if err != nil {
		http.Error(w, "init wikis: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// scan wikis on adminifier
	initWikis()

	// redirect to the root
	http.Redirect(w, r, "../", http.StatusSeeOther)
}
