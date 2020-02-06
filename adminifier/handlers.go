package adminifier

import (
	"log"
	"net/http"
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
	if !parsePost(w, r, "username", "password") {
		return
	}
	log.Println("got login request")
}

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
