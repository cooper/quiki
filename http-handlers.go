// Copyright (c) 2017, Mitchell Cooper
package main

import (
	wikiclient "github.com/cooper/go-wikiclient"
	"net/http"
	"regexp"
	"time"
)

var imageRegex = regexp.MustCompile("")

// master handler
func handleRoot(w http.ResponseWriter, r *http.Request) {
	var delayedWiki wikiInfo

	// try each wiki
	for _, wiki := range wikis {

		// wrong root
		wikiRoot := wiki.conf.Get("root.wiki")
		if r.URL.Path != wikiRoot && r.URL.Path != wikiRoot+"/" {
			continue
		}

		// wrong host
		if wiki.host != r.Host {

			// if the wiki host is empty, it is the fallback wiki.
			// delay it until we've checked all other wikis.
			if wiki.host == "" && delayedWiki.name == "" {
				delayedWiki = wiki
			}

			continue
		}

		// host matches
		if handleMainPage(wiki, w, r) {
			return
		}
	}

	// try the delayed wiki
	if delayedWiki.name != "" {
		if handleMainPage(delayedWiki, w, r) {
			return
		}
	}

	// anything else is a 404
	http.NotFound(w, r)
}

func handleMainPage(wiki wikiInfo, w http.ResponseWriter, r *http.Request) bool {

	// no main page configured
	mainPage := wiki.conf.Get("main_page")
	if mainPage == "" {
		return false
	}

	wiki.client = wikiclient.NewClient(tr, wiki.defaultSess, 3*time.Second)
	handlePage(wiki, mainPage, w, r)
	return true
}

// page request
func handlePage(wiki wikiInfo, relPath string, w http.ResponseWriter, r *http.Request) {
	res, err := wiki.client.DisplayPage(relPath)

	// wikiclient error or 'not found' response
	if handleError(wiki, err, w, r) || handleError(wiki, res, w, r) {
		return
	}

	// other response
	switch res.Get("type") {

	// page redirect
	case "redirect":
		http.Redirect(w, r, wiki.conf.Get("root.page")+"/"+res.Get("name"), 301)

	// page content
	case "page":
		renderTemplate(wiki, w, "page", wikiPage{
			Res:        res,
			Title:      res.Get("title"),
			WikiTitle:  wiki.title,
			WikiLogo:   wiki.template.logo,
			WikiRoot:   wiki.conf.Get("root.wiki"),
			StaticRoot: wiki.template.staticRoot,
			navigation: wiki.conf.GetSlice("navigation"),
		})
	}

	http.NotFound(w, r)
}

// image request
func handleImage(wiki wikiInfo, relPath string, w http.ResponseWriter, r *http.Request) {
	res, err := wiki.client.DisplayImage(relPath, 0, 0)
	if handleError(wiki, err, w, r) || handleError(wiki, res, w, r) {
		return
	}
	http.ServeFile(w, r, res.Get("path"))
}

// this is set true when calling handlePage for the error page. this way, if an
// error occurs when trying to display the error page, we don't infinitely loop
// between handleError and handlePage
var useLowLevelError bool

func handleError(wiki wikiInfo, errMaybe interface{}, w http.ResponseWriter, r *http.Request) bool {
	msg := "An unknown error has occurred"
	switch err := errMaybe.(type) {

	// if there's no error, stop
	case nil:
		return false

	// message of type "not found" is an error; otherwise, stop
	case wikiclient.Message:
		if err.Get("type") != "not found" {
			return false
		}
		msg = err.Get("error")

		// string
	case string:
		msg = err

		// error
	case error:
		msg = err.Error()

	}

	// if we have an error page, use it
	errorPage := wiki.conf.Get("error_page")
	if !useLowLevelError && errorPage != "" {
		useLowLevelError = true
		handlePage(wiki, errorPage, w, r)
		useLowLevelError = false
		return true
	}

	// generic error response
	http.Error(w, msg, http.StatusNotFound)
	return true
}

func renderTemplate(wiki wikiInfo, w http.ResponseWriter, templateName string, p wikiPage) {
	err := wiki.template.template.ExecuteTemplate(w, templateName+".tpl", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
