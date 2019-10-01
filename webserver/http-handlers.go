package webserver

// Copyright (c) 2019, Mitchell Cooper

import (
	"bytes"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/cooper/quiki/wiki"
)

// master handler
func handleRoot(w http.ResponseWriter, r *http.Request) {
	var delayedWiki *wikiInfo

	// try each wiki
	for _, w := range wikis {

		// wrong root
		wikiRoot := w.Opt.Root.Wiki
		if r.URL.Path != wikiRoot && !strings.HasPrefix(r.URL.Path, wikiRoot+"/") {
			continue
		}

		// wrong host
		if w.host != r.Host {

			// if the wiki host is empty, it is the fallback wiki.
			// delay it until we've checked all other wikis.
			if w.host == "" && delayedWiki == nil {
				delayedWiki = w
			}

			continue
		}

		// host matches
		delayedWiki = w
		break
	}

	// a wiki matches this
	if delayedWiki != nil {

		// show the main page for the delayed wiki
		wikiRoot := delayedWiki.Opt.Root.Wiki
		mainPage := delayedWiki.Opt.MainPage
		if mainPage != "" && (r.URL.Path == wikiRoot || r.URL.Path == wikiRoot+"/") {

			// main page redirect is enabled
			// FIXME:
			// if delayedWiki.conf.GetBool("main_redirect") {
			// 	http.Redirect(
			// 		w, r,
			// 		delayedWiki.conf.Get("root.page")+
			// 			"/"+mainPage,
			// 		http.StatusMovedPermanently,
			// 	)
			// 	return
			// }

			// display main page
			handlePage(delayedWiki, mainPage, w, r)
			return
		}

		// show the 404 page for the delayed wiki
		handleError(delayedWiki, "Page not found.", w, r)
		return
	}

	// anything else is a generic 404
	http.NotFound(w, r)
}

// page request
func handlePage(wi *wikiInfo, relPath string, w http.ResponseWriter, r *http.Request) {
	res := wi.DisplayPage(relPath)

	// other response
	switch res := res.(type) {

	// error
	case wiki.DisplayError:
		handleError(wi, res, w, r)

	// page redirect
	case wiki.DisplayRedirect:
		http.Redirect(w, r, res.Redirect, http.StatusMovedPermanently)

	// page content
	case wiki.DisplayPage:
		renderTemplate(wi, w, "page", wikiPageFromRes(wi, res))

	// anything else
	default:
		http.NotFound(w, r)
	}
}

// image request
func handleImage(wi *wikiInfo, relPath string, w http.ResponseWriter, r *http.Request) {
	res := wi.DisplayImage(relPath)

	// other response
	switch res := res.(type) {

	// error
	case wiki.DisplayError:
		handleError(wi, res, w, r)

	// image content
	case wiki.DisplayImage:
		http.ServeFile(w, r, res.Path)

	// anything else
	default:
		http.NotFound(w, r)
	}
}

// topic request
func handleCategoryPosts(wi *wikiInfo, relPath string, w http.ResponseWriter, r *http.Request) {

	// // extract page number from relPath
	// pageN := 1
	// catName := relPath
	// split := strings.SplitN(relPath, "/", 2)
	// if len(split) == 2 {
	// 	if i, err := strconv.Atoi(split[1]); err == nil {
	// 		pageN = i
	// 	}
	// 	catName = split[0]
	// }

	// // error
	// res, err := wiki.client.DisplayCategoryPosts(catName, pageN)
	// if handleError(wiki, err, w, r) || handleError(wiki, res, w, r) {
	// 	return
	// }

	// // pages is a map of page numbers to arrays of page refs
	// pagesMap, ok := res.Args["pages"].(map[string]interface{})
	// if !ok {
	// 	handleError(wiki, "invalid response", w, r)
	// 	return
	// }

	// // get the page with the requested number
	// aSlice, ok := pagesMap[strconv.Itoa(pageN)].([]interface{})
	// if !ok {
	// 	log.Printf("problem: %+v", pagesMap[strconv.Itoa(pageN)])
	// 	handleError(wiki, "invalid page number", w, r)
	// 	return
	// }

	// // add each page
	// page := wikiPageFromRes(wiki, res)
	// for _, argMap := range aSlice {
	// 	msg := wikiclient.Message{Args: argMap.(map[string]interface{})}
	// 	page.Pages = append(page.Pages, wikiPageFromRes(wiki, msg))
	// }

	// renderTemplate(wiki, w, "posts", page)
}

// this is set true when calling handlePage for the error page. this way, if an
// error occurs when trying to display the error page, we don't infinitely loop
// between handleError and handlePage
var useLowLevelError bool

func handleError(wi *wikiInfo, errMaybe interface{}, w http.ResponseWriter, r *http.Request) {
	msg := "An unknown error has occurred"
	switch err := errMaybe.(type) {

	// if there's no error, stop
	case nil:
		return

		// display error
	case wiki.DisplayError:
		log.Println(err)
		msg = err.Error

	// string
	case string:
		msg = err

	// error
	case error:
		msg = err.Error()

	}

	//TODO:
	// // if we have an error page for this wiki, use it
	// errorPage := wiki.conf.Get("error_page")
	// if !useLowLevelError && errorPage != "" {
	// 	useLowLevelError = true
	// 	w.WriteHeader(http.StatusNotFound)
	// 	handlePage(wiki, errorPage, w, r)
	// 	useLowLevelError = false
	// 	return true
	// }

	// if the template provides an error page, fall back to that

	if errTmpl := wi.template.template.Lookup("error.tpl"); errTmpl != nil {
		var buf bytes.Buffer
		w.WriteHeader(http.StatusNotFound)
		page := wikiPageWith(wi)
		page.Name = "Error"
		page.Title = "Error"
		page.Message = msg
		errTmpl.Execute(&buf, page)
		w.Header().Set("Content-Length", strconv.FormatInt(int64(buf.Len()), 10))
		w.Write(buf.Bytes())
		return
	}

	// finally, fall back to generic error response
	http.Error(w, msg, http.StatusNotFound)
}

func renderTemplate(wi *wikiInfo, w http.ResponseWriter, templateName string, dot wikiPage) {
	var buf bytes.Buffer
	err := wi.template.template.ExecuteTemplate(&buf, templateName+".tpl", dot)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Length", strconv.FormatInt(int64(buf.Len()), 10))
	w.Write(buf.Bytes())
}

func wikiPageFromRes(w *wikiInfo, res wiki.DisplayPage) wikiPage {
	page := wikiPageWith(w)
	page.Res = res
	page.File = res.File
	page.Name = res.Name
	page.Title = res.Title
	return page
}

func wikiPageWith(w *wikiInfo) wikiPage {
	return wikiPage{
		WikiTitle: w.title,
		// WikiLogo:   w.getLogo(), FIXME:
		WikiRoot:   w.Opt.Root.Wiki,
		StaticRoot: w.template.staticRoot,
		Navigation: w.Opt.Navigation,
	}
}
