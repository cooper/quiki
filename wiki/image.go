package wiki

import (
	"fmt"
	_ "image/jpeg" // for jpegs
	_ "image/png"  // for pngs
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	httpdate "github.com/Songmu/go-httpdate"
	"github.com/cooper/quiki/adminifier/utils"
	"github.com/cooper/quiki/wikifier"
)

var (
	imageNameRegex  = regexp.MustCompile(`^(\d+)x(\d+)-(.+)$`)
	imageScaleRegex = regexp.MustCompile(`^(.+)\@(\d+)x$`)
)

// ImageInfo represents a full-size image on the wiki.
type ImageInfo struct {
	File       string     `json:"file"`               // filename
	Base       string     `json:"base,omitempty"`     // base name
	Width      int        `json:"width,omitempty"`    // full-size width
	Height     int        `json:"height,omitempty"`   // full-size height
	Created    *time.Time `json:"created,omitempty"`  // creation time
	Modified   *time.Time `json:"modified,omitempty"` // modify time
	Dimensions [][]int    `json:"-"`                  // dimensions used throughout the wiki
}

// SizedImage represents an image in specific dimensions.
type SizedImage struct {
	// for example mydir/100x200-myimage@3x.png
	Width, Height int    // 100, 200 (dimensions as requested)
	Scale         int    // 3 (scale as requested)
	Prefix        string // mydir
	RelNameNE     string // myimage (name without extension)
	Ext           string // png (extension)
	zeroByZero    bool   // true when created from 0x0-name
}

// SizedImageFromName returns a SizedImage given an image name.
func SizedImageFromName(name string) SizedImage {
	w, h := 0, 0
	zeroByZero := false

	// before all else, separate name and prefix
	pfx := ""
	name = filepath.ToSlash(name) // just in case
	lastSlash := strings.LastIndexByte(name, '/')
	if lastSlash != -1 {
		pfx = name[:lastSlash+1]
		name = name[lastSlash+1:]
	}

	// width and height were given, so it's a resized image
	if matches := imageNameRegex.FindStringSubmatch(name); len(matches) != 0 {
		w, _ = strconv.Atoi(matches[1])
		h, _ = strconv.Atoi(matches[2])
		zeroByZero = w == 0 && h == 0
		name = matches[3]
	}

	// extract extension
	nameNE := name
	ext := ""
	lastDot := strings.LastIndexByte(name, '.')
	if lastDot != -1 && lastDot < len(name) {
		nameNE = name[:lastDot]
		ext = name[lastDot+1:]
	}

	// if this is a retina request, calculate scaled dimensions
	scale := 1
	if matches := imageScaleRegex.FindStringSubmatch(nameNE); len(matches) != 0 {
		nameNE = matches[1]
		scale, _ = strconv.Atoi(matches[2])
		// name = matches[1] + "." + ext
	}

	// put it all together
	return SizedImage{
		Width:      w,
		Height:     h,
		Scale:      scale,
		Prefix:     pfx,
		RelNameNE:  nameNE,
		Ext:        ext,
		zeroByZero: zeroByZero,
	}
}

// TrueWidth returns the actual image width when the Scale is taken into consideration.
func (img SizedImage) TrueWidth() int {
	return img.Width * img.Scale
}

// TrueHeight returns the actual image height when the Scale is taken into consideration.
func (img SizedImage) TrueHeight() int {
	return img.Height * img.Scale
}

// FullSizeName returns the name of the full-size image.
func (img SizedImage) FullSizeName() string {
	return img.Prefix + img.RelNameNE + "." + img.Ext
}

// TrueNameNE is like TrueName but without the extension.
func (img SizedImage) TrueNameNE() string {
	if img.Width == 0 && img.Height == 0 {
		return img.Prefix + img.RelNameNE
	}
	return fmt.Sprintf(
		"%s%dx%d-%s",
		img.Prefix,
		img.TrueWidth(),
		img.TrueHeight(),
		img.RelNameNE,
	)
}

