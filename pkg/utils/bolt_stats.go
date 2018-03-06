package utils

import (
	"github.com/coreos/bbolt"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	boltFreePageN = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bolt_free_page_n",
		Help: "total number of free pages on the freelist",
	})
	boltPendingPageN = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bolt_pending_page_n",
		Help: "total number of pending pages on the freelist",
	})
	boltFreeAlloc = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bolt_freelist_alloc",
		Help: "total bytes allocated in free pages",
	})
	boltFreelistInUse = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bolt_freelist_in_use",
		Help: "total bytes used by the freelist",
	})

	boltTxN = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bolt_transactions",
		Help: "total number of started read transactions",
	})
	boltOpenTxN = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "bolt_open_transactions",
		Help: "number of currently open read transactions",
	})
)

func init() {
	prometheus.MustRegister(boltFreePageN)
	prometheus.MustRegister(boltPendingPageN)
	prometheus.MustRegister(boltFreeAlloc)
	prometheus.MustRegister(boltFreelistInUse)

	prometheus.MustRegister(boltTxN)
	prometheus.MustRegister(boltOpenTxN)
}

// ReportBoltStats uses an instance of bolt stats and export these to prometheus
func ReportBoltStats(stats *bolt.Stats) {
	boltFreePageN.Set(float64(stats.FreePageN))
	boltPendingPageN.Set(float64(stats.PendingPageN))
	boltFreeAlloc.Set(float64(stats.FreeAlloc))
	boltFreelistInUse.Set(float64(stats.FreelistInuse))

	boltTxN.Set(float64(stats.TxN))
	boltOpenTxN.Set(float64(stats.OpenTxN))
}
