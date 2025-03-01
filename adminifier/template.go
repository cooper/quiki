package adminifier

import (
	"log"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

func handleTemplate(w http.ResponseWriter, r *http.Request, dot any) {
	relPath := strings.TrimPrefix(r.URL.Path, root)
	err := tmpl.ExecuteTemplate(w, relPath+".tpl", dot)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(errors.Wrap(err, "execute template"))
	}
}
