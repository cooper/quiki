// Copyright (c) 2017, Mitchell Cooper
package quikiserver

import (
	"net/http"
	"log"
)

func Run() {
    http.Handle("/image/", http.StripPrefix("/file/", http.FileServer(http.Dir("."))))
	log.Fatal(http.ListenAndServe(":12345", nil))
}
