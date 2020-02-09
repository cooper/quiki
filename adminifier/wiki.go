package adminifier

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/cooper/quiki/webserver"
)

var javascriptTemplates string

var frameHandlers = map[string]func(string, *webserver.WikiInfo) interface{}{
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
	Static    string // adminifier-static root
	AdminRoot string // adminifier root
	Root      string // wiki root
}

// TODO: verify session on ALL wiki handlers

func setupWikiHandlers(shortcode string, wi *webserver.WikiInfo) {

	// each of these URLs generates wiki.tpl
	for _, which := range []string{"dashboard", "pages", "categories", "images", "models", "settings", "help"} {
		mux.HandleFunc(host+root+shortcode+"/"+which, func(w http.ResponseWriter, r *http.Request) {
			handleWiki(shortcode, w, r)
		})
	}

	// frames to load via ajax
	frameRoot := host + root + shortcode + "/frame/"
	mux.HandleFunc(frameRoot, func(w http.ResponseWriter, r *http.Request) {
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
			dot = handler(shortcode, wi)
		}

		// execute frame template
		err := tmpl.ExecuteTemplate(w, tmplName, dot)

		// error occurred in template execution
		if err != nil {
			panic(err)
		}
	})
}

func handleWiki(shortcode string, w http.ResponseWriter, r *http.Request) {

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
		getWikiTemplate(shortcode),
	})
	if err != nil {
		panic(err)
	}
}

func handleDashboardFrame(shortcode string, wi *webserver.WikiInfo) interface{} {
	return nil
}

func handlePagesFrame(shortcode string, wi *webserver.WikiInfo) interface{} {
	return struct {
		Type  string
		Order string
		wikiTemplate
	}{
		Type:         "Pages",
		Order:        "m-",
		wikiTemplate: getWikiTemplate(shortcode),
	}
}

func handleCategoriesFrame(shortcode string, wi *webserver.WikiInfo) interface{} {
	return nil
}

func handleImagesFrame(shortcode string, wi *webserver.WikiInfo) interface{} {
	return nil
}

func handleModelsFrame(shortcode string, wi *webserver.WikiInfo) interface{} {
	return nil
}

func handleSettingsFrame(shortcode string, wi *webserver.WikiInfo) interface{} {
	return nil
}

func handleEditPageFrame(shortcode string, wi *webserver.WikiInfo) interface{} {
	return struct {
		Model   bool   // true if editing a model
		Title   string // page title or filename
		File    string // filename
		Content string // file content
		wikiTemplate
	}{
		Model:        false,
		Title:        "Title",
		File:         "File",
		Content:      "Content",
		wikiTemplate: getWikiTemplate(shortcode),
	}
}

func getWikiTemplate(shortcode string) wikiTemplate {
	return wikiTemplate{
		AdminRoot: root,
		Static:    root + "adminifier-static",
		Root:      root + shortcode,
	}
}
