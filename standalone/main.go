package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cooper/quiki/wikifier"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("wrong # of args")
	}
	page := wikifier.NewPage(os.Args[1])

	// parse
	err := page.Parse()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(page.HTML())
}
