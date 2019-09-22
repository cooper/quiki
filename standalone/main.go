package main

import (
	"log"
	"os"

	"github.com/cooper/quiki/wikifier"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("wrong # of args")
	}
	page := wikifier.NewPage(os.Args[1])
	err := page.Parse()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(page.HTML())
}