// TrueName returns the image name with true dimensions.
func (img SizedImage) TrueName() string {
	return img.TrueNameNE() + "." + img.Ext
}

// ScaleName returns the image name with dimensions and scale.
func (img SizedImage) ScaleName() string {
	if img.Scale <= 1 {
		return img.TrueName()
	}
	return fmt.Sprintf("%s%dx%d-%s@%dx.%s",
		img.Prefix,
		img.Width,
		img.Height,
		img.RelNameNE,
		img.Scale,
		img.Ext,
	)
}

// DisplayImage represents an image to display.
type DisplayImage struct {

	// basename of the scaled image file
	File string `json:"file,omitempty"`

	// absolute path to the scaled image.
	// this file should be served to the user
	Path string `json:"path,omitempty"`

	// absolute path to the full-size image.
	// if the full-size image is being displayed, same as Path
	FullsizePath string `json:"fullsize_path,omitempty"`

	// image type
	// 'png' or 'jpeg'
	ImageType string `json:"image_type,omitempty"`

	// mime 'image/png' or 'image/jpeg'
	// suitable for the Content-Type header
	Mime string `json:"mime,omitempty"`

	// bytelength of image data
	// suitable for use in the Content-Length header
	Length int64 `json:"length,omitempty"`

	// time when the image was last modified.
	// if Generated is true, this is the current time.
	// if FromCache is true, this is the modified date of the cache file.
	// otherwise, this is the modified date of the image file itself.
	Modified     *time.Time `json:"modified,omitempty"`
	ModifiedHTTP string     `json:"modified_http,omitempty"` // HTTP format for Last-Modified

	// true if the content being sered was read from a cache file.
	// opposite of Generated
	FromCache bool `json:"cached,omitempty"`

	// true if the content being served was just generated.
	// opposite of FromCache
	Generated bool `json:"generated,omitempty"`

	// true if the content generated in order to fulfill this request was
	// written to cache. this can only been true when Generated is true
	CacheGenerated bool `json:"cache_gen,omitempty"`
}

// DisplayImage returns the display result for an image.
func (w *Wiki) DisplayImage(name string) any {
	return w.DisplaySizedImageGenerate(SizedImageFromName(name), false)
}

// DisplaySizedImage returns the display result for an image in specific dimensions.
func (w *Wiki) DisplaySizedImage(img SizedImage) any {
	return w.DisplaySizedImageGenerate(img, false)
}

// DisplaySizedImageGenerate returns the display result for an image in specific dimensions
// and allows images to be generated in any dimension.
func (w *Wiki) DisplaySizedImageGenerate(img SizedImage, generateOK bool) any {
	return w.DisplaySizedImageGenerateInternal(img, generateOK, true, false)
}

