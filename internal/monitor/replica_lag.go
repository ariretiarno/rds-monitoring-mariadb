package monitor

import (
	"database/sql"
	"fmt"
	"time"

	"mariadb-encryption-monitor/internal/database"
)

// ReplicaLagMetric represents replica lag measurement
type ReplicaLagMetric struct {
	Timestamp  time.Time
	LagSeconds float64
	Status     string
	Error      error
}

// ReplicaLagMonitor monitors replication lag
type ReplicaLagMonitor struct {
	connMgr *database.ConnectionManager
}

// NewReplicaLagMonitor creates a new replica lag monitor
func NewReplicaLagMonitor(connMgr *database.ConnectionManager) *ReplicaLagMonitor {
	return &ReplicaLagMonitor{
		connMgr: connMgr,
	}
}

// MeasureLag measures the current replication lag
func (rlm *ReplicaLagMonitor) MeasureLag() (*ReplicaLagMetric, error) {
	metric := &ReplicaLagMetric{
		Timestamp: time.Now(),
		Status:    "unknown",
	}

	targetConn, err := rlm.connMgr.GetTargetConnection()
	if err != nil {
		metric.Error = err
		metric.Status = "error"
		return metric, err
	}

	// Query SHOW SLAVE STATUS
	var secondsBehindMaster sql.NullFloat64
	var slaveIORunning, slaveSQLRunning sql.NullString

	query := "SHOW SLAVE STATUS"
	rows, err := targetConn.Query(query)
	if err != nil {
		metric.Error = fmt.Errorf("failed to query slave status: %w", err)
		metric.Status = "error"
		return metric, metric.Error
	}
	defer rows.Close()

	if !rows.Next() {
		// No replication configured
		metric.Status = "no_replication"
		metric.LagSeconds = 0
		return metric, nil
	}

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		metric.Error = fmt.Errorf("failed to get columns: %w", err)
		metric.Status = "error"
		return metric, metric.Error
	}

	// Create a slice to hold the values
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Scan the row
	if err := rows.Scan(valuePtrs...); err != nil {
		metric.Error = fmt.Errorf("failed to scan slave status: %w", err)
		metric.Status = "error"
		return metric, metric.Error
	}

	// Find the indices of the columns we need
	columnMap := make(map[string]int)
	for i, col := range columns {
		columnMap[col] = i
	}

	// Extract values
	if idx, ok := columnMap["Slave_IO_Running"]; ok {
		if val, ok := values[idx].([]byte); ok {
			slaveIORunning.String = string(val)
			slaveIORunning.Valid = true
		}
	}

	if idx, ok := columnMap["Slave_SQL_Running"]; ok {
		if val, ok := values[idx].([]byte); ok {
			slaveSQLRunning.String = string(val)
			slaveSQLRunning.Valid = true
		}
	}

	if idx, ok := columnMap["Seconds_Behind_Master"]; ok {
		if values[idx] != nil {
			switch v := values[idx].(type) {
			case int64:
				secondsBehindMaster.Float64 = float64(v)
				secondsBehindMaster.Valid = true
			case float64:
				secondsBehindMaster.Float64 = v
				secondsBehindMaster.Valid = true
			case []byte:
				// Try to parse as float
				var f float64
				if _, err := fmt.Sscanf(string(v), "%f", &f); err == nil {
					secondsBehindMaster.Float64 = f
					secondsBehindMaster.Valid = true
				}
			}
		}
	}

	// Check replication status
	if slaveIORunning.Valid && slaveSQLRunning.Valid {
		if slaveIORunning.String != "Yes" || slaveSQLRunning.String != "Yes" {
			metric.Status = "replication_stopped"
			metric.Error = fmt.Errorf("replication not running (IO: %s, SQL: %s)", slaveIORunning.String, slaveSQLRunning.String)
			return metric, metric.Error
		}
	}

	if secondsBehindMaster.Valid {
		metric.LagSeconds = secondsBehindMaster.Float64
		metric.Status = "ok"
	} else {
		metric.Status = "unknown"
		metric.Error = fmt.Errorf("seconds_behind_master is NULL")
	}

	return metric, nil
}
