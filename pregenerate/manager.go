// Package pregenerate provides a unified queue-only architecture for page and image generation.
//
// ARCHITECTURE OVERVIEW:
// This package eliminates the complexity of dual on-demand/pregeneration systems by using
// a single generation path through configurable worker pools. All requests (HTTP, background)
// go through the same queuing system, providing:
//
// - Worker pools that scale with CPU cores for optimal concurrency
// - Large buffered priority channels to prevent blocking on urgent requests
// - Automatic failover to direct generation when queues are full
// - Request timeouts to protect user experience
// - Result channels for synchronous operations with deduplication
// - Memory management with periodic cleanup of tracking maps
//
// QUEUE FLOW:
// 1. HTTP requests use GeneratePageSync/GenerateImageSync with high priority
// 2. Background scanning uses QueuePage/QueueImage with normal priority
// 3. Workers pull from priority queues first, then background queues
// 4. Full queues trigger direct generation to never block users
// 5. Result channels eliminate duplicate work for concurrent requests
//
// This design provides the concurrency benefits of per-request goroutines while
// maintaining the architectural simplicity and deduplication of centralized generation.
package pregenerate

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/cooper/quiki/wiki"
)

// max returns the larger of two integers (for Go versions < 1.21)
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Manager handles intelligent, background pregeneration of pages and images
type Manager struct {
	wiki              *wiki.Wiki
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
	priorityCh        chan string // high priority pages - recently accessed
	backgroundCh      chan string // low priority - background pregeneration
	priorityImageCh   chan string // high priority images - recently accessed
	backgroundImageCh chan string // low priority - background image pregeneration
	stats             Stats
	mu                sync.RWMutex
	options           Options
	processingPages   map[string]bool // tracks pages currently being processed
	processingImages  map[string]bool // tracks images currently being processed
	completedPages    map[string]bool // tracks pages already pregenerated this session
	completedImages   map[string]bool // tracks images already pregenerated this session

	// promotion tracking - tracks items promoted from background to priority
	promotedPages  map[string]bool
	promotedImages map[string]bool

	// synchronous operation support
	pageResults  map[string]chan any // channels waiting for page results
	imageResults map[string]chan any // channels waiting for image results
}

// Options configures pregeneration behavior
type Options struct {
	RateLimit                time.Duration // time between background generations
	ProgressInterval         int           // show progress every N pages
	PriorityQueueSize        int           // size of priority queue
	BackgroundQueueSize      int           // size of background queue
	ImagePriorityQueueSize   int           // size of priority image queue
	ImageBackgroundQueueSize int           // size of background image queue
	PriorityWorkers          int           // number of priority workers
	BackgroundWorkers        int           // number of background workers
	ImagePriorityWorkers     int           // number of priority image workers
	ImageBackgroundWorkers   int           // number of background image workers
	RequestTimeout           time.Duration // max time to wait for sync page requests
	ImageRequestTimeout      time.Duration // max time to wait for sync image requests
	ForceGen                 bool          // whether to force regeneration bypassing cache
	LogVerbose               bool          // enable verbose logging
	EnableImages             bool          // whether to pregenerate common image sizes
	CleanupInterval          time.Duration // how often to clean up tracking maps (0 = disabled)
	MaxTrackingEntries       int           // max entries in tracking maps before cleanup
}

// DefaultOptions returns sensible defaults
func DefaultOptions() Options {
	numCPU := runtime.NumCPU()
	return Options{
		RateLimit:                10 * time.Millisecond, // ~100 pages per second
		ProgressInterval:         10,
		PriorityQueueSize:        500,  // large buffer for urgent requests
		BackgroundQueueSize:      2000, // substantial background capacity
		ImagePriorityQueueSize:   200,
		ImageBackgroundQueueSize: 500,
		PriorityWorkers:          max(2, numCPU/2),  // minimum 2, scale with CPU
		BackgroundWorkers:        max(1, numCPU/4),  // background uses fewer resources
		ImagePriorityWorkers:     max(2, numCPU/2),  // images need good parallelism
		ImageBackgroundWorkers:   max(1, numCPU/4),  // background image processing
		RequestTimeout:           30 * time.Second,  // reasonable timeout for users
		ImageRequestTimeout:      120 * time.Second, // longer timeout for large image processing
		EnableImages:             true,              // pregenerate common image sizes
		CleanupInterval:          30 * time.Minute,  // clean up tracking maps every 30 minutes
		MaxTrackingEntries:       10000,             // max 10k entries before forced cleanup
	}
}

