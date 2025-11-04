package storage

import (
	"sync"
	"time"
)

// ConnectionStatus represents database connection status
type ConnectionStatus struct {
	SourceConnected bool
	TargetConnected bool
	LastChecked     time.Time
}

// ReplicaLagMetric represents replica lag measurement
type ReplicaLagMetric struct {
	DatabasePair string
	Timestamp    time.Time
	LagSeconds   float64
	Status       string
	Error        error
}

// ChecksumResult represents the result of a checksum validation
type ChecksumResult struct {
	DatabasePair   string
	TableName      string
	SourceChecksum string
	TargetChecksum string
	Match          bool
	Timestamp      time.Time
	Error          error
}

// ConsistencyResult represents the result of a consistency check
type ConsistencyResult struct {
	DatabasePair   string
	TableName      string
	SourceRowCount int64
	TargetRowCount int64
	Consistent     bool
	Timestamp      time.Time
	Error          error
}

// CurrentMetrics represents the current state of all metrics
type CurrentMetrics struct {
	ReplicaLag         map[string]*ReplicaLagMetric      // key: database_pair
	ChecksumResults    map[string]*ChecksumResult        // key: database_pair:table_name
	ConsistencyResults map[string]*ConsistencyResult     // key: database_pair:table_name
	ConnectionStatus   map[string]ConnectionStatus       // key: database_pair
	LastUpdated        time.Time
}

// MetricsStorage stores monitoring metrics in memory
type MetricsStorage struct {
	mu                  sync.RWMutex
	replicaLagHistory   []ReplicaLagMetric
	checksumResults     map[string]*ChecksumResult        // key: database_pair:table_name
	consistencyResults  map[string]*ConsistencyResult     // key: database_pair:table_name
	connectionStatus    map[string]ConnectionStatus       // key: database_pair
	maxHistorySize      int
	historyDuration     time.Duration
}

// NewMetricsStorage creates a new metrics storage
func NewMetricsStorage() *MetricsStorage {
	return &MetricsStorage{
		replicaLagHistory:   make([]ReplicaLagMetric, 0),
		checksumResults:     make(map[string]*ChecksumResult),
		consistencyResults:  make(map[string]*ConsistencyResult),
		connectionStatus:    make(map[string]ConnectionStatus),
		maxHistorySize:      8640, // 24 hours at 10-second intervals
		historyDuration:     24 * time.Hour,
	}
}

// StoreReplicaLag stores a replica lag metric
func (ms *MetricsStorage) StoreReplicaLag(metric *ReplicaLagMetric) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.replicaLagHistory = append(ms.replicaLagHistory, *metric)

	// Trim history to maintain 24-hour window
	cutoff := time.Now().Add(-ms.historyDuration)
	for i, m := range ms.replicaLagHistory {
		if m.Timestamp.After(cutoff) {
			ms.replicaLagHistory = ms.replicaLagHistory[i:]
			break
		}
	}

	// Also enforce max size
	if len(ms.replicaLagHistory) > ms.maxHistorySize {
		ms.replicaLagHistory = ms.replicaLagHistory[len(ms.replicaLagHistory)-ms.maxHistorySize:]
	}
}

// StoreChecksumResult stores a checksum result
func (ms *MetricsStorage) StoreChecksumResult(result *ChecksumResult) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	key := result.DatabasePair + ":" + result.TableName
	ms.checksumResults[key] = result
}

// StoreConsistencyResult stores a consistency result
func (ms *MetricsStorage) StoreConsistencyResult(result *ConsistencyResult) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	key := result.DatabasePair + ":" + result.TableName
	ms.consistencyResults[key] = result
}

// GetReplicaLagHistory returns replica lag history for the specified duration
func (ms *MetricsStorage) GetReplicaLagHistory(duration time.Duration) []ReplicaLagMetric {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	cutoff := time.Now().Add(-duration)
	result := make([]ReplicaLagMetric, 0)

	for _, metric := range ms.replicaLagHistory {
		if metric.Timestamp.After(cutoff) {
			result = append(result, metric)
		}
	}

	return result
}

// GetCurrentMetrics returns the current state of all metrics
func (ms *MetricsStorage) GetCurrentMetrics() *CurrentMetrics {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	// Get latest replica lag for each database pair
	latestReplicaLag := make(map[string]*ReplicaLagMetric)
	for i := len(ms.replicaLagHistory) - 1; i >= 0; i-- {
		lag := ms.replicaLagHistory[i]
		if _, exists := latestReplicaLag[lag.DatabasePair]; !exists {
			lagCopy := lag
			latestReplicaLag[lag.DatabasePair] = &lagCopy
		}
	}

	return &CurrentMetrics{
		ReplicaLag:         latestReplicaLag,
		ChecksumResults:    ms.checksumResults,
		ConsistencyResults: ms.consistencyResults,
		ConnectionStatus:   ms.connectionStatus,
		LastUpdated:        time.Now(),
	}
}

// UpdateConnectionStatus updates the connection status for a database pair
func (ms *MetricsStorage) UpdateConnectionStatus(pairName string, status ConnectionStatus) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.connectionStatus[pairName] = status
}
