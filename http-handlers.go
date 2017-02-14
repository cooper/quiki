// Copyright (c) 2017, Mitchell Cooper
package main

import (
	"fmt"
	"github.com/cooper/quiki/wikiclient"
	"net/http"
)

func handlePage(c wikiclient.Client, relPath string, w http.ResponseWriter, r *http.Request) {
	res, err := c.DisplayPage(relPath)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprint(w, res)
}

func handleImage(c wikiclient.Client, relPath string, w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, relPath, c, r)
}

func handleError(c wikiclient.Client, relPath string, w http.ResponseWriter, r *http.Request) {
}
