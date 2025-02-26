package adminifier

import (
	"encoding/json"
	"html/template"
	"io/fs"
	"log"
	"maps"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/cooper/quiki/authenticator"
	"github.com/cooper/quiki/resources"
	"github.com/cooper/quiki/webserver"
	"github.com/cooper/quiki/wiki"
	"github.com/cooper/quiki/wikifier"
	"github.com/pkg/errors"
)

var javascriptTemplates string

var frameHandlers = map[string]func(*wikiRequest){
	"dashboard":     handleDashboardFrame,
	"pages":         handlePagesFrame,
	"categories":    handleCategoriesFrame,
	"images":        handleImagesFrame,
	"models":        handleModelsFrame,
	"settings":      handleSettingsFrame,
	"edit-page":     handleEditPageFrame,
	"edit-category": handleEditCategoryFrame,
	"edit-model":    handleEditModelFrame,
	"switch-branch": handleSwitchBranchFrame,
	"help":          handleWikiHelpFrame,
	"help/":         handleWikiHelpFrame,
}

var wikiFuncHandlers = map[string]func(*wikiRequest){
	"switch-branch/": handleSwitchBranch,
	"create-branch":  handleCreateBranch,
	"write-page":     handleWritePage,
	"write-model":    handleWriteModel,
	"write-config":   handleWriteWikiConfig,
	"image/":         handleImage,
}

// wikiTemplate members are available to all wiki templates
type wikiTemplate struct {
	User              *authenticator.User  // user
	ServerPanelAccess bool                 // whether user can access main panel
	Shortcode         string               // wiki shortcode
	Title             string               // wiki title
	Branch            string               // selected branch
	Static            string               // adminifier static root
	QStatic           string               // webserver static root
	AdminRoot         string               // adminifier root
	Root              string               // wiki root
	WikiRoots         wikifier.PageOptRoot // wiki roots
}

type wikiRequest struct {
	shortcode string
	wikiRoot  string
	wi        *webserver.WikiInfo
	w         http.ResponseWriter
	r         *http.Request
	tmplName  string
	dot       any
	err       error
}

type editorOpts struct {
	page   bool // true if editing a page
	model  bool // true if editing a model
	config bool // true if editing the config
	cat    bool // true if editing a category
	info   any  // PageInfo or ModelInfo
}

var loadedWikiShortcodes = make(map[string]bool)

// TODO: verify session on ALL wiki handlers

func initWikis() {
	for shortcode, wi := range webserver.Wikis {
		if loadedWikiShortcodes[shortcode] {
			continue
		}
		setupWikiHandlers(shortcode, wi)
		loadedWikiShortcodes[shortcode] = true
	}
}

