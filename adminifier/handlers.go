package adminifier

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/cooper/quiki/authenticator"
	"github.com/cooper/quiki/router"
	"github.com/cooper/quiki/webserver"
	"github.com/cooper/quiki/wiki"
)

// permissionChecker is the shared permission checker for adminifier
var permissionChecker *webserver.PermissionChecker

// handlers that call functions
var adminUnauthenticatedHandlers = map[string]func(w http.ResponseWriter, r *http.Request){
	"login":       handleLoginPage,
	"create-user": handleCreateUserPage,
	"logout":      handleLogout,
}

var adminUnauthenticatedFuncHandlers = map[string]func(w http.ResponseWriter, r *http.Request){
	"login":       handleLogin,
	"create-user": handleCreateUser,
	"create-wiki": handleCreateWiki,
}

// todo: separate authenticated func handlers

var adminFrameHandlers = map[string]func(*adminRequest){
	"sites":  handleSitesFrame,
	"routes": handleRoutesFrame,
	"help":   handleAdminHelpFrame,
	"help/":  handleAdminHelpFrame,
}

type adminTemplate struct {
	User      *authenticator.User
	Wikis     map[string]*webserver.WikiInfo
	Templates []string
	Title     string // server title
	Static    string // adminifier static root
	QStatic   string // webserver static root
	AdminRoot string // adminifier root'
}

type adminRequest struct {
	w        http.ResponseWriter
	r        *http.Request
	tmplName string
	dot      any
	err      error
}

// withSecurityHeaders wraps a handler function to add security headers
func withSecurityHeaders(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		webserver.SetSecurityHeaders(w)
		handler(w, r)
	}
}

