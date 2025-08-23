package wiki

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"sync"
	"time"

	"github.com/cooper/imaging"
)

// ImageProcessorInterface defines the common interface for all image processors
type ImageProcessorInterface interface {
	// direct file-to-file processing (avoids loading into memory)
	ResizeImageDirect(inputPath, outputPath string, width, height, quality int) error
}

// ProcessorStats tracks which processor is being used
type ProcessorStats struct {
	VipsSuccess        int
	VipsFailed         int
	ImageMagickSuccess int
	ImageMagickFailed  int
	PureGoUsed         int
	TotalProcessed     int
}

// ImageProcessor handles safe, concurrent image processing with resource limits
type ImageProcessor struct {
	semaphore   chan struct{} // limits concurrent processing
	maxMemoryMB int64         // max memory per image in MB
	timeout     time.Duration // max processing time per image
	mu          sync.RWMutex
	processing  map[string]bool // tracks images currently being processed
}

// ImageProcessorOptions configures the image processor
type ImageProcessorOptions struct {
	MaxConcurrent int           // max concurrent image operations
	MaxMemoryMB   int64         // max memory per image (width * height * 4 bytes)
	Timeout       time.Duration // max time per image operation
}

// DefaultImageProcessorOptions returns sensible defaults for cheap VPS
func DefaultImageProcessorOptions() ImageProcessorOptions {
	return ImageProcessorOptions{
		MaxConcurrent: 2,                // conservative for cheap VPS
		MaxMemoryMB:   256,              // 256MB per image for 2-4GB RAM systems
		Timeout:       20 * time.Second, // 20 seconds for cheaper hardware
	}
}

// global processor instance
var globalImageProcessor *ImageProcessor
var imageProcessorOnce sync.Once

// GetImageProcessor returns the global image processor instance
func GetImageProcessor() *ImageProcessor {
	imageProcessorOnce.Do(func() {
		opts := DefaultImageProcessorOptions()
		globalImageProcessor = NewImageProcessor(opts)
	})
	return globalImageProcessor
}

// GetPureGoImageProcessorForWiki returns an image processor configured for a specific wiki
func GetPureGoImageProcessorForWiki(w *Wiki) *ImageProcessor {
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

	return NewImageProcessor(opts)
}

// NewImageProcessor creates a new image processor with the given options
func NewImageProcessor(opts ImageProcessorOptions) *ImageProcessor {
	return &ImageProcessor{
		semaphore:   make(chan struct{}, opts.MaxConcurrent),
		maxMemoryMB: opts.MaxMemoryMB,
		timeout:     opts.Timeout,
		processing:  make(map[string]bool),
	}
}

// safeImageOpen opens an image with memory and timeout limits
func (p *ImageProcessor) safeImageOpen(path string) (image.Image, error) {
	// check if already processing this image
	p.mu.Lock()
	if p.processing[path] {
		p.mu.Unlock()
		return nil, fmt.Errorf("image already being processed: %s", path)
	}
	p.processing[path] = true
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		delete(p.processing, path)
		p.mu.Unlock()
	}()

	// acquire semaphore to limit concurrent processing
	select {
	case p.semaphore <- struct{}{}:
		defer func() { <-p.semaphore }()
	default:
		return nil, fmt.Errorf("too many concurrent image operations")
	}

	// first, check image dimensions without fully decoding
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image config: %v", err)
	}

	// estimate memory usage (width * height * 4 bytes for RGBA)
	estimatedMB := int64(config.Width * config.Height * 4 / (1024 * 1024))
	if estimatedMB > p.maxMemoryMB {
		return nil, fmt.Errorf("image too large: %dx%d (~%dMB) exceeds limit of %dMB",
			config.Width, config.Height, estimatedMB, p.maxMemoryMB)
	}

	// create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	// channel to receive result
	resultCh := make(chan struct {
		img image.Image
		err error
	}, 1)

	// process image in goroutine
	go func() {
		img, err := imaging.Open(path)
		resultCh <- struct {
			img image.Image
			err error
		}{img, err}
	}()

	// wait for result or timeout
	select {
	case result := <-resultCh:
		return result.img, result.err
	case <-ctx.Done():
		return nil, fmt.Errorf("image processing timeout: %s", path)
	}
}

// safeImageResize resizes an image with timeout protection
func (p *ImageProcessor) safeImageResize(img image.Image, width, height int) (image.Image, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	resultCh := make(chan struct {
		img image.Image
		err error
	}, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultCh <- struct {
					img image.Image
					err error
				}{nil, fmt.Errorf("resize panic: %v", r)}
			}
		}()

		resized := imaging.Resize(img, width, height, imaging.Lanczos)
		resultCh <- struct {
			img image.Image
			err error
		}{resized, nil}
	}()

	select {
	case result := <-resultCh:
		return result.img, result.err
	case <-ctx.Done():
		return nil, fmt.Errorf("image resize timeout")
	}
}

// GetImageDimensionsFromFile efficiently reads image dimensions from file header without loading the full image
func GetImageDimensionsFromFile(path string) (width, height int, err error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to open image file: %v", err)
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode image config: %v", err)
	}

	return config.Width, config.Height, nil
}

// GetImageDimensionsSafe safely gets image dimensions without loading the full image
func (p *ImageProcessor) GetImageDimensionsSafe(path string) (width, height int, err error) {
	return GetImageDimensionsFromFile(path)
}

// ResizeImageDirect implements ImageProcessorInterface - file-to-file resize using pure go
func (p *ImageProcessor) ResizeImageDirect(inputPath, outputPath string, width, height, quality int) error {
	// for pure go processor, we have to load into memory
	img, err := p.safeImageOpen(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open image: %v", err)
	}

	resized, err := p.safeImageResize(img, width, height)
	if err != nil {
		return fmt.Errorf("failed to resize image: %v", err)
	}

	return imaging.Save(resized, outputPath)
}

// withConcurrencyControl executes a function with proper concurrency control
func (p *ImageProcessor) withConcurrencyControl(inputPath string, fn func() error) error {
	// check if already processing this image
	p.mu.Lock()
	if p.processing[inputPath] {
		p.mu.Unlock()
		return fmt.Errorf("image already being processed: %s", inputPath)
	}
	p.processing[inputPath] = true
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		delete(p.processing, inputPath)
		p.mu.Unlock()
	}()

	// acquire semaphore to limit concurrent processing (blocking)
	p.semaphore <- struct{}{}
	defer func() { <-p.semaphore }()

	return fn()
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
