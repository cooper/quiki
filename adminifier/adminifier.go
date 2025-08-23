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

	// simple rule: if root is empty, it means domain root ("/")
	// if root is non-empty, it must start with "/"
	if root != "" && !strings.HasPrefix(root, "/") {
		root = "/" + root
	}
	if root != "" && !strings.HasSuffix(root, "/") {
		root += "/"
	}

	// for session cookie path: use "/" if root is empty, otherwise use root
	cookiePath := root
	if cookiePath == "" {
		cookiePath = "/"
	}

	// configure session manager
	sessMgr = webserver.SessMgr
	sessMgr.Cookie.SameSite = http.SameSiteStrictMode
	sessMgr.Cookie.Path = cookiePath

	// setup static files
	staticRoot := root
	if staticRoot == "" {
		staticRoot = "/"
	}

	staticPattern := staticRoot + "static/"
	qstaticPattern := staticRoot + "qstatic/"

	if err := setupStatic(resources.Adminifier, host+staticPattern, staticPattern); err != nil {
		log.Fatal(errors.Wrap(err, "setup adminifier static"))
	}

	if err := setupStatic(resources.Webserver, host+qstaticPattern, qstaticPattern); err != nil {
		log.Fatal(errors.Wrap(err, "setup adminifier qstatic"))
	} // create template
	tmpl = template.Must(template.ParseFS(resources.Adminifier, "template/*.tpl"))

	// register main handler - simple approach
	pattern := host + root
	if root == "" {
		pattern = host + "/"
	}
	mux.RegisterFunc(pattern, "adminifier root", handleRoot)
	log.Println("registered adminifier root: " + pattern) // admin handlers
	setupAdminHandlers()

	// handlers for each site at shortcode/
	initWikis()

	// if there are no users yet, let them know token
	if tok, _ := conf.Get("adminifier.token"); tok != nil && len(webserver.Auth.Users) == 0 {
		userURL := host + root
		if root == "" && host != "" {
			userURL = host + "/"
		}
		log.Printf("no admin users exist yet, visit %screate-user to create one", userURL)
		log.Printf("your setup token: %s", tok)
	}
}
func setupStatic(efs fs.FS, pattern, stripPrefix string) error {
	subFS, err := fs.Sub(efs, "static")
	if err != nil {
		return errors.Wrap(err, "creating static sub filesystem")
	}
	fileServer := http.FileServer(http.FS(subFS))
	mux.Register(pattern, "adminifier static files", http.StripPrefix(stripPrefix, fileServer))
	return nil
}
