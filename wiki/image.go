package wiki

import (
	"fmt"
	"image"
	_ "image/jpeg" // for jpegs
	_ "image/png"  // for pngs
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	httpdate "github.com/Songmu/go-httpdate"
	"github.com/disintegration/imaging"
)

var (
	imageNameRegex  = regexp.MustCompile(`^(\d+)x(\d+)-(.+)$`)
	imageScaleRegex = regexp.MustCompile(`^(.+)\@(\d+)x$`)
)

// SizedImage represents an image in specific dimensions.
type SizedImage struct {
	// for example 100x200-myimage@3x.png
	Width, Height int    // 100, 200
	Scale         int    // 3
	Name          string // myimage
	Ext           string // png
}

func SizedImageFromName(name string) SizedImage {
	w, h := 0, 0

	// width and height were given, so it's a resized image
	if matches := imageNameRegex.FindStringSubmatch(name); len(matches) != 0 {
		w, _ = strconv.Atoi(matches[1])
		h, _ = strconv.Atoi(matches[2])
		name = matches[3]
	}

	// default true width and height to the same
	trueW, trueH := w, h

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
		trueW *= scale
		trueH *= scale
	}

	// put it all together
	return SizedImage{
		Width:  w,
		Height: h,
		Scale:  scale,
		Name:   nameNE,
		Ext:    ext,
	}
}

func (img SizedImage) TrueWidth() int {
	return img.Width * img.Scale
}

func (img SizedImage) TrueHeight() int {
	return img.Height * img.Scale
}

func (img SizedImage) FullSizeName() string {
	return img.Name + "." + img.Ext
}

func (img SizedImage) FullNameNE() string {
	return fmt.Sprintf(
		"%dx%d-%s",
		img.TrueWidth(),
		img.TrueHeight(),
		img.Name,
	)
}

