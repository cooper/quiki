package wiki

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cooper/quiki/adminifier/utils"
	"github.com/cooper/quiki/wikifier"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
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
		Email: "quiki@quiki.rlygd.net",
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

// the "and commit" portion of the *andCommit functions
func (w *Wiki) andCommit(wt *git.Worktree, comment string, commit CommitOpts) error {

	// comment overrides default
	if commit.Comment != "" {
		comment = commit.Comment
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

// WritePage writes a page file.

// writeFile writes a file in the wiki.
//
// The path must be absolute within the wiki directory.
//
// If the file does not exist and createOK is false, an error is returned.
// If the file exists and is a symbolic link, an error is returned.
//
// This is a low-level API that allows writing any file within the wiki
// directory, so it should not be utilized directly by frontends.
// Use WritePage, WriteModel, WriteImage, or WriteConfig instead.
func (w *Wiki) writeFile(path string, content []byte, createOK bool, commit CommitOpts) error {
	relPath, err := filepath.Rel(w.Dir(), path)
	if err != nil {
		return errors.Wrap(err, "filepath:Rel")
	}

	if createOK {
		// ensure parent directories exist
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
	}

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
		return errors.New("refusing to write to symlinked file")
	}

	// write file all at once
	if err := os.WriteFile(path, content, 0644); err != nil {
		return err
	}

	// commit the change
	return w.addAndCommit(relPath, commit)
}

// WritePage writes a page file.
func (w *Wiki) WritePage(name string, content []byte, createOK bool, commit CommitOpts) error {
	// Coordinate with file monitor to prevent race conditions
	return w.writeFileWithMonitorCoordination(w.PathForPage(name), content, createOK, commit, name)
}

// writeFileWithMonitorCoordination wraps writeFile with monitor coordination
func (w *Wiki) writeFileWithMonitorCoordination(path string, content []byte, createOK bool, commit CommitOpts, resourceName string) error {
	// Get page lock to coordinate with monitor
	if resourceName != "" {
		pageLock := w.GetPageLock(resourceName)
		pageLock.Lock()
		defer pageLock.Unlock()
	}

	// TODO: Pause monitor during write to prevent race conditions
	// This would require importing the monitor package, which would create a cycle
	// For now, the locking provides basic coordination

	return w.writeFile(path, content, createOK, commit)
}

// WriteModel writes a model file.
func (w *Wiki) WriteModel(name string, content []byte, createOK bool, commit CommitOpts) error {
	return w.writeFile(w.PathForModel(name), content, createOK, commit)
}

// WriteImage writes an image file.
func (w *Wiki) WriteImage(name string, content []byte, createOK bool, commit CommitOpts) error {
	return w.writeFile(w.PathForImage(name), content, createOK, commit)
}

// WriteConfig writes the wiki configuration file.
func (w *Wiki) WriteConfig(content []byte, commit CommitOpts) error {
	return w.writeFile(w.ConfigFile, content, true, commit)
}

// DeleteFile deletes a file in the wiki.
//
// The filename must be relative to the wiki directory.
// If the file does not exist, an error is returned.
// If the file exists and is a symbolic link, the link itself is deleted,
// not the target file.
//
// This is a low-level API that allows deleting any file within the wiki
// directory, so it should not be utilized directly by frontends.
// Use DeletePage, DeleteModel, or DeleteImage instead.
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

// GetLatestCommitHash returns the most recent commit hash.
func (w *Wiki) GetLatestCommitHash() (string, error) {
	repo, err := w.repo()
	if err != nil {
		return "", err
	}

	ref, err := repo.Head()
	if err != nil {
		return "", errors.Wrap(err, "git:repo:Head")
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return "", errors.Wrap(err, "git:repo:CommitObject")
	}

	return commit.Hash.String(), nil
}

// CreatePage creates a new page file.
// If content is empty, a default page is created.
func (w *Wiki) CreatePage(where string, title string, content []byte, commit CommitOpts) (string, error) {
	if len(content) == 0 {
		content = []byte("@page.title: " + utils.EscFmt(title) + ";\n@page.draft;\n\n")
	}
	name := wikifier.PageName(strings.Replace(title, "/", "_", -1))
	if where != "" && !strings.HasSuffix(where, "/") {
		where += "/"
	}
	return name, w.WritePage(where+name, content, true, commit)
}

// CreateModel creates a new model file.
func (w *Wiki) CreateModel(title string, content []byte, commit CommitOpts) (string, error) {
	name := wikifier.ModelName(title)
	return name, w.WriteModel(name, content, true, commit)
}

// RevisionInfo contains information about a specific revision.
type RevisionInfo struct {
	Id      string    `json:"id"`
	Author  string    `json:"author"`
	Date    time.Time `json:"date"`
	Message string    `json:"message"`
}

// RevisionsMatchingPage returns a list of commit infos matching a page file.
func (w *Wiki) RevisionsMatchingPage(nameOrPath string) ([]RevisionInfo, error) {
	return w._revisionsMatchingFile(w.RelPath(w.PathForPage(nameOrPath)))
}

// Diff returns the diff between two revisions.
// NOTE: For now, this includes all changes, not just those to a specific file.
func (w *Wiki) Diff(from, to string) (*object.Patch, error) {
	repo, err := w.repo()
	if err != nil {
		return nil, err
	}

	// get the latest commit
	var toHash plumbing.Hash
	if to == "" {
		ref, err := repo.Head()
		if err != nil {
			return nil, err
		}
		toHash = ref.Hash()
	} else {
		toHash = plumbing.NewHash(to)
	}

	// get the commits
	fromCommit, err := repo.CommitObject(plumbing.NewHash(from))
	if err != nil {
		return nil, err
	}
	toCommit, err := repo.CommitObject(toHash)
	if err != nil {
		return nil, err
	}

	// TODO: create a utility that generates a diff string from the
	// []Chunk returned by Patch.FilePatches()[n].Files()[n].Chunks().
	// unfortunately go-git does not publicize this utility.
	return fromCommit.Patch(toCommit)
}

// _revisionsMatchingFile returns a list of commit infos matching a file path.
func (w *Wiki) _revisionsMatchingFile(path string) ([]RevisionInfo, error) {
	repo, err := w.repo()
	if err != nil {
		return nil, err
	}

	logOptions := &git.LogOptions{
		FileName: &path,
	}

	commitIter, err := repo.Log(logOptions)
	if err != nil {
		return nil, err
	}

	var revisions []RevisionInfo
	err = commitIter.ForEach(func(c *object.Commit) error {
		revisions = append(revisions, RevisionInfo{
			Id:      c.Hash.String(),
			Author:  c.Author.Name,
			Date:    c.Author.When,
			Message: c.Message,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return revisions, nil
}
