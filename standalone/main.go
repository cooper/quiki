package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cooper/quiki/wikifier"
)

func main() {
	var page *wikifier.Page

	if len(os.Args) > 1 {
		page = wikifier.NewPage(os.Args[1])
	} else {
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		page = wikifier.NewPageSource(string(input))
	}

	// parse
	err := page.Parse()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(page.HTML())
}
