package wiki

import (
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
	}
}

// acquire waits for a processing slot and returns a func to release it.
// it gives up after the processor timeout rather than failing immediately, so
// a burst of requests (e.g. the adminifier image grid) queues up instead of
// producing a pile of broken images.
func (p *ImageProcessor) acquire() (release func(), err error) {
	timer := time.NewTimer(p.timeout)
	defer timer.Stop()
	select {
	case p.semaphore <- struct{}{}:
		return func() { <-p.semaphore }, nil
	case <-timer.C:
		return nil, fmt.Errorf("timed out waiting for an image processing slot")
	}
}

// checkImageMemory rejects images that would use too much memory when decoded.
// it reads only the header, so the image itself is never loaded.
func (p *ImageProcessor) checkImageMemory(path string) error {
	width, height, err := GetImageDimensionsFromFile(path)
	if err != nil {
		return err
	}

	// estimate memory usage (width * height * 4 bytes for RGBA)
	estimatedMB := int64(width) * int64(height) * 4 / (1024 * 1024)
	if estimatedMB > p.maxMemoryMB {
		return fmt.Errorf("image too large: %dx%d (~%dMB) exceeds limit of %dMB",
			width, height, estimatedMB, p.maxMemoryMB)
	}
	return nil
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

// ResizeImageDirect implements ImageProcessorInterface - file-to-file resize using pure go.
// unlike the external processors, this has to decode the whole image into memory,
// so it is guarded by the memory estimate, the concurrency limit, and the timeout.
func (p *ImageProcessor) ResizeImageDirect(inputPath, outputPath string, width, height, quality int) error {
	if err := p.checkImageMemory(inputPath); err != nil {
		return err
	}

	release, err := p.acquire()
	if err != nil {
		return err
	}

	// do the work in a goroutine so the caller can give up after the timeout.
	// the slot is held until the work actually finishes, since the memory is in
	// use until then whether or not anyone is still waiting for the result.
	// a panic in the decoder must be recovered here: an unrecovered panic in a
	// goroutine takes down the whole server, not just the request.
	done := make(chan error, 1)
	go func() {
		defer release()
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("image processing panic: %v", r)
			}
		}()
		done <- resizeWithImaging(inputPath, outputPath, width, height, quality)
	}()

	timer := time.NewTimer(p.timeout)
	defer timer.Stop()
	select {
	case err := <-done:
		return err
	case <-timer.C:
		return fmt.Errorf("image processing timeout: %s", inputPath)
	}
}

// resizeWithImaging does the actual decode, resize, and encode in memory
func resizeWithImaging(inputPath, outputPath string, width, height, quality int) error {
	img, err := imaging.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open image: %v", err)
	}
	resized := imaging.Resize(img, width, height, imaging.Lanczos)

	var opts []imaging.EncodeOption
	if quality > 0 {
		opts = append(opts, imaging.JPEGQuality(quality))
	}
	return imaging.Save(resized, outputPath, opts...)
}

// withConcurrencyControl executes fn (an external resize process) with memory-aware
// concurrency control
func (p *ImageProcessor) withConcurrencyControl(fn func() error) error {
	// try to acquire worker with memory awareness using global monitor
	memMonitor := GetMemoryMonitor()
	maxRetries := 50                    // increased from 10
	retryDelay := 10 * time.Millisecond // reduced from 100ms

	for i := 0; i < maxRetries; i++ {
		if memMonitor.acquireWorker() {
			defer memMonitor.releaseWorker()
			return fn()
		}

		// if we can't get a worker due to memory constraints, wait and retry
		// much more aggressive retry for better parallelization
		if i < maxRetries-1 {
			time.Sleep(retryDelay)
			// much slower exponential backoff - only increase every 5 attempts
			if i > 0 && i%5 == 0 {
				retryDelay = min(retryDelay*2, 100*time.Millisecond) // cap at 100ms
			}
		}
	}

	// if we still can't get a worker after retries, fall back to the processor's
	// own concurrency limit, waiting for a slot rather than failing outright
	release, err := p.acquire()
	if err != nil {
		return fmt.Errorf("system overloaded: %v", err)
	}
	defer release()
	return fn()
}
