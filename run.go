package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/cooper/quiki/pregenerate"
	"github.com/cooper/quiki/wiki"
	"github.com/cooper/quiki/wikifier"
)

// reads from stdin
func interactiveMode() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	page := wikifier.NewPageSource(string(input))
	runPageAndExit(page)
}

// runs a page
func runPageAndExit(page *wikifier.Page) {
	err := page.Parse()
	if err != nil {
		log.Fatal(err)
	}
	if jsonOutput {
		json.NewEncoder(os.Stdout).Encode(page)
		os.Exit(0)
	}
	fmt.Println(page.HTMLAndCSS())
	os.Exit(0)
}

// runs a wiki page
func runWikiPageAndExit(w *wiki.Wiki, pagePath string) {
	pregen := pregenerate.New(w)
	defer pregen.Stop()

	res := pregen.GeneratePageSync(pagePath, true)
	if jsonOutput {
		json.NewEncoder(os.Stdout).Encode(res)
		os.Exit(0)
	}
	switch r := res.(type) {
	case wiki.DisplayPage:
		if r.CSS != "" {
			fmt.Println(`<style type="text/css">`)
			fmt.Println(r.CSS)
			fmt.Println(`</style>`)
		}
		fmt.Println(r.Content)
	case wiki.DisplayError:
		log.Fatal(r.Error)
	case wiki.DisplayRedirect:
		log.Fatal("Redirect: " + r.Redirect)
	default:
		log.Fatal("unsupported response type from wiki.DisplayPage()")
	}
	os.Exit(0)
}
