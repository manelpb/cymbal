package cmd

import (
	"time"

	"github.com/1broseidon/cymbal/internal/index"
	"github.com/1broseidon/cymbal/internal/metrics"
)

func timeQuery(cmdName string, fn func() error) error {
	start := time.Now()
	err := fn()
	dur := time.Since(start).Seconds()

	if dbPath := metrics.GetDBPath(); dbPath != "" {
		store, openErr := index.OpenStore(dbPath)
		if openErr == nil {
			store.RecordMetric("query_"+cmdName+"_total", 1)
			if err != nil {
				store.RecordMetric("query_"+cmdName+"_errors", 1)
			}
			store.RecordMetric("query_"+cmdName+"_duration_sec", dur)
			store.Close()
		}
	}

	metrics.QueryDuration.WithLabelValues(cmdName).Observe(dur)
	if err != nil {
		metrics.QueryErrors.WithLabelValues(cmdName).Inc()
	}
	metrics.QueryTotal.WithLabelValues(cmdName).Inc()
	return err
}
