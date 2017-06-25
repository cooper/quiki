// Copyright (c) 2017, Mitchell Cooper
package main

import (
	"bytes"
	wikiclient "github.com/cooper/go-wikiclient"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// master handler
func handleRoot(w http.ResponseWriter, r *http.Request) {
	var delayedWiki wikiInfo

	// try each wiki
	for _, wiki := range wikis {

		// wrong root
		wikiRoot := wiki.conf.Get("root.wiki")
		if r.URL.Path != wikiRoot && !strings.HasPrefix(r.URL.Path, wikiRoot+"/") {
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
		delayedWiki = wiki
		break
	}

	// a wiki matches this
	if delayedWiki.name != "" {

		// show the main page for the delayed wiki
		wikiRoot := delayedWiki.conf.Get("root.wiki")
		mainPage := delayedWiki.conf.Get("main_page")
		if mainPage != "" && (r.URL.Path == wikiRoot || r.URL.Path == wikiRoot+"/") {

			// main page redirect is enabled
			if delayedWiki.conf.GetBool("main_redirect") {
				http.Redirect(
					w, r,
					delayedWiki.conf.Get("root.page")+
						"/"+mainPage,
					http.StatusMovedPermanently,
				)
				return
			}

			// display main page
			delayedWiki.client = wikiclient.NewClient(tr, delayedWiki.defaultSess, 60*time.Second)
			handlePage(delayedWiki, mainPage, w, r)
			return
		}

		// show the 404 page for the delayed wiki
		http.Error(w, "404 page not found for "+delayedWiki.name, http.StatusNotFound)
	}

	// anything else is a generic 404
	http.NotFound(w, r)
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
		http.Redirect(w, r, res.Get("redirect"), http.StatusMovedPermanently)

	// page content
	case "page":
		renderTemplate(wiki, w, "page", wikiPageFromRes(wiki, res))

	// anything else
	default:
		http.NotFound(w, r)
	}
}

// image request
func handleImage(wiki wikiInfo, relPath string, w http.ResponseWriter, r *http.Request) {
	res, err := wiki.client.DisplayImage(relPath, 0, 0)
	if handleError(wiki, err, w, r) || handleError(wiki, res, w, r) {
		return
	}
	http.ServeFile(w, r, res.Get("path"))
}

// topic request
func handleCategoryPosts(wiki wikiInfo, relPath string, w http.ResponseWriter, r *http.Request) {

	// extract page number from relPath
	pageN := 1
	catName := relPath
	split := strings.SplitN(relPath, "/", 2)
	if len(split) == 2 {
		if i, err := strconv.Atoi(split[1]); err == nil {
			pageN = i
		}
		catName = split[0]
	}

	// error
	res, err := wiki.client.DisplayCategoryPosts(catName, pageN)
	if handleError(wiki, err, w, r) || handleError(wiki, res, w, r) {
		return
	}

	// pages is a map of page numbers to arrays of page refs
	pagesMap, ok := res.Args["pages"].(map[string]interface{})
	if !ok {
		handleError(wiki, "invalid response", w, r)
		return
	}

	// get the page with the requested number
	aSlice, ok := pagesMap[strconv.Itoa(pageN)].([]interface{})
	if !ok {
		log.Printf("problem: %+v", pagesMap[strconv.Itoa(pageN)])
		handleError(wiki, "invalid page number", w, r)
		return
	}

	// add each page
	page := wikiPageFromRes(wiki, res)
	for _, argMap := range aSlice {
		msg := wikiclient.Message{Args: argMap.(map[string]interface{})}
		page.Pages = append(page.Pages, wikiPageFromRes(wiki, msg))
	}

	renderTemplate(wiki, w, "posts", page)
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

func renderTemplate(wiki wikiInfo, w http.ResponseWriter, templateName string, dot wikiPage) {
	var buf bytes.Buffer
	err := wiki.template.template.ExecuteTemplate(&buf, templateName+".tpl", dot)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Length", strconv.FormatInt(int64(buf.Len()), 10))
	w.Write(buf.Bytes())
}

func wikiPageFromRes(wiki wikiInfo, res wikiclient.Message) wikiPage {
	return wikiPage{
		Res:        res,
		File:       res.Get("file"),
		Name:       res.Get("name"),
		Title:      res.Get("title"),
		WikiTitle:  wiki.title,
		WikiLogo:   wiki.getLogo(),
		WikiRoot:   wiki.conf.Get("root.wiki"),
		StaticRoot: wiki.template.staticRoot,
		navigation: wiki.conf.GetSlice("navigation"),
	}
}
