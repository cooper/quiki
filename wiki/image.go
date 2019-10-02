package wiki

import (
	"fmt"
	"image"
	_ "image/jpeg" // for jpegs
	_ "image/png"  // for pngs
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	httpdate "github.com/Songmu/go-httpdate"
	"github.com/cooper/quiki/wikifier"
	"github.com/disintegration/imaging"
)

var (
	imageNameRegex  = regexp.MustCompile(`^(\d+)x(\d+)-(.+)$`)
	imageScaleRegex = regexp.MustCompile(`^(.+)\@(\d+)x$`)
)

// ImageInfo represents a full-size image on the wiki.
type ImageInfo struct {
	File       string     `json:"file"`               // filename
	Width      int        `json:"width,omitempty"`    // full-size width
	Height     int        `json:"height,omitempty"`   // full-size height
	Created    *time.Time `json:"created,omitempty"`  // creation time
	Modified   *time.Time `json:"modified,omitempty"` // modify time
	Dimensions [][]int    `json:"-"`                  // dimensions used throughout the wiki
}

// SizedImage represents an image in specific dimensions.
type SizedImage struct {
	// for example 100x200-myimage@3x.png
	Width, Height int    // 100, 200 (dimensions as requested)
	Scale         int    // 3 (scale as requested)
	Name          string // myimage (name without extension)
	Ext           string // png (extension)
	zeroByZero    bool   // true when created from 0x0-name
}

// SizedImageFromName returns a SizedImage given an image name.
func SizedImageFromName(name string) SizedImage {
	w, h := 0, 0
	zeroByZero := false

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
		name = matches[1] + "." + ext
	}

	// put it all together
	return SizedImage{
		Width:      w,
		Height:     h,
		Scale:      scale,
		Name:       nameNE,
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
	return img.Name + "." + img.Ext
}

// FullNameNE is like FullName but without the extension.
func (img SizedImage) FullNameNE() string {
	if img.Width == 0 && img.Height == 0 {
		return img.Name
	}
	return fmt.Sprintf(
		"%dx%d-%s",
		img.TrueWidth(),
		img.TrueHeight(),
		img.Name,
	)
}

// FullName returns the image name with true dimensions.
func (img SizedImage) FullName() string {
	return img.FullNameNE() + "." + img.Ext
}

