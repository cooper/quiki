// Copyright (c) 2017, Mitchell Cooper
package main

import "net/http"
import "log"

func handler(rootType, root string, w http.ResponseWriter, r *http.Request) {
	log.Println(rootType, root)
}