func (img SizedImage) FullName() string {
	return img.FullNameNE() + "." + img.Ext
}

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

	// UNIX timestamp of when the image was last modified.
	// if Generated is true, this is the current time.
	// if FromCache is true, this is the modified date of the cache file.
	// otherwise, this is the modified date of the image file itself.
	ModUnix int64 `json:"mod_unix,omitempty"`

	// like ModUnix except in HTTP date format
	// suitable for Last-Modified header
	Modified string `json:"modified,omitempty"`

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
	// my $result = {};
	var r DisplayImage

	// # if $image_name is an array ref, it's [ name, width, height, scale ]
	// # if both dimensions are 0, parse the image name normally
	// if (ref $image_name eq 'ARRAY') {
	//     my ($name, $w, $h, $s) = @$image_name;
	//     $w //= 0;
	//     $h //= 0;
	//     $image_name  = "${w}x${h}-$name";
	//     $image_name  = $name    if !$w && !$h;
	//     $image_name .= "@${s}x" if $s;
	// }

	// # parse the image name.
	// my $image = $wiki->parse_image_name($image_name);
	// $image_name = $image->{name};

	// # check if the file exists.
	// my $big_path = $wiki->path_for_image($image_name);
	// return display_error('Image does not exist.')
	//     if !-f $big_path;

	// check if the file exists
	path := w.pathForImage(img.FullSizeName())
	fi, err := os.Lstat(path)
	if err != nil {
		return DisplayError{
			Error:         "Image does not exist.",
			DetailedError: "Image '" + path + "' error: " + err.Error(),
		}
	}

	// image name and full path
	r.Path = path
	r.FullsizePath = path
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
			DetailedError: "Image '" + path + "' is neither png nor jpeg",
		}
	}

	// TODO: update image category as needed

	// if both dimensions are missing, display the full-size version of the image
	if img.Width == 0 && img.Height == 0 {
		r.Modified = httpdate.Time2Str(fi.ModTime())
		r.ModUnix = fi.ModTime().Unix()
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
			scaledImage := img // copy
			img.Scale = scale  // set scale
			w.DisplaySizedImageGenerate(scaledImage, true)
		}
	}

	// #=========================#
	// #=== Find cached image ===#
	// #=========================#

	// look for cached version
	cachePath := w.Opt.Dir.Cache + "/image/" + img.FullName()
	cacheFi, err := os.Lstat(cachePath)

	// it exists
	if err == nil && cacheFi.ModTime().After(fi.ModTime()) {
		if cacheFi.ModTime().Before(fi.ModTime()) {

			// the original is newer, so forget the cached file
			os.Remove(cachePath)

		} else {

			// it exists and the cache file is newer
			r.Path = cachePath
			r.File = filepath.Base(cachePath)
			r.FromCache = true
			r.Modified = httpdate.Time2Str(cacheFi.ModTime())
			r.ModUnix = cacheFi.ModTime().Unix()
			r.Length = cacheFi.Size()

			// symlink if necessary
			w.symlinkScaledImage(img, img.FullName())

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
	// if !generateOK {
	// 	dimensions := strconv.Itoa(img.TrueWidth()) + "x" + strconv.Itoa(img.TrueHeight())
	// 	return DisplayError{Error: "Image does not exist at " + dimensions + "."}
	// }

	// # generate the image
	// my $err = $wiki->_generate_image($image, $cache_path, $result);
	// return $err if $err;

	// generate the image
	dispErr, useFullSize := w.generateImage(img, cachePath, &r)
	if dispErr != nil {
		return dispErr
	}

	// # the generator says to use the full-size image.
	// return $wiki->_get_image_full_size($image, $result, \@stat, \%opts)
	//     if delete $result->{use_fullsize};
	if useFullSize {

	}

	// delete $result->{content} if $opts{dont_open};
	// return $result;

	return r
}

func (w *Wiki) generateImage(img SizedImage, cachePath string, r *DisplayImage) (dispErr interface{}, useFullSize bool) {
	// my ($wiki, $image, $cache_path, $result) = @_;

	// # if we are restricting to only sizes used in the wiki, check.
	// my ($width, $height, $r_width, $r_height) =
	//     @$image{ qw(width height r_width r_height) };

	// # create GD instance with this full size image.
	// GD::Image->trueColor(1);
	// my $full_image = GD::Image->new($result->{fullsize_path});
	// return display_error("Couldn't handle image $$result{fullsize_path}")
	//     if !$full_image;
	// my ($fi_width, $fi_height) = $full_image->getBounds();

	width, height := img.TrueWidth(), img.TrueHeight()

	// open the full size image
	fsPath := w.pathForImage(img.FullSizeName())
	fsFile, err := os.Open(fsPath)
	defer fsFile.Close()
	if err != nil {
		dispErr = DisplayError{
			Error:         "Image does not exist.",
			DetailedError: "Open image '" + fsPath + "' error: " + err.Error(),
		}
		return
	}

	// decode it
	fsConfig, _, _ := image.DecodeConfig(fsFile)
	fsImage, err := imaging.Open(fsPath)
	if err != nil {
		dispErr = DisplayError{
			Error:         "Image does not exist.",
			DetailedError: "Decode image '" + fsPath + "' error: " + err.Error(),
		}
		return
	}

	// the request is to generate an image the same or larger than the original
	if width >= fsConfig.Width && height >= fsConfig.Height {
		useFullSize = true

		// symlink this to the full-size image
		w.symlinkScaledImage(img, img.FullSizeName())

		return // success
	}

	// create resized image
	// note: if either width or height is 0, the aspect is preserved
	newImage := imaging.Resize(fsImage, width, height, imaging.Lanczos)

	// generate the image in the source format and write
	newImagePath := w.Opt.Dir.Cache + "/image/" + img.FullName()
	err = imaging.Save(newImage, newImagePath)
	if err != nil {
		dispErr = DisplayError{
			Error:         "Failed to generate image.",
			DetailedError: "Save image '" + fsPath + "' error: " + err.Error(),
		}
		return
	}
	newImageFi, _ := os.Lstat(newImagePath)

	// inject info from the newly generated image
	r.Path = cachePath
	r.File = filepath.Base(cachePath)
	r.Generated = true
	r.Modified = httpdate.Time2Str(newImageFi.ModTime())
	r.ModUnix = newImageFi.ModTime().Unix()
	r.Length = newImageFi.Size()
	r.CacheGenerated = true

	return // success
}

func (w *Wiki) symlinkScaledImage(img SizedImage, name string) {
	if img.Scale <= 1 {
		return
	}
	scalePath := w.Opt.Dir.Cache + "/image/" + img.ScaleName()
	_, err := os.Lstat(scalePath)
	if err != nil {
		os.Symlink(name, scalePath)
	}
}
