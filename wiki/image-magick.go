package wiki

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ImageMagickProcessor handles image processing using imagemagick convert command
type ImageMagickProcessor struct {
	*ImageProcessor        // embed for common functionality
	convertPath     string // path to imagemagick convert binary
}

// ImageMagickOptions configures the imagemagick processor
type ImageMagickOptions struct {
	MaxConcurrent int           // max concurrent image operations
	Timeout       time.Duration // max time per image operation
	ConvertPath   string        // path to convert binary (empty = auto-detect)
	Quality       int           // JPEG quality (1-100, 0 = default)
	MaxPixels     int64         // max pixels to prevent zip bombs
}

// DefaultImageMagickOptions returns sensible defaults
func DefaultImageMagickOptions() ImageMagickOptions {
	return ImageMagickOptions{
		MaxConcurrent: 4,                // can handle more with external process
		Timeout:       60 * time.Second, // longer timeout for complex operations
		Quality:       85,               // good quality/size balance
		MaxPixels:     100_000_000,      // 100MP limit (e.g., 10000x10000)
	}
}

// NewImageMagickProcessor creates a new imagemagick-based processor
func NewImageMagickProcessor(opts ImageMagickOptions) (*ImageMagickProcessor, error) {
	// find convert binary
	convertPath := opts.ConvertPath
	if convertPath == "" {
		var err error
		convertPath, err = exec.LookPath("convert")
		if err != nil {
			// try magick command (imagemagick 7+)
			convertPath, err = exec.LookPath("magick")
			if err != nil {
				return nil, fmt.Errorf("imagemagick not found: install with 'brew install imagemagick' or 'apt-get install imagemagick'")
			}
		}
	}

	// create base processor with common functionality
	baseOpts := ImageProcessorOptions{
		MaxConcurrent: opts.MaxConcurrent,
		Timeout:       opts.Timeout,
		MaxMemoryMB:   512, // not used for external processing but needed for interface
	}
	baseProcessor := NewImageProcessor(baseOpts)

	return &ImageMagickProcessor{
		ImageProcessor: baseProcessor,
		convertPath:    convertPath,
	}, nil
}

// ResizeImage resizes an image using ImageMagick convert
func (p *ImageMagickProcessor) ResizeImage(inputPath, outputPath string, width, height int, quality int) error {
	return p.withConcurrencyControl(inputPath, func() error {
		// ensure output directory exists
		outputDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}

		// build ImageMagick command
		args := []string{}

		// check if using ImageMagick 7+ (magick command)
		if strings.Contains(p.convertPath, "magick") {
			args = append(args, "convert") // subcommand for ImageMagick 7+
		}

		args = append(args,
			inputPath,
			"-resize", fmt.Sprintf("%dx%d>", width, height), // > means "only shrink, never enlarge"
			"-strip",              // remove metadata for smaller files
			"-interlace", "Plane", // progressive JPEG
		)

		// set quality if specified
		if quality > 0 {
			args = append(args, "-quality", strconv.Itoa(quality))
		}

		args = append(args, outputPath)

		// create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
		defer cancel()

		// run ImageMagick convert
		cmd := exec.CommandContext(ctx, p.convertPath, args...)

		// capture stderr for better error messages
		var stderr strings.Builder
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("image processing timeout: %s", inputPath)
			}
			return fmt.Errorf("imagemagick failed: %v, stderr: %s", err, stderr.String())
		}

		return nil
	})
}

// GetImageDimensions gets image dimensions using ImageMagick identify
func (p *ImageMagickProcessor) GetImageDimensions(path string) (width, height int, err error) {
	// build identify command
	args := []string{}

	if strings.Contains(p.convertPath, "magick") {
		args = append(args, "identify") // subcommand for ImageMagick 7+
	} else {
		// for ImageMagick 6, use separate identify command
		identifyPath, err := exec.LookPath("identify")
		if err != nil {
			return 0, 0, fmt.Errorf("imagemagick identify not found")
		}
		args = []string{path, "-format", "%wx%h"}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, identifyPath, args...)
		output, err := cmd.Output()
		if err != nil {
			return 0, 0, fmt.Errorf("failed to get image dimensions: %v", err)
		}

		parts := strings.Split(strings.TrimSpace(string(output)), "x")
		if len(parts) != 2 {
			return 0, 0, fmt.Errorf("unexpected identify output: %s", output)
		}

		width, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid width: %s", parts[0])
		}

		height, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid height: %s", parts[1])
		}

		return width, height, nil
	}

	// ImageMagick 7+ path
	args = append(args, "identify", path, "-format", "%wx%h")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, p.convertPath, args...)
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get image dimensions: %v", err)
	}

	parts := strings.Split(strings.TrimSpace(string(output)), "x")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected identify output: %s", output)
	}

	width, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid width: %s", parts[0])
	}

	height, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid height: %s", parts[1])
	}

	return width, height, nil
}

// CheckImageMagickAvailable checks if imagemagick is available
func CheckImageMagickAvailable() error {
	// try convert first
	if _, err := exec.LookPath("convert"); err == nil {
		return nil
	}

	// try magick command (ImageMagick 7+)
	if _, err := exec.LookPath("magick"); err == nil {
		return nil
	}

	return fmt.Errorf("imagemagick not found: install with 'brew install imagemagick' or 'apt-get install imagemagick'")
}

// ResizeImageDirect implements ImageProcessorInterface - direct file-to-file processing
func (p *ImageMagickProcessor) ResizeImageDirect(inputPath, outputPath string, width, height, quality int) error {
	return p.ResizeImage(inputPath, outputPath, width, height, quality)
}
