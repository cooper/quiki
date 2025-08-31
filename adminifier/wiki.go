package adminifier

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"maps"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/cooper/quiki/authenticator"
	"github.com/cooper/quiki/webserver"
	"github.com/cooper/quiki/wiki"
	"github.com/cooper/quiki/wikifier"
	"github.com/pkg/errors"
)

var wikiFrameHandlers = map[string]func(*wikiRequest){
	// "switch-branch": handleSwitchBranchFrame,
	"dashboard":     handleDashboardFrame,
	"pages":         handlePagesFrame,
	"pages/":        handlePagesFrame,
	"categories":    handleCategoriesFrame,
	"images":        handleImagesFrame,
	"images/":       handleImagesFrame,
	"models":        handleModelsFrame,
	"models/":       handleModelsFrame,
	"settings":      handleSettingsFrame,
	"edit-page":     handleEditPageFrame,
	"edit-category": handleEditCategoryFrame,
	"edit-model":    handleEditModelFrame,
	"help":          handleWikiHelpFrame,
	"help/":         handleWikiHelpFrame,
}

var wikiFuncHandlers = map[string]func(*wikiRequest){
	// "switch-branch/":      handleSwitchBranch,
	// "create-branch":       handleCreateBranch,
	"write-page":          handleWritePage,
	"write-model":         handleWriteModel,
	"write-config":        handleWriteWikiConfig,
	"image/":              handleImage,
	"page-revisions":      handlePageRevisions,
	"page-diff":           handlePageDiff,
	"create-page":         handleCreatePage,
	"create-model":        handleCreateModel,
	"create-page-folder":  handleCreatePageFolder,
	"create-model-folder": handleCreateModelFolder,
	"create-image-folder": handleCreateImageFolder,
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

// canUserReadWiki checks if a user can access a wiki for reading
func canUserReadWiki(r *http.Request, shortcode string, wi *webserver.WikiInfo) bool {
	// if wiki doesn't require auth, anyone can access it
	if !wi.RequireAuth {
		return true
	}

	// wiki requires auth, check if user is logged in
	if !sessMgr.GetBool(r.Context(), "loggedIn") {
		return false
	}

	// user is logged in, check if they have permission to read this wiki
	return permissionChecker.HasWikiPermission(r, shortcode, "read.wiki")
}

// canUserWriteWiki checks if a user can edit a wiki
func canUserWriteWiki(r *http.Request, shortcode string) bool {
	// editing always requires login
	if !sessMgr.GetBool(r.Context(), "loggedIn") {
		return false
	}

	// check if user has write permission for this wiki
	return permissionChecker.HasWikiPermission(r, shortcode, "write.wiki")
}

type editorOpts struct {
	page   bool // true if editing a page
	model  bool // true if editing a model
	config bool // true if editing the config
	cat    bool // true if editing a category
	info   any  // PageInfo or ModelInfo
}

var loadedWikiShortcodes = make(map[string]bool)

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
	wikiRoot := root + wikiDelimeter + shortcode

	// each of these URLs generates wiki.tpl
	mux.RegisterFunc(host+wikiRoot, "adminifier site root", func(w http.ResponseWriter, r *http.Request) {
		handleWikiRoot(shortcode, wi, w, r)
	})
	wikiRoot += "/"
	mux.HandleFunc(host+wikiRoot, func(w http.ResponseWriter, r *http.Request) {
		handleWikiRoot(shortcode, wi, w, r)
	})
	for which := range wikiFrameHandlers {
		mux.HandleFunc(host+wikiRoot+which, func(w http.ResponseWriter, r *http.Request) {
			handleWiki(shortcode, wi, w, r)
		})
	}

	// frames to load via ajax
	frameRoot := wikiRoot + "frame/"
	mux.HandleFunc(host+frameRoot, func(w http.ResponseWriter, r *http.Request) {

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
		tmplName := "frame-" + frameName + ".tpl"

		// call func to create template params
		var dot any = nil

		if handler, exist := wikiFrameHandlers[frameNameFull]; exist {
			// create wiki request
			wr := &wikiRequest{
				shortcode: shortcode,
				wikiRoot:  wikiRoot,
				w:         w,
				r:         r,
			}
			dot = wr

			wr.wi = wi
			// DISABLED: possibly switch wikis
			// switchUserWiki(wr, wi)
			// if wr.err != nil {
			// 	// FIXME: don't panic
			// 	panic(wr.err)
			// }

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

			// check if user can edit this wiki
			//
			// TODO: everything in func/ will be JSON,
			// so return a "not logged in" error to present login popup
			// rather than redirecting
			//
			if !canUserWriteWiki(r, shortcode) {
				// if not logged in, redirect to login
				if !sessMgr.GetBool(r.Context(), "loggedIn") {
					redirectIfNotLoggedIn(w, r)
				} else {
					// logged in but no permission
					http.Error(w, "insufficient permissions to edit this wiki", http.StatusForbidden)
				}
				return
			}

			// create wiki request
			wr := &wikiRequest{
				shortcode: shortcode,
				wikiRoot:  wikiRoot,
				w:         w,
				r:         r,
			}

			wr.wi = wi
			// DISABLED: possibly switch wikis
			// switchUserWiki(wr, wi)
			// if wr.err != nil {
			// 	// FIXME: don't panic
			// 	panic(wr.err)
			// }

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

func handleWikiRoot(shortcode string, wi *webserver.WikiInfo, w http.ResponseWriter, r *http.Request) {
	// ensure it is the root
	wikiRoot := root + wikiDelimeter + shortcode
	if r.URL.Path != wikiRoot && r.URL.Path != wikiRoot+"/" {
		http.NotFound(w, r)
		return
	}
	handleWiki(shortcode, wi, w, r)
}

func handleWiki(shortcode string, wi *webserver.WikiInfo, w http.ResponseWriter, r *http.Request) {
	// check if user can access this wiki (public vs private wiki)
	if !canUserReadWiki(r, shortcode, wi) {
		// wiki requires auth but user not logged in; redirect to login
		if wi.RequireAuth && !sessMgr.GetBool(r.Context(), "loggedIn") {
			redirectIfNotLoggedIn(w, r)
		} else {
			// either logged in but no permission, or wiki is public (shouldn't happen)
			http.Error(w, "insufficient permissions to view this wiki", http.StatusForbidden)
		}
		return
	}

	err := tmpl.ExecuteTemplate(w, "wiki.tpl", getGenericTemplate(&wikiRequest{
		shortcode: shortcode,
		wi:        wi,
		w:         w,
		r:         r,
	}))
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
func getSortFunc(wr *wikiRequest, defaultFunc wiki.SortFunc, defaultDescending bool) (bool, wiki.SortFunc) {
	descending := defaultDescending
	sortFunc := defaultFunc
	s := wr.r.URL.Query().Get("sort")
	if len(s) != 0 {
		sortFunc = sorters[string(s[0])]
		descending = len(s) > 1 && s[1] == '-'
	}

	return descending, sortFunc
}

func handlePagesFrame(wr *wikiRequest) {
	descending, sortFunc := getSortFunc(wr, wiki.SortModified, true)
	dir := strings.TrimPrefix(strings.TrimPrefix(wr.r.URL.Path, wr.wikiRoot+"frame/pages"), "/")
	pages, dirs := wr.wi.PagesAndDirsSorted(dir, descending, sortFunc, wiki.SortTitle)
	handleFileFrames(wr, "pages", struct {
		Pages []wikifier.PageInfo `json:"pages"`
		Dirs  []string            `json:"dirs"`
		Cd    string              `json:"cd"`
	}{pages, dirs, dir})
}

func handleImagesFrame(wr *wikiRequest) {
	descending, sortFunc := getSortFunc(wr, wiki.SortTitle, false)
	dir := strings.TrimPrefix(strings.TrimPrefix(wr.r.URL.Path, wr.wikiRoot+"frame/images"), "/")

	// get images asynchronously to prevent blocking on large image operations
	imagesCh := make(chan []wiki.ImageInfo, 1)
	dirsCh := make(chan []string, 1)
	errCh := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				errCh <- fmt.Errorf("panic while loading images: %v", r)
			}
		}()

		images, dirs := wr.wi.ImagesAndDirsSorted(dir, descending, sortFunc, wiki.SortTitle)
		imagesCh <- images
		dirsCh <- dirs
	}()

	// wait for results with timeout
	select {
	case images := <-imagesCh:
		dirs := <-dirsCh
		handleFileFrames(wr, "images", struct {
			Images []wiki.ImageInfo `json:"images"`
			Dirs   []string         `json:"dirs"`
			Cd     string           `json:"cd"`
		}{images, dirs, dir}, "d")
	case err := <-errCh:
		wr.err = err
	case <-time.After(30 * time.Second):
		wr.err = fmt.Errorf("timeout loading images in directory: %s", dir)
	}
}

func handleModelsFrame(wr *wikiRequest) {
	descending, sortFunc := getSortFunc(wr, wiki.SortModified, true)
	dir := strings.TrimPrefix(strings.TrimPrefix(wr.r.URL.Path, wr.wikiRoot+"frame/models"), "/")
	models, dirs := wr.wi.ModelsAndDirsSorted(dir, descending, sortFunc, wiki.SortTitle)
	handleFileFrames(wr, "models", struct {
		Models []wikifier.ModelInfo `json:"models"`
		Dirs   []string             `json:"dirs"`
		Cd     string               `json:"cd"`
	}{models, dirs, dir})
}

func handleCategoriesFrame(wr *wikiRequest) {
	descending, sortFunc := getSortFunc(wr, wiki.SortModified, true)
	cats := wr.wi.CategoriesSorted(descending, sortFunc, wiki.SortTitle)
	handleFileFrames(wr, "categories", cats)
}

func handleFileFrames(wr *wikiRequest, typ string, results any, extras ...string) {

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
		if typ == "images" {
			s = "t+"
		} else {
			s = "m-"
		}
	}

	wr.dot = struct {
		JSON  template.HTML
		Order string // sort
		List  bool   // for images, show file list rather than grid
		Cd    string
		wikiTemplate
	}{
		JSON:         template.HTML("<!--JSON\n" + string(res) + "\n-->"),
		Order:        s,
		List:         wr.r.URL.Query().Get("mode") == "list",
		Cd:           strings.TrimPrefix(strings.TrimPrefix(wr.r.URL.Path, wr.wikiRoot+"frame/"+typ), "/"),
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
	display := wr.wi.DisplayPageDraft(pageName, true)
	switch d := display.(type) {
	case wiki.DisplayError:
		err := d.ErrorAsWarning()
		pageErr = &err
	case wiki.DisplayPage:
		warnings = d.Warnings
	}

	maps.Copy(res, map[string]any{
		"warnings":      warnings,
		"displayError":  pageErr,
		"displayResult": display,
		"displayType":   wiki.DisplayType(display),
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
		si.Height = h
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

func handlePageRevisions(wr *wikiRequest) {
	if !parsePost(wr.w, wr.r, "page") {
		return
	}

	pageName := wr.r.Form.Get("page")
	revisions, err := wr.wi.RevisionsMatchingPage(pageName)
	if err != nil {
		wr.err = err
		return
	}

	jsonWriter := json.NewEncoder(wr.w)
	wr.err = jsonWriter.Encode(map[string]any{
		"success": true,
		"revs":    revisions,
	})
}

func handlePageDiff(wr *wikiRequest) {
	if !parsePost(wr.w, wr.r, "page", "from") {
		return
	}

	_, from, to := wr.r.Form.Get("page"), wr.r.Form.Get("from"), wr.r.Form.Get("to")
	diff, err := wr.wi.Diff(from, to)
	if err != nil {
		wr.err = err
		return
	}

	jsonWriter := json.NewEncoder(wr.w)
	wr.err = jsonWriter.Encode(map[string]any{
		"success": true,
		"diff":    diff.String(),
	})
}

func handleCreatePage(wr *wikiRequest) {
	handleCreate("page", wr, func(dir, title string) (string, error) {
		return wr.wi.CreatePage(dir, title, nil, getCommitOpts(wr, "Create page: "+title))
	})
}

func handleCreatePageFolder(wr *wikiRequest) {
	handleCreateFolder("page", wr, func(dir, name string) (string, error) {
		return wr.wi.CreatePageFolder(dir, name)
	})
}

func handleCreateImageFolder(wr *wikiRequest) {
	handleCreateFolder("image", wr, func(dir, name string) (string, error) {
		return wr.wi.CreateImageFolder(dir, name)
	})
}

func handleCreateModelFolder(wr *wikiRequest) {
	handleCreateFolder("model", wr, func(dir, name string) (string, error) {
		return wr.wi.CreateModelFolder(dir, name)
	})
}

func handleCreateFolder(typ string, wr *wikiRequest, createFunc func(dir, name string) (string, error)) {
	if !parsePost(wr.w, wr.r, "name") {
		return
	}

	name, dir := wr.r.Form.Get("name"), wr.r.Form.Get("dir")
	_, wr.err = createFunc(dir, name)
	if wr.err != nil {
		return
	}

	// redirect to list
	http.Redirect(wr.w, wr.r, path.Join(wr.wikiRoot+typ+"s", dir), http.StatusTemporaryRedirect)
}

func handleCreateModel(wr *wikiRequest) {
	handleCreate("model", wr, func(dir, title string) (string, error) {
		return wr.wi.CreateModel(title, nil, getCommitOpts(wr, "Create model: "+title))
	})
}

func handleCreate(typ string, wr *wikiRequest, createFunc func(dir, title string) (string, error)) {
	if !parsePost(wr.w, wr.r, "title") {
		return
	}

	title, dir := wr.r.Form.Get("title"), wr.r.Form.Get("dir")
	var filename string
	filename, wr.err = createFunc(dir, title)
	if wr.err != nil {
		return
	}

	// redirect to edit page
	http.Redirect(wr.w, wr.r, wr.wikiRoot+"edit-"+typ+"?page="+path.Join(dir, filename), http.StatusTemporaryRedirect)
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
		Root:              root + wikiDelimeter + wr.shortcode,
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
