package wiki

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/cooper/quiki/wikifier"
	"gopkg.in/src-d/go-billy.v4"

	"github.com/cooper/go-git/v4"
	"github.com/cooper/go-git/v4/config"
	"github.com/cooper/go-git/v4/plumbing"
	"github.com/cooper/go-git/v4/plumbing/object"
	"github.com/pkg/errors"
)

// CommitOpts describes the options for a wiki revision.
type CommitOpts struct {

	// Comment is the commit description.
	Comment string

	// Name is the fullname of the user committing changes.
	Name string

	// Email is the email address of the user committing changes.
	Email string

	// Time is the timestamp to associate with the revision.
	// If unspecified, current time is used.
	Time time.Time
}

// default options for commits made by quiki itself
var quikiCommitOpts = &git.CommitOptions{
	Author: &object.Signature{
		Name:  "quiki",
		Email: "quiki@quiki.app",
		When:  time.Now(),
	},
}

// repo fetches the wiki's git repository, creating it if needed.
func (w *Wiki) repo() (repo *git.Repository, err error) {

	// we've already loaded the repository
	if w._repo != nil {
		repo = w._repo
		return
	}

	// open it
	repo, err = git.PlainOpen(w.Dir())

	// it doesn't exist- let's initialize it
	if err == git.ErrRepositoryNotExists {
		repo, err = w.createRepo()
	} else if err != nil {
		// error in open other than nonexist

		err = errors.Wrap(err, "git:PlainOpen")
		return
	}

	// success
	w._repo = repo
	return
}

// create new repository
func (w *Wiki) createRepo() (repo *git.Repository, err error) {

	/// initialize new repo
	repo, err = git.PlainInit(w.Dir(), false)

	// error in init
	if err != nil {
		return nil, errors.Wrap(err, "git:PlainInit")
	}

	// initialized ok

	// create master branch
	err = repo.CreateBranch(&config.Branch{
		Name:  "master",
		Merge: plumbing.Master,
	})
	if err != nil {
		return nil, errors.Wrap(err, "git:repo:CreateBranch")
	}

	// TODO: default .gitignore

	// add all files and initial commit
	wt, err := repo.Worktree()
	if err != nil {
		return nil, errors.Wrap(err, "git:repo:Worktree")
	}

	// add .
	_, err = wt.Add(".")
	if err != nil {
		return nil, err
	}

	// commit
	_, err = wt.Commit("Initial commit", quikiCommitOpts)
	if err != nil {
		return nil, errors.Wrap(err, "git:worktree:Commit")
	}

	return repo, nil
}

// // IsRepo returns true if the wiki directory is versioned by git.
// func (w *Wiki) IsRepo() bool {
// 	_, err := os.Stat(w.Dir(".git"))
// 	return err == nil
// }

// // InitRepo initializes a git repository for the wiki.
// // If the wiki is already versioned by git, no error is produced.
// func (w *Wiki) InitRepo() error {
// 	if w.IsRepo() {
// 		return nil
// 	}
// 	_, err := git.Init(ginit.Quiet, ginit.Directory(w.Dir())
// 	return err
// }

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

// ensure a branch exists in git
func (w *Wiki) hasBranch(name string) (bool, error) {
	names, err := w.BranchNames()
	if err != nil {
		return false, err
	}
	for _, branchName := range names {
		if branchName == name {
			return true, nil
		}
	}
	return false, nil
}

// checks out a branch in another directory. returns the directory
func (w *Wiki) checkoutBranch(name string) (string, error) {

	// never checkout master in a linked repo
	if name == "master" {
		return "", errors.New("cannot check out master in a linked repo")
	}

	// TODO: make sure name is a simple string with no path elements

	// make cache/branch/ if needed
	wikifier.MakeDir(filepath.Join(w.Opt.Dir.Cache, "branch"), "")

	// e.g. cache/branch/mybranchname
	targetDir := filepath.Join(w.Opt.Dir.Cache, "branch", name)

	// directory already exists, so I'm good with saying the branch is there
	if fi, err := os.Stat(targetDir); err == nil && fi.IsDir() {
		return targetDir, nil
	}

	repo, err := w.repo()
	if err != nil {
		return "", err
	}

	// create the linked repository
	if _, err = repo.PlainAddWorktree(name, targetDir, &git.AddWorktreeOptions{}); err != nil {
		return "", err
	}

	return targetDir, nil
}

// the "and commit" portion of the *andCommit functions
func (w *Wiki) andCommit(wt *git.Worktree, comment string, commit CommitOpts) error {
	if commit.Comment != "" {
		comment += ": " + commit.Comment
	}

	// time defaults to now
	if commit.Time.IsZero() {
		commit.Time = time.Now()
	}

	// commit
	_, err := wt.Commit(comment, &git.CommitOptions{
		Author: &object.Signature{
			Name:  commit.Name,
			Email: commit.Email,
			When:  commit.Time,
		},
	})
	if err != nil {
		return errors.Wrap(err, "git:worktree:Commit")
	}

	return nil
}

