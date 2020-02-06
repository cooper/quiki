package adminifier

import (
	"net/http"
	"strings"
)

// handlers that go straight to templates
var tmplHandlers = []string{"login"}

func handleTemplate(w http.ResponseWriter, r *http.Request) {
	relPath := strings.TrimPrefix(r.URL.Path, root)
	err := tmpl.ExecuteTemplate(w, relPath+".tpl", nil)
	if err != nil {
		// TODO: internal server error
		panic(err)
	}
}
