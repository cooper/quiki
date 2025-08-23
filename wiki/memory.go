package wiki

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
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
		fmt.Printf("memory: creating global instance with %d max workers\n", maxConcurrency)
		globalMemoryMonitor = NewMemoryMonitor(maxConcurrency)
	})
	return globalMemoryMonitor
}

func NewMemoryMonitor(maxConcurrency int) *MemoryMonitor {
	fmt.Printf("memory: initializing with max concurrency %d\n", maxConcurrency)
	m := &MemoryMonitor{
		maxConcurrency: maxConcurrency,
	}
	m.updateMemoryStats()
	return m
}

func (m *MemoryMonitor) updateMemoryStats() {
	// get actual system memory statistics
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		// fallback to conservative estimate if we can't get system info
		m.mu.Lock()
		m.availableMemoryMB = 400 // assume 400MB available
		m.lastCheck = time.Now()
		m.mu.Unlock()
		fmt.Printf("memory stats: failed to get system memory, assuming 400MB available\n")
		return
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	goHeapMB := int64(memStats.Alloc) / (1024 * 1024)

	totalMB := int64(memInfo.Total) / (1024 * 1024)
	usedMB := int64(memInfo.Used) / (1024 * 1024)
	availableMB := int64(memInfo.Available) / (1024 * 1024)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.availableMemoryMB = availableMB
	m.lastCheck = time.Now()

	fmt.Printf("memory stats: total=%dMB, used=%dMB, available=%dMB, go_heap=%dMB\n",
		totalMB, usedMB, availableMB, goHeapMB)
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
	if availableMB < 200 {
		safeConcurrency = 1
	} else if availableMB < 400 {
		safeConcurrency = max(2, safeConcurrency/2)
	}

	fmt.Printf("memory: available=%dMB, calculated=%d, final=%d, active=%d, max=%d\n",
		availableMB, originalSafe, safeConcurrency, m.currentActive, m.maxConcurrency)

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

		if availableMB < 200 {
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