// FastOptions returns options optimized for speed (essentially no rate limiting)
func FastOptions() Options {
	opts := DefaultOptions()
	opts.RateLimit = 1 * time.Millisecond // essentially unlimited (~1000 pages per second)
	opts.ProgressInterval = 100
	// use more aggressive worker scaling for speed
	numCPU := runtime.NumCPU()
	opts.PriorityWorkers = max(4, numCPU)      // more priority workers for speed
	opts.ImagePriorityWorkers = max(4, numCPU) // more image workers for speed
	opts.RequestTimeout = 10 * time.Second     // shorter timeout for fast mode
	return opts
}

// SlowOptions returns options that are gentle on system resources
func SlowOptions() Options {
	opts := DefaultOptions()
	opts.RateLimit = 500 * time.Millisecond // 2 pages per second
	opts.PriorityQueueSize = 50
	opts.BackgroundQueueSize = 200
	return opts
}

type Stats struct {
	TotalPages     int
	PregenedPages  int
	FailedPages    int
	LastPregenTime time.Time
	AverageGenTime time.Duration
}

// GetStats returns current pregeneration statistics
func (m *Manager) GetStats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

// New creates a new intelligent pregeneration manager with default options
func New(w *wiki.Wiki) *Manager {
	return NewWithOptions(w, DefaultOptions())
}

// NewWithOptions creates a new intelligent pregeneration manager with custom options
func NewWithOptions(w *wiki.Wiki, opts Options) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Manager{
		wiki:              w,
		ctx:               ctx,
		cancel:            cancel,
		priorityCh:        make(chan string, opts.PriorityQueueSize),
		backgroundCh:      make(chan string, opts.BackgroundQueueSize),
		priorityImageCh:   make(chan string, opts.ImagePriorityQueueSize),
		backgroundImageCh: make(chan string, opts.ImageBackgroundQueueSize),
		options:           opts,
		processingPages:   make(map[string]bool),
		processingImages:  make(map[string]bool),
		completedPages:    make(map[string]bool),
		completedImages:   make(map[string]bool),
		promotedPages:     make(map[string]bool),
		promotedImages:    make(map[string]bool),
		pageResults:       make(map[string]chan any),
		imageResults:      make(map[string]chan any),
	}

	// start worker pools
	// priority page workers
	m.wg.Add(opts.PriorityWorkers)
	for i := 0; i < opts.PriorityWorkers; i++ {
		go m.priorityWorker()
	}

	// background page workers
	m.wg.Add(opts.BackgroundWorkers)
	m.wiki.Log(fmt.Sprintf("pregenerate: starting %d background workers, queue size: %d, rate limit: %v",
		opts.BackgroundWorkers, opts.BackgroundQueueSize, opts.RateLimit))
	for i := 0; i < opts.BackgroundWorkers; i++ {
		go m.backgroundWorker()
	}

	if opts.EnableImages {
		// priority image workers
		m.wg.Add(opts.ImagePriorityWorkers)
		for i := 0; i < opts.ImagePriorityWorkers; i++ {
			go m.priorityImageWorker()
		}

		// background image workers
		m.wg.Add(opts.ImageBackgroundWorkers)
		for i := 0; i < opts.ImageBackgroundWorkers; i++ {
			go m.backgroundImageWorker()
		}
	}

	// start cleanup worker if enabled
	if opts.CleanupInterval > 0 {
		m.wg.Add(1)
		go m.cleanupWorker()
	}

	return m
}

// Stop gracefully shuts down the pregeneration manager
func (m *Manager) Stop() {
	m.cancel()
	m.wg.Wait()
	close(m.priorityCh)
	close(m.backgroundCh)
	if m.priorityImageCh != nil {
		close(m.priorityImageCh)
	}
	if m.backgroundImageCh != nil {
		close(m.backgroundImageCh)
	}
}

