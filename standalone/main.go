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

	// set some variables
	page.Set("myKey", "myValue")
	if val, err := page.GetStr("myKey"); err != nil {
		log.Fatal("Got error: ", err)
	} else {
		log.Println("Got:", val)
	}

	// fetch map property from page
	if val, err := page.GetStr("myMap.a"); err != nil {
		log.Fatal("Got map error: ", err)
	} else {
		log.Println("Got from map:", val)
	}

	// fetch object var

	page.Set("page", page)

	if val, err := page.GetObj("page.page"); err != nil {
		log.Fatal("Got error: ", err)
	} else {
		log.Println("Got:", val)
	}

	// fetch recursive var

	if val, err := page.GetStr("page.myKey"); err != nil {
		log.Fatal("Got error: ", err)
	} else {
		log.Println("Got:", val)
	}

	// // parse
	// err := page.Parse()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // spit out html
	// log.Println(page.HTML())
}
