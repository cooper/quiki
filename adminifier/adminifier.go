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

const wikiDelimeter = "-/"

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

	if !strings.HasPrefix(root, "/") {
		root = "/" + root
	}
	if !strings.HasSuffix(root, "/") {
		root += "/"
	}

	// configure session manager
	sessMgr = webserver.SessMgr
	sessMgr.Cookie.SameSite = http.SameSiteStrictMode
	sessMgr.Cookie.Path = root

	// setup adminifier static files server
	if err := setupStatic(resources.Adminifier, root+"static/"); err != nil {
		log.Fatal(errors.Wrap(err, "setup adminifier static"))
	}

	// setup webserver static files server
	if err := setupStatic(resources.Webserver, root+"qstatic/"); err != nil {
		log.Fatal(errors.Wrap(err, "setup adminifier qstatic"))
	}

	// create template
	tmpl = template.Must(template.ParseFS(resources.Adminifier, "template/*.tpl"))

	// main handler
	mux.RegisterFunc(host+root, "adminifier root", handleRoot)
	log.Println("registered adminifier root: " + host + root)

	// admin handlers
	setupAdminHandlers()

	// handlers for each site at shortcode/
	initWikis()

	// if there are no users yet, let them know token
	if tok, _ := conf.Get("adminifier.token"); tok != nil && len(webserver.Auth.Users) == 0 {
		log.Printf("no admin users exist yet, visit %screate-user to create one", host+root)
		log.Printf("your setup token: %s", tok)
	}
}
func setupStatic(efs fs.FS, staticRoot string) error {
	subFS, err := fs.Sub(efs, "static")
	if err != nil {
		return errors.Wrap(err, "creating static sub filesystem")
	}
	fileServer := http.FileServer(http.FS(subFS))
	mux.Register(host+staticRoot, "adminifier static files", http.StripPrefix(staticRoot, fileServer))
	return nil
}
