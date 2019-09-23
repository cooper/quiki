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

	// parse
	err := page.Parse()
	if err != nil {
		log.Fatal(err)
	}

	// spit out html
	log.Println(page.HTML())

	// fetch map
	if m, err := page.GetObj("myMap"); err != nil {
		log.Fatal("myMap error: ", err)
	} else {
		log.Println("myMap:", m)
	}

	// fetch map
	if val, err := page.GetStr("myMap.A"); err != nil {
		log.Fatal("myMap.A error: ", err)
	} else {
		log.Println("myMap.A:", val)
	}

	// fetch nested map
	if val, err := page.GetStr("myMap.E.F"); err != nil {
		log.Fatal("myMap.E.F error: ", err)
	} else {
		log.Println("myMap.E.F:", val)
	}

}
