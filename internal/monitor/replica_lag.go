package monitor

import (
	"database/sql"
	"fmt"
	"log"
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
		metric.Status = "connection_error"
		return metric, err
	}

	// Try SHOW SLAVE STATUS first (MySQL/MariaDB traditional replication)
	query := "SHOW SLAVE STATUS"
	rows, err := targetConn.Query(query)
	if err != nil {
		metric.Error = fmt.Errorf("failed to query slave status: %w", err)
		metric.Status = "query_error"
		return metric, metric.Error
	}
	defer rows.Close()

	if !rows.Next() {
		// No replication configured - this is normal for non-replica databases
		metric.Status = "no_replication"
		metric.LagSeconds = 0
		metric.Error = fmt.Errorf("no replication configured (SHOW SLAVE STATUS returned no rows)")
		return metric, nil
	}

	// Declare variables for replication status
	var secondsBehindMaster sql.NullFloat64
	var slaveIORunning, slaveSQLRunning sql.NullString

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

	// Extract values with detailed logging
	if idx, ok := columnMap["Slave_IO_Running"]; ok {
		if val, ok := values[idx].([]byte); ok {
			slaveIORunning.String = string(val)
			slaveIORunning.Valid = true
			log.Printf("DEBUG: Slave_IO_Running = %s", slaveIORunning.String)
		} else {
			log.Printf("DEBUG: Slave_IO_Running value type: %T, value: %v", values[idx], values[idx])
		}
	} else {
		log.Printf("DEBUG: Slave_IO_Running column not found in SHOW SLAVE STATUS")
	}

	if idx, ok := columnMap["Slave_SQL_Running"]; ok {
		if val, ok := values[idx].([]byte); ok {
			slaveSQLRunning.String = string(val)
			slaveSQLRunning.Valid = true
			log.Printf("DEBUG: Slave_SQL_Running = %s", slaveSQLRunning.String)
		} else {
			log.Printf("DEBUG: Slave_SQL_Running value type: %T, value: %v", values[idx], values[idx])
		}
	} else {
		log.Printf("DEBUG: Slave_SQL_Running column not found in SHOW SLAVE STATUS")
	}

	if idx, ok := columnMap["Seconds_Behind_Master"]; ok {
		log.Printf("DEBUG: Seconds_Behind_Master raw value type: %T, value: %v", values[idx], values[idx])
		if values[idx] != nil {
			switch v := values[idx].(type) {
			case int64:
				secondsBehindMaster.Float64 = float64(v)
				secondsBehindMaster.Valid = true
				log.Printf("DEBUG: Parsed Seconds_Behind_Master as int64: %.2f (Valid=%v)", secondsBehindMaster.Float64, secondsBehindMaster.Valid)
			case uint64:
				secondsBehindMaster.Float64 = float64(v)
				secondsBehindMaster.Valid = true
				log.Printf("DEBUG: Parsed Seconds_Behind_Master as uint64: %.2f (Valid=%v)", secondsBehindMaster.Float64, secondsBehindMaster.Valid)
			case int32:
				secondsBehindMaster.Float64 = float64(v)
				secondsBehindMaster.Valid = true
				log.Printf("DEBUG: Parsed Seconds_Behind_Master as int32: %.2f (Valid=%v)", secondsBehindMaster.Float64, secondsBehindMaster.Valid)
			case uint32:
				secondsBehindMaster.Float64 = float64(v)
				secondsBehindMaster.Valid = true
				log.Printf("DEBUG: Parsed Seconds_Behind_Master as uint32: %.2f (Valid=%v)", secondsBehindMaster.Float64, secondsBehindMaster.Valid)
			case int:
				secondsBehindMaster.Float64 = float64(v)
				secondsBehindMaster.Valid = true
				log.Printf("DEBUG: Parsed Seconds_Behind_Master as int: %.2f (Valid=%v)", secondsBehindMaster.Float64, secondsBehindMaster.Valid)
			case uint:
				secondsBehindMaster.Float64 = float64(v)
				secondsBehindMaster.Valid = true
				log.Printf("DEBUG: Parsed Seconds_Behind_Master as uint: %.2f (Valid=%v)", secondsBehindMaster.Float64, secondsBehindMaster.Valid)
			case float64:
				secondsBehindMaster.Float64 = v
				secondsBehindMaster.Valid = true
				log.Printf("DEBUG: Parsed Seconds_Behind_Master as float64: %.2f (Valid=%v)", secondsBehindMaster.Float64, secondsBehindMaster.Valid)
			case float32:
				secondsBehindMaster.Float64 = float64(v)
				secondsBehindMaster.Valid = true
				log.Printf("DEBUG: Parsed Seconds_Behind_Master as float32: %.2f (Valid=%v)", secondsBehindMaster.Float64, secondsBehindMaster.Valid)
			case []byte:
				strVal := string(v)
				log.Printf("DEBUG: Seconds_Behind_Master as bytes: '%s'", strVal)
				// Try to parse as float
				var f float64
				if _, err := fmt.Sscanf(strVal, "%f", &f); err == nil {
					secondsBehindMaster.Float64 = f
					secondsBehindMaster.Valid = true
					log.Printf("DEBUG: Parsed Seconds_Behind_Master from bytes: %.2f (Valid=%v)", secondsBehindMaster.Float64, secondsBehindMaster.Valid)
				} else {
					log.Printf("DEBUG: Failed to parse Seconds_Behind_Master from bytes: '%s', error: %v", strVal, err)
				}
			case string:
				log.Printf("DEBUG: Seconds_Behind_Master as string: '%s'", v)
				var f float64
				if _, err := fmt.Sscanf(v, "%f", &f); err == nil {
					secondsBehindMaster.Float64 = f
					secondsBehindMaster.Valid = true
					log.Printf("DEBUG: Parsed Seconds_Behind_Master from string: %.2f (Valid=%v)", secondsBehindMaster.Float64, secondsBehindMaster.Valid)
				} else {
					log.Printf("DEBUG: Failed to parse Seconds_Behind_Master from string: '%s', error: %v", v, err)
				}
			default:
				log.Printf("DEBUG: Unexpected type for Seconds_Behind_Master: %T, value: %v", v, v)
			}
		} else {
			log.Printf("DEBUG: Seconds_Behind_Master is NULL (nil value)")
		}
	} else {
		log.Printf("DEBUG: Seconds_Behind_Master column not found in SHOW SLAVE STATUS")
		log.Printf("DEBUG: Available columns: %v", columns)
	}

	// Check replication status
	if slaveIORunning.Valid && slaveSQLRunning.Valid {
		if slaveIORunning.String != "Yes" || slaveSQLRunning.String != "Yes" {
			metric.Status = "replication_stopped"
			metric.LagSeconds = 0
			metric.Error = fmt.Errorf("replication not running (IO: %s, SQL: %s)", slaveIORunning.String, slaveSQLRunning.String)
			return metric, metric.Error
		}
	} else {
		// Couldn't determine replication status
		metric.Status = "status_unknown"
		metric.LagSeconds = 0
		metric.Error = fmt.Errorf("could not determine replication status (Slave_IO_Running or Slave_SQL_Running not found)")
		return metric, nil
	}

	log.Printf("DEBUG: Final check - secondsBehindMaster.Valid=%v, secondsBehindMaster.Float64=%.2f", secondsBehindMaster.Valid, secondsBehindMaster.Float64)
	
	if secondsBehindMaster.Valid {
		metric.LagSeconds = secondsBehindMaster.Float64
		metric.Status = "ok"
		log.Printf("DEBUG: Setting status to 'ok' with lag %.2f seconds", metric.LagSeconds)
	} else {
		// Replication is running but Seconds_Behind_Master is NULL
		// This can happen when replication just started or has issues
		metric.Status = "lag_unknown"
		metric.LagSeconds = 0
		metric.Error = fmt.Errorf("seconds_behind_master is NULL (replication may be initializing)")
		log.Printf("DEBUG: Setting status to 'lag_unknown' because Valid=false")
	}

	return metric, nil
}
