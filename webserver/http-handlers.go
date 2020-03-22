package webserver

// Copyright (c) 2020, Mitchell Cooper

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/cooper/quiki/wiki"
)

// master handler
func handleRoot(w http.ResponseWriter, r *http.Request) {
	var delayedWiki *WikiInfo

	// try each wiki
	for _, w := range Wikis {

		// wrong root
		wikiRoot := w.Opt.Root.Wiki
		if r.URL.Path != wikiRoot && !strings.HasPrefix(r.URL.Path, wikiRoot+"/") {
			continue
		}

		// wrong host
		if w.Host != r.Host {

			// if the wiki host is empty, it is the fallback wiki.
			// delay it until we've checked all other wikis.
			if w.Host == "" && delayedWiki == nil {
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
			if delayedWiki.Opt.MainRedirect {
				http.Redirect(
					w, r,
					delayedWiki.Opt.Root.Page+
						"/"+mainPage,
					http.StatusMovedPermanently,
				)
				return
			}

			// display main page
			handlePage(delayedWiki, mainPage, w, r)
			return
		}

		// if the page root is blank, this may be a page
		if delayedWiki.Opt.Root.Page == "" {
			relPath := strings.TrimLeft(strings.TrimPrefix(r.URL.Path, wikiRoot), "/")
			handlePage(delayedWiki, relPath, w, r)
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
func handlePage(wi *WikiInfo, relPath string, w http.ResponseWriter, r *http.Request) {
	handleResponse(wi, wi.DisplayPage(relPath), w, r)
}

// image request
func handleImage(wi *WikiInfo, relPath string, w http.ResponseWriter, r *http.Request) {
	handleResponse(wi, wi.DisplayImage(relPath), w, r)
}

// topic request
func handleCategoryPosts(wi *WikiInfo, relPath string, w http.ResponseWriter, r *http.Request) {

	// extract page number from relPath
	pageN := 0
	catName := relPath
	split := strings.SplitN(relPath, "/", 2)
	if len(split) == 2 {
		if i, err := strconv.Atoi(split[1]); err == nil {
			pageN = i - 1
		}
		catName = split[0]
	}

	handleResponse(wi, wi.DisplayCategoryPosts(catName, pageN), w, r)
}

func handleResponse(wi *WikiInfo, res interface{}, w http.ResponseWriter, r *http.Request) {
	switch res := res.(type) {

	// page content
	case wiki.DisplayPage:
		renderTemplate(wi, w, "page", wikiPageFromRes(wi, res))

	// image content
	case wiki.DisplayImage:
		http.ServeFile(w, r, res.Path)

	// posts
	case wiki.DisplayCategoryPosts:

		// create template page
		page := wikiPageWith(wi)
		page.PageCSS = template.CSS(res.CSS)
		page.File = res.File
		page.Name = res.Name
		page.Title = res.Title
		page.PageN = res.PageN + 1
		page.NumPages = res.NumPages

		// add each page result as a wikiPage
		for _, dispPage := range res.Pages {
			page.Pages = append(page.Pages, wikiPageFromRes(wi, dispPage))
		}

		renderTemplate(wi, w, "posts", page)

	// error
	case wiki.DisplayError:
		handleError(wi, res, w, r)

	// redirect
	case wiki.DisplayRedirect:
		http.Redirect(w, r, res.Redirect, http.StatusMovedPermanently)

	// anything else
	default:
		http.NotFound(w, r)
	}
}

// this is set true when calling handlePage for the error page. this way, if an
// error occurs when trying to display the error page, we don't infinitely loop
// between handleError and handlePage
var useLowLevelError bool

func handleError(wi *WikiInfo, errMaybe interface{}, w http.ResponseWriter, r *http.Request) {
	status := http.StatusNotFound
	msg := "An unknown error has occurred"
	switch err := errMaybe.(type) {

	// if there's no error, stop
	case nil:
		return

	// display error
	case wiki.DisplayError:
		log.Println(err)
		msg = err.Error
		if err.Status != 0 {
			status = err.Status
		}

	// string
	case string:
		msg = err

	// error
	case error:
		msg = err.Error()

	}

	// if we have an error page for this wiki, use it
	errorPage := wi.Opt.ErrorPage
	if !useLowLevelError && errorPage != "" {
		useLowLevelError = true
		w.WriteHeader(status)
		handlePage(wi, errorPage, w, r)
		useLowLevelError = false
		return
	}

	// if the template provides an error page, fall back to that

	if errTmpl := wi.template.template.Lookup("error.tpl"); errTmpl != nil {
		var buf bytes.Buffer
		w.WriteHeader(status)
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
	http.Error(w, msg, status)
}

func renderTemplate(wi *WikiInfo, w http.ResponseWriter, templateName string, dot wikiPage) {
	var buf bytes.Buffer
	err := wi.template.template.ExecuteTemplate(&buf, templateName+".tpl", dot)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Length", strconv.FormatInt(int64(buf.Len()), 10))
	w.Write(buf.Bytes())
}

func wikiPageFromRes(wi *WikiInfo, res wiki.DisplayPage) wikiPage {
	page := wikiPageWith(wi)
	page.HTMLContent = template.HTML(res.Content)
	page.PageCSS = template.CSS(res.CSS)
	page.File = res.File
	page.Name = res.Name
	page.Title = res.Title
	page.Description = res.Description
	page.Keywords = res.Keywords
	page.Author = res.Author
	return page
}

func wikiPageWith(wi *WikiInfo) wikiPage {
	return wikiPage{
		WikiTitle:  wi.Title,
		WikiLogo:   wi.Logo,
		WikiRoot:   wi.Opt.Root.Wiki,
		Root:       wi.Opt.Root,
		StaticRoot: wi.template.staticRoot,
		Navigation: wi.Opt.Navigation,
		retina:     wi.Opt.Image.Retina,
	}
}
