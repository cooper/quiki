package adminifier

import (
	"net/http"

	"github.com/cooper/quiki/authenticator"
	"github.com/cooper/quiki/webserver"
)

// handlers that call functions
var funcHandlers = map[string]func(w http.ResponseWriter, r *http.Request){
	"func/login": handleLogin,
}

func handleRoot(w http.ResponseWriter, r *http.Request) {

	// if not logged in, temp redirect to login page
	if !sessMgr.GetBool(r.Context(), "loggedIn") {
		http.Redirect(w, r, root+"login", http.StatusTemporaryRedirect)
		return
	}

	tmpl.ExecuteTemplate(w, "server.tpl", struct {
		User  authenticator.User
		Wikis map[string]*webserver.WikiInfo
	}{
		User:  sessMgr.Get(r.Context(), "user").(authenticator.User),
		Wikis: webserver.Wikis,
	})
	// TODO: if user has only one site and no admin privs, go straight to site dashboard
	// and deny access to the server admin panel
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

	// redirect to dashboard, which is now located at adminifier root
	http.Redirect(w, r, "../", http.StatusTemporaryRedirect)
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
