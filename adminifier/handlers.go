package adminifier

import (
	"net/http"

	"github.com/cooper/quiki/webserver"
)

// handlers that call functions
var funcHandlers = map[string]func(w http.ResponseWriter, r *http.Request){
	"func/login": handleLogin,
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	// TODO: if not logged in, temp redirect to login page
	http.Redirect(w, r, root+"login", http.StatusTemporaryRedirect)
	// TODO: if user has multiple sites OR server admin privs, go to server dashboard
	// TODO: if user has only one site and no admin privs, go straight to site dashboard
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

	sessMgr.Put(r.Context(), "user", &user)
	sessMgr.Put(r.Context(), "loggedIn", true)
	w.Write([]byte("Good job"))
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
