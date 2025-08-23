package wiki

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/cooper/imaging"
)

// AutoImageProcessor tries processors in order: libvips -> imagemagick -> pure go
type AutoImageProcessor struct {
	vips        *VipsProcessor
	imagemagick *ImageMagickProcessor
	pureGo      *ImageProcessor
	mu          sync.RWMutex
	stats       ProcessorStats
}

// cache availability checks so we only log not-found once during startup
var (
	libvipsOnce sync.Once
	libvipsErr  error

	magickOnce sync.Once
	magickErr  error
)

func checkLibvipsOnce() error {
	libvipsOnce.Do(func() {
		libvipsErr = CheckLibvipsAvailable()
		if libvipsErr != nil {
			// log not-found only once
			log.Printf("image processor: libvips not available: %v", libvipsErr)
		}
	})
	return libvipsErr
}

func checkImageMagickOnce() error {
	magickOnce.Do(func() {
		magickErr = CheckImageMagickAvailable()
		if magickErr != nil {
			// log not-found only once
			log.Printf("image processor: ImageMagick not available: %v", magickErr)
		}
	})
	return magickErr
}

// NewAutoImageProcessor creates a processor that tries libvips -> imagemagick -> pure go
func NewAutoImageProcessor(opts ImageProcessorOptions) *AutoImageProcessor {
	c := &AutoImageProcessor{
		pureGo: NewImageProcessor(opts), // always available as final fallback
	}

	// try to initialize libvips (highest priority)
	if err := checkLibvipsOnce(); err == nil {
		vipsOpts := DefaultVipsOptions()
		vipsOpts.MaxConcurrent = opts.MaxConcurrent
		vipsOpts.Timeout = opts.Timeout

		if processor, err := NewVipsProcessor(vipsOpts); err == nil {
			c.vips = processor
			log.Printf("image processor: using libvips auto (libvips -> imagemagick -> pure go)")
		} else {
			log.Printf("image processor: libvips initialization failed: %v", err)
		}
	}

	// try to initialize ImageMagick (second priority)
	if err := checkImageMagickOnce(); err == nil {
		magickOpts := DefaultImageMagickOptions()
		magickOpts.MaxConcurrent = opts.MaxConcurrent
		magickOpts.Timeout = opts.Timeout

		if processor, err := NewImageMagickProcessor(magickOpts); err == nil {
			c.imagemagick = processor
			if c.vips == nil {
				log.Printf("image processor: using imagemagick auto (imagemagick -> pure go)")
			}
		} else {
			log.Printf("image processor: ImageMagick initialization failed: %v", err)
		}
	}

	if c.vips == nil && c.imagemagick == nil {
		log.Printf("image processor: using pure go fallback (warning: poor performance with large images)")
		log.Printf("image processor: install libvips ('brew install vips') or imagemagick ('brew install imagemagick') for better performance")
	}

	return c
}

// ResizeImageDirect resizes directly from file to file without loading into memory
func (c *AutoImageProcessor) ResizeImageDirect(inputPath, outputPath string, width, height, quality int) error {
	c.mu.Lock()
	c.stats.TotalProcessed++
	c.mu.Unlock()

	// try libvips first if available (fastest and most memory efficient)
	if c.vips != nil {
		err := c.vips.ResizeImageVips(inputPath, outputPath, width, height, quality)
		if err == nil {
			c.mu.Lock()
			c.stats.VipsSuccess++
			c.mu.Unlock()
			return nil
		}

		c.mu.Lock()
		c.stats.VipsFailed++
		c.mu.Unlock()
		log.Printf("image processor: libvips resize failed (%v), trying imagemagick", err)
	}

	// try imagemagick second
	if c.imagemagick != nil {
		err := c.imagemagick.ResizeImage(inputPath, outputPath, width, height, quality)
		if err == nil {
			c.mu.Lock()
			c.stats.ImageMagickSuccess++
			c.mu.Unlock()
			return nil
		}

		c.mu.Lock()
		c.stats.ImageMagickFailed++
		c.mu.Unlock()
		log.Printf("image processor: imagemagick resize failed (%v), falling back to pure go", err)
	}

	// fallback to pure go (requires loading into memory)
	c.mu.Lock()
	c.stats.PureGoUsed++
	c.mu.Unlock()

	return c.resizeWithPureGo(inputPath, outputPath, width, height)
}

