// Copyright (c) 2017, Mitchell Cooper
package main

import (
	wikiclient "github.com/cooper/go-wikiclient"
	"net/http"
	"regexp"
)

var imageRegex = regexp.MustCompile("")

// master handler for the wiki root
func handleWikiRoot(wiki wikiInfo, relPath string, w http.ResponseWriter, r *http.Request) {

	// main page
	mainPage := wiki.conf.Get("main_page")
	if relPath == "" && mainPage != "" {
		handlePage(wiki, mainPage, w, r)
		return
	}

	// anything else is a 404
	http.NotFound(w, r)
}

// page request
func handlePage(wiki wikiInfo, relPath string, w http.ResponseWriter, r *http.Request) {
	res, err := wiki.client.DisplayPage(relPath)
	if handleError(err, w, r) || handleError(res, w, r) {
		return
	}
	renderTemplate(wiki, w, "page", wikiPage{
		Res:        res,
		Title:      res.Get("title"),
		WikiTitle:  wiki.title,
		WikiLogo:   wiki.template.logo,
		StaticRoot: wiki.template.staticRoot,
	})
}

// image request
func handleImage(wiki wikiInfo, relPath string, w http.ResponseWriter, r *http.Request) {
	res, err := wiki.client.DisplayImage(relPath, 0, 0)
	if handleError(err, w, r) || handleError(res, w, r) {
		return
	}
	http.ServeFile(w, r, res.Get("path"))
}

// func handleResponse(res wikiclient.Message, w http.ResponseWriter, r *http.Request) {
// 	if res.Get("type") == "not found" {
// 		handleError(res, w, r)
// 		return
// 	}
// 	w.Header().Set("Content-Type", res.Get("mime"))
// 	w.Header().Set("Content-Length", res.Get("length"))
// 	w.Header().Set("Last-Modified", res.Get("modified"))
// 	w.Write([]byte(res.Get("content")))
// }

func handleError(errMaybe interface{}, w http.ResponseWriter, r *http.Request) bool {
	var msg string
	switch err := errMaybe.(type) {
	case nil:
		return false
	case wikiclient.Message:
		if err.Get("type") != "not found" {
			return false
		}
		msg = err.Get("error")
	}
	http.Error(w, msg, http.StatusNotFound)
	return true
}

func renderTemplate(wiki wikiInfo, w http.ResponseWriter, templateName string, p wikiPage) {
	err := wiki.template.template.ExecuteTemplate(w, templateName+".tpl", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