// DisplaySizedImageGenerateInternal is the internal implementation that can skip locking
func (w *Wiki) DisplaySizedImageGenerateInternal(img SizedImage, generateOK bool, needLock bool, nonBlocking bool) any {
	var r DisplayImage
	logName := img.ScaleName()
	w.Debug("display image:", logName)

	// check if the file exists
	bigPath := w.PathForImage(img.FullSizeName())
	fi, err := os.Lstat(bigPath)
	if err != nil {
		return DisplayError{
			Error:         "Image does not exist.",
			DetailedError: "Image '" + bigPath + "' error: " + err.Error(),
		}
	}

	// one dimension is missing
	var bigW, bigH int
	oldName := img.TrueName()
	if (img.Width == 0 && img.Height != 0) || (img.Height == 0 && img.Width != 0) {
		w.Debugf("display image: %s: missing a dimension; have to open", logName)

		// get full size dimensions
		bigW, bigH = getImageDimensions(bigPath)

		// find missing dimension
		// note: we haven't checked if both are 0 yet, but this will return 0, 0 in that case
		img.Width, img.Height = calculateImageDimensions(bigW, bigH, img.Width, img.Height)
	}

	// check if the name has changed after this adjustment.
	// if so, redirect
	trueName := img.TrueName()
	if trueName != oldName || img.zeroByZero {
		w.Debugf("display image: %s: redirect %s -> %s", logName, oldName, trueName)
		return DisplayRedirect{filepath.Base(trueName)}
	}

	// image name and full path
	r.Path = bigPath
	r.FullsizePath = bigPath
	r.File = filepath.Base(r.Path)

	// image type and mime type
	if img.Ext == "jpg" || img.Ext == "jpeg" {
		r.ImageType = "jpeg"
		r.Mime = "image/jpeg"
	} else if img.Ext == "png" {
		r.ImageType = "png"
		r.Mime = "image/png"
	} else {
		return DisplayError{
			Error:         "Unknown image type.",
			DetailedError: "Image '" + bigPath + "' is neither png nor jpeg",
		}
	}

	// create or update image category
	// consider: do we need to do this here, and does it write every time?
	w.GetSpecialCategory(r.File, CategoryTypeImage).addImage(w, r.File, nil, nil)

	// if both dimensions are missing, display the full-size version of the image
	if img.Width == 0 && img.Height == 0 {
		w.Debugf("display image: %s: using full-size", logName)
		mod := fi.ModTime()
		r.Modified = &mod
		r.ModifiedHTTP = httpdate.Time2Str(mod)
		r.Length = fi.Size()
		return r
	}

	// at this point, at least one dimension was provided, and both
	// dimensions have been determined

	// #============================#
	// #=== Retina scale support ===#
	// #============================#

	// this is not a retina request, but retina is enabled, and
	// this is a pregeneration request of the normal-scale image.
	// so, commit a pregeneration request for each scaled version.
	if img.Scale <= 1 && generateOK {
		for _, scale := range w.Opt.Image.Retina {
			w.Debugf("display image: %s: also generating retina @%dx", logName, scale)
			scaledImage := img        // copy
			scaledImage.Scale = scale // set scale
			w.DisplaySizedImageGenerateInternal(scaledImage, generateOK, needLock, false)
		}
	}

	// #=========================#
	// #=== Find cached image ===#
	// #=========================#

	// look for cached version
	cachePath := w.Opt.Dir.Cache + "/image/" + trueName
	wikifier.MakeDir(w.Opt.Dir.Cache+"/image/", trueName)
	cacheFi, err := os.Lstat(cachePath)

	// it exists
	if err == nil && cacheFi.ModTime().After(fi.ModTime()) {
		if cacheFi.ModTime().Before(fi.ModTime()) {

			// the original is newer, so forget the cached file
			w.Debugf("display image: %s: purging outdated cache", logName)
			os.Remove(cachePath)

		} else {

			// it exists and the cache file is newer
			w.Debugf("display image: %s: using cached version", logName)
			mod := cacheFi.ModTime()
			r.Path = cachePath
			r.File = filepath.Base(cachePath)
			r.FromCache = true
			r.Modified = &mod
			r.ModifiedHTTP = httpdate.Time2Str(mod)
			r.Length = cacheFi.Size()

			w.symlinkScaledImage(img, trueName)
			return r
		}
	}

	// #======================#
	// #=== Generate image ===#
	// #======================#

	// so if we made it all the way down to here, we need to
	// generate the image in specific dimensions

	// use image-specific locking to prevent duplicate generation (if needed)
	if needLock {
		imageLock := w.GetImageLock(trueName)
		imageLock.Lock()
		defer imageLock.Unlock()
	}

	// check cache again after acquiring lock (another process might have generated it)
	cacheFi, err = os.Lstat(cachePath)
	if err == nil && cacheFi.ModTime().After(fi.ModTime()) {
		w.Debugf("display image: %s: using cached version (generated while waiting for lock)", logName)
		mod := cacheFi.ModTime()
		r.Path = cachePath
		r.File = filepath.Base(cachePath)
		r.FromCache = true
		r.Modified = &mod
		r.ModifiedHTTP = httpdate.Time2Str(mod)
		r.Length = cacheFi.Size()
		w.symlinkScaledImage(img, trueName)
		return r
	}

	// after acquiring lock and checking cache, verify this is an allowed size
	// for non-pregeneration requests when arbitrary sizes are disabled
	if !generateOK && !w.Opt.Image.ArbitrarySizes {
		// get original image dimensions if not already known
		if bigW == 0 || bigH == 0 {
			bigW, bigH = getImageDimensions(bigPath)
		}

		// allow full-size images (either 0x0 or matching original dimensions)
		isFullSize := (img.Width == 0 && img.Height == 0) ||
			(img.TrueWidth() == bigW && img.TrueHeight() == bigH)

		if !isFullSize && !w.isImageSizeReferenced(img) {
			dimensions := strconv.Itoa(img.TrueWidth()) + "x" + strconv.Itoa(img.TrueHeight())
			return DisplayError{
				Error:         "Image size not allowed.",
				DetailedError: "Image '" + img.TrueName() + "' at " + dimensions + " is not referenced in wiki content.",
			}
		}
	}

	// generate the image
	// note: bigW and bigH might still be empty
	if nonBlocking {
		// in non-blocking mode, return an error immediately and queue for background generation
		// TODO: queue image for background generation here
		return DisplayError{
			Error:         "Image being generated.",
			DetailedError: "Image '" + img.TrueName() + "' is being generated in the background.",
		}
	}

	if dispErr := w.generateImage(img, bigPath, bigW, bigH, &r); dispErr != nil {
		return dispErr
	}

	w.symlinkScaledImage(img, trueName)
	return r
}

