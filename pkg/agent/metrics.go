package agent

import (
	"sync"

	"github.com/eirsyl/apollo/pkg/agent/redis"
)

// Metrics stores the collected state of the redis node
type Metrics struct {
	registry map[string]float64
	lock     sync.Mutex
}

// NewMetrics returns a new instance of the metric struct
func NewMetrics() (*Metrics, error) {
	m := Metrics{
		registry: map[string]float64{},
	}
	return &m, nil
}

// RegisterMetric stores a scrape result inside the registry
func (m *Metrics) RegisterMetric(scrape redis.ScrapeResult) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.registry[scrape.Name] = scrape.Value
}

// ExportMetrics returns a map of all metrics stored in the registry
func (m *Metrics) ExportMetrics() map[string]float64 {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.registry
}
