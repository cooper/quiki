package wiki

import "fmt"

// Pregenerate simulates requests for all wiki resources
// such that content caches can be pregenerated and stored.
func (w *Wiki) Pregenerate() {

	// cache page content
	for _, pageName := range w.allPageFiles() {
		w.DisplayPageDraft(pageName, true)
	}

	// generate images in all the sizes used
	for _, image := range w.Images() {
		for _, d := range image.Dimensions {
			sized := SizedImageFromName(fmt.Sprintf("%dx%d-%s", d[0], d[1], image.File))
			w.DisplaySizedImageGenerate(sized, true)
		}
	}
}