// resizeWithPureGo handles file-to-file resize using pure go (loads into memory)
func (c *AutoImageProcessor) resizeWithPureGo(inputPath, outputPath string, width, height int) error {
	// load image into memory (only for pure go fallback)
	img, err := c.pureGo.safeImageOpen(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open image: %v", err)
	}

	// resize in memory
	resized, err := c.pureGo.safeImageResize(img, width, height)
	if err != nil {
		return fmt.Errorf("failed to resize image: %v", err)
	}

	// save to output file
	return imaging.Save(resized, outputPath)
}

// GetStats returns processor usage statistics
func (c *AutoImageProcessor) GetStats() ProcessorStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

// GetImageProcessorForWiki returns a processor configured for a specific wiki
func GetImageProcessorForWiki(w *Wiki) ImageProcessorInterface {
	// check what processor is configured
	processor := w.Opt.Image.Processor
	if processor == "" {
		processor = "auto" // default to auto selection
	}

	opts := DefaultImageProcessorOptions()

	// override with wiki-specific settings
	if w.Opt.Image.MaxConcurrent > 0 {
		opts.MaxConcurrent = w.Opt.Image.MaxConcurrent
	}
	if w.Opt.Image.MaxMemoryMB > 0 {
		opts.MaxMemoryMB = w.Opt.Image.MaxMemoryMB
	}
	if w.Opt.Image.TimeoutSeconds > 0 {
		opts.Timeout = time.Duration(w.Opt.Image.TimeoutSeconds) * time.Second
	}

	switch processor {
	case "go":
		w.Log("image processor: using pure go implementation; consider installing libvips or imagemagick for better performance")
		return NewImageProcessor(opts)

	case "vips":
		if err := checkLibvipsOnce(); err == nil {
			vipsOpts := DefaultVipsOptions()
			vipsOpts.MaxConcurrent = opts.MaxConcurrent
			vipsOpts.Timeout = opts.Timeout
			if w.Opt.Image.Quality > 0 {
				vipsOpts.Quality = w.Opt.Image.Quality
			}

			if processor, err := NewVipsProcessor(vipsOpts); err == nil {
				w.Log("image processor: using libvips for highest performance")
				return processor
			} else {
				w.Log(fmt.Sprintf("image processor: libvips failed: %v, falling back to pure go", err))
				w.Log("image processor: ensure libvips is properly installed: 'brew install vips' or 'apt-get install libvips-tools'")
				return NewImageProcessor(opts)
			}
		}
		// libvips not available (already logged once by checkLibvipsOnce) - fall back to pure go
		return NewImageProcessor(opts)

	case "imagemagick":
		if err := checkImageMagickOnce(); err == nil {
			magickOpts := DefaultImageMagickOptions()
			magickOpts.MaxConcurrent = opts.MaxConcurrent
			magickOpts.Timeout = opts.Timeout
			if w.Opt.Image.Quality > 0 {
				magickOpts.Quality = w.Opt.Image.Quality
			}

			if processor, err := NewImageMagickProcessor(magickOpts); err == nil {
				w.Log("image processor: using imagemagick for high performance")
				return processor
			} else {
				w.Log(fmt.Sprintf("image processor: imagemagick failed: %v, falling back to pure go", err))
				w.Log("image processor: ensure imagemagick is properly installed: 'brew install imagemagick' or 'apt-get install imagemagick'")
				return NewImageProcessor(opts)
			}
		}
		// ImageMagick not available (already logged once by checkImageMagickOnce) - fall back to pure go
		return NewImageProcessor(opts)

	case "auto":
		fallthrough
	default:
		w.Log("image processor: using auto selection (libvips -> imagemagick -> pure go)")
		return NewAutoImageProcessor(opts)
	}
}
