package wiki

import (
	"log"

	"gopkg.in/src-d/go-git.v4"
)

func (w *Wiki) repo() *git.Repository {

	// we've already loaded the repository
	if w._repo != nil {
		return w._repo
	}

	// open it
	repo, err := git.PlainOpen(w.Opt.Dir.Wiki)

	// it doesn't exist- let's initialize it
	if err == git.ErrRepositoryNotExists {
		repo, err = git.PlainInit(w.Opt.Dir.Wiki, false)

		// error in init
		if err != nil {
			// TODO: better logging
			log.Println("git:PlainInit error:", err)
			return nil
		}

		// TODO: default .gitignore
		// TODO: add all files and initial commit

	} else if err != nil {
		// error in open other than nonexist

		// TODO: better logging
		log.Println("git:PlainOpen error:", err)
		return nil
	}

	// success
	w._repo = repo
	return repo
}

// // IsRepo returns true if the wiki directory is versioned by git.
// func (w *Wiki) IsRepo() bool {
// 	_, err := os.Stat(filepath.Join(w.Opt.Dir.Wiki, ".git"))
// 	return err == nil
// }

// // InitRepo initializes a git repository for the wiki.
// // If the wiki is already versioned by git, no error is produced.
// func (w *Wiki) InitRepo() error {
// 	if w.IsRepo() {
// 		return nil
// 	}
// 	_, err := git.Init(ginit.Quiet, ginit.Directory(w.Opt.Dir.Wiki))
// 	return err
// }

// // Branches returns the git branches available.
// func (w *Wiki) Branches() ([]string, error) {
// 	if !w.IsRepo() {
// 		return nil, nil
// 	}
// 	branches, err := git.Branch(branch.List)
// 	return strings.Split(branches, "\n"), err
// }
