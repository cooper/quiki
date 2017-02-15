// Copyright (c) 2017, Mitchell Cooper
package main

import (
	"fmt"
	"github.com/cooper/quiki/wikiclient"
	"net/http"
)

// page request
func handlePage(wiki wikiInfo, relPath string, w http.ResponseWriter, r *http.Request) {
	res, err := wiki.client.DisplayPage(relPath)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	if res.String("type") == "not found" {
		handleError(res, w, r)
		return
	}
	renderTemplate(wiki, w, "page", wikiPage{
		Title:      res.String("title"),
		WikiTitle:  wiki.name,
		Res:        res,
		StaticRoot: "TODO",
	})
}

// image request
func handleImage(wiki wikiInfo, relPath string, w http.ResponseWriter, r *http.Request) {
}

// func handleResponse(res wikiclient.Message, w http.ResponseWriter, r *http.Request) {
// 	if res.String("type") == "not found" {
// 		handleError(res, w, r)
// 		return
// 	}
// 	w.Header().Set("Content-Type", res.String("mime"))
// 	w.Header().Set("Content-Length", res.String("length"))
// 	w.Header().Set("Last-Modified", res.String("modified"))
// 	w.Write([]byte(res.String("content")))
// }

func handleError(res wikiclient.Message, w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func renderTemplate(wiki wikiInfo, w http.ResponseWriter, templateName string, p wikiPage) {
	err := wiki.template.template.ExecuteTemplate(w, templateName+".tpl", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
