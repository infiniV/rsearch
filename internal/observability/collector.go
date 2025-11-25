package observability

import (
	"sync"
	"time"
)

// Collector manages background system metrics collection
type Collector struct {
	metrics  *Metrics
	stopChan chan struct{}
	wg       sync.WaitGroup
	mu       sync.Mutex
	running  bool
}

// NewCollector creates a new metrics collector
func NewCollector(metrics *Metrics) *Collector {
	return &Collector{
		metrics:  metrics,
		stopChan: make(chan struct{}),
	}
}

// Start begins collecting system metrics in the background
func (c *Collector) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return
	}

	c.running = true
	c.wg.Add(1)

	go func() {
		defer c.wg.Done()

		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		// Update immediately on start
		c.metrics.UpdateSystemMetrics()

		for {
			select {
			case <-ticker.C:
				c.metrics.UpdateSystemMetrics()
			case <-c.stopChan:
				return
			}
		}
	}()
}

// Stop halts the background metrics collection
func (c *Collector) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return
	}

	c.running = false
	close(c.stopChan)
	c.wg.Wait()

	// Recreate stopChan for potential restart
	c.stopChan = make(chan struct{})
}