func setupWikiHandlers(shortcode string, wi *webserver.WikiInfo) {
	wikiRoot := root + "sites/" + shortcode + "/"

	// each of these URLs generates wiki.tpl
	for which := range frameHandlers {
		mux.HandleFunc(host+wikiRoot+which, func(w http.ResponseWriter, r *http.Request) {
			handleWiki(shortcode, wi, w, r)
		})
	}

	// frames to load via ajax
	frameRoot := wikiRoot + "frame/"
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
		tmplName := "frame-" + frameName + ".tpl"

		// call func to create template params
		var dot any = nil

		if handler, exist := frameHandlers[frameNameFull]; exist {
			// create wiki request
			wr := &wikiRequest{
				shortcode: shortcode,
				wikiRoot:  wikiRoot,
				w:         w,
				r:         r,
			}
			dot = wr

			// possibly switch wikis
			switchUserWiki(wr, wi)
			if wr.err != nil {
				// FIXME: don't panic
				panic(wr.err)
			}

			// call handler
			handler(wr)

			// handler returned an error
			if wr.err != nil {
				http.Error(w, wr.err.Error(), http.StatusInternalServerError)
				return
			}

			// handler was successful
			if wr.dot != nil {
				dot = wr.dot
			}
			if wr.tmplName != "" {
				tmplName = wr.tmplName
			}
		}

		// frame template does not exist
		if exist := tmpl.Lookup(tmplName); exist == nil {
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

	// functions
	funcRoot := wikiRoot + "func/"
	for funcName, thisHandler := range wikiFuncHandlers {
		handler := thisHandler
		mux.HandleFunc(host+funcRoot+funcName, func(w http.ResponseWriter, r *http.Request) {

			// check logged in
			//
			// TODO: everything in func/ will be JSON,
			// so return a "not logged in" error to present login popup
			// rather than redirecting
			//
			if !sessMgr.GetBool(r.Context(), "loggedIn") {
				http.Redirect(w, r, root+"login", http.StatusTemporaryRedirect)
				return
			}

			// create wiki request
			wr := &wikiRequest{
				shortcode: shortcode,
				wikiRoot:  wikiRoot,
				w:         w,
				r:         r,
			}

			// possibly switch wikis
			switchUserWiki(wr, wi)
			if wr.err != nil {
				// FIXME: don't panic
				panic(wr.err)
			}

			// call handler
			handler(wr)

			// handler returned an error
			if wr.err != nil {
				// FIXME: don't panic
				panic(wr.err)
			}
		})
	}
}

func handleWiki(shortcode string, wi *webserver.WikiInfo, w http.ResponseWriter, r *http.Request) {

	// check logged in
	if !sessMgr.GetBool(r.Context(), "loggedIn") {
		http.Redirect(w, r, root+"login", http.StatusTemporaryRedirect)
		return
	}

	// load javascript templates
	if javascriptTemplates == "" {
		files, err := fs.ReadDir(resources.Adminifier, "js-tmpl")
		if err != nil {
			log.Printf("error reading js-tmpl directory: %v", err)
		}
		for _, file := range files {
			data, err := fs.ReadFile(resources.Adminifier, "js-tmpl/"+file.Name())
			if err != nil {
				log.Printf("error reading js-tmpl file %s: %v", file.Name(), err)
			}
			javascriptTemplates += string(data)
		}
	}

	err := tmpl.ExecuteTemplate(w, "wiki.tpl", struct {
		JSTemplates template.HTML
		wikiTemplate
	}{
		template.HTML(javascriptTemplates),
		getGenericTemplate(&wikiRequest{
			shortcode: shortcode,
			wi:        wi,
			w:         w,
			r:         r,
		}),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		panic(err)
	}
}

func handleDashboardFrame(wr *wikiRequest) {

	// wiki logs
	logs, _ := os.ReadFile(wr.wi.Dir("cache", "wiki.log"))

	// pages with errors and warnings
	var errors []wikifier.PageInfo
	var warnings []wikifier.PageInfo
	for _, info := range wr.wi.PagesSorted(false, wiki.SortModified, wiki.SortTitle) {
		if info.Error != nil {
			errors = append(errors, info)
		}
		if info.Warnings != nil {
			warnings = append(warnings, info)
		}
	}

	wr.dot = struct {
		Logs     string
		Errors   []wikifier.PageInfo
		Warnings []wikifier.PageInfo
	}{
		Logs:     string(logs),
		Errors:   errors,
		Warnings: warnings,
	}
}

var sorters map[string]wiki.SortFunc = map[string]wiki.SortFunc{
	"t": wiki.SortTitle,
	"a": wiki.SortAuthor,
	"c": wiki.SortCreated,
	"m": wiki.SortModified,
	"d": wiki.SortDimensions,
}

// find sort from query
func getSortFunc(wr *wikiRequest) (bool, wiki.SortFunc) {
	descending := true
	sortFunc := wiki.SortModified
	s := wr.r.URL.Query().Get("sort")
	if len(s) != 0 {
		sortFunc = sorters[string(s[0])]
		descending = len(s) > 1 && s[1] == '-'
	}

	return descending, sortFunc
}

func handlePagesFrame(wr *wikiRequest) {
	descending, sortFunc := getSortFunc(wr)
	pages := wr.wi.PagesSorted(descending, sortFunc, wiki.SortTitle)
	handleFileFrames(wr, pages)
}

func handleImagesFrame(wr *wikiRequest) {
	descending, sortFunc := getSortFunc(wr)
	images := wr.wi.ImagesSorted(descending, sortFunc, wiki.SortTitle)
	handleFileFrames(wr, images, "d")
}

func handleModelsFrame(wr *wikiRequest) {
	descending, sortFunc := getSortFunc(wr)
	models := wr.wi.ModelsSorted(descending, sortFunc, wiki.SortTitle)
	handleFileFrames(wr, models)
}

func handleCategoriesFrame(wr *wikiRequest) {
	descending, sortFunc := getSortFunc(wr)
	cats := wr.wi.CategoriesSorted(descending, sortFunc, wiki.SortTitle)
	handleFileFrames(wr, cats)
}

func handleFileFrames(wr *wikiRequest, results any, extras ...string) {

	// json stuffs
	res, err := json.Marshal(map[string]any{
		"sort_types": append([]string{"t", "a", "c", "m"}, extras...),
		"results":    results,
	})
	if err != nil {
		wr.err = err
		return
	}

	// determine sort
	// consider: should we validate sort here also
	s := wr.r.URL.Query().Get("sort")
	if s == "" {
		s = "m-"
	}

	wr.dot = struct {
		JSON  template.HTML
		Order string // sort
		List  bool   // for images, show file list rather than grid
		wikiTemplate
	}{
		JSON:         template.HTML("<!--JSON\n" + string(res) + "\n-->"),
		Order:        s,
		List:         wr.r.URL.Query().Get("mode") == "list",
		wikiTemplate: getGenericTemplate(wr),
	}
}

func handleSettingsFrame(wr *wikiRequest) {
	// serve editor for the config file
	handleEditor(wr, wr.wi.ConfigFile, "wiki.conf", "Configuration file", editorOpts{config: true})
}

func handleEditPageFrame(wr *wikiRequest) {
	q := wr.r.URL.Query()

	// no page filename provided
	name := q.Get("page")
	if name == "" {
		wr.err = errors.New("no page filename provided")
		return
	}

	// find the page. if File is empty, it doesn't exist
	info := wr.wi.PageInfo(name)
	if info.File == "" {
		wr.err = errors.New("page does not exist")
		return
	}

	// serve editor
	handleEditor(wr, info.Path, info.File, info.Title, editorOpts{page: true, info: info})
}

func handleEditModelFrame(wr *wikiRequest) {
	q := wr.r.URL.Query()

	// no page filename provided
	name := q.Get("page")
	if name == "" {
		wr.err = errors.New("no model filename provided")
		return
	}

	// find the model. if File is empty, it doesn't exist
	info := wr.wi.ModelInfo(name)
	if info.File == "" {
		wr.err = errors.New("model does not exist")
		return
	}

	// serve editor
	handleEditor(wr, info.Path, info.File, info.Title, editorOpts{model: true, info: info})
}

func handleEditCategoryFrame(wr *wikiRequest) {
	q := wr.r.URL.Query()

	// no page filename provided
	name := q.Get("cat")
	if name == "" {
		wr.err = errors.New("no page filename provided")
		return
	}

	// find the category
	metaPath := wr.wi.PathForCategory(name, false)
	info := wr.wi.CategoryInfo(name)
	if !info.Exists() {
		wr.err = errors.New("category does not exist")
		return
	}

	// serve editor
	handleEditor(wr, metaPath, info.File, info.Title, editorOpts{cat: true, info: info})
}

func handleWikiHelpFrame(wr *wikiRequest) {
	wr.dot, wr.err = handleHelpFrame(wr.wikiRoot, wr.w, wr.r)
}

func handleEditor(wr *wikiRequest, path, file, title string, o editorOpts) {
	wr.tmplName = "frame-editor.tpl"

	// call DisplayFile to get the content
	var fileRes wiki.DisplayFile
	switch r := wr.wi.DisplayFile(path).(type) {
	case wiki.DisplayFile:
		fileRes = r
	case wiki.DisplayError:
		wr.err = errors.New(r.DetailedError)
		return
	default:
		wr.err = errors.New("unknown error occurred in DisplayFile")
		return
	}

	// json stuff
	jsonData, err := json.Marshal(struct {
		Page     bool `json:"page"`
		Model    bool `json:"model"`
		Config   bool `json:"config"`
		Category bool `json:"category"`
		Info     any  `json:"info,omitempty"` // PageInfo or ModelInfo
		wiki.DisplayFile
	}{
		Page:        o.page,
		Model:       o.model,
		Config:      o.config,
		Category:    o.cat,
		Info:        o.info,
		DisplayFile: fileRes,
	})
	if err != nil {
		wr.err = err
		return
	}

	// template stuff
	wr.dot = struct {
		Found    bool
		JSON     template.HTML
		Page     bool   // true if editing a page
		Model    bool   // true if editing a model
		Config   bool   // true if editing config
		Category bool   // true if editing a category
		Info     any    // PageInfo, ModelInfo, or CategoryInfo
		Title    string // page title or filename
		File     string // filename
		Content  string // file content
		wikiTemplate
	}{
		Found:        true,
		JSON:         template.HTML("<!--JSON\n" + string(jsonData) + "\n-->"),
		Page:         o.page,
		Model:        o.model,
		Config:       o.config,
		Category:     o.cat,
		Info:         o.info,
		Title:        title,
		File:         file,
		Content:      fileRes.Content,
		wikiTemplate: getGenericTemplate(wr),
	}
}

func handleSwitchBranchFrame(wr *wikiRequest) {
	branches, err := wr.wi.BranchNames()
	if err != nil {
		wr.err = err
		return
	}
	wr.dot = struct {
		Branches []string
		wikiTemplate
	}{
		Branches:     branches,
		wikiTemplate: getGenericTemplate(wr),
	}
}

func handleSwitchBranch(wr *wikiRequest) {
	branchName := strings.TrimPrefix(wr.r.URL.Path, wr.wikiRoot+"func/switch-branch/")
	if branchName == "" {
		wr.err = errors.New("no branch selected")
		return
	}

	// bad branch name
	if !wiki.ValidBranchName(branchName) {
		wr.err = errors.New("invalid branch name: " + branchName)
		return
	}

	// fetch the branch
	_, wr.err = wr.wi.Branch(branchName)
	if wr.err != nil {
		return
	}

	// set branch
	sessMgr.Put(wr.r.Context(), "branch", branchName)

	// TODO: when this request is submitted by JS, the UI can just reload
	// the current frame so the user stays on the same page, just in new branch

	// redirect back to dashboard
	http.Redirect(wr.w, wr.r, wr.wikiRoot+"dashboard", http.StatusTemporaryRedirect)
}

func handleCreateBranch(wr *wikiRequest) {

	// TODO: need a different version of parsePost that returns JSON errors
	if !parsePost(wr.w, wr.r, "branch") {
		return
	}

	// bad branch name
	branchName := wr.r.Form.Get("branch")
	if !wiki.ValidBranchName(branchName) {
		wr.err = errors.New("invalid branch name: " + branchName)
		return
	}

	// create or switch branches
	_, err := wr.wi.NewBranch(branchName)
	if err != nil {
		wr.err = err
		return
	}
	sessMgr.Put(wr.r.Context(), "branch", branchName)

	// redirect back to dashboard
	http.Redirect(wr.w, wr.r, wr.wikiRoot+"dashboard", http.StatusTemporaryRedirect)
}

func handleWritePage(wr *wikiRequest) {
	if !parsePost(wr.w, wr.r, "name", "content") {
		return
	}

	pageName, content, message := wr.r.Form.Get("name"), wr.r.Form.Get("content"), wr.r.Form.Get("message")

	// write the page
	res := handleWriteFile(wr, func() error {
		return wr.wi.WritePage(pageName, []byte(content), true, getCommitOpts(wr, message))
	})

	// display the page
	var warnings []wikifier.Warning
	var pageErr *wikifier.Warning
	switch res := wr.wi.DisplayPage(pageName).(type) {
	case wiki.DisplayError:
		err := res.ErrorAsWarning()
		pageErr = &err
	case wiki.DisplayPage:
		warnings = res.Warnings
	}

	maps.Copy(res, map[string]any{
		"warnings":     warnings,
		"displayError": pageErr,
	})
	json.NewEncoder(wr.w).Encode(res)
}

func handleWriteModel(wr *wikiRequest) {
	if !parsePost(wr.w, wr.r, "name", "content") {
		return
	}
	modelName, content, message := wr.r.Form.Get("name"), wr.r.Form.Get("content"), wr.r.Form.Get("message")
	res := handleWriteFile(wr, func() error {
		return wr.wi.WriteModel(modelName, []byte(content), true, getCommitOpts(wr, message))
	})
	json.NewEncoder(wr.w).Encode(res)
}

func handleWriteWikiConfig(wr *wikiRequest) {
	if !parsePost(wr.w, wr.r, "content") {
		return
	}
	content, message := wr.r.Form.Get("content"), wr.r.Form.Get("message")
	res := handleWriteFile(wr, func() error {
		return wr.wi.WriteConfig([]byte(content), getCommitOpts(wr, message))
	})
	json.NewEncoder(wr.w).Encode(res)
}

func handleWriteFile(wr *wikiRequest, writeFunc func() error) map[string]any {
	if err := writeFunc(); err != nil {
		return map[string]any{
			"success": false,
			"error":   err.Error(),
		}
	}

	// fetch latest commit hash
	hash, err := wr.wi.GetLatestCommitHash()
	if err != nil {
		log.Printf("error getting latest commit hash: %v", err)
	}

	return map[string]any{
		"success":       true,
		"revLatestHash": hash,
	}
}

func handleImage(wr *wikiRequest) {
	imageName := strings.TrimPrefix(wr.r.URL.Path, wr.wikiRoot+"func/image/")
	si := wiki.SizedImageFromName(imageName)

	// get dimensions from query params
	w, _ := strconv.Atoi(wr.r.URL.Query().Get("width"))
	h, _ := strconv.Atoi(wr.r.URL.Query().Get("height"))
	if w != 0 {
		si.Width = w
	}
	if h != 0 {
		si.Height = 0
	}

	// display/generate image
	switch res := wr.wi.DisplaySizedImageGenerate(si, true).(type) {

	// image
	case wiki.DisplayImage:
		http.ServeFile(wr.w, wr.r, res.Path)

	// redirect to true image name
	case wiki.DisplayRedirect:
		http.Redirect(wr.w, wr.r, res.Redirect, http.StatusMovedPermanently)

	// error/other
	default:
		http.NotFound(wr.w, wr.r)
	}
}

// possibly switch wiki branches
func switchUserWiki(wr *wikiRequest, wi *webserver.WikiInfo) {
	userWiki := wi
	branchName := sessMgr.GetString(wr.r.Context(), "branch")
	if branchName != "" {
		branchWiki, err := wi.Branch(branchName)
		if err != nil {
			wr.err = err
			return
		}
		userWiki = wi.Copy(branchWiki)
	}
	wr.wi = userWiki
}

func getGenericTemplate(wr *wikiRequest) wikiTemplate {

	// prepend external root to all wiki roots
	roots := wr.wi.Opt.Root
	roots = wikifier.PageOptRoot{
		Wiki:     path.Join(roots.Ext, roots.Wiki),
		Image:    path.Join(roots.Ext, roots.Image),
		Category: path.Join(roots.Ext, roots.Category),
		Page:     path.Join(roots.Ext, roots.Page),
		File:     path.Join(roots.Ext, roots.File),
		Ext:      roots.Ext,
	}

	return wikiTemplate{
		User:              sessMgr.Get(wr.r.Context(), "user").(*authenticator.User),
		ServerPanelAccess: true, // TODO
		Branch:            sessMgr.GetString(wr.r.Context(), "branch"),
		Shortcode:         wr.shortcode,
		Title:             wr.wi.Title,
		AdminRoot:         strings.TrimRight(root, "/"),
		Static:            root + "static",
		QStatic:           root + "qstatic",
		Root:              root + "sites/" + wr.shortcode,
		WikiRoots:         roots,
	}
}

func getCommitOpts(wr *wikiRequest, comment string) wiki.CommitOpts {
	user := sessMgr.Get(wr.r.Context(), "user").(*authenticator.User)
	return wiki.CommitOpts{
		Comment: comment,
		Name:    user.DisplayName,
		Email:   user.Email,
	}
}