// GeneratePageSync generates a single page synchronously (blocks until complete)
// This is the new unified entry point for all page generation
func (m *Manager) GeneratePageSync(pageName string, highPriority bool) any {
	// first check if already cached and fresh
	page := m.wiki.FindPage(pageName)
	if page != nil && page.Exists() && !m.options.ForceGen {
		cacheModify := page.CacheModified()
		pageModified := page.Modified()
		if !pageModified.After(cacheModify) {
			// cache is fresh, use unified generation to return cached result
			return m.pregeneratePage(pageName, highPriority)
		}
	}

	// create result channel for this request
	resultCh := make(chan any, 1)

	m.mu.Lock()
	// check if already being processed with result channel
	if existingCh, exists := m.pageResults[pageName]; exists {
		m.mu.Unlock()
		// someone else is already generating this, wait for their result with timeout
		select {
		case result := <-existingCh:
			return result
		case <-time.After(m.options.RequestTimeout):
			// timeout - return error
			return wiki.DisplayError{Error: "Request timeout: page generation took too long"}
		}
	}

	// register our result channel
	m.pageResults[pageName] = resultCh
	m.mu.Unlock()

	// queue for generation
	if highPriority {
		select {
		case m.priorityCh <- pageName:
			// queued for priority
		default:
			// priority queue full, fall back to background but mark as promoted
			m.mu.Lock()
			m.promotedPages[pageName] = true
			m.mu.Unlock()
			select {
			case m.backgroundCh <- pageName:
				// queued for background
			default:
				// both queues full, generate directly
				m.mu.Lock()
				delete(m.pageResults, pageName)
				m.mu.Unlock()
				close(resultCh)
				return m.generatePageDirect(pageName)
			}
		}
	} else {
		select {
		case m.backgroundCh <- pageName:
			// queued for background
		default:
			// background queue full, generate directly
			m.mu.Lock()
			delete(m.pageResults, pageName)
			m.mu.Unlock()
			close(resultCh)
			return m.generatePageDirect(pageName)
		}
	}

	// wait for result with timeout
	select {
	case result := <-resultCh:
		return result
	case <-time.After(m.options.RequestTimeout):
		// timeout - clean up and return error
		m.mu.Lock()
		delete(m.pageResults, pageName)
		m.mu.Unlock()
		close(resultCh)
		return wiki.DisplayError{Error: "Request timeout: page generation took too long"}
	}
}

// GenerateImageSync generates image thumbnails synchronously (blocks until complete)
// This is the new unified entry point for all image generation
func (m *Manager) GenerateImageSync(imageName string, highPriority bool) any {
	if !m.options.EnableImages {
		// even with images disabled, use unified generation for consistency
		return m.pregenerateImage(imageName)
	}

	// create result channel for this request
	resultCh := make(chan any, 1)

	m.mu.Lock()
	// check if already being processed with result channel
	if existingCh, exists := m.imageResults[imageName]; exists {
		m.mu.Unlock()
		// someone else is already generating this, wait for their result with timeout
		select {
		case result := <-existingCh:
			return result
		case <-time.After(m.options.ImageRequestTimeout):
			// timeout - return error
			return wiki.DisplayError{Error: "Request timeout: image generation took too long"}
		}
	}

	// register our result channel
	m.imageResults[imageName] = resultCh
	m.mu.Unlock()

	// queue for generation
	if highPriority {
		select {
		case m.priorityImageCh <- imageName:
			// queued for priority
		default:
			// priority queue full, fall back to background but mark as promoted
			m.mu.Lock()
			m.promotedImages[imageName] = true
			m.mu.Unlock()
			select {
			case m.backgroundImageCh <- imageName:
				// queued for background
			default:
				// both queues full, generate directly
				m.mu.Lock()
				delete(m.imageResults, imageName)
				m.mu.Unlock()
				close(resultCh)
				return m.generateImageDirect(imageName)
			}
		}
	} else {
		select {
		case m.backgroundImageCh <- imageName:
			// queued for background
		default:
			// background queue full, generate directly
			m.mu.Lock()
			delete(m.imageResults, imageName)
			m.mu.Unlock()
			close(resultCh)
			return m.generateImageDirect(imageName)
		}
	}

	// wait for result with timeout
	select {
	case result := <-resultCh:
		return result
	case <-time.After(m.options.ImageRequestTimeout):
		// timeout - clean up and return error
		m.mu.Lock()
		delete(m.imageResults, imageName)
		m.mu.Unlock()
		close(resultCh)
		return wiki.DisplayError{Error: "Request timeout: image generation took too long"}
	}
}

// generatePageDirect generates a page directly without queuing (emergency fallback for full queues)
// This still goes through the same generation logic, just bypasses the worker queue
func (m *Manager) generatePageDirect(pageName string) any {
	return m.pregeneratePage(pageName, true) // use the same core logic
}

