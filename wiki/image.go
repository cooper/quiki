package wiki

// DisplayImage represents an image to display.
type DisplayImage struct {

	// basename of the scaled image file
	File string

	// absolute path to the scaled image.
	// this file should be served to the user
	Path string

	// absolute path to the full-size image.
	// if the full-size image is being displayed, same as Path
	FullsizePath string

	// image type
	// 'png' or 'jpeg'
	ImageType string

	// mime 'image/png' or 'image/jpeg'
	// suitable for the Content-Type header
	Mime string

	// binary image data
	// omitted with dont_open option
	Content string

	// bytelength of image data
	// suitable for use in the Content-Length header
	Length int64

	// UNIX timestamp of when the image was last modified.
	// if Generated is true, this is the current time.
	// if FromCache is true, this is the modified date of the cache file.
	// otherwise, this is the modified date of the image file itself.
	ModUnix int64

	// like ModUnix except in HTTP date format
	// suitable for Last-Modified header
	Modified string

	// true if the content being sered was read from a cache file.
	// opposite of Generated
	FromCache bool

	// true if the content being served was just generated.
	// opposite of FromCache
	Generated bool

	// true if the content generated in order to fulfill this request was
	// written to cache. this can only been true when Generated is true
	CacheGenerated bool
}
