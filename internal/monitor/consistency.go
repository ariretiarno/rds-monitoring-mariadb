package monitor

import (
	"database/sql"
	"fmt"
	"time"

	"mariadb-encryption-monitor/internal/database"
)

// ConsistencyResult represents the result of a consistency check
type ConsistencyResult struct {
	TableName      string
	SourceRowCount int64
	TargetRowCount int64
	Consistent     bool
	Timestamp      time.Time
	Error          error
}

// ConsistencyChecker checks data consistency between databases
type ConsistencyChecker struct {
	connMgr *database.ConnectionManager
}

// NewConsistencyChecker creates a new consistency checker
func NewConsistencyChecker(connMgr *database.ConnectionManager) *ConsistencyChecker {
	return &ConsistencyChecker{
		connMgr: connMgr,
	}
}

// CheckTable checks consistency for a single table
func (cc *ConsistencyChecker) CheckTable(tableName string) (*ConsistencyResult, error) {
	result := &ConsistencyResult{
		TableName: tableName,
		Timestamp: time.Now(),
	}

	sourceConn, err := cc.connMgr.GetSourceConnection()
	if err != nil {
		result.Error = fmt.Errorf("source connection error: %w", err)
		return result, result.Error
	}

	targetConn, err := cc.connMgr.GetTargetConnection()
	if err != nil {
		result.Error = fmt.Errorf("target connection error: %w", err)
		return result, result.Error
	}

	// Get row count from source
	sourceCount, err := cc.getRowCount(sourceConn, tableName)
	if err != nil {
		result.Error = fmt.Errorf("source row count error: %w", err)
		return result, result.Error
	}
	result.SourceRowCount = sourceCount

	// Get row count from target
	targetCount, err := cc.getRowCount(targetConn, tableName)
	if err != nil {
		result.Error = fmt.Errorf("target row count error: %w", err)
		return result, result.Error
	}
	result.TargetRowCount = targetCount

	// Compare counts
	result.Consistent = (sourceCount == targetCount)

	return result, nil
}

// CheckAllTables checks consistency for multiple tables
func (cc *ConsistencyChecker) CheckAllTables(tables []string) ([]*ConsistencyResult, error) {
	results := make([]*ConsistencyResult, 0, len(tables))

	for _, table := range tables {
		result, err := cc.CheckTable(table)
		if err != nil {
			// Continue with other tables even if one fails
			results = append(results, result)
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// getRowCount gets the row count for a table
func (cc *ConsistencyChecker) getRowCount(conn interface{ QueryRow(string, ...interface{}) *sql.Row }, tableName string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)
	var count int64
	err := conn.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get row count: %w", err)
	}
	return count, nil
}
