package wiki

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"
)

// writeTestPNG writes a solid-color PNG of the given size and returns its path
func writeTestPNG(t *testing.T, dir string, width, height int) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for i := range img.Pix {
		img.Pix[i] = 0x80
	}
	img.Set(0, 0, color.White)
	path := filepath.Join(dir, strconv.Itoa(width)+"x"+strconv.Itoa(height)+".png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	return path
}

// a burst of concurrent requests should all succeed by queuing on the
// concurrency limit rather than failing with "too many concurrent operations"
func TestImageProcessorConcurrentBurst(t *testing.T) {
	dir := t.TempDir()
	input := writeTestPNG(t, dir, 400, 300)
	p := NewImageProcessor(ImageProcessorOptions{
		MaxConcurrent: 2,
		MaxMemoryMB:   64,
		Timeout:       30 * time.Second,
	})

	const n = 20
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			out := filepath.Join(dir, "out-"+strconv.Itoa(i)+".png")
			errs[i] = p.ResizeImageDirect(input, out, 40, 30, 85)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("resize %d: %v", i, err)
			continue
		}
		out := filepath.Join(dir, "out-"+strconv.Itoa(i)+".png")
		w, h, err := GetImageDimensionsFromFile(out)
		if err != nil {
			t.Errorf("resize %d: %v", i, err)
		} else if w != 40 || h != 30 {
			t.Errorf("resize %d: got %dx%d, want 40x30", i, w, h)
		}
	}

	// all slots must have been released
	if len(p.semaphore) != 0 {
		t.Errorf("%d slots still held after all work finished", len(p.semaphore))
	}
}

// images over the memory limit are rejected from the header alone
func TestImageProcessorMemoryLimit(t *testing.T) {
	dir := t.TempDir()
	input := writeTestPNG(t, dir, 1000, 1000) // ~3.8MB decoded
	p := NewImageProcessor(ImageProcessorOptions{
		MaxConcurrent: 1,
		MaxMemoryMB:   1,
		Timeout:       time.Second,
	})
	err := p.ResizeImageDirect(input, filepath.Join(dir, "out.png"), 10, 10, 85)
	if err == nil {
		t.Fatal("expected an error for an image over the memory limit")
	}
	if len(p.semaphore) != 0 {
		t.Error("a slot was held for a rejected image")
	}
}

// when every slot is busy, a request waits for the timeout and then gives up
// with a clear error instead of failing instantly
func TestImageProcessorWaitsForSlot(t *testing.T) {
	dir := t.TempDir()
	input := writeTestPNG(t, dir, 20, 20)
	p := NewImageProcessor(ImageProcessorOptions{
		MaxConcurrent: 1,
		MaxMemoryMB:   64,
		Timeout:       100 * time.Millisecond,
	})

	// occupy the only slot
	p.semaphore <- struct{}{}

	start := time.Now()
	err := p.ResizeImageDirect(input, filepath.Join(dir, "out.png"), 10, 10, 85)
	if err == nil {
		t.Fatal("expected a timeout error while the slot is held")
	}
	if elapsed := time.Since(start); elapsed < 100*time.Millisecond {
		t.Errorf("gave up after %v; should have waited for the %v timeout", elapsed, p.timeout)
	}

	// once the slot frees up, the same request succeeds
	<-p.semaphore
	if err := p.ResizeImageDirect(input, filepath.Join(dir, "out.png"), 10, 10, 85); err != nil {
		t.Fatalf("resize after slot freed: %v", err)
	}
}

// the processor must be shared across calls for the same wiki,
// otherwise its concurrency limit means nothing
func TestGetImageProcessorForWikiIsShared(t *testing.T) {
	w := &Wiki{}
	if GetImageProcessorForWiki(w) != GetImageProcessorForWiki(w) {
		t.Fatal("GetImageProcessorForWiki returned a different processor on each call")
	}
}
