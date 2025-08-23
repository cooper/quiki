package wiki

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// VipsProcessor handles image processing using libvips (requires libvips installation)
type VipsProcessor struct {
	*ImageProcessor     // embed for common functionality
	quality         int // JPEG quality
}

// VipsOptions configures the libvips processor
type VipsOptions struct {
	MaxConcurrent int           // max concurrent image operations
	Timeout       time.Duration // max time per image operation
	Quality       int           // JPEG quality (1-100)
	MaxPixels     int64         // max pixels to prevent issues
}

// DefaultVipsOptions returns sensible defaults
func DefaultVipsOptions() VipsOptions {
	return VipsOptions{
		MaxConcurrent: 8,                // libvips is very efficient with concurrency
		Timeout:       20 * time.Second, // libvips is much faster than imagemagick
		Quality:       85,
		MaxPixels:     500_000_000, // 500MP limit - libvips handles very large images efficiently
	}
}

// NewVipsProcessor creates a new libvips-based processor
func NewVipsProcessor(opts VipsOptions) (*VipsProcessor, error) {
	// check if vips is available
	if err := CheckLibvipsAvailable(); err != nil {
		return nil, err
	}

	// create base processor with common functionality
	baseOpts := ImageProcessorOptions{
		MaxConcurrent: opts.MaxConcurrent,
		Timeout:       opts.Timeout,
		MaxMemoryMB:   512, // not used for external processing but needed for interface
	}
	baseProcessor := NewImageProcessor(baseOpts)

	return &VipsProcessor{
		ImageProcessor: baseProcessor,
		quality:        opts.Quality,
	}, nil
}

// CheckLibvipsAvailable checks if libvips is available via vips command
func CheckLibvipsAvailable() error {
	// try vips command first
	if _, err := exec.LookPath("vips"); err == nil {
		return nil
	}

	// try vipsthumbnail as alternative
	if _, err := exec.LookPath("vipsthumbnail"); err == nil {
		return nil
	}

	return fmt.Errorf("libvips not found: install with 'brew install vips' or 'apt-get install libvips-tools'")
}

// ResizeImageVips resizes an image using libvips
func (p *VipsProcessor) ResizeImageVips(inputPath, outputPath string, width, height int, quality int) error {
	return p.withConcurrencyControl(inputPath, func() error {
		// ensure output directory exists
		outputDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}

		// use vipsthumbnail for optimal performance
		// vipsthumbnail is specifically designed for generating thumbnails efficiently
		vipsPath, err := exec.LookPath("vipsthumbnail")
		if err != nil {
			// fallback to vips command
			vipsPath, err = exec.LookPath("vips")
			if err != nil {
				return fmt.Errorf("libvips commands not found")
			}
			return p.resizeWithVipsCommand(vipsPath, inputPath, outputPath, width, height, quality)
		}

		return p.resizeWithVipsThumbnail(vipsPath, inputPath, outputPath, width, height, quality)
	})
}

// resizeWithVipsThumbnail uses vipsthumbnail command (fastest option)
func (p *VipsProcessor) resizeWithVipsThumbnail(vipsPath, inputPath, outputPath string, width, height, quality int) error {
	// vipsthumbnail is optimized for thumbnail generation
	// format: vipsthumbnail input.jpg --size=800x600 --output=output.jpg[Q=85]

	// create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	// build size parameter - vips uses format like "800x600>"
	sizeParam := fmt.Sprintf("%dx%d>", width, height) // > means "only shrink, never enlarge"

	// build output with quality
	outputParam := outputPath
	if quality > 0 && (filepath.Ext(outputPath) == ".jpg" || filepath.Ext(outputPath) == ".jpeg") {
		outputParam = fmt.Sprintf("%s[Q=%d]", outputPath, quality)
	}

	cmd := exec.CommandContext(ctx, vipsPath, inputPath, "--size="+sizeParam, "--output="+outputParam)

	// capture stderr for better error messages
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("libvips processing timeout: %s", inputPath)
		}
		return fmt.Errorf("vipsthumbnail failed: %v, stderr: %s", err, stderr.String())
	}

	return nil
}

// resizeWithVipsCommand uses general vips command as fallback
func (p *VipsProcessor) resizeWithVipsCommand(vipsPath, inputPath, outputPath string, width, height, quality int) error {
	// format: vips resize input.jpg output.jpg 0.5 --kernel=lanczos3

	// we need to calculate scale factor, but first get image dimensions
	origWidth, origHeight, err := GetImageDimensionsFromFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to get image dimensions: %v", err)
	}

	// calculate scale to fit within width x height
	scaleX := float64(width) / float64(origWidth)
	scaleY := float64(height) / float64(origHeight)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// don't upscale
	if scale > 1.0 {
		scale = 1.0
	}

	// create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	// use vips resize command
	cmd := exec.CommandContext(ctx, vipsPath, "resize", inputPath, outputPath, fmt.Sprintf("%.6f", scale), "--kernel=lanczos3")

	// capture stderr for better error messages
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("libvips processing timeout: %s", inputPath)
		}
		return fmt.Errorf("vips resize failed: %v, stderr: %s", err, stderr.String())
	}

	return nil
}

// ResizeImageDirect implements ImageProcessorInterface - direct file-to-file processing
func (p *VipsProcessor) ResizeImageDirect(inputPath, outputPath string, width, height, quality int) error {
	return p.ResizeImageVips(inputPath, outputPath, width, height, quality)
}
