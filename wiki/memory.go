package wiki

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// MemoryMonitor provides application-wide protection against memory exhaustion
type MemoryMonitor struct {
	mu                sync.RWMutex
	lastCheck         time.Time
	availableMemoryMB int64
	maxConcurrency    int
	currentActive     int
	activeMu          sync.Mutex
}

var globalMemoryMonitor *MemoryMonitor
var memoryMonitorOnce sync.Once

func GetMemoryMonitor() *MemoryMonitor {
	memoryMonitorOnce.Do(func() {
		maxConcurrency := max(4, runtime.NumCPU()*2)
		globalMemoryMonitor = NewMemoryMonitor(maxConcurrency)
	})
	return globalMemoryMonitor
}

func NewMemoryMonitor(maxConcurrency int) *MemoryMonitor {
	m := &MemoryMonitor{
		maxConcurrency: maxConcurrency,
	}
	m.updateMemoryStats()
	return m
}

func (m *MemoryMonitor) updateMemoryStats() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	totalMemoryMB := int64(memStats.Sys) / (1024 * 1024)
	usedMemoryMB := int64(memStats.Alloc) / (1024 * 1024)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.availableMemoryMB = totalMemoryMB - usedMemoryMB
	m.lastCheck = time.Now()
}

func (m *MemoryMonitor) shouldAllowNewWorker() bool {
	m.activeMu.Lock()
	defer m.activeMu.Unlock()

	if m.currentActive == 0 {
		return true
	}

	if time.Since(m.lastCheck) > 3*time.Second {
		m.updateMemoryStats()
	}

	m.mu.RLock()
	availableMB := m.availableMemoryMB
	m.mu.RUnlock()

	// external processors (libvips/imagemagick) use minimal process memory
	estimatedMemoryPerWorkerMB := int64(25)

	safeConcurrency := max(1, int(availableMB/estimatedMemoryPerWorkerMB))

	if safeConcurrency > m.maxConcurrency {
		safeConcurrency = m.maxConcurrency
	}

	originalSafe := safeConcurrency
	if availableMB < 100 {
		safeConcurrency = 1
	} else if availableMB < 200 {
		safeConcurrency = max(2, safeConcurrency/2)
	}

	if originalSafe != safeConcurrency {
		fmt.Printf("memory: adjusted concurrency from %d to %d due to low memory (%dMB)\n",
			originalSafe, safeConcurrency, availableMB)
	}

	return m.currentActive < safeConcurrency
}

func (m *MemoryMonitor) acquireWorker() bool {
	if !m.shouldAllowNewWorker() {
		m.mu.RLock()
		availableMB := m.availableMemoryMB
		m.mu.RUnlock()

		m.activeMu.Lock()
		currentActive := m.currentActive
		m.activeMu.Unlock()

		fmt.Printf("memory: blocking worker - available: %dMB, active: %d, max: %d\n",
			availableMB, currentActive, m.maxConcurrency)

		if availableMB < 100 {
			fmt.Printf("memory: forcing GC due to critically low memory (%dMB)\n", availableMB)
			runtime.GC()
			time.Sleep(50 * time.Millisecond)
			m.updateMemoryStats()

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

	fmt.Printf("memory: acquired worker - active: %d/%d\n", m.currentActive, m.maxConcurrency)
	return true
}

// releaseWorker releases a worker slot
func (m *MemoryMonitor) releaseWorker() {
	m.activeMu.Lock()
	if m.currentActive > 0 {
		m.currentActive--
	}
	released := m.currentActive
	m.activeMu.Unlock()

	fmt.Printf("memory: released worker - active: %d/%d\n", released, m.maxConcurrency)
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
