// Package adminifier provides an administrative panel with a web-based editor.
package adminifier

import (
	"net/http"

	"github.com/cooper/quiki/webserver"
	_ "github.com/cooper/quiki/webserver" // access existing ServeMux and config
	"github.com/cooper/quiki/wikifier"
)

var mux *http.ServeMux
var conf *wikifier.Page

// Configure sets up adminifier on webserver.ServeMux using webserver.Conf.
func Configure() {
	conf = webserver.Conf
	mux = webserver.Mux
}
