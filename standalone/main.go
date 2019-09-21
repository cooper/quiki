package main

import (
	"log"

	"github.com/cooper/quiki/wikifier"
)

func main() {
	log.Fatal(wikifier.Parse("/* Hello */ there [lady] {} Person"))
}
