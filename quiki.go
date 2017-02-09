// Copyright (c) 2017, Mitchell Cooper
package main

import (
	"net/http"
	"log"
)

func main() {
    http.Handle("/image/", http.StripPrefix("/image/", http.FileServer(http.Dir("."))))
	log.Fatal(http.ListenAndServe(":12345", nil))
}
