package adminifier

import (
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cooper/quiki/resources"
	"github.com/cooper/quiki/webserver"
	"github.com/cooper/quiki/wiki"
	"github.com/pkg/errors"
)

var helpWiki *wiki.Wiki

func handleHelpFrame(helpRoot string, w http.ResponseWriter, r *http.Request) (any, error) {

	var dot struct {
		Title   string
		Content template.HTML
	}

	// create the help wiki if not already loaded
	if helpWiki == nil {

		// find dir of config file
		helpDir := filepath.Join(filepath.Dir(webserver.Conf.Path()), "help")

		// check if help dir exists already
		if _, err := os.Stat(helpDir); err != nil {

			// copy help resources
			subFs, err := fs.Sub(resources.Adminifier, "help")
			if err != nil {
				return nil, errors.Wrap(err, "reading help resource")
			}
			err = os.CopyFS(helpDir, subFs)
			if err != nil {
				return nil, errors.Wrap(err, "copying help resource")
			}
		}

		var err error
		helpWiki, err = wiki.NewWiki(helpDir)
		if err != nil {
			return nil, err
		}
	}

	// determine page
	helpPage := strings.TrimPrefix(strings.TrimPrefix(r.URL.Path, helpRoot+"frame/help"), "/")
	if helpPage == "" {
		helpPage = helpWiki.Opt.MainPage
	}

	// display the page
	res := helpWiki.DisplayPage(helpPage)
	switch res := res.(type) {

	// page content
	case wiki.DisplayPage:
		dot.Title = res.Title
		content := string(res.Content)
		content = strings.ReplaceAll(content, `"/pagereplace/`, `"`+helpRoot+"help/")
		dot.Content = template.HTML(content)
		return dot, nil

	// error
	case wiki.DisplayError:
		return nil, errors.New(res.Error)

	// something else
	default:
		return nil, errors.New("unknown response")
	}
}
