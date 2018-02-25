package agent

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	prom_strutil "github.com/prometheus/prometheus/util/strutil"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

// Exporter is reponsible for exporting redis gauges to prometheus.
type Exporter struct {
	addr         string
	metrics      *Metrics
	duration     prometheus.Gauge
	scrapeErrors prometheus.Gauge
	totalScrapes prometheus.Counter
	gauges       map[string]*prometheus.GaugeVec

	gaugesMtx sync.RWMutex
	sync.RWMutex
}

type scrapeResult struct {
	Name  string
	Value float64
}

var (
	metricMap = map[string]string{
		// # Server
		"uptime_in_seconds": "uptime_in_seconds",
		"process_id":        "process_id",

		// # Clients
		"connected_clients":          "connected_clients",
		"client_longest_output_list": "client_longest_output_list",
		"client_biggest_input_buf":   "client_biggest_input_buf",
		"blocked_clients":            "blocked_clients",

		// #
		// Memory
		"used_memory":             "memory_used_bytes",
		"used_memory_rss":         "memory_used_rss_bytes",
		"used_memory_peak":        "memory_used_peak_bytes",
		"used_memory_lua":         "memory_used_lua_bytes",
		"maxmemory":               "memory_max_bytes",
		"mem_fragmentation_ratio": "memory_fragmentation_ratio",

		// #
		// Persistence
		"rdb_changes_since_last_save":  "rdb_changes_since_last_save",
		"rdb_last_bgsave_time_sec":     "rdb_last_bgsave_duration_sec",
		"rdb_current_bgsave_time_sec":  "rdb_current_bgsave_duration_sec",
		"aof_enabled":                  "aof_enabled",
		"aof_rewrite_in_progress":      "aof_rewrite_in_progress",
		"aof_rewrite_scheduled":        "aof_rewrite_scheduled",
		"aof_last_rewrite_time_sec":    "aof_last_rewrite_duration_sec",
		"aof_current_rewrite_time_sec": "aof_current_rewrite_duration_sec",

		// #
		// Stats
		"total_connections_received": "connections_received_total",
		"total_commands_processed":   "commands_processed_total",
		"instantaneous_ops_per_sec":  "instantaneous_ops_per_sec",
		"total_net_input_bytes":      "net_input_bytes_total",
		"total_net_output_bytes":     "net_output_bytes_total",
		"instantaneous_input_kbps":   "instantaneous_input_kbps",
		"instantaneous_output_kbps":  "instantaneous_output_kbps",
		"rejected_connections":       "rejected_connections_total",
		"expired_keys":               "expired_keys_total",
		"evicted_keys":               "evicted_keys_total",
		"keyspace_hits":              "keyspace_hits_total",
		"keyspace_misses":            "keyspace_misses_total",
		"pubsub_channels":            "pubsub_channels",
		"pubsub_patterns":            "pubsub_patterns",
		"latest_fork_usec":           "latest_fork_usec",

		// #
		// Replication
		"loading":                    "loading_dump_file",
		"connected_slaves":           "connected_slaves",
		"repl_backlog_size":          "replication_backlog_bytes",
		"master_last_io_seconds_ago": "master_last_io_seconds",
		"master_repl_offset":         "master_repl_offset",

		// #
		// CPU
		"used_cpu_sys":           "used_cpu_sys",
		"used_cpu_user":          "used_cpu_user",
		"used_cpu_sys_children":  "used_cpu_sys_children",
		"used_cpu_user_children": "used_cpu_user_children",

		// #
		// Cluster
		"cluster_stats_messages_sent":     "cluster_messages_sent_total",
		"cluster_stats_messages_received": "cluster_messages_received_total",
	}
)

// NewExporter returns a new struct implementing the prometheus Metric interface
func NewExporter(addr string, m *Metrics) (*Exporter, error) {
	e := Exporter{
		addr:    addr,
		metrics: m,
		duration: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "redis_exporter_last_scrape_duration_seconds",
			Help: "The last scrape duration.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "redis_exporter_scrapes_total",
			Help: "Current total redis scrapes.",
		}),
		scrapeErrors: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "redis_exporter_last_scrape_error",
			Help: "The last scrape error status.",
		}),
	}
	e.initGauges()
	return &e, nil
}

func (e *Exporter) initGauges() {
	e.gaugesMtx.Lock()
	e.gauges = map[string]*prometheus.GaugeVec{}
	e.gauges["db_keys"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "redis_db_keys",
		Help: "Total number of keys by DB",
	}, []string{"addr"})
	e.gauges["db_avg_ttl_seconds"] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "redis_db_avg_ttl_seconds",
		Help: "Avg TTL in seconds",
	}, []string{"addr"})
	e.gaugesMtx.Unlock()
}

// Describe outputs Redis metric descriptions.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range e.gauges {
		m.Describe(ch)
	}

	ch <- e.duration.Desc()
	ch <- e.totalScrapes.Desc()
	ch <- e.scrapeErrors.Desc()
}

// Collect fetches new gauges from the RedisHost and updates the appropriate
// gauges.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	scrapes := make(chan scrapeResult)

	e.Lock()
	defer e.Unlock()

	//e.initGauges()
	go e.scrape(scrapes)
	e.setMetrics(scrapes)

	ch <- e.duration
	ch <- e.totalScrapes
	ch <- e.scrapeErrors
	e.collectMetrics(ch)
}

func (e *Exporter) scrape(scrapes chan<- scrapeResult) {
	defer close(scrapes)
	now := time.Now().UnixNano()
	e.totalScrapes.Inc()

	errorCount := 0
	var up float64 = 1
	scrapes <- scrapeResult{Name: "up", Value: up}

	var metricChan = make(chan Metric, 1)

	go e.metrics.ExportMetrics(&metricChan)

	for metric := range metricChan {
		name := metric.Name
		if e.includeMetric(name) {

			if newName, ok := metricMap[name]; ok {
				name = newName
			}
			scrapes <- scrapeResult{Name: name, Value: metric.Value}
		} else {
			log.Debugf("Ignored metric: %s", name)
		}
	}

	e.scrapeErrors.Set(float64(errorCount))
	e.duration.Set(float64(time.Now().UnixNano()-now) / 1000000000)
}

func (e *Exporter) setMetrics(scrapes <-chan scrapeResult) {
	for scr := range scrapes {
		name := scr.Name
		if _, ok := e.gauges[name]; !ok {
			e.gaugesMtx.Lock()
			e.gauges[name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: e.sanitizeMetricName(name),
			}, []string{"addr"})
			e.gaugesMtx.Unlock()

		}
		var labels prometheus.Labels = map[string]string{"addr": e.addr}
		e.gauges[name].With(labels).Set(scr.Value)
	}
}

func (e *Exporter) collectMetrics(gauges chan<- prometheus.Metric) {
	for _, m := range e.gauges {
		m.Collect(gauges)
	}
}

func (e *Exporter) includeMetric(s string) bool {
	if strings.HasPrefix(s, "db") || strings.HasPrefix(s, "cmdstat_") || strings.HasPrefix(s, "cluster_") {
		return true
	}
	_, ok := metricMap[s]
	return ok
}

func (e *Exporter) sanitizeMetricName(n string) string {
	s := prom_strutil.SanitizeLabelName(n)
	return fmt.Sprintf("redis_%s", s)
}