func setupAdminHandlers() {
	for name, function := range adminUnauthenticatedHandlers {
		mux.HandleFunc(host+root+name, "adminifier "+name, withSecurityHeaders(function))
	}
	for name, function := range adminUnauthenticatedFuncHandlers {
		mux.HandleFunc(host+root+"func/"+name, "adminifier func "+name, withSecurityHeaders(function))
	}

	// authenticated handlers

	// each of these generates admin.tpl
	for which := range adminFrameHandlers {
		mux.HandleFunc(host+root+which, "adminifier "+which, withSecurityHeaders(handleAdmin))
	}

	// frames to load via ajax
	frameRoot := root + "frame/"
	mux.HandleFunc(host+frameRoot, "adminifier frame", withSecurityHeaders(func(w http.ResponseWriter, r *http.Request) {

		// check logged in
		if redirectIfNotLoggedIn(w, r) {
			return
		}

		frameNameFull := strings.TrimPrefix(r.URL.Path, frameRoot)
		frameName := frameNameFull
		if i := strings.IndexByte(frameNameFull, '/'); i != -1 {
			frameNameFull = frameName[:i+1]
			frameName = frameNameFull[:i]
		}
		tmplName := "admin-frame-" + frameName + ".tpl"

		// call func to create template params
		var dot any = nil

		if handler, exist := adminFrameHandlers[frameNameFull]; exist {
			// create wiki request
			ar := &adminRequest{w: w, r: r}
			dot = ar

			// call handler
			handler(ar)

			// handler returned an error
			if ar.err != nil {
				http.Error(w, ar.err.Error(), http.StatusInternalServerError)
				return
			}

			// handler was successful
			if ar.dot != nil {
				dot = ar.dot
			}
			if ar.tmplName != "" {
				tmplName = ar.tmplName
			}
		}

		// frame template does not exist
		if exist := tmpl.Lookup(tmplName); exist == nil {
			log.Printf("frame template not found: %s", tmplName)
			http.NotFound(w, r)
			return
		}

		// execute frame template with dot
		err := tmpl.ExecuteTemplate(w, tmplName, dot)

		// error occurred in template execution
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	// make sure that this is admin root
	if strings.TrimPrefix(r.URL.Path, root) != "" {
		http.NotFound(w, r)
		return
	}

	handleAdmin(w, r)
}

func handleAdmin(w http.ResponseWriter, r *http.Request) {
	if redirectIfNotLoggedIn(w, r) {
		return
	}

	// check if user has server access
	if !permissionChecker.HasServerPermission(r, "read.server.config") {
		http.Error(w, "insufficient permissions for server administration", http.StatusForbidden)
		return
	}

	err := tmpl.ExecuteTemplate(w, "admin.tpl", createAdminTemplate(r))

	if err != nil {
		panic(err)
	}

	// TODO: if user has only one site and no admin privs, go straight to site dashboard
	// and deny access to the server admin panel
}

func createAdminTemplate(r *http.Request) adminTemplate {
	session, _ := sessMgr.Get(r.Context(), "user").(*webserver.Session)
	var user *authenticator.User
	if session != nil {
		user = &session.User
	}
	return adminTemplate{
		User:      user,
		Wikis:     make(map[string]*webserver.WikiInfo),
		Templates: []string{},
		Title:     "quiki",
		AdminRoot: strings.TrimRight(root, "/"),
		Static:    root + "static",
		QStatic:   root + "qstatic",
	}
}

func handleConfigJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	data := createAdminTemplate(r)

	err := tmpl.ExecuteTemplate(w, "config.js.tpl", data)
	if err != nil {
		http.Error(w, "failed to generate config", http.StatusInternalServerError)
		return
	}
}

func handleLoginPage(w http.ResponseWriter, r *http.Request) {
	// if no users exist, redirect to create-user
	if len(webserver.Auth.Users) == 0 {
		http.Redirect(w, r, "create-user", http.StatusTemporaryRedirect)
		return
	}

	r.ParseForm()

	// get server title from config
	serverTitle := "quiki"
	if title, err := webserver.Conf.GetStr("server.name"); err == nil {
		serverTitle = title
	}

	data := struct {
		Title        string
		Heading      string
		Redirect     string
		Static       string
		SharedStatic string
		WikiName     string
		WikiLogo     string
		WikiTitle    string
		Error        string
		Success      string
		ShowLinks    bool
		CSRFToken    string
	}{
		Title:        serverTitle + " login",
		Heading:      serverTitle + " login",
		Redirect:     r.Form.Get("redirect"),
		Static:       root + "static",
		SharedStatic: root + "shared",
		WikiName:     serverTitle,
		WikiLogo:     "image/favicon.png",
		WikiTitle:    serverTitle,
		ShowLinks:    false,
		CSRFToken:    webserver.GetOrCreateCSRFToken(r),
	}

	handleTemplate(w, r, data)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {

	// missing parameters or malformed request
	if !parsePost(w, r, "username", "password") {
		return
	}

	username := webserver.SanitizeInput(r.Form.Get("username"))
	password := r.Form.Get("password") // don't sanitize passwords

	// validate form fields
	if err := webserver.ValidateAuthForm(username, password); err != nil {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	// validate csrf token
	csrfToken := webserver.GetOrCreateCSRFToken(r)
	if !webserver.ValidateCSRFToken(r, csrfToken) {
		http.Error(w, "security token validation failed, please try again", http.StatusBadRequest)
		return
	}

	// check modern multi-factor rate limiting
	if webserver.CheckRateLimit(r, username) {
		http.Error(w, "too many failed attempts, please try again later", http.StatusTooManyRequests)
		return
	}

	// any login attempt voids the current session
	sessMgr.Destroy(r.Context())

	// attempt login
	user, err := webserver.Auth.Login(username, password)
	if err != nil {
		webserver.HandleAuthError(w, err, r, username)
		return
	}

	// login successful - clear failed attempts
	webserver.ClearSuccessfulLogin(r, username)

	// start session and remember user info
	session := webserver.NewSession(&user)
	sessMgr.Put(r.Context(), "user", session)
	sessMgr.Put(r.Context(), "loggedIn", true)
	sessMgr.Put(r.Context(), "branch", "master") // FIXME: derive default branch

	// regenerate session id after successful login for security
	if err := sessMgr.RenewToken(r.Context()); err != nil {
		log.Printf("failed to renew session token: %v", err)
	}

	// redirect to dashboard, which is now located at adminifier root
	redirect := r.Form.Get("redirect")
	http.Redirect(w, r, path.Join(root, redirect), http.StatusTemporaryRedirect)
}

func handleCreateUserPage(w http.ResponseWriter, r *http.Request) {
	// if users exist, redirect to login
	if len(webserver.Auth.Users) != 0 {
		http.Redirect(w, r, "login", http.StatusTemporaryRedirect)
		return
	}

	// get server title from config
	serverTitle := "quiki"
	if title, err := webserver.Conf.GetStr("server.name"); err == nil {
		serverTitle = title
	}

	data := struct {
		Title        string
		Heading      string
		Redirect     string
		Static       string
		SharedStatic string
		WikiName     string
		WikiLogo     string
		WikiTitle    string
		Error        string
		Success      string
		ShowLinks    bool
		CSRFToken    string
	}{
		Title:        serverTitle + " setup wizard",
		Heading:      "Create Iniital User",
		Static:       root + "static",
		SharedStatic: root + "shared",
		WikiName:     serverTitle,
		WikiLogo:     "image/favicon.png",
		WikiTitle:    serverTitle,
		CSRFToken:    webserver.GetOrCreateCSRFToken(r),
	}

	handleTemplate(w, r, data)
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

	// validate csrf token
	csrfToken := webserver.GetOrCreateCSRFToken(r)
	if !webserver.ValidateCSRFToken(r, csrfToken) {
		http.Error(w, "security token validation failed, please try again", http.StatusBadRequest)
		return
	}

	// validate token
	expectedToken, err := webserver.Conf.GetStr("adminifier.token")
	if err != nil || expectedToken == "" {
		http.Error(w, "no token set", http.StatusInternalServerError)
		return
	}
	if r.Form.Get("token") != expectedToken {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	// remove the token from config
	err = webserver.Conf.Unset("adminifier.token")
	if err != nil {
		http.Error(w, "unsetting token in config: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = webserver.Conf.Write()
	if err != nil {
		http.Error(w, "remove token from config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// rehash server after config change
	err = webserver.Rehash()
	if err != nil {
		http.Error(w, "rehash server: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// create user using the modern interface
	username := r.Form.Get("username")
	user, err := webserver.Auth.CreateUser(username, r.Form.Get("password"))

	// error occurred
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// set admin details on the created user
	user.DisplayName = r.Form.Get("display")
	user.Email = r.Form.Get("email")
	user.Permissions = []string{"read.*", "write.*"} // first user gets full admin permissions
	err = webserver.Auth.UpdateUser(username, user)
	if err != nil {
		http.Error(w, "failed to set user details: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// log 'em in by simulating a request to /func/login
	handleLogin(w, r)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	// get username before destroying session for cache cleanup
	var username string
	if session, ok := sessMgr.Get(r.Context(), "user").(*webserver.Session); ok && session != nil && session.Username != "" {
		username = session.Username
	}

	// destroy session
	sessMgr.Destroy(r.Context())

	// clear permission cache if we had a user
	if username != "" {
		permissionChecker.ClearPermissionCache(r, username)
	}

	// redirect to login
	http.Redirect(w, r, root+"login", http.StatusTemporaryRedirect)
}

func handleSitesFrame(ar *adminRequest) {
	// check if user can view wikis
	if !permissionChecker.HasServerPermission(ar.r, "read.server.wikis") {
		ar.err = errors.New("insufficient permissions to view wikis")
		return
	}

	ar.dot = struct {
		Wikis         map[string]*webserver.WikiInfo
		Templates     []string
		BaseWikis     []string
		WikiDelimeter string
		adminTemplate
	}{
		Wikis:         webserver.Wikis,
		Templates:     webserver.TemplateNames(),
		BaseWikis:     wiki.AvailableBaseWikis(),
		WikiDelimeter: wikiDelimeter,
		adminTemplate: createAdminTemplate(ar.r),
	}
}

func handleRoutesFrame(ar *adminRequest) {
	ar.dot = struct {
		Routes []router.Route
		adminTemplate
	}{
		Routes:        webserver.Router.Routes(),
		adminTemplate: createAdminTemplate(ar.r),
	}
}

func handleAdminHelpFrame(ar *adminRequest) {
	ar.dot, ar.err = handleHelpFrame(root, ar.w, ar.r)
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

func redirectIfNotLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	// if not logged in, temp redirect to login page
	if !sessMgr.GetBool(r.Context(), "loggedIn") {
		redirect := strings.TrimPrefix(r.URL.Path, root)
		if r.URL.RawQuery != "" {
			redirect += url.QueryEscape("?" + r.URL.RawQuery)
		}
		if redirect == "/" {
			redirect = ""
		} else if redirect != "" {
			redirect = "?redirect=/" + redirect
		}
		http.Redirect(w, r, root+"login"+redirect, http.StatusTemporaryRedirect)
		return true
	}
	return false
}
