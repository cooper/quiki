package pregenerate

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cooper/quiki/wiki"
)

// Manager handles intelligent, background pregeneration of pages
type Manager struct {
	wiki         *wiki.Wiki
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	priorityCh   chan string // high priority pages - recently accessed
	backgroundCh chan string // low priority - background pregeneration
	stats        Stats
	mu           sync.RWMutex
	options      Options
}

// Options configures pregeneration behavior
type Options struct {
	RateLimit           time.Duration // time between background generations
	ProgressInterval    int           // show progress every N pages
	PriorityQueueSize   int           // size of priority queue
	BackgroundQueueSize int           // size of background queue
	ForceGen            bool          // whether to force regeneration bypassing cache
	LogVerbose          bool          // enable verbose logging
}

// DefaultOptions returns sensible defaults
func DefaultOptions() Options {
	return Options{
		RateLimit:           10 * time.Millisecond, // ~100 pages per second
		ProgressInterval:    10,
		PriorityQueueSize:   50,
		BackgroundQueueSize: 200,
	}
}

// FastOptions returns options optimized for speed (essentially no rate limiting)
func FastOptions() Options {
	opts := DefaultOptions()
	opts.RateLimit = 1 * time.Millisecond // essentially unlimited (~1000 pages per second)
	opts.ProgressInterval = 100
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
		wiki:         w,
		ctx:          ctx,
		cancel:       cancel,
		priorityCh:   make(chan string, opts.PriorityQueueSize),
		backgroundCh: make(chan string, opts.BackgroundQueueSize),
		options:      opts,
	}

	// start workers
	m.wg.Add(2)
	go m.priorityWorker()
	go m.backgroundWorker()

	return m
}

// Stop gracefully shuts down the pregeneration manager
func (m *Manager) Stop() {
	m.cancel()
	m.wg.Wait()
	close(m.priorityCh)
	close(m.backgroundCh)
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

// StartBackground begins low-priority background pregeneration of all pages
func (m *Manager) StartBackground() {
	m.pregenerateAll(false)
}

// PregenerateSync synchronously pregenerates all pages
func (m *Manager) PregenerateSync() Stats {
	return m.pregenerateAll(true)
}

func (m *Manager) pregenerateAll(synchronous bool) Stats {
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
		m.wiki.Log("skipping pregeneration: " + err.Error())
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
			m.pregeneratePage(pageName, true)
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
			m.pregeneratePage(pageName, false)
			<-ticker.C // rate limit
		case <-m.ctx.Done():
			return
		}
	}
}

// pregeneratePage generates a single page and updates statistics
func (m *Manager) pregeneratePage(pageName string, isHighPriority bool) {
	start := time.Now()

	// check if page exists
	page := m.wiki.FindPage(pageName)
	if !page.Exists() {
		return
	}

	originalForceGen := m.wiki.Opt.Page.ForceGen
	if m.options.ForceGen {
		m.wiki.Opt.Page.ForceGen = true
	}

	result := m.wiki.DisplayPageDraft(pageName, true)

	m.wiki.Opt.Page.ForceGen = originalForceGen

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
}
