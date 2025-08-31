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

// CreateWizardConfig creates a new server config file given the options.
func CreateWizardConfig(opts Options) {

	// default options
	if opts.Config == "" {
		opts.Config = filepath.Join(os.Getenv("HOME"), "quiki", "quiki.conf")
	}
	if opts.Port == "" {
		opts.Port = "8080"
	}

	// config already exists
	if _, err := os.Stat(opts.Config); err == nil {
		log.Printf("config found at %s", opts.Config)
		return
	}

	// make the quiki dir if needed
	quikiDir := filepath.Dir(opts.Config)
	if err := os.MkdirAll(quikiDir, 0755); err != nil {
		log.Fatal(errors.Wrap(err, "make quiki dir"))
	}

	// make the wikis dir if needed
	if err := os.MkdirAll(opts.WikisDir, 0755); err != nil {
		log.Fatal(errors.Wrap(err, "make wikis dir"))
	}

	// create config page
	token := generateWizardToken()
	conf := wikifier.NewPage(opts.Config)
	vars := map[string]any{
		"server.http.bind":            opts.Bind,
		"server.http.port":            opts.Port,
		"server.dir.wiki":             opts.WikisDir,
		"server.enable.pregeneration": true,
		"server.pregen.mode":          "default", // options: default, fast, slow
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
