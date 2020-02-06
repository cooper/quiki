package adminifier

import "net/http"

// handlers that call functions
var funcHandlers = map[string]func(w http.ResponseWriter, r *http.Request){}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	// TODO: if not logged in, temp redirect to login page
	// TODO: if user has multiple sites OR server admin privs, go to server dashboard
	// TODO: if user has only one site and no admin privs, go straight to site dashboard
}
