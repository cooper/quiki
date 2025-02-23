package wiki

// Pregenerate simulates requests for all wiki resources
// such that content caches can be pregenerated and stored.
func (w *Wiki) Pregenerate() (results []any) {
	w.pregenerating = true
	allPageFiles := w.allPageFiles()
	results = make([]any, len(allPageFiles))
	for i, pageName := range allPageFiles {
		results[i] = w.DisplayPageDraft(pageName, true)
		if dp, ok := results[i].(DisplayPage); ok {
			if !dp.FromCache {
				w.Log("pregenerated " + pageName)
			}
		}
	}
	w.pregenerating = false
	return
}
