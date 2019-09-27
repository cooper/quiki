package wiki

import (
	"github.com/cooper/quiki/wikifier"
)

// A Wiki represents a quiki website.
type Wiki struct {
	ConfigFile        string
	PrivateConfigFile string
	Opt               wikifier.PageOpts
	conf              *wikifier.Page
	pconf             *wikifier.Page
}