// addAndCommit adds a file and then commits changes
func (w *Wiki) addAndCommit(path string, commit CommitOpts) error {

	// get repo
	repo, err := w.repo()
	if err != nil {
		return err
	}

	// get worktree
	wt, err := repo.Worktree()
	if err != nil {
		return errors.Wrap(err, "git:repo:Worktree")
	}

	// add the file
	_, err = wt.Add(path)
	if err != nil {
		return err
	}

	return w.andCommit(wt, "Update "+filepath.Base(path), commit)
}

// removeAndCommit removes a file and then commits changes
func (w *Wiki) removeAndCommit(path string, commit CommitOpts) error {

	// get repo
	repo, err := w.repo()
	if err != nil {
		return err
	}

	// get worktree
	wt, err := repo.Worktree()
	if err != nil {
		return errors.Wrap(err, "git:repo:Worktree")
	}

	// remove the file
	_, err = wt.Remove(path)
	if err != nil {
		return err
	}

	return w.andCommit(wt, "Delete "+filepath.Base(path), commit)
}

// Branch returns a Wiki instance for this wiki at another branch.
// If the branch does not exist, an error is returned.
func (w *Wiki) Branch(name string) (*Wiki, error) {

	// never checkout master in a linked repo
	if name == "master" {
		return w, nil
	}

	// find branch
	if exist, err := w.hasBranch(name); !exist {
		if err != nil {
			return nil, err
		}
		return nil, git.ErrBranchNotFound
	}

	// check out the branch in cache/branch/<name>;
	// if it already has been checked out, this does nothing
	dir, err := w.checkoutBranch(name)
	if err != nil {
		return nil, err
	}

	// create a new Wiki at this location
	return NewWiki(dir)
}

// NewBranch is like Branch, except it creates the branch at the
// current master revision if it does not yet exist.
func (w *Wiki) NewBranch(name string) (*Wiki, error) {
	repo, err := w.repo()
	if err != nil {
		return nil, err
	}

	// find branch
	if exist, err := w.hasBranch(name); !exist {
		if err != nil {
			return nil, err
		}

		// try to create it
		err := repo.CreateBranch(&config.Branch{
			Name:  name,
			Merge: plumbing.NewBranchReferenceName(name),
		})
		if err != nil {
			return nil, err
		}

		// determine where master is at
		fs := repo.Storer.(interface{ Filesystem() billy.Filesystem }).Filesystem()
		f1, err := fs.Open(fs.Join("refs", "heads", "master"))
		if err != nil {
			return nil, err
		}
		defer f1.Close()
		masterRef, err := ioutil.ReadAll(f1)
		if err != nil {
			return nil, err
		}

		// set refs/heads/<name> to same as master
		f2, err := fs.Create(fs.Join("refs", "heads", name))
		if err != nil {
			return nil, err
		}
		defer f2.Close()
		_, err = fmt.Fprintf(f2, "%s\n", string(masterRef))
		if err != nil {
			return nil, err
		}
	}

	// now that it exists, fetch it
	return w.Branch(name)
}

var branchNameRgx = regexp.MustCompile(`^[\w]+[\w\-/]*[\w]+$`)

// ValidBranchName returns whether a branch name is valid.
//
// quiki branch names may contain word-like characters `\w` and
// forward slash (`/`) but may not start or end with a slash.
//
func ValidBranchName(name string) bool {
	return branchNameRgx.MatchString(name)
}

// WritePage writes a page file.

// WriteFile writes a file in the wiki.
//
// The filename must be relative to the wiki directory.
// If the file does not exist and createOK is false, an error is returned.
// If the file exists and is a symbolic link, an error is returned.
//
// This is a low-level API that allows writing any file within the wiki
// directory, so it should not be utilized directly by frontends.
// Use WritePage, WriteModel, WriteImage, or WriteConfig instead.
//
func (w *Wiki) WriteFile(name string, content []byte, createOK bool, commit CommitOpts) error {
	path := w.UnresolvedAbsFilePath(name)
	fi, err := os.Lstat(path)

	if err != nil {
		// some error occurred

		if os.IsNotExist(err) {
			// file doesn't exist-- only care if createOK is false
			if !createOK {
				return err
			}
		} else {
			// all other errors are always bad
			return err
		}
	} else if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		// no errors occurred, so file exists
		// check if it's a symlink
		return errors.New("symlink cannot be written with WriteFile")
	}

	// write file all at once
	if err := ioutil.WriteFile(path, content, 0644); err != nil {
		return err
	}

	// commit the change
	return w.addAndCommit(name, commit)
}

// DeleteFile deletes a file in the wiki.
//
// The filename must be relative to the wiki directory.
// If the file does not exist, an error is returned.
//
// This is a low-level API that allows deleting any file within the wiki
// directory, so it should not be utilized directly by frontends.
// Use DeletePage, DeleteModel, or DeleteImage instead.
//
func (w *Wiki) DeleteFile(name string, commit CommitOpts) error {

	// error running lstat on file, might not exist or whatev
	path := w.UnresolvedAbsFilePath(name)
	_, err := os.Lstat(path)
	if err != nil {
		return err
	}

	// delete the file and commit the change
	return w.removeAndCommit(path, commit)
}
