package wiki

// Pregenerate simulates requests for all wiki resources
// such that content caches can be pregenerated and stored.
func (w *Wiki) Pregenerate() {

	// cache page content
	for _, pageName := range w.allPageFiles() {
		w.Log("pregen page:", pageName)
		w.DisplayPageDraft(pageName, true)
	}

	// generate images in all the sizes used
	for _, image := range w.Images() {
		for _, d := range image.Dimensions {
			sized := SizedImageFromName(image.File)
			sized.Width = d[0]
			sized.Height = d[1]
			w.Log("pregen image:", sized.ScaleName())
			w.DisplaySizedImageGenerate(sized, true)
		}
	}
}