// Images returns info about all the images in the wiki.
func (w *Wiki) Images() []ImageInfo {
	imageNames := w.allImageFiles()
	return w.imagesIn("", imageNames)
}

// ImagesInDir returns info about all the images in the specified directory.
func (w *Wiki) ImagesInDir(where string) []ImageInfo {
	imageNames := w.imageFilesInDir(where)
	return w.imagesIn(where, imageNames)
}

func (w *Wiki) imagesIn(prefix string, imageNames []string) []ImageInfo {
	images := make([]ImageInfo, 0, len(imageNames))
	for _, name := range imageNames {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Error processing image %s: %v", filepath.Join(prefix, name), r)
				}
			}()
			info := w.ImageInfo(filepath.Join(prefix, name))
			if info.File != "" { // only add if we got valid info
				images = append(images, info)
			}
		}()
	}
	return images
}

type sortableImageInfo ImageInfo

func (ii sortableImageInfo) SortInfo() SortInfo {
	return SortInfo{
		// TODO: Author
		Title:      ii.Base,
		Created:    *ii.Created,
		Modified:   *ii.Modified,
		Dimensions: []int{ii.Width, ii.Height},
	}
}

// ImagesSorted returns info about all the images in the wiki, sorted as specified.
// Accepted sort functions are SortTitle, SortAuthor, SortCreated, SortModified, and SortDimensions.
func (w *Wiki) ImagesSorted(descend bool, sorters ...SortFunc) []ImageInfo {
	return _imagesSorted(w.Images(), descend, sorters...)
}

func _imagesSorted(images []ImageInfo, descend bool, sorters ...SortFunc) []ImageInfo {
	// convert to []Sortable
	sorted := make([]Sortable, len(images))
	for i, pi := range images {
		sorted[i] = sortableImageInfo(pi)
	}

	// sort
	var sorter sort.Interface = sorter(sorted, sorters...)
	if descend {
		sorter = sort.Reverse(sorter)
	}
	sort.Sort(sorter)

	// convert back to []ImageInfo
	for i, si := range sorted {
		images[i] = ImageInfo(si.(sortableImageInfo))
	}

	return images
}

