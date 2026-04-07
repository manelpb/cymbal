package collectors

import (
	"github.com/1broseidon/cymbal/internal/index"
	"github.com/prometheus/client_golang/prometheus"
)

type DBCollector struct {
	dbPath string
}

func NewDBCollector(dbPath string) prometheus.Collector {
	return &DBCollector{dbPath: dbPath}
}

func (c *DBCollector) Describe(ch chan<- *prometheus.Desc) {}

func (c *DBCollector) Collect(ch chan<- prometheus.Metric) {
	store, err := index.OpenStore(c.dbPath)
	if err != nil {
		return
	}
	defer store.Close()

	// Repo stats (file/symbol counts)
	stats, err := store.RepoStats()
	if err == nil && stats.Path != "" {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("cymbal_repo_file_count", "", nil, nil),
			prometheus.GaugeValue, float64(stats.FileCount),
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("cymbal_repo_symbol_count", "", nil, nil),
			prometheus.GaugeValue, float64(stats.SymbolCount),
		)
		for lang, count := range stats.Languages {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("cymbal_repo_language_files", "", []string{"language"}, nil),
				prometheus.GaugeValue, float64(count), lang,
			)
		}
	}

	// Persisted metrics from DB (index ops + query ops)
	allMetrics, err := store.GetAllMetrics()
	if err != nil {
		return
	}

	// Map DB metric names to Prometheus metric descriptors.
	// Index metrics (counters)
	for _, name := range []string{
		"index_files_indexed", "index_files_skipped",
		"index_parse_errors", "index_write_errors",
		"index_symbols_found", "index_stale_removed",
	} {
		if v, ok := allMetrics[name]; ok {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("cymbal_"+name+"_total", "", nil, nil),
				prometheus.CounterValue, v,
			)
		}
	}

	// Query metrics (counters + histogram sum for avg calculation)
	for _, cmd := range []string{
		"investigate", "search", "refs", "outline", "impact", "trace",
	} {
		if total, ok := allMetrics["query_"+cmd+"_total"]; ok {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("cymbal_query_total", "", []string{"command"}, nil),
				prometheus.CounterValue, total, cmd,
			)
		}
		if errs, ok := allMetrics["query_"+cmd+"_errors"]; ok {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("cymbal_query_errors_total", "", []string{"command"}, nil),
				prometheus.CounterValue, errs, cmd,
			)
		}
		if dur, ok := allMetrics["query_"+cmd+"_duration_sec"]; ok {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc("cymbal_query_duration_seconds_sum", "", []string{"command"}, nil),
				prometheus.CounterValue, dur, cmd,
			)
		}
	}
}