// generateImageDirect generates image thumbnails directly without queuing (emergency fallback for full queues)
// This still goes through the same generation logic, just bypasses the worker queue
func (m *Manager) generateImageDirect(imageName string) any {
	return m.pregenerateImage(imageName) // use the same core logic
}

// QueueAllContentAtStartup discovers and queues all pages and images for background pregeneration
func (m *Manager) QueueAllContentAtStartup() {
	// always log this important startup activity
	m.wiki.Log("pregenerate: discovering and queuing all content for background pregeneration")

	// queue all pages for background pregeneration
	go func() {
		allPages := m.wiki.AllPageFiles()
		m.wiki.Log(fmt.Sprintf("pregenerate: found %d total pages, starting to queue them", len(allPages)))
		if m.options.LogVerbose {
			m.wiki.Log(fmt.Sprintf("queuing %d pages for background pregeneration", len(allPages)))
		}
		if m.options.LogVerbose {
			m.wiki.Log(fmt.Sprintf("queuing %d pages for background pregeneration", len(allPages)))
		}

		queuedCount := 0
		skippedCount := 0
		for _, pageName := range allPages {
			// for background startup queueing, only check if currently processing
			// don't check completedPages since background processing should be repeatable
			m.mu.Lock()
			isProcessing := m.processingPages[pageName]
			isPromoted := m.promotedPages[pageName]
			if isProcessing || isPromoted {
				skippedCount++
				m.wiki.Log(fmt.Sprintf("pregenerate: skipping page %s - processing:%v promoted:%v",
					pageName, isProcessing, isPromoted))
				m.mu.Unlock()
				continue
			}
			m.mu.Unlock()

			select {
			case m.backgroundCh <- pageName:
				queuedCount++
				m.wiki.Log("pregenerate: queued page for background processing: " + pageName)
			case <-m.ctx.Done():
				m.wiki.Log("pregenerate: context canceled while queueing pages")
				return
			default:
				// background queue full, skip for now
				skippedCount++
				m.wiki.Log(fmt.Sprintf("pregenerate: background queue FULL, skipping: %s (queue size: %d)", pageName, cap(m.backgroundCh)))
			}
		}
		m.wiki.Log(fmt.Sprintf("pregenerate: finished queueing pages - queued: %d, skipped: %d", queuedCount, skippedCount))
	}()

	// queue all images for background pregeneration
	if m.options.EnableImages {
		go func() {
			allImages := m.wiki.AllImageFiles()
			if m.options.LogVerbose {
				m.wiki.Log(fmt.Sprintf("queuing %d images for background pregeneration", len(allImages)))
			}

			for _, imageName := range allImages {
				// check if already processed or queued
				m.mu.Lock()
				if m.processingImages[imageName] || m.completedImages[imageName] || m.promotedImages[imageName] {
					m.mu.Unlock()
					continue
				}
				m.mu.Unlock()

				select {
				case m.backgroundImageCh <- imageName:
					// queued
				case <-m.ctx.Done():
					return
				default:
					// background queue full, skip for now
					if m.options.LogVerbose {
						m.wiki.Log("background image queue full, skipping: " + imageName)
					}
				}
			}
		}()
	}
}

// PromotePageToPriority marks a page as promoted and adds it to priority queue
// Background workers will skip promoted items to avoid duplicate processing
func (m *Manager) PromotePageToPriority(pageName string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	// if already processing or completed, no need to promote
	if m.processingPages[pageName] || m.completedPages[pageName] {
		return false
	}

	// mark as promoted so background workers will skip it
	m.promotedPages[pageName] = true

	// try to put in priority queue (non-blocking)
	select {
	case m.priorityCh <- pageName:
		return true
	default:
		// priority queue full, unmark promotion since we couldn't queue it
		delete(m.promotedPages, pageName)
		return false
	}
}

// PromoteImageToPriority marks an image as promoted and adds it to priority queue
// Background workers will skip promoted items to avoid duplicate processing
func (m *Manager) PromoteImageToPriority(imageName string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	// if already processing or completed, no need to promote
	if m.processingImages[imageName] || m.completedImages[imageName] {
		return false
	}

	// mark as promoted so background workers will skip it
	m.promotedImages[imageName] = true

	// try to put in priority queue (non-blocking)
	select {
	case m.priorityImageCh <- imageName:
		return true
	default:
		// priority queue full, unmark promotion since we couldn't queue it
		delete(m.promotedImages, imageName)
		return false
	}
}

