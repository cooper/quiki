// Copyright (c) 2017, Mitchell Cooper
package main

import (
	"github.com/cooper/quiki/wikiclient"
	"log"
	"net/http"
)

func handlePage(c wikiclient.Client, w http.ResponseWriter, r *http.Request) {
	log.Println(c, r)
}

func handleImage(c wikiclient.Client, w http.ResponseWriter, r *http.Request) {
	log.Println(c, r)
}
