// Copyright (c) 2017, Mitchell Cooper
package main

import "net/http"
import "log"

func handlePage(root string, w http.ResponseWriter, r *http.Request) {
	log.Println(r)
}

func handleImage(root string, w http.ResponseWriter, r *http.Request) {
	log.Println(r)
}
