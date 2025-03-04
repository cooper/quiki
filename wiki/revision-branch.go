package wiki

import "github.com/go-git/go-git/v5/plumbing"

// BranchNames returns the revision branches available.
func (w *Wiki) BranchNames() ([]string, error) {
	repo, err := w.repo()
	var names []string
	if err != nil {
		return nil, err
	}
	branches, err := repo.Branches()
	if err != nil {
		return nil, err
	}
	branches.ForEach(func(ref *plumbing.Reference) error {
		names = append(names, ref.Name().Short())
		return nil
	})
	return names, nil
}

// working with multiple branches is disabled until go-git supports it officially
// see: https://github.com/cooper/quiki/issues/156

// // ensure a branch exists in git
// func (w *Wiki) hasBranch(name string) (bool, error) {
// 	names, err := w.BranchNames()
// 	if err != nil {
// 		return false, err
// 	}
// 	for _, branchName := range names {
// 		if branchName == name {
// 			return true, nil
// 		}
// 	}
// 	return false, nil
// }

// // checks out a branch in another directory. returns the directory
// func (w *Wiki) checkoutBranch(name string) (string, error) {

// 	// never checkout master in a linked repo
// 	if name == "master" {
// 		return "", errors.New("cannot check out master in a linked repo")
// 	}

// 	// make sure name is a simple wordlike string with no path elements
// 	if !ValidBranchName(name) {
// 		return "", errors.New("invalid branch name")
// 	}

// 	// make cache/branch/ if needed
// 	wikifier.MakeDir(filepath.Join(w.Opt.Dir.Cache, "branch"), "")

// 	// e.g. cache/branch/mybranchname
// 	targetDir := filepath.Join(w.Opt.Dir.Cache, "branch", name)

// 	// directory already exists, so I'm good with saying the branch is there
// 	if fi, err := os.Stat(targetDir); err == nil && fi.IsDir() {
// 		return targetDir, nil
// 	}

// 	repo, err := w.repo()
// 	if err != nil {
// 		return "", err
// 	}

// 	// create the linked repository
// 	if _, err = repo.PlainAddWorktree(name, targetDir, &git.AddWorktreeOptions{}); err != nil {
// 		return "", err
// 	}

// 	return targetDir, nil
// }

// // Branch returns a Wiki instance for this wiki at another branch.
// // If the branch does not exist, an error is returned.
// func (w *Wiki) Branch(name string) (*Wiki, error) {

// 	// never checkout master in a linked repo
// 	if name == "master" {
// 		return w, nil
// 	}

// 	// find branch
// 	if exist, err := w.hasBranch(name); !exist {
// 		if err != nil {
// 			return nil, err
// 		}
// 		return nil, git.ErrBranchNotFound
// 	}

// 	// check out the branch in cache/branch/<name>;
// 	// if it already has been checked out, this does nothing
// 	dir, err := w.checkoutBranch(name)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// create a new Wiki at this location
// 	return NewWiki(dir)
// }

// // NewBranch is like Branch, except it creates the branch at the
// // current master revision if it does not yet exist.
// func (w *Wiki) NewBranch(name string) (*Wiki, error) {
// 	repo, err := w.repo()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// find branch
// 	if exist, err := w.hasBranch(name); !exist {
// 		if err != nil {
// 			return nil, err
// 		}

// 		// try to create it
// 		err := repo.CreateBranch(&config.Branch{
// 			Name:  name,
// 			Merge: plumbing.NewBranchReferenceName(name),
// 		})
// 		if err != nil {
// 			return nil, err
// 		}

// 		// determine where master is at
// 		fs := repo.Storer.(interface{ Filesystem() billy.Filesystem }).Filesystem()
// 		f1, err := fs.Open(fs.Join("refs", "heads", "master"))
// 		if err != nil {
// 			return nil, err
// 		}
// 		defer f1.Close()
// 		masterRef, err := io.ReadAll(f1)
// 		if err != nil {
// 			return nil, err
// 		}

// 		// set refs/heads/<name> to same as master
// 		f2, err := fs.Create(fs.Join("refs", "heads", name))
// 		if err != nil {
// 			return nil, err
// 		}
// 		defer f2.Close()
// 		_, err = fmt.Fprintf(f2, "%s\n", string(masterRef))
// 		if err != nil {
// 			return nil, err
// 		}
// 	}

// 	// now that it exists, fetch it
// 	return w.Branch(name)
// }

// var branchNameRgx = regexp.MustCompile(`^[\w-]+$`)

// // ValidBranchName returns whether a branch name is valid.
// //
// // quiki branch names may contain word-like characters `\w` and
// // forward slash (`/`) but may not start or end with a slash.
// func ValidBranchName(name string) bool {
// 	return branchNameRgx.MatchString(name)
// }
