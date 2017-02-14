// Copyright (c) 2017, Mitchell Cooper
package main

import (
	"fmt"
	"github.com/cooper/quiki/wikiclient"
	"net/http"
)

// page request
func handlePage(c wikiclient.Client, relPath string, w http.ResponseWriter, r *http.Request) {
	res, err := c.DisplayPage(relPath)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	handleResponse(res, w, r)
}

// image request
func handleImage(c wikiclient.Client, relPath string, w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, relPath, c, r)
}

func handleResponse(res wikiclient.Message, w http.ResponseWriter, r *http.Request) {
	if res.Command == "error" {
		handleError(res, w, r)
		return
	}
	w.Header().Set("Content-Type", res.String("mime"))
	w.Header().Set("Content-Length", res.String("length"))
	w.Header().Set("Last-Modified", res.String("modified"))
	w.Write([]byte(res.String("content")))
}

func handleError(res wikiclient.Message, w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}