// ScaleName returns the image name with dimensions and scale.
func (img SizedImage) ScaleName() string {
	if img.Scale <= 1 {
		return img.FullName()
	}
	return fmt.Sprintf("%dx%d-%s@%dx.%s",
		img.Width,
		img.Height,
		img.Name,
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
func (w *Wiki) DisplayImage(name string) interface{} {
	return w.DisplaySizedImageGenerate(SizedImageFromName(name), false)
}

// DisplaySizedImage returns the display result for an image in specific dimensions.
func (w *Wiki) DisplaySizedImage(img SizedImage) interface{} {
	return w.DisplaySizedImageGenerate(img, false)
}

// DisplaySizedImageGenerate returns the display result for an image in specific dimensions
// and allows images to be generated in any dimension.
func (w *Wiki) DisplaySizedImageGenerate(img SizedImage, generateOK bool) interface{} {
	var r DisplayImage

	// check if the file exists
	bigPath := w.pathForImage(img.FullSizeName())
	fi, err := os.Lstat(bigPath)
	if err != nil {
		return DisplayError{
			Error:         "Image does not exist.",
			DetailedError: "Image '" + bigPath + "' error: " + err.Error(),
		}
	}

	// get full size dimensions
	bigW, bigH := getImageDimensions(bigPath)

	// find missing dimension
	// note: we haven't checked if both are 0 yet, but this will return 0, 0 in that case
	oldName := img.FullName()
	img.Width, img.Height = calculateImageDimensions(bigW, bigH, img.Width, img.Height)

	// check if the name has changed after this adjustment.
	// if so, redirect
	fullName := img.FullName()
	if fullName != oldName || img.zeroByZero {
		return DisplayRedirect{Redirect: w.Opt.Root.Image + "/" + fullName}
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
	w.GetSpecialCategory(r.File, CategoryTypeImage).addImage(w, r.File, nil, nil)

	// if both dimensions are missing, display the full-size version of the image
	if img.Width == 0 && img.Height == 0 {
		mod := fi.ModTime()
		r.Modified = &mod
		r.ModifiedHTTP = httpdate.Time2Str(mod)
		r.Length = fi.Size()
		return r
	}

	// at this point, at least one dimension is present

	// #============================#
	// #=== Retina scale support ===#
	// #============================#

	// this is not a retina request, but retina is enabled, and
	// this is a pregeneration request of the normal-scale image.
	// so, commit a pregeneration request for each scaled version.
	if img.Scale <= 1 && generateOK {
		for _, scale := range w.Opt.Image.Retina {
			scaledImage := img        // copy
			scaledImage.Scale = scale // set scale
			w.DisplaySizedImageGenerate(scaledImage, true)
		}
	}

	// #=========================#
	// #=== Find cached image ===#
	// #=========================#

	// look for cached version
	cachePath := w.Opt.Dir.Cache + "/image/" + fullName
	wikifier.MakeDir(w.Opt.Dir.Cache+"/image/", fullName)
	cacheFi, err := os.Lstat(cachePath)

	// it exists
	if err == nil && cacheFi.ModTime().After(fi.ModTime()) {
		if cacheFi.ModTime().Before(fi.ModTime()) {

			// the original is newer, so forget the cached file
			os.Remove(cachePath)

		} else {

			// it exists and the cache file is newer
			mod := cacheFi.ModTime()
			r.Path = cachePath
			r.File = filepath.Base(cachePath)
			r.FromCache = true
			r.Modified = &mod
			r.ModifiedHTTP = httpdate.Time2Str(mod)
			r.Length = cacheFi.Size()

			w.symlinkScaledImage(img, fullName)
			return r
		}
	}

	// #======================#
	// #=== Generate image ===#
	// #======================#

	// so if we made it all the way down to here, we need to
	// generate the image in specific dimensions

	// we're not allowed to do this if this is a legit (non-pregeneration)
	// request. because like, we would've served a cached image if it were
	// actually used somewhere on the wiki

	// FIXME: this is disabled for now
	// if !generateOK {
	// 	dimensions := strconv.Itoa(img.TrueWidth()) + "x" + strconv.Itoa(img.TrueHeight())
	// 	return DisplayError{Error: "Image does not exist at " + dimensions + "."}
	// }

	// generate the image
	if dispErr := w.generateImage(img, bigPath, bigW, bigH, &r); dispErr != nil {
		return dispErr
	}

	w.symlinkScaledImage(img, fullName)
	return r
}

// Images returns info about all the images in the wiki.
func (w *Wiki) Images() map[string]ImageInfo {
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
	path := w.pathForImage(name)
	imgFi, err := os.Stat(path)
	if err != nil {
		return
	}

	mod := imgFi.ModTime()
	info.File = name
	info.Modified = &mod // actual image mod time

	// find image category
	imageCat := w.GetSpecialCategory(name, CategoryTypeImage)

	// it doesn't exist. let's create it
	if !imageCat.Exists() {
		imageCat.addImage(w, name, nil, nil)
	}

	// it should exist at this point
	if imageCat.Exists() {
		info.Width = imageCat.ImageInfo.Width
		info.Height = imageCat.ImageInfo.Height
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

func (w *Wiki) generateImage(img SizedImage, bigPath string, bigW, bigH int, r *DisplayImage) interface{} {
	width, height := img.TrueWidth(), img.TrueHeight()

	// open the full-size image
	bigImage, err := imaging.Open(bigPath)
	if err != nil {
		return DisplayError{
			Error:         "Image does not exist.",
			DetailedError: "Decode image '" + bigPath + "' error: " + err.Error(),
		}
	}

	// the request is to generate an image the same or larger than the original
	if width >= bigW || height >= bigH {

		// symlink this to the full-size image
		w.symlinkScaledImage(img, img.FullSizeName())

		return nil // success
	}

	// create resized image
	newImage := imaging.Resize(bigImage, width, height, imaging.Lanczos)

	// generate the image in the source format and write
	newImagePath := w.Opt.Dir.Cache + "/image/" + img.FullName()
	err = imaging.Save(newImage, newImagePath)
	if err != nil {
		return DisplayError{
			Error:         "Failed to generate image.",
			DetailedError: "Save image '" + bigPath + "' error: " + err.Error(),
		}
	}

	newImageFi, _ := os.Lstat(newImagePath)

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

func (w *Wiki) symlinkScaledImage(img SizedImage, name string) {
	if img.Scale <= 1 {
		return
	}
	scalePath := w.Opt.Dir.Cache + "/image/" + img.ScaleName()
	os.Symlink(name, scalePath)
}

func getImageDimensions(path string) (w, h int) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return
	}
	c, _, _ := image.DecodeConfig(file)
	w = c.Width
	h = c.Height
	return
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
