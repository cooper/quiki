// Copyright (c) 2017, Mitchell Cooper
package main

import (
	"github.com/cooper/quiki/config"
	"log"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "" {
		log.Fatal("Please provide the wikiserver configuration file as the first argument")
	}

	conf := config.New(os.Args[1])
	if err := conf.Parse(); err != nil {
		log.Fatal("Configuration error: ", err)
	}
	log.Println(conf)

	http.Handle("/image/", http.StripPrefix("/image/", http.FileServer(http.Dir("."))))
	log.Fatal(http.ListenAndServe(":12345", nil))
}
