package wiki

import (
	"os"
	"path/filepath"

	"github.com/ldez/go-git-cmd-wrapper/git"
	ginit "github.com/ldez/go-git-cmd-wrapper/init"
)

// IsRepo returns true if the wiki directory is versioned by git.
func (w *Wiki) IsRepo() bool {
	_, err := os.Stat(filepath.Join(w.Opt.Dir.Wiki, ".git"))
	return err == nil
}

// InitRepo initializes a git repository for the wiki.
// If the wiki is already versioned by git, no error is produced.
func (w *Wiki) InitRepo() error {
	if w.IsRepo() {
		return nil
	}
	_, err := git.Init(ginit.Quiet, ginit.Directory(w.Opt.Dir.Wiki))
	return err
}
