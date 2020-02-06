// Package adminifier provides an administrative panel with a web-based editor.
package adminifier

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cooper/quiki/webserver"
	_ "github.com/cooper/quiki/webserver" // access existing ServeMux and config
	"github.com/cooper/quiki/wikifier"
	"github.com/pkg/errors"
)

var mux *http.ServeMux
var conf *wikifier.Page
var dirAdminifier string

// Configure sets up adminifier on webserver.ServeMux using webserver.Conf.
func Configure() {
	conf = webserver.Conf
	mux = webserver.Mux

	// do nothing if not enabled
	if enable, _ := conf.GetBool("adminifier.enable"); !enable {
		return
	}

	// extract strings
	var host, root string
	for key, ptr := range map[string]*string{
		"server.dir.adminifier": &dirAdminifier,
		"adminifier.host":       &host,
		"adminifier.root":       &root,
	} {
		str, err := conf.GetStr(key)
		if err != nil {
			log.Fatal(err)
		}
		*ptr = str
	}

	// setup static files server
	if err := setupStatic(host+root, filepath.Join(dirAdminifier, "adminifier-static")); err != nil {
		log.Fatal(errors.Wrap(err, "setup adminifier-static"))
	}
}

func setupStatic(hostRoot, staticPath string) error {
	hostRoot += "/adminifier-static/"
	if stat, err := os.Stat(staticPath); err != nil || !stat.IsDir() {
		if err == nil {
			err = errors.New("not a directory")
		}
		return err
	}
	fileServer := http.FileServer(http.Dir(staticPath))
	mux.Handle(hostRoot, http.StripPrefix(hostRoot, fileServer))
	log.Println("registered adminifier root: " + hostRoot)
	return nil
}