// RequestPregeneration requests pregeneration of a specific page (with high priority)
func (m *Manager) RequestPregeneration(pageName string) {
	select {
	case m.priorityCh <- pageName:
		// queued
	default:
		// priority queue is full, skip
		m.wiki.Log("priority queue full, skipping: " + pageName)
	}
}

// RequestPagePregeneration requests high-priority pregeneration of a specific page
// This is called when a user manually requests a page
func (m *Manager) RequestPagePregeneration(pageName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// skip if already processing, completed, or promoted
	if m.processingPages[pageName] || m.completedPages[pageName] || m.promotedPages[pageName] {
		return
	}

	select {
	case m.priorityCh <- pageName:
		// queued for high priority
		m.promotedPages[pageName] = true // mark as promoted to prevent background processing
	default:
		// priority queue is full, try background queue
		select {
		case m.backgroundCh <- pageName:
			// queued for background
		default:
			// both queues full, skip
			if m.options.LogVerbose {
				m.wiki.Log("all page queues full, skipping: " + pageName)
			}
		}
	}
}

// RequestImagePregeneration requests pregeneration of a specific image
// This is called when a user manually requests an image
func (m *Manager) RequestImagePregeneration(imageName string) {
	if !m.options.EnableImages {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// skip if already processing, completed, or promoted
	if m.processingImages[imageName] || m.completedImages[imageName] || m.promotedImages[imageName] {
		return
	}

	select {
	case m.priorityImageCh <- imageName:
		// queued for high priority
		m.promotedImages[imageName] = true // mark as promoted to prevent background processing
	default:
		// priority queue is full, try background queue
		select {
		case m.backgroundImageCh <- imageName:
			// queued for background
		default:
			// both queues full, skip
			if m.options.LogVerbose {
				m.wiki.Log("all image queues full, skipping: " + imageName)
			}
		}
	}
}

// StartBackground begins low-priority background pregeneration of all pages and images
func (m *Manager) StartBackground() *Manager {
	// queue all content for background pregeneration at startup
	m.QueueAllContentAtStartup()
	return m
}

// StartWorkers starts the worker goroutines without queuing content at startup
func (m *Manager) StartWorkers() *Manager {
	return m
}

// pregenerateAllImages handles both synchronous and asynchronous image pregeneration
func (m *Manager) pregenerateAllImages(synchronous bool) {
	if !m.options.EnableImages {
		return
	}

	imageFiles := m.wiki.AllImageFiles()
	if len(imageFiles) == 0 {
		return
	}

	if synchronous {
		m.wiki.Log(fmt.Sprintf("synchronously pregenerating %d images...", len(imageFiles)))

		for i, imageName := range imageFiles {
			if m.options.ProgressInterval > 0 && i%m.options.ProgressInterval == 0 && i > 0 {
				m.wiki.Log(fmt.Sprintf("pregenerated %d/%d images", i, len(imageFiles)))
			}

			m.pregenerateImage(imageName)

			// apply rate limiting if configured
			if m.options.RateLimit > 0 && i < len(imageFiles)-1 {
				time.Sleep(m.options.RateLimit * 2) // slower rate for images
			}
		}

		m.wiki.Log(fmt.Sprintf("synchronous image pregeneration complete: %d images processed", len(imageFiles)))

	} else {
		if m.backgroundImageCh == nil {
			return
		}

		m.wiki.Log(fmt.Sprintf("queuing %d images for background pregeneration", len(imageFiles)))

		go func() {
			for _, imageName := range imageFiles {
				select {
				case m.backgroundImageCh <- imageName:
					// queued
				case <-m.ctx.Done():
					// shutting down
					return
				}
			}
		}()
	}
} // PregenerateSync synchronously pregenerates all pages and images
func (m *Manager) PregenerateSync() Stats {
	stats := m.pregenerateAllPages(true)
	m.pregenerateAllImages(true)
	return stats
}

func (m *Manager) pregenerateAllPages(synchronous bool) Stats {
	// Use wiki-level locking for cross-process coordination
	err := m.wiki.WithWikiLock(func() error {
		allPages := m.wiki.AllPageFiles()
		m.mu.Lock()
		m.stats.TotalPages = len(allPages)
		m.mu.Unlock()

		if synchronous {
			m.wiki.Log(fmt.Sprintf("synchronously pregenerating %d pages...", len(allPages)))

			// for sync pregen, process pages directly
			for i, pageName := range allPages {
				if m.options.ProgressInterval > 0 && i%m.options.ProgressInterval == 0 && i > 0 {
					m.wiki.Log(fmt.Sprintf("pregenerated %d/%d pages", i, len(allPages)))
				}

				m.pregeneratePage(pageName, true)

				// apply rate limiting if configured
				if m.options.RateLimit > 0 && i < len(allPages)-1 {
					time.Sleep(m.options.RateLimit)
				}
			}

			m.wiki.Log(fmt.Sprintf("synchronous pregeneration complete: %d pages processed, %d generated, %d failed",
				m.stats.TotalPages, m.stats.PregenedPages, m.stats.FailedPages))

		} else {
			m.wiki.Log(fmt.Sprintf("starting background pregeneration of %d pages", len(allPages)))

			go func() {
				for _, pageName := range allPages {
					select {
					case m.backgroundCh <- pageName:
						// queued
					case <-m.ctx.Done():
						// shutting down
						return
					}
				}
			}()
		}
		return nil
	})

	if err != nil {
		m.wiki.Log("skipping pregenerate: " + err.Error())
	}

	// return current stats
	m.mu.RLock()
	currentStats := m.stats
	m.mu.RUnlock()
	return currentStats
}

// priorityWorker handles high-priority page pregeneration
func (m *Manager) priorityWorker() {
	defer m.wg.Done()

	for {
		select {
		case pageName := <-m.priorityCh:
			m.mu.Lock()
			if m.processingPages[pageName] || m.completedPages[pageName] {
				// clean up promotion tracking if item was already handled
				delete(m.promotedPages, pageName)
				resultCh, hasWaiter := m.pageResults[pageName]
				delete(m.pageResults, pageName)
				m.mu.Unlock()

				// notify waiter with cached result if available
				if hasWaiter {
					go func() {
						result := m.wiki.DisplayPageDraft(pageName, true)
						select {
						case resultCh <- result:
						default:
						}
						close(resultCh)
					}()
				}
				continue // skip if already handled
			}
			m.processingPages[pageName] = true
			resultCh, hasWaiter := m.pageResults[pageName]
			delete(m.pageResults, pageName)
			m.mu.Unlock()

			// generate the page
			result := m.pregeneratePage(pageName, true)

			// notify any waiters
			if hasWaiter {
				select {
				case resultCh <- result:
				default:
				}
				close(resultCh)
			}

			m.mu.Lock()
			delete(m.processingPages, pageName)
			delete(m.promotedPages, pageName) // clean up promotion tracking
			m.completedPages[pageName] = true
			m.mu.Unlock()
		case <-m.ctx.Done():
			return
		}
	}
}

// backgroundWorker handles low-priority background pregeneration
func (m *Manager) backgroundWorker() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.options.RateLimit)
	defer ticker.Stop()

	for {
		select {
		case pageName := <-m.backgroundCh:
			m.wiki.Log("pregenerate: background worker received page: " + pageName)
			// check if already being processed, completed, or promoted
			m.mu.Lock()
			if m.processingPages[pageName] || m.completedPages[pageName] || m.promotedPages[pageName] {
				m.wiki.Log("pregenerate: background worker skipping page (already processed/processing): " + pageName)
				m.mu.Unlock()
				<-ticker.C // rate limit even for skipped items to prevent queue flooding
				continue
			}
			m.processingPages[pageName] = true
			m.mu.Unlock()

			m.wiki.Log("pregenerate: background worker generating page: " + pageName)
			result := m.pregeneratePage(pageName, false)

			// notify any waiting result channels
			m.mu.Lock()
			if resultCh, exists := m.pageResults[pageName]; exists {
				select {
				case resultCh <- result:
				default:
				}
				delete(m.pageResults, pageName)
			}

			// mark as completed
			delete(m.processingPages, pageName)
			m.completedPages[pageName] = true
			m.mu.Unlock()

			<-ticker.C // rate limit
		case <-m.ctx.Done():
			return
		}
	}
}

