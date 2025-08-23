package wiki

import (
	"runtime"
	"sync"
	"time"
)

// MemoryMonitor tracks system memory and adjusts concurrency accordingly
// this is application-wide protection against memory exhaustion from image processing
type MemoryMonitor struct {
	mu                sync.RWMutex
	lastCheck         time.Time
	availableMemoryMB int64
	maxConcurrency    int
	currentActive     int
	activeMu          sync.Mutex
}

// global memory monitor instance
var globalMemoryMonitor *MemoryMonitor
var memoryMonitorOnce sync.Once

// GetMemoryMonitor returns the global memory monitor instance
func GetMemoryMonitor() *MemoryMonitor {
	memoryMonitorOnce.Do(func() {
		// default to reasonable concurrency based on CPU cores
		maxConcurrency := max(4, runtime.NumCPU())
		globalMemoryMonitor = NewMemoryMonitor(maxConcurrency)
	})
	return globalMemoryMonitor
}

// NewMemoryMonitor creates a memory monitor with intelligent concurrency control
func NewMemoryMonitor(maxConcurrency int) *MemoryMonitor {
	m := &MemoryMonitor{
		maxConcurrency: maxConcurrency,
	}
	m.updateMemoryStats()
	return m
}

// updateMemoryStats gets current system memory stats and calculates safe concurrency
func (m *MemoryMonitor) updateMemoryStats() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// get total system memory (approximate based on heap limit)
	// this is conservative - we assume we can use up to 70% of available memory
	totalMemoryMB := int64(memStats.Sys) / (1024 * 1024)
	usedMemoryMB := int64(memStats.Alloc) / (1024 * 1024)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.availableMemoryMB = totalMemoryMB - usedMemoryMB
	m.lastCheck = time.Now()
}

// shouldAllowNewWorker checks if we can safely start another concurrent operation
func (m *MemoryMonitor) shouldAllowNewWorker() bool {
	m.activeMu.Lock()
	defer m.activeMu.Unlock()

	// always allow at least one worker
	if m.currentActive == 0 {
		return true
	}

	// update memory stats if stale (every 3 seconds for more responsive adjustment)
	if time.Since(m.lastCheck) > 3*time.Second {
		m.updateMemoryStats()
	}

	m.mu.RLock()
	availableMB := m.availableMemoryMB
	m.mu.RUnlock()

	// very conservative estimates for image processing memory usage per operation
	// different processors can use 200-800MB per large image during processing
	// we use worst-case scenario to prevent crashes
	estimatedMemoryPerWorkerMB := int64(400) // conservative estimate for safety

	// calculate safe concurrency based on available memory
	safeConcurrency := max(1, int(availableMB/estimatedMemoryPerWorkerMB))

	// never exceed configured maximum
	if safeConcurrency > m.maxConcurrency {
		safeConcurrency = m.maxConcurrency
	}

	// if we're low on memory, be very conservative
	if availableMB < 1000 { // less than 1GB available
		safeConcurrency = 1 // only one at a time
	} else if availableMB < 2000 { // less than 2GB available
		safeConcurrency = max(1, safeConcurrency/2)
	}

	return m.currentActive < safeConcurrency
}

// acquireWorker tries to acquire a worker slot, returns false if memory is too low
func (m *MemoryMonitor) acquireWorker() bool {
	if !m.shouldAllowNewWorker() {
		// if memory is low, try forcing garbage collection and check again
		m.mu.RLock()
		availableMB := m.availableMemoryMB
		m.mu.RUnlock()

		if availableMB < 1000 { // less than 1GB available
			runtime.GC()                      // force garbage collection to free up memory
			time.Sleep(50 * time.Millisecond) // give GC time to work
			m.updateMemoryStats()             // refresh stats after GC

			// try again after GC
			if !m.shouldAllowNewWorker() {
				return false
			}
		} else {
			return false
		}
	}

	m.activeMu.Lock()
	m.currentActive++
	m.activeMu.Unlock()
	return true
}

// releaseWorker releases a worker slot
func (m *MemoryMonitor) releaseWorker() {
	m.activeMu.Lock()
	if m.currentActive > 0 {
		m.currentActive--
	}
	m.activeMu.Unlock()
}

// GetStats returns current memory statistics for monitoring
func (m *MemoryMonitor) GetStats() (availableMB int64, activeWorkers int, maxWorkers int) {
	m.mu.RLock()
	availableMB = m.availableMemoryMB
	m.mu.RUnlock()

	m.activeMu.Lock()
	activeWorkers = m.currentActive
	m.activeMu.Unlock()

	maxWorkers = m.maxConcurrency
	return
}
