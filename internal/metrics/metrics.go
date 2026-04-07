package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	IndexFilesIndexed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cymbal_index_files_indexed_total",
			Help: "Total files successfully indexed",
		},
		[]string{"language"},
	)

	IndexFilesSkipped = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cymbal_index_files_skipped_total",
			Help: "Total files skipped (unchanged since last index)",
		},
		[]string{"language"},
	)

	IndexErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cymbal_index_errors_total",
			Help: "Total indexing errors",
		},
		[]string{"type"},
	)

	IndexDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cymbal_index_duration_seconds",
			Help:    "Time spent indexing per file",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"language"},
	)

	IndexSymbolsFound = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cymbal_index_symbols_found_total",
			Help: "Total symbols extracted during indexing",
		},
		[]string{"language"},
	)

	IndexStaleRemoved = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cymbal_index_stale_removed_total",
			Help: "Total stale files removed during reindex",
		},
	)

	QueryTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cymbal_query_total",
			Help: "Total query commands executed",
		},
		[]string{"command"},
	)

	QueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cymbal_query_duration_seconds",
			Help:    "Time spent handling query commands",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5},
		},
		[]string{"command"},
	)

	QueryErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cymbal_query_errors_total",
			Help: "Total query errors",
		},
		[]string{"command"},
	)

	RepoFileCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "cymbal_repo_file_count",
			Help: "Number of indexed files in the current repo",
		},
	)

	RepoSymbolCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "cymbal_repo_symbol_count",
			Help: "Number of indexed symbols in the current repo",
		},
	)

	RepoLanguages = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cymbal_repo_language_files",
			Help: "Number of indexed files per language",
		},
		[]string{"language"},
	)
)