// pregeneratePage generates a single page and updates statistics
func (m *Manager) pregeneratePage(pageName string, isHighPriority bool) any {
	start := time.Now()
	m.wiki.Log(fmt.Sprintf("pregenerate: starting generation for page: %s (priority: %v)", pageName, isHighPriority))

	// check if page exists
	page := m.wiki.FindPage(pageName)
	if !page.Exists() {
		m.wiki.Log("pregenerate: page does not exist: " + pageName)
		return wiki.DisplayError{Error: "Page not found"}
	}

	// check if page is already cached and fresh - avoid unnecessary work
	if !m.options.ForceGen {
		cacheModify := page.CacheModified()
		pageModified := page.Modified()
		if !pageModified.After(cacheModify) {
			m.wiki.Log(fmt.Sprintf("pregenerate: page %s already cached and fresh", pageName))
			if m.options.LogVerbose {
				m.wiki.Log(fmt.Sprintf("page %s already cached and fresh", pageName))
			}
			// return the cached result
			return m.wiki.DisplayPageDraft(pageName, true)
		}
	}

	m.wiki.Log("pregenerate: calling DisplayPageDraft for: " + pageName)
	// temporarily modify ForceGen in a thread-safe way
	var result any
	originalForceGen := m.wiki.Opt.Page.ForceGen
	if m.options.ForceGen {
		// use atomic operations or separate lock for options modification
		func() {
			// create a brief critical section for ForceGen modification
			m.wiki.Opt.Page.ForceGen = true
			result = m.wiki.DisplayPageDraft(pageName, true)
			m.wiki.Opt.Page.ForceGen = originalForceGen
		}()
	} else {
		result = m.wiki.DisplayPageDraft(pageName, true)
	}

	m.wiki.Log("pregenerate: DisplayPageDraft completed for: " + pageName)

	// update stats
	m.mu.Lock()
	duration := time.Since(start)
	m.stats.LastPregenTime = time.Now()

	if dp, ok := result.(wiki.DisplayPage); ok && !dp.FromCache {
		m.stats.PregenedPages++
		if m.stats.AverageGenTime == 0 {
			m.stats.AverageGenTime = duration
		} else {
			m.stats.AverageGenTime = (m.stats.AverageGenTime + duration) / 2
		}
		if m.options.LogVerbose {
			m.wiki.Log(fmt.Sprintf("pregenerated %s (%.2fms, priority=%v)",
				pageName, float64(duration.Nanoseconds())/1000000, isHighPriority))
		}
	} else if _, isError := result.(wiki.DisplayError); isError {
		m.stats.FailedPages++
		m.wiki.Log(fmt.Sprintf("failed to pregenerate %s: %v", pageName, result))
	}
	m.mu.Unlock()

	return result
}

