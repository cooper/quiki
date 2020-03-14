package wiki

// Pregenerate simulates requests for all wiki resources
// such that content caches can be pregenerated and stored.
func (w *Wiki) Pregenerate() {
	w.pregenerating = true

	for _, pageName := range w.allPageFiles() {
		w.Debug("pregen page:", pageName)
		w.DisplayPageDraft(pageName, true)
	}

	w.pregenerating = false
}
