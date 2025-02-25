package adminifier

import (
	"log"
	"net/http"
	"strings"

	"github.com/cooper/quiki/authenticator"
	"github.com/cooper/quiki/webserver"
)

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

var adminFrameHandlers = map[string]func(*adminRequest){
	"sites": handleSitesFrame,
	"help":  handleAdminHelpFrame,
	"help/": handleAdminHelpFrame,
}

type adminTemplate struct {
	User        *authenticator.User
	Wikis       map[string]*webserver.WikiInfo
	Templates   []string
	Title       string // server title
	Static      string // adminifier static root
	QStatic     string // webserver static root
	AdminRoot   string // adminifier root'
	JSTemplates string
}

type adminRequest struct {
	w        http.ResponseWriter
	r        *http.Request
	tmplName string
	dot      any
	err      error
}

func setupAdminHandlers() {
	for name, function := range adminUnauthenticatedHandlers {
		mux.HandleFunc(host+root+name, function)
	}
	for name, function := range adminUnauthenticatedFuncHandlers {
		mux.HandleFunc(host+root+"func/"+name, function)
	}

	// authenticated handlers

	// each of these generates admin.tpl
	for which := range adminFrameHandlers {
		mux.HandleFunc(host+root+which, handleAdmin)
	}

	// frames to load via ajax
	frameRoot := root + "frame/"
	mux.HandleFunc(host+frameRoot, func(w http.ResponseWriter, r *http.Request) {

		// check logged in
		if !sessMgr.GetBool(r.Context(), "loggedIn") {
			http.Redirect(w, r, root+"login", http.StatusTemporaryRedirect)
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
	})
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

	// if not logged in, temp redirect to login page
	if !sessMgr.GetBool(r.Context(), "loggedIn") {
		http.Redirect(w, r, root+"login", http.StatusTemporaryRedirect)
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
	return adminTemplate{
		User:        sessMgr.Get(r.Context(), "user").(*authenticator.User),
		Title:       "quiki",
		AdminRoot:   strings.TrimRight(root, "/"),
		Static:      root + "static",
		QStatic:     root + "qstatic",
		JSTemplates: "",
	}
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

	handleTemplate(w, r)
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

	// log 'em in by simulating a request to /func/login
	handleLogin(w, r)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	// destory session
	sessMgr.Destroy(r.Context())

	// redirect to login
	http.Redirect(w, r, root+"login", http.StatusTemporaryRedirect)
}

func handleSitesFrame(ar *adminRequest) {
	ar.dot = struct {
		Wikis     map[string]*webserver.WikiInfo
		Templates []string
		adminTemplate
	}{
		Wikis:         webserver.Wikis,
		Templates:     webserver.TemplateNames(),
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