// ImagesAndDirs returns info about all the images and directories in a directory.
func (w *Wiki) ImagesAndDirs(where string) ([]ImageInfo, []string) {
	images := w.ImagesInDir(where)
	dirs := utils.DirsInDir(filepath.Join(w.Opt.Dir.Image, where))
	return images, dirs
}

// ImagesAndDirsSorted returns info about all the images and directories in a directory, sorted as specified.
// Accepted sort functions are SortTitle, SortAuthor, SortCreated, SortModified, and SortDimensions.
// Directories are always sorted alphabetically (but still respect the descend flag).
func (w *Wiki) ImagesAndDirsSorted(where string, descend bool, sorters ...SortFunc) ([]ImageInfo, []string) {
	images, dirs := w.ImagesAndDirs(where)
	images = _imagesSorted(images, descend, sorters...)
	if descend {
		sort.Sort(sort.Reverse(sort.StringSlice(dirs)))
	} else {
		sort.Strings(dirs)
	}
	return images, dirs
}

// ImageMap returns a map of image filename to ImageInfo for all images in the wiki.
func (w *Wiki) ImageMap() map[string]ImageInfo {
	imageNames := w.allImageFiles()
	images := make(map[string]ImageInfo, len(imageNames))

	// images individually
	for _, name := range imageNames {
		images[name] = w.ImageInfo(name)
	}

	return images
}

// ImageInfo returns info for an image given its full-size name.
func (w *Wiki) ImageInfo(name string) (info ImageInfo) {

	// the image does not exist
	path := w.PathForImage(name)
	imgFi, err := os.Stat(path)
	if err != nil {
		return
	}

	mod := imgFi.ModTime()
	info.File = name
	info.Base = filepath.Base(name)
	info.Modified = &mod // actual image mod time

	// find image category
	imageCat := w.GetSpecialCategory(name, CategoryTypeImage)

	// it doesn't exist. let's create it
	if !imageCat.Exists() {
		imageCat.addImage(w, name, nil, nil)
	}

	// it should exist at this point
	if imageCat.Exists() {
		if imageCat.ImageInfo != nil {
			info.Width = imageCat.ImageInfo.Width
			info.Height = imageCat.ImageInfo.Height
		}
		info.Created = imageCat.Created // category creation time, not image
		for _, entry := range imageCat.Pages {
			info.Dimensions = append(info.Dimensions, entry.Dimensions...)
		}
		return
	}

	// image category still doesn't exist???
	// let's read the dimensions manually
	info.Width, info.Height = getImageDimensions(path)

	return
}

func (w *Wiki) generateImage(img SizedImage, bigPath string, bigW, bigH int, r *DisplayImage) any {
	width, height := img.TrueWidth(), img.TrueHeight()

	// use the auto processor for optimal performance
	imageProc := GetImageProcessorForWiki(w)

	// get dimensions efficiently without loading full image into memory
	bigWidth, bigHeight, err := GetImageDimensionsFromFile(bigPath)
	if err != nil {
		return DisplayError{
			Error:         "Failed to get image dimensions.",
			DetailedError: "Get dimensions for '" + bigPath + "' error: " + err.Error(),
		}
	}

	// figure out full-size dimensions if we haven't already
	if bigW == 0 || bigH == 0 {
		bigW = bigWidth
		bigH = bigHeight
	}

	// the request is to generate an image the same or larger than the original
	if width >= bigW || height >= bigH {

		// symlink this to the full-size image
		w.symlinkScaledImage(img, img.FullSizeName())

		return nil // success
	}

	// safe point - we will resize the image

	w.Debug("generate image:", img.TrueName())

	// create resized image directly from file to avoid loading into memory
	newImagePath := filepath.FromSlash(w.Opt.Dir.Cache + "/image/" + img.TrueName())

	// resize directly from file to file without loading into memory
	quality := w.Opt.Image.Quality
	if quality <= 0 {
		quality = 85 // default quality
	}
	err = imageProc.ResizeImageDirect(bigPath, newImagePath, width, height, quality)
	if err != nil {
		return DisplayError{
			Error:         "Failed to resize image.",
			DetailedError: "Resize image '" + bigPath + "' error: " + err.Error(),
		}
	}

	newImageFi, err := os.Lstat(newImagePath)
	if err != nil || newImageFi == nil {
		return DisplayError{
			Error:         "Failed to stat generated image.",
			DetailedError: "Stat image '" + newImagePath + "' error: " + err.Error(),
		}
	}

	// inject info from the newly generated image
	mod := newImageFi.ModTime()
	r.Path = newImagePath
	r.File = filepath.Base(newImagePath)
	r.Generated = true
	r.Modified = &mod
	r.ModifiedHTTP = httpdate.Time2Str(mod)
	r.Length = newImageFi.Size()
	r.CacheGenerated = true

	return nil // success
}