// priorityImageWorker handles high-priority image pregeneration
func (m *Manager) priorityImageWorker() {
	defer m.wg.Done()

	for {
		select {
		case imageName := <-m.priorityImageCh:
			// check if already being processed
			m.mu.Lock()
			if m.processingImages[imageName] || m.completedImages[imageName] {
				// clean up promotion tracking if item was already handled
				delete(m.promotedImages, imageName)
				m.mu.Unlock()
				continue
			}
			m.processingImages[imageName] = true
			m.mu.Unlock()

			result := m.pregenerateImage(imageName)

			// notify any waiting result channels
			m.mu.Lock()
			if resultCh, exists := m.imageResults[imageName]; exists {
				select {
				case resultCh <- result:
				default:
				}
				delete(m.imageResults, imageName)
			}

			// mark as completed
			delete(m.processingImages, imageName)
			delete(m.promotedImages, imageName) // clean up promotion tracking
			m.completedImages[imageName] = true
			m.mu.Unlock()

		case <-m.ctx.Done():
			return
		}
	}
}

func (m *Manager) backgroundImageWorker() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.options.RateLimit * 2) // slower rate for images
	defer ticker.Stop()

	for {
		select {
		case imageName := <-m.backgroundImageCh:
			m.wiki.Log("pregenerate: background image worker received image: " + imageName)
			// check if already being processed, completed, or promoted
			m.mu.Lock()
			if m.processingImages[imageName] || m.completedImages[imageName] || m.promotedImages[imageName] {
				m.wiki.Log("pregenerate: background image worker skipping image (already processed/processing): " + imageName)
				m.mu.Unlock()
				<-ticker.C // rate limit even for skipped items to prevent queue flooding
				continue
			}
			m.processingImages[imageName] = true
			m.mu.Unlock()

			m.wiki.Log("pregenerate: background image worker generating image: " + imageName)
			result := m.pregenerateImage(imageName)

			// notify any waiting result channels
			m.mu.Lock()
			if resultCh, exists := m.imageResults[imageName]; exists {
				select {
				case resultCh <- result:
				default:
				}
				delete(m.imageResults, imageName)
			}

			// mark as completed
			delete(m.processingImages, imageName)
			m.completedImages[imageName] = true
			m.mu.Unlock()

			<-ticker.C // rate limit
		case <-m.ctx.Done():
			return
		}
	}
}

