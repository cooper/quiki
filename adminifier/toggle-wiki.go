package adminifier

import (
	"encoding/json"
	"net/http"

	"github.com/cooper/quiki/webserver"
)

func handleEnableWiki(w http.ResponseWriter, r *http.Request) {
	setWikiEnabled(w, r, true)
}

func handleDisableWiki(w http.ResponseWriter, r *http.Request) {
	setWikiEnabled(w, r, false)
}

// setWikiEnabled flips server.wiki.<name>.enable in the server config and
// rehashes so the change takes effect immediately
func setWikiEnabled(w http.ResponseWriter, r *http.Request, enabled bool) {
	if redirectIfNotLoggedIn(w, r) {
		return
	}
	if !permissionChecker.HasServerPermission(r, "write.server.wikis") {
		http.Error(w, "insufficient permissions to manage wikis", http.StatusForbidden)
		return
	}
	if !parsePost(w, r, "name") {
		return
	}

	// only wikis that are actually configured can be toggled
	name := r.Form.Get("name")
	key := "server.wiki." + name
	if existing, _ := webserver.Conf.Get(key); existing == nil {
		http.Error(w, "no such wiki: "+name, http.StatusNotFound)
		return
	}

	respond := func(err error) {
		res := map[string]any{"success": true, "enabled": enabled}
		if err != nil {
			res = map[string]any{"success": false, "error": err.Error()}
		}
		json.NewEncoder(w).Encode(res)
	}

	webserver.Conf.Set(key+".enable", enabled)
	if err := webserver.Conf.Write(); err != nil {
		respond(err)
		return
	}
	if err := webserver.Rehash(); err != nil {
		respond(err)
		return
	}

	// adminifier handlers for a newly enabled wiki
	if enabled {
		initWikis()
	}
	respond(nil)
}
