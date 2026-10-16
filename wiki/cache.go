package wiki

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// PurgeCache removes all generated page and image caches. categories and
// other metadata are kept; they are brought up to date as content is
// regenerated.
func (w *Wiki) PurgeCache() error {
	for _, sub := range []string{"page", "image"} {
		dir := filepath.Join(w.Opt.Dir.Cache, sub)
		if err := os.RemoveAll(dir); err != nil {
			return errors.Wrap(err, "purge "+sub+" cache")
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.Wrap(err, "recreate "+sub+" cache")
		}
	}
	w.Log("cache purged")
	return nil
}