// cleanupWorker periodically cleans up tracking maps to prevent memory leaks
func (m *Manager) cleanupWorker() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.options.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanupTrackingMaps()
		case <-m.ctx.Done():
			return
		}
	}
}

// cleanupTrackingMaps removes old entries from tracking maps
func (m *Manager) cleanupTrackingMaps() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// if tracking maps are getting too large, clear them entirely
	// this is safe because they're just optimization to avoid duplicate work
	if len(m.completedPages) > m.options.MaxTrackingEntries {
		if m.options.LogVerbose {
			m.wiki.Log(fmt.Sprintf("clearing page tracking maps (%d entries)", len(m.completedPages)))
		}
		m.completedPages = make(map[string]bool)
		m.promotedPages = make(map[string]bool)
	}

	if len(m.completedImages) > m.options.MaxTrackingEntries {
		if m.options.LogVerbose {
			m.wiki.Log(fmt.Sprintf("clearing image tracking maps (%d entries)", len(m.completedImages)))
		}
		m.completedImages = make(map[string]bool)
		m.promotedImages = make(map[string]bool)
	}
}

// pregenerateImage generates an image in all sizes that are actually used in the wiki
func (m *Manager) pregenerateImage(imageName string) any {
	if m.options.LogVerbose {
		m.wiki.Log(fmt.Sprintf("pregenerating image: %s", imageName))
	}
	m.wiki.Log(fmt.Sprintf("DEBUG: pregenerateImage called with imageName='%s'", imageName))

	// use image-specific locking to coordinate with on-demand generation
	imageLock := m.wiki.GetImageLock(imageName)
	imageLock.Lock()
	defer imageLock.Unlock()

	// Get the image category that tracks all references to this image
	imageCat := m.wiki.GetSpecialCategory(imageName, wiki.CategoryTypeImage)

	// No category means no references exist, so nothing to pregenerate
	if imageCat == nil || !imageCat.Exists() {
		if m.options.LogVerbose {
			m.wiki.Log(fmt.Sprintf("no references found for image: %s", imageName))
		}
		m.wiki.Log(fmt.Sprintf("DEBUG: category lookup failed for %s - imageCat=%v exists=%v", imageName, imageCat != nil, imageCat != nil && imageCat.Exists()))
		return nil // no error, just nothing to generate
	}

	// Collect all unique dimensions that are actually used
	usedSizes := make(map[[2]int]bool)

	// Look through all pages that reference this image
	for _, pageEntry := range imageCat.Pages {
		// pageEntry.Dimensions contains the dimensions as [][]int
		for _, dimensionPair := range pageEntry.Dimensions {
			if len(dimensionPair) >= 2 {
				width, height := dimensionPair[0], dimensionPair[1]
				usedSizes[[2]int{width, height}] = true
			}
		}
	}

	// always include configurable thumbnail sizes from PregenThumbs setting
	// this ensures adminifier and other interfaces load quickly without generating on-demand
	if imageCat.ImageInfo != nil {
		origWidth := imageCat.ImageInfo.Width
		origHeight := imageCat.ImageInfo.Height

		// parse thumbnail sizes from wiki config
		thumbnailSizes := m.wiki.ParseThumbnailSizes(m.wiki.Opt.Image.PregenThumbs, origWidth, origHeight)
		for _, size := range thumbnailSizes {
			usedSizes[[2]int{size[0], size[1]}] = true
		}
	}

	if m.options.LogVerbose {
		m.wiki.Log(fmt.Sprintf("found %d unique sizes for image: %s", len(usedSizes), imageName))
	}

	// Generate images for each actually-used size
	for size := range usedSizes {
		img := wiki.SizedImageFromName(imageName)
		img.Width = size[0]
		img.Height = size[1]

		// generate the image
		result := m.wiki.DisplaySizedImageGenerate(img, true)

		// check if generation was successful
		if _, isError := result.(wiki.DisplayError); isError && m.options.LogVerbose {
			m.wiki.Log(fmt.Sprintf("failed to pregenerate %s at %dx%d: %v",
				imageName, size[0], size[1], result))
		}
	}

	img := wiki.SizedImageFromName(imageName)
	return m.wiki.DisplaySizedImageGenerate(img, true)
}