// symlinks scaled cache file e.g. 100x200-asdf@2x.jpg -> 200x400-asdf.jpg
func (w *Wiki) symlinkScaledImage(img SizedImage, name string) {

	// dumb request
	if img.Scale <= 1 {
		return
	}

	// only symlink if this is a supported scale
	ok := false
	for _, scale := range w.Opt.Image.Retina {
		if scale == img.Scale {
			ok = true
			break
		}
	}
	if !ok {
		return
	}

	w.Debugf("symlink image: %s -> %s", name, img.ScaleName())
	scalePath := filepath.FromSlash(w.Opt.Dir.Cache + "/image/" + img.ScaleName())
	os.Symlink(filepath.Base(name), scalePath)
}

func getImageDimensions(path string) (w, h int) {
	width, height, err := GetImageDimensionsFromFile(path)
	if err != nil {
		log.Printf("getImageDimensions: error getting dimensions for '%s': %v", path, err)
		return 0, 0
	}
	return width, height
}

// isImageSizeReferenced checks if the given image size is referenced in wiki content.
// This prevents arbitrary image size generation by checking the image category tracking.
func (w *Wiki) isImageSizeReferenced(img SizedImage) bool {
	// Full-size images (0x0) are always allowed
	if img.Width == 0 && img.Height == 0 {
		return true
	}

	// Check if this matches the original image dimensions (also considered full-size)
	bigPath := w.PathForImage(img.FullSizeName())
	if bigW, bigH := getImageDimensions(bigPath); bigW > 0 && bigH > 0 {
		if img.TrueWidth() == bigW && img.TrueHeight() == bigH {
			return true // original dimensions are always allowed
		}
	}

	// Get the image category that tracks all references to this image
	imageCat := w.GetSpecialCategory(img.FullSizeName(), CategoryTypeImage)

	// No category means no references exist
	if imageCat == nil || !imageCat.Exists() {
		return false
	}

	// Check if any page references this specific size
	targetW, targetH := img.TrueWidth(), img.TrueHeight()

	// Look through all pages that reference this image
	for _, pageEntry := range imageCat.Pages {
		// pageEntry.Dimensions contains the dimensions as [][]int
		for _, dimensionPair := range pageEntry.Dimensions {
			if len(dimensionPair) >= 2 {
				refW, refH := dimensionPair[0], dimensionPair[1]
				if refW == targetW && refH == targetH {
					return true
				}
			}
		}
	}

	return false
}

// determine missing dimensions
func calculateImageDimensions(bigW, bigH, width, height int) (int, int) {
	if width == 0 && height == 0 {
		return 0, 0
	}
	if width == 0 {
		tmpW := float64(height) * float64(bigW) / float64(bigH)
		width = int(math.Max(1.0, math.Floor(tmpW+0.5)))
	}
	if height == 0 {
		tmpH := float64(width) * float64(bigH) / float64(bigW)
		height = int(math.Max(1.0, math.Floor(tmpH+0.5)))
	}
	return width, height
}
