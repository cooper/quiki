package webserver

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"

	"github.com/cooper/quiki/wikifier"
	"github.com/pkg/errors"
)

func CreateWizardConfig(opts Options) {
	if opts.Port == "" {
		opts.Port = "8080"
	}

	// config already exists
	if _, err := os.Stat(opts.Config); err == nil {
		log.Fatalf("config file already exists: %s; please run without -w", opts.Config)
	}

	// make the directory path if needed
	if err := os.MkdirAll(filepath.Dir(opts.Config), 0755); err != nil {
		log.Fatal(errors.Wrap(err, "make config dir"))
	}

	// create config page
	token := generateWizardToken()
	conf := wikifier.NewPage(opts.Config)
	vars := map[string]any{
		"server.http.bind":            opts.Bind,
		"server.http.port":            opts.Port,
		"server.enable.pregeneration": true,
		"adminifier.enable":           true,
		"adminifier.host":             "",
		"adminifier.root":             "/admin",
		"adminifier.token":            token,
	}
	for k, v := range vars {
		conf.Set(k, v)
	}

	// write config
	conf.VarsOnly = true
	if err := conf.Write(); err != nil {
		log.Fatal(errors.Wrap(err, "write wizard config"))
	}

	// print wizard instructions
	host := opts.Bind
	if host == "" {
		host = "localhost"
	}
	log.Printf("config written to %s", opts.Config)
}

func generateWizardToken() string {
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		log.Fatal(errors.Wrap(err, "generate wizard token"))
	}
	return hex.EncodeToString(tokenBytes)
}
