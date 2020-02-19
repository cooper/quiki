package adminifier

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/cooper/quiki/authenticator"
	"github.com/cooper/quiki/webserver"
	"github.com/cooper/quiki/wiki"
	"github.com/pkg/errors"
)

var javascriptTemplates string

var frameHandlers = map[string]func(*wikiRequest){
	"dashboard":  handleDashboardFrame,
	"pages":      handlePagesFrame,
	"categories": handleCategoriesFrame,
	"images":     handleImagesFrame,
	"models":     handleModelsFrame,
	"settings":   handleSettingsFrame,
	"edit-page":  handleEditPageFrame,
}

// wikiTemplate members are available to all wiki templates
type wikiTemplate struct {
	User      authenticator.User // user
	Shortcode string             // wiki shortcode
	WikiTitle string             // wiki title
	Static    string             // adminifier-static root
	AdminRoot string             // adminifier root
	Root      string             // wiki root
}

type wikiRequest struct {
	shortcode string
	wi        *webserver.WikiInfo
	w         http.ResponseWriter
	r         *http.Request
	dot       interface{}
	err       error
}

// TODO: verify session on ALL wiki handlers

func setupWikiHandlers(shortcode string, wi *webserver.WikiInfo) {

	// each of these URLs generates wiki.tpl
	for _, which := range []string{
		"dashboard", "pages", "categories",
		"images", "models", "settings", "help", "edit-page",
	} {
		mux.HandleFunc(host+root+shortcode+"/"+which, func(w http.ResponseWriter, r *http.Request) {
			handleWiki(shortcode, wi, w, r)
		})
	}

	// frames to load via ajax
	frameRoot := root + shortcode + "/frame/"
	log.Println(frameRoot)
	mux.HandleFunc(host+frameRoot, func(w http.ResponseWriter, r *http.Request) {

		// check logged in
		if !sessMgr.GetBool(r.Context(), "loggedIn") {
			http.Redirect(w, r, root+"login", http.StatusTemporaryRedirect)
			return
		}

		frameName := strings.TrimPrefix(r.URL.Path, frameRoot)
		tmplName := "frame-" + frameName + ".tpl"

		// frame template does not exist
		if exist := tmpl.Lookup(tmplName); exist == nil {
			http.NotFound(w, r)
			return
		}

		// call func to create template params
		var dot interface{} = nil
		if handler, exist := frameHandlers[frameName]; exist {

			// create wiki request
			wr := &wikiRequest{
				shortcode: shortcode,
				wi:        wi,
				w:         w,
				r:         r,
			}

			// TODO: if working in another branch, override wr.wi to
			// the wiki instance for that branch

			// call handler
			handler(wr)

			// handler returned an error
			if wr.err != nil {
				panic(wr.err)
			}

			// handler was successful
			dot = wr.dot
		}

		// execute frame template with dot
		err := tmpl.ExecuteTemplate(w, tmplName, dot)

		// error occurred in template execution
		if err != nil {
			panic(err)
		}
	})
}

func handleWiki(shortcode string, wi *webserver.WikiInfo, w http.ResponseWriter, r *http.Request) {

	// check logged in
	if !sessMgr.GetBool(r.Context(), "loggedIn") {
		http.Redirect(w, r, root+"login", http.StatusTemporaryRedirect)
		return
	}

	// load javascript templates
	if javascriptTemplates == "" {
		files, _ := filepath.Glob(dirAdminifier + "/template/js-tmpl/*.tpl")
		for _, fileName := range files {
			data, _ := ioutil.ReadFile(fileName)
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
		panic(err)
	}
}

func handleDashboardFrame(wr *wikiRequest) {
}

func handlePagesFrame(wr *wikiRequest) {
	handleFileFrames(wr, wr.wi.Pages())
}

func handleImagesFrame(wr *wikiRequest) {
	handleFileFrames(wr, wr.wi.Images(), "d")
}

func handleModelsFrame(wr *wikiRequest) {
	handleFileFrames(wr, wr.wi.Models())
}

func handleCategoriesFrame(wr *wikiRequest) {
	handleFileFrames(wr, wr.wi.Categories())
}

func handleFileFrames(wr *wikiRequest, results interface{}, extras ...string) {
	res, err := json.Marshal(map[string]interface{}{
		"sort_types": append([]string{"a", "c", "u", "m"}, extras...),
		"results":    results,
	})
	if err != nil {
		wr.err = err
		return
	}
	wr.dot = struct {
		JSON  template.HTML
		Order string
		wikiTemplate
	}{
		JSON:         template.HTML("<!--JSON\n" + string(res) + "\n-->"),
		Order:        "m-",
		wikiTemplate: getGenericTemplate(wr),
	}
}

func handleSettingsFrame(wr *wikiRequest) {
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
	handleEditor(wr, info.Path, info.File, info.Title)
}

func handleEditor(wr *wikiRequest, path, file, title string) {

	// call DisplayFile to get the content
	res := wr.wi.DisplayFile(path)
	fileRes, ok := res.(wiki.DisplayFile)
	if !ok {
		wr.err = errors.New("error occurred in DisplayFile")
		return
	}

	wr.dot = struct {
		Found   bool
		JSON    template.HTML
		Model   bool   // true if editing a model
		Title   string // page title or filename
		File    string // filename
		Content string // file content
		wikiTemplate
	}{
		Found:        true,
		JSON:         template.HTML("<!--JSON\n{}\n-->"), // TODO
		Model:        false,
		Title:        title,
		File:         file,
		Content:      fileRes.Content,
		wikiTemplate: getGenericTemplate(wr),
	}
}

func getGenericTemplate(wr *wikiRequest) wikiTemplate {
	return wikiTemplate{
		User:      sessMgr.Get(wr.r.Context(), "user").(authenticator.User),
		Shortcode: wr.shortcode,
		WikiTitle: wr.wi.Title,
		AdminRoot: strings.TrimRight(root, "/"),
		Static:    root + "adminifier-static",
		Root:      root + wr.shortcode,
	}
}
