package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/1broseidon/cymbal/internal/metrics/collectors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Expose Prometheus metrics via HTTP server",
	Long: `Starts a lightweight HTTP server that serves Prometheus metrics
at /metrics for the currently indexed repository.

Metrics include index stats (files indexed, parse errors, symbols found),
query stats (counts and latencies by command), and repo-level gauges
(file/symbol counts by language).

Examples:
  cymbal metrics --port 9090
  cymbal metrics --port 9090 --db ./index.db`,
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		dbPath := getDBPath(cmd)
		ensureFresh(dbPath)

		registry := prometheus.NewRegistry()
		registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
		registry.MustRegister(prometheus.NewGoCollector())
		registry.MustRegister(collectors.NewDBCollector(dbPath))

		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
		mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok\n"))
		})

		addr := fmt.Sprintf(":%d", port)
		fmt.Fprintf(os.Stderr, "Prometheus metrics server listening on %s\n", addr)
		fmt.Fprintf(os.Stderr, "  /metrics  — Prometheus scrape endpoint\n")
		fmt.Fprintf(os.Stderr, "  /health   — health check\n")
		return http.ListenAndServe(addr, mux)
	},
}

func init() {
	metricsCmd.Flags().IntP("port", "p", 9090, "port to listen on")
	rootCmd.AddCommand(metricsCmd)
}
