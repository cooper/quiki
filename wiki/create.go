package wiki

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cooper/quiki/resources"
	"github.com/cooper/quiki/wikifier"
	"github.com/pkg/errors"
)

type CreateWikiOpts struct {
	WikiName     string
	TemplateName string
	MainPage     string
	ErrorPage    string
}

// CreateWiki creates a new wiki at the specified path using a base wiki directory.
func CreateWiki(path, basePath string, opts CreateWikiOpts) error {
	return CreateWikiFS(path, os.DirFS(basePath), opts)
}

// CreateWikiFromResource creates a new wiki at the specified path using a base wiki resource.
func CreateWikiFromResource(path, resourceName string, opts CreateWikiOpts) error {
	if resourceName == "" {
		resourceName = "default"
	}
	baseFs, err := fs.Sub(resources.Wikis, resourceName)
	if err != nil {
		return errors.Wrap(err, "get base wiki fs")
	}
	return CreateWikiFS(path, baseFs, opts)
}

// CreateWiki creates a new wiki at the specified path using a base wiki fs.
func CreateWikiFS(path string, fsys fs.FS, opts CreateWikiOpts) error {

	// derive name from path if not specified
	if opts.WikiName == "" {
		opts.WikiName = filepath.Base(path)
	}
	normalizedName := wikifier.PageNameLink(opts.WikiName)

	// walk the fs to ensure it's not empty
	empty := true
	err := fs.WalkDir(fsys, ".", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if filePath != "." {
			empty = false
			return fs.SkipDir // stop walking as soon as we find a non-root entry
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "walk base wiki fs")
	}
	if empty {
		return errors.New("base wiki is empty")
	}

	// copy base wiki
	if err := os.CopyFS(path, fsys); err != nil {
		return errors.Wrap(err, "copy new wiki")
	}

	conf := wikifier.NewPage(filepath.Join(path, "wiki.conf"))

	// if the base wiki has a config, parse it
	if conf.Exists() {
		err := conf.Parse()
		if err != nil {
			return errors.Wrap(err, "parse wiki config")
		}
	}

	// write new config vars
	vars := map[string]any{
		"name":       opts.WikiName,
		"root.wiki":  normalizedName,
		"template":   opts.TemplateName,
		"main_page":  opts.MainPage,
		"error_page": opts.ErrorPage,
	}
	for k, v := range vars {
		if v == "" {
			continue
		}
		conf.Set(k, v)
	}

	// write wiki config
	conf.VarsOnly = true
	if err := conf.Write(); err != nil {
		return errors.Wrap(err, "write wiki config")
	}

	return nil
}
