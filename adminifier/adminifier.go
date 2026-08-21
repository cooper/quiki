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
	"github.com/cooper/quiki/router"
	"github.com/cooper/quiki/webserver"
	"github.com/cooper/quiki/wikifier"
	"github.com/pkg/errors"
)

var tmpl *template.Template
var mux *router.Router
var conf *wikifier.Page
var sessMgr *scs.SessionManager
var host, root string

const wikiDelimeter = "-/"

// Configure sets up adminifier on webserver.ServeMux using webserver.Conf.
func Configure() {
	conf = webserver.Conf
	mux = webserver.Router

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

	// configure session manager and permission checker
	sessMgr = webserver.SessMgr
	sessMgr.Cookie.SameSite = http.SameSiteStrictMode

	if sessMgr.Cookie.Domain == "" {
		sessMgr.Cookie.Path = root
	}

	permissionChecker = webserver.NewPermissionChecker(sessMgr)

	// setup adminifier static files server
	if err := setupStatic(resources.Adminifier, root+"static/"); err != nil {
		log.Fatal(errors.Wrap(err, "setup adminifier static"))
	}

	// setup shared static files server
	if err := setupStatic(resources.Shared, root+"shared/"); err != nil {
		log.Fatal(errors.Wrap(err, "setup shared static"))
	}

	// setup webserver static files server
	if err := setupStatic(resources.Webserver, root+"qstatic/"); err != nil {
		log.Fatal(errors.Wrap(err, "setup adminifier qstatic"))
	}

	// create template
	tmpl = template.Must(template.New("").ParseFS(resources.Adminifier, "template/*.tpl"))
	tmpl = template.Must(tmpl.ParseFS(resources.Shared, "template/*.tpl"))

	// main handler
	mux.HandleFunc(host+root, "adminifier root", func(w http.ResponseWriter, r *http.Request) {
		handleRoot(&adminifierLoginRedirectWriter{ResponseWriter: w}, r)
	})
	mux.HandleFunc(host+root+"static/config.js", "adminifier config.js", handleConfigJS)
	log.Println("registered adminifier root: " + host + root)

	// admin handlers
	setupAdminHandlers()

	// handlers for each site at shortcode/
	initWikis()

	// if there are no users yet, let them know token
	if tok, _ := conf.Get("adminifier.token"); tok != nil && len(webserver.Auth.Users) == 0 {
		const (
			red    = "\033[31m"
			yellow = "\033[33m"
			bold   = "\033[1m"
			reset  = "\033[0m"
		)
		log.Printf("%s%sno server users exist yet!%s", bold, yellow, reset)
		log.Println()
		log.Printf("%s%s/!\\          YOUR SETUP TOKEN           /!\\%s", bold, red, reset)
		log.Printf("%s%s/!\\   %s%s%s%s  /!\\%s", bold, red, reset, bold, tok, red, reset)
		log.Printf("%s%s/!\\                                     /!\\%s", bold, red, reset)
		log.Println()
		log.Printf("%s%svisit %s%s to create the first user%s", bold, yellow, host+root, reset, reset)
		log.Printf("%s%scopy+paste the above token%s into browser to claim your server", bold, yellow, reset)
	}
}

// adminifierLoginRedirectWriter rewrites relative login redirects to absolute adminifier paths.
type adminifierLoginRedirectWriter struct {
	http.ResponseWriter
}

func (l *adminifierLoginRedirectWriter) WriteHeader(status int) {
	if status >= 300 && status < 400 {
		loc := l.ResponseWriter.Header().Get("Location")
		if loc == "login" || strings.HasPrefix(loc, "login?") ||
			loc == "./login" || strings.HasPrefix(loc, "./login?") {
			if strings.HasPrefix(loc, "./") {
				loc = strings.TrimPrefix(loc, "./")
			}
			l.ResponseWriter.Header().Set("Location", root+loc)
		}
	}
	l.ResponseWriter.WriteHeader(status)
}

func setupStatic(efs fs.FS, staticRoot string) error {
	subFS, err := fs.Sub(efs, "static")
	if err != nil {
		return errors.Wrap(err, "creating static sub filesystem")
	}
	fileServer := http.FileServer(http.FS(subFS))
	mux.Handle(host+staticRoot, "adminifier static files", http.StripPrefix(staticRoot, fileServer))
	return nil
}
