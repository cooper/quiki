// Package adminifier provides an administrative panel with a web-based editor.
package adminifier

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/alexedwards/scs/v2"
	"github.com/cooper/quiki/resources"
	"github.com/cooper/quiki/webserver"
	"github.com/cooper/quiki/wikifier"
	"github.com/pkg/errors"
)

var tmpl *template.Template
var mux *webserver.ServeMux
var conf *wikifier.Page
var sessMgr *scs.SessionManager
var host, root string

// Configure sets up adminifier on webserver.ServeMux using webserver.Conf.
func Configure() {
	conf = webserver.Conf
	mux = webserver.Mux

	// do nothing if not enabled
	if enable, _ := conf.GetBool("adminifier.enable"); !enable {
		return
	}

	// extract strings
	for key, ptr := range map[string]*string{
		"adminifier.host": &host,
		"adminifier.root": &root,
	} {
		str, err := conf.GetStr(key)
		if err != nil {
			log.Fatal(err)
		}
		*ptr = str
	}

	if !strings.HasSuffix(root, "/") {
		root += "/"
	}

	// configure session manager
	sessMgr = webserver.SessMgr
	sessMgr.Cookie.SameSite = http.SameSiteStrictMode
	sessMgr.Cookie.Path = root

	// create template
	tmpl = template.Must(template.ParseFS(resources.Adminifier, "template/*.tpl"))

	// main handler
	subFS, err := fs.Sub(resources.Adminifier, "app")
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to get app filesystem"))
	}
	fileServer := http.FileServer(http.FS(subFS))
	mux.Register(host+root, "adminifier root", http.StripPrefix(root, fileServer))
	log.Println("registered adminifier root: " + host + root)

	// handlers for each site at shortcode/
	initWikis()

	// if there are no users yet, let them know token
	if tok, _ := conf.Get("adminifier.token"); tok != nil && len(webserver.Auth.Users) == 0 {
		log.Printf("no admin users exist yet, visit %screate-user to create one", host+root)
		log.Printf("your setup token: %s", tok)
	}
}
