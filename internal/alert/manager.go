package alert

import (
	"fmt"
	"sync"
	"time"

	"mariadb-encryption-monitor/internal/config"
)

// Alert represents an alert
type Alert struct {
	ID        string
	Timestamp time.Time
	Severity  string
	Type      string
	Message   string
	Resolved  bool
}

// AlertManager manages alerts
type AlertManager struct {
	config       *config.Config
	alerts       []Alert
	activeAlerts map[string]*Alert
	mu           sync.RWMutex
}

// NewAlertManager creates a new alert manager
func NewAlertManager(cfg *config.Config) *AlertManager {
	return &AlertManager{
		config:       cfg,
		alerts:       make([]Alert, 0),
		activeAlerts: make(map[string]*Alert),
	}
}

// ReplicaLagMetric represents replica lag data for alert evaluation
type ReplicaLagMetric struct {
	LagSeconds float64
	Status     string
	Error      error
}

// EvaluateReplicaLag evaluates replica lag and generates alerts if needed
func (am *AlertManager) EvaluateReplicaLag(pairName string, metric *ReplicaLagMetric) {
	if metric == nil {
		return
	}

	alertKey := fmt.Sprintf("replica_lag_%s", pairName)

	// Check if lag exceeds threshold
	if metric.Status == "ok" && metric.LagSeconds > am.config.ReplicaLagThreshold.Seconds() {
		alert := Alert{
			ID:        fmt.Sprintf("%s_%d", alertKey, time.Now().Unix()),
			Timestamp: time.Now(),
			Severity:  "WARNING",
			Type:      "replica_lag",
			Message:   fmt.Sprintf("[%s] Replica lag (%.2f seconds) exceeds threshold (%.2f seconds)", pairName, metric.LagSeconds, am.config.ReplicaLagThreshold.Seconds()),
			Resolved:  false,
		}
		am.addAlert(alertKey, alert)
	} else if metric.Status == "replication_stopped" {
		alert := Alert{
			ID:        fmt.Sprintf("%s_%d", alertKey, time.Now().Unix()),
			Timestamp: time.Now(),
			Severity:  "CRITICAL",
			Type:      "replication_stopped",
			Message:   fmt.Sprintf("[%s] Replication stopped: %v", pairName, metric.Error),
			Resolved:  false,
		}
		am.addAlert(alertKey, alert)
	} else {
		// Resolve alert if it exists
		am.resolveAlert(alertKey)
	}
}

// ChecksumResult represents checksum data for alert evaluation
type ChecksumResult struct {
	TableName      string
	SourceChecksum string
	TargetChecksum string
	Match          bool
	Error          error
}

// EvaluateChecksum evaluates checksum results and generates alerts if needed
func (am *AlertManager) EvaluateChecksum(pairName string, result *ChecksumResult) {
	if result == nil {
		return
	}

	alertKey := fmt.Sprintf("checksum_%s_%s", pairName, result.TableName)

	if !result.Match && result.Error == nil {
		alert := Alert{
			ID:        fmt.Sprintf("%s_%d", alertKey, time.Now().Unix()),
			Timestamp: time.Now(),
			Severity:  "CRITICAL",
			Type:      "checksum_mismatch",
			Message:   fmt.Sprintf("[%s] Checksum mismatch for table %s (source: %s, target: %s)", pairName, result.TableName, result.SourceChecksum, result.TargetChecksum),
			Resolved:  false,
		}
		am.addAlert(alertKey, alert)
	} else if result.Error != nil {
		alert := Alert{
			ID:        fmt.Sprintf("%s_%d", alertKey, time.Now().Unix()),
			Timestamp: time.Now(),
			Severity:  "WARNING",
			Type:      "checksum_error",
			Message:   fmt.Sprintf("[%s] Checksum validation error for table %s: %v", pairName, result.TableName, result.Error),
			Resolved:  false,
		}
		am.addAlert(alertKey, alert)
	} else {
		// Resolve alert if it exists
		am.resolveAlert(alertKey)
	}
}

// ConsistencyResult represents consistency data for alert evaluation
type ConsistencyResult struct {
	TableName      string
	SourceRowCount int64
	TargetRowCount int64
	Consistent     bool
	Error          error
}

// EvaluateConsistency evaluates consistency results and generates alerts if needed
func (am *AlertManager) EvaluateConsistency(pairName string, result *ConsistencyResult) {
	if result == nil {
		return
	}

	alertKey := fmt.Sprintf("consistency_%s_%s", pairName, result.TableName)

	if !result.Consistent && result.Error == nil {
		alert := Alert{
			ID:        fmt.Sprintf("%s_%d", alertKey, time.Now().Unix()),
			Timestamp: time.Now(),
			Severity:  "CRITICAL",
			Type:      "consistency_mismatch",
			Message:   fmt.Sprintf("[%s] Row count mismatch for table %s (source: %d, target: %d)", pairName, result.TableName, result.SourceRowCount, result.TargetRowCount),
			Resolved:  false,
		}
		am.addAlert(alertKey, alert)
	} else if result.Error != nil {
		alert := Alert{
			ID:        fmt.Sprintf("%s_%d", alertKey, time.Now().Unix()),
			Timestamp: time.Now(),
			Severity:  "WARNING",
			Type:      "consistency_error",
			Message:   fmt.Sprintf("[%s] Consistency check error for table %s: %v", pairName, result.TableName, result.Error),
			Resolved:  false,
		}
		am.addAlert(alertKey, alert)
	} else {
		// Resolve alert if it exists
		am.resolveAlert(alertKey)
	}
}

// addAlert adds or updates an alert
func (am *AlertManager) addAlert(key string, alert Alert) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Check if alert already exists to avoid duplicates
	if existing, exists := am.activeAlerts[key]; exists {
		if existing.Message == alert.Message {
			return // Duplicate alert, don't add
		}
	}

	am.activeAlerts[key] = &alert
	am.alerts = append(am.alerts, alert)
}

// resolveAlert resolves an active alert
func (am *AlertManager) resolveAlert(key string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if alert, exists := am.activeAlerts[key]; exists {
		alert.Resolved = true
		delete(am.activeAlerts, key)
	}
}

// GetActiveAlerts returns all active alerts
func (am *AlertManager) GetActiveAlerts() []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	active := make([]Alert, 0, len(am.activeAlerts))
	for _, alert := range am.activeAlerts {
		active = append(active, *alert)
	}

	return active
}

// GetAlertHistory returns all alerts (including resolved)
func (am *AlertManager) GetAlertHistory() []Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	// Return last 100 alerts
	start := 0
	if len(am.alerts) > 100 {
		start = len(am.alerts) - 100
	}

	return am.alerts[start:]
}
