package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/1broseidon/cymbal/internal/index"
	"github.com/spf13/cobra"
)

var gainCmd = &cobra.Command{
	Use:   "gain",
	Short: "Print terminal usage report",
	Long:  "Prints a terminal-native usage report showing index and query statistics.",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := getDBPath(cmd)
		ensureFresh(dbPath)
		return printGainReport(dbPath)
	},
}

func init() {
	rootCmd.AddCommand(gainCmd)
}

func formatK(n float64) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", n/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", n/1_000)
	}
	return fmt.Sprintf("%.0f", n)
}

func bar(pct float64) string {
	const full = "█"
	const empty = "░"
	const total = 20
	filled := int(pct / 100 * float64(total))
	return strings.Repeat(full, filled) + strings.Repeat(empty, total-filled)
}

func shortenPath(p string) string {
	if len(p) <= 50 {
		return p
	}
	const sep = string(filepath.Separator)
	parts := strings.Split(p, sep)
	if len(parts) <= 2 {
		return p[:47] + "..."
	}
	return sep + strings.Join(parts[len(parts)-2:], sep)
}

func printGainReport(dbPath string) error {
	store, err := index.OpenStore(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open store: %w", err)
	}
	defer store.Close()

	allMetrics, err := store.GetAllMetrics()
	if err != nil {
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	indexFilesIndexed := allMetrics["index_files_indexed"]
	indexFilesSkipped := allMetrics["index_files_skipped"]
	indexParseErrors := allMetrics["index_parse_errors"]
	indexWriteErrors := allMetrics["index_write_errors"]
	indexSymbolsFound := allMetrics["index_symbols_found"]
	indexStaleRemoved := allMetrics["index_stale_removed"]

	commands := []string{"investigate", "search", "refs", "outline", "impact", "trace"}
	type cmdStats struct {
		name     string
		count    float64
		errs     float64
		totalDur float64
	}
	var cmds []cmdStats
	for _, name := range commands {
		count := allMetrics["query_"+name+"_total"]
		errs := allMetrics["query_"+name+"_errors"]
		dur := allMetrics["query_"+name+"_duration_sec"]
		if count > 0 || errs > 0 || dur > 0 {
			cmds = append(cmds, cmdStats{
				name:     name,
				count:    count,
				errs:     errs,
				totalDur: dur * 1000,
			})
		}
	}

	totalQueries := 0.0
	for _, c := range cmds {
		totalQueries += c.count
	}

	cacheHitRate := 0.0
	if indexFilesIndexed+indexFilesSkipped > 0 {
		cacheHitRate = indexFilesSkipped / (indexFilesIndexed + indexFilesSkipped) * 100
	}

	repoStats, err := store.RepoStats()
	if err != nil {
		repoStats = &index.RepoStatsResult{}
	}

	cwd, _ := os.Getwd()
	shortRepo := shortenPath(cwd)

	fmt.Printf("Cymbal Usage Statistics\n")
	fmt.Printf("════════════════════════════════════════════════════════════\n")
	fmt.Printf("Repo:           %s\n", shortRepo)
	fmt.Printf("Total commands: %.0f\n", totalQueries)
	fmt.Printf("Files indexed:  %s  (%s skipped, %s parse err, %s write err)\n",
		formatK(indexFilesIndexed),
		formatK(indexFilesSkipped),
		formatK(indexParseErrors),
		formatK(indexWriteErrors))
	fmt.Printf("Cache hit rate: %.1f%%  %s\n\n", cacheHitRate, bar(cacheHitRate))

	fmt.Printf("By Language\n")
	fmt.Printf("────────────────────────────────────────────────────────────────────────\n")
	fmt.Printf("  %-12s  %8s  %8s\n", "Language", "Files", "Symbols")
	fmt.Printf("  %-12s  %8s  %8s\n", "---------", "-----", "-------")

	type langStats struct {
		lang    string
		files   int
		symbols int
	}
	var langs []langStats
	for lang, count := range repoStats.Languages {
		langs = append(langs, langStats{lang: lang, files: count})
	}
	sort.Slice(langs, func(i, j int) bool { return langs[i].files > langs[j].files })
	for _, ls := range langs {
		unusedSymbols := indexSymbolsFound
		_ = unusedSymbols
		fmt.Printf("  %-12s  %8d  %8d\n", ls.lang, ls.files, ls.symbols)
	}

	fmt.Printf("\nBy Command\n")
	fmt.Printf("────────────────────────────────────────────────────────────────────────\n")
	fmt.Printf("  %-12s  %6s  %6s  %8s  %10s\n", "Command", "Count", "Errors", "Avg ms", "Total ms")
	fmt.Printf("  %-12s  %6s  %6s  %8s  %10s\n", "-------", "-----", "------", "-------", "---------")

	sort.Slice(cmds, func(i, j int) bool { return cmds[i].count > cmds[j].count })
	for _, cs := range cmds {
		avgMs := 0.0
		if cs.count > 0 {
			avgMs = cs.totalDur / cs.count
		}
		fmt.Printf("  %-12s  %6.0f  %6.0f  %8.1f  %10.1f\n",
			cs.name, cs.count, cs.errs, avgMs, cs.totalDur)
	}

	_ = indexStaleRemoved

	return nil
}
