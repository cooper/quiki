package wiki

// Pregenerate simulates requests for all wiki resources
// such that content caches can be pregenerated and stored.
func (w *Wiki) Pregenerate() {
	for _, pageName := range w.allPageFiles() {
		w.DisplayPageDraft(pageName, true)
	}
}
