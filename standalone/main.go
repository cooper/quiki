package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/cooper/quiki/wikifier"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("wrong # of args")
	}
	content, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	err = wikifier.Parse(string(content))
	if err != nil {
		log.Fatal(err)
	}
}
