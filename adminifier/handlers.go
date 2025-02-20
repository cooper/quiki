package adminifier

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cooper/quiki/authenticator"
	"github.com/cooper/quiki/webserver"
)

// handlers that call functions
var funcHandlers = map[string]func(w http.ResponseWriter, r *http.Request){
	"login":            handleLoginPage,
	"func/login":       handleLogin,
	"create-user":      handleCreateUserPage,
	"func/create-user": handleCreateUser,
	"logout":           handleLogout,
}

func handleRoot(w http.ResponseWriter, r *http.Request) {

	// if not logged in, temp redirect to login page
	if !sessMgr.GetBool(r.Context(), "loggedIn") {
		http.Redirect(w, r, root+"login", http.StatusTemporaryRedirect)
		return
	}

	// make sure that this is admin root
	if strings.TrimPrefix(r.URL.Path, root) != "" {
		http.NotFound(w, r)
		return
	}

	tmpl.ExecuteTemplate(w, "server.tpl", struct {
		User  *authenticator.User
		Wikis map[string]*webserver.WikiInfo
	}{
		User:  sessMgr.Get(r.Context(), "user").(*authenticator.User),
		Wikis: webserver.Wikis,
	})
	// TODO: if user has only one site and no admin privs, go straight to site dashboard
	// and deny access to the server admin panel
}

func handleLoginPage(w http.ResponseWriter, r *http.Request) {

	// if no users exist, redirect to create-user
	if len(webserver.Auth.Users) == 0 {
		http.Redirect(w, r, "create-user", http.StatusTemporaryRedirect)
		return
	}

	handleTemplate(w, r)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {

	// missing parameters or malformed request
	if !parsePost(w, r, "username", "password") {
		return
	}

	// any login attempt voids the current session
	sessMgr.Destroy(r.Context())

	// attempt login
	user, err := webserver.Auth.Login(r.Form.Get("username"), r.Form.Get("password"))
	if err != nil {
		w.Write([]byte("Bad password"))
		return
	}

	// start session and remember user info
	sessMgr.Put(r.Context(), "user", &user)
	sessMgr.Put(r.Context(), "loggedIn", true)
	sessMgr.Put(r.Context(), "branch", "master")

	// redirect to dashboard, which is now located at adminifier root
	http.Redirect(w, r, "../", http.StatusTemporaryRedirect)
}

func handleCreateUserPage(w http.ResponseWriter, r *http.Request) {

	// if users exist, redirect to login
	if len(webserver.Auth.Users) != 0 {
		http.Redirect(w, r, "login", http.StatusTemporaryRedirect)
		return
	}

	// get default wikis path if not configured
	wikiPathConfigured, _ := webserver.Conf.GetStr("server.dir.wiki")
	var defaultWikiPath string
	if wikiPathConfigured == "" {
		// use whatever dir the webserver.Conf.Path() is in as the default
		defaultWikiPath = filepath.Join(filepath.Dir(webserver.Conf.Path()), "wikis")
	}

	// render template with default wikis path if not configured
	tmpl.ExecuteTemplate(w, "create-user.tpl", struct {
		DefaultWikiPath string
	}{
		DefaultWikiPath: defaultWikiPath,
	})
}

func handleCreateUser(w http.ResponseWriter, r *http.Request) {

	// missing parameters or malformed request
	if !parsePost(w, r, "display", "email", "username", "password", "token") {
		return
	}

	// for now, you can only create a user if none exist
	// not authorized otherwise
	if len(webserver.Auth.Users) != 0 {
		http.Error(w, "user already exists", http.StatusUnauthorized)
		return
	}

	// validate token
	expectedToken, err := webserver.Conf.GetStr("adminifier.token")
	if err != nil || expectedToken == "" {
		http.Error(w, "no token set", http.StatusInternalServerError)
		return
	}
	if r.Form.Get("token") != expectedToken {
		log.Printf("expected token: %s, got: %s", expectedToken, r.Form.Get("token"))
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	// create the wiki path if not exists
	wikiPath := r.Form.Get("path")
	if wikiPath == "" {
		wikiPath, err = webserver.Conf.GetStr("server.dir.wiki")
		if err != nil || wikiPath == "" {
			http.Error(w, "no wiki path configured", http.StatusInternalServerError)
			return
		}
	}
	if err = os.MkdirAll(wikiPath, 0755); err != nil {
		http.Error(w, "creating wikis directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// create user
	// TODO: validate things
	err = webserver.Auth.NewUser(authenticator.User{
		Username:    r.Form.Get("username"),
		DisplayName: r.Form.Get("display"),
		Email:       r.Form.Get("email"),
	}, r.Form.Get("password"))

	// error occurred
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// remove token from config, add wiki path
	// do this last in case other errors occurred first
	webserver.Conf.Set("server.dir.wiki", wikiPath)
	webserver.Conf.Unset("adminifier.token")
	err = webserver.Conf.Write()
	if err != nil {
		http.Error(w, "remove token from config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// log 'em in by simulating a request to /func/login
	handleLogin(w, r)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	// destory session
	sessMgr.Destroy(r.Context())

	// redirect to login
	handleRoot(w, r)
}

// parsePost confirms POST requests are well-formed and parameters satisfied
func parsePost(w http.ResponseWriter, r *http.Request, required ...string) bool {

	// check that it is a POST request
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return false
	}

	// parse the parameters
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return false
	}

	// check that required parameters are present
	for _, req := range required {
		if _, ok := r.PostForm[req]; !ok {
			http.Error(w, "missing parameter: "+req, http.StatusUnprocessableEntity)
			return false
		}
	}

	return true
}
