package agent

import (
	"sync"

	"github.com/eirsyl/apollo/pkg/agent/redis"
)

/**
 * This file contains the internal metric storage used by the agent.
 * The values stored here is a key value pair consisting of a string
 * and a float64.
 */

// Metrics stores the collected state of the redis node
type Metrics struct {
	registry map[string]float64
	lock     sync.Mutex
}

// Metric returns a metric, used by a channel when exporting metrics
type Metric struct {
	Name  string
	Value float64
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
func (m *Metrics) ExportMetrics(exportChan *chan Metric) {
	m.lock.Lock()
	defer m.lock.Unlock()

	for name, value := range m.registry {
		m := Metric{
			Name:  name,
			Value: value,
		}
		*exportChan <- m
	}

	close(*exportChan)
}
