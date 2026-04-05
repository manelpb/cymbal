package metrics

import (
	"sync"
)

var (
	metricDBPath string
	metricMu     sync.RWMutex
)

// SetDBPath records the database path for metric persistence.
// Called once by the index command after ensureFresh.
func SetDBPath(path string) {
	metricMu.Lock()
	metricDBPath = path
	metricMu.Unlock()
}

// GetDBPath returns the stored database path for metric operations.
func GetDBPath() string {
	metricMu.RLock()
	defer metricMu.RUnlock()
	return metricDBPath
}
