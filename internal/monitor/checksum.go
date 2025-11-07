package monitor

import (
	"database/sql"
	"fmt"
	"time"

	"mariadb-encryption-monitor/internal/database"
)

// ChecksumResult represents the result of a checksum validation
type ChecksumResult struct {
	TableName      string
	SourceChecksum string
	TargetChecksum string
	Match          bool
	Timestamp      time.Time
	Error          error
}

// ChecksumValidator validates data integrity using checksums
type ChecksumValidator struct {
	connMgr *database.ConnectionManager
}

// NewChecksumValidator creates a new checksum validator
func NewChecksumValidator(connMgr *database.ConnectionManager) *ChecksumValidator {
	return &ChecksumValidator{
		connMgr: connMgr,
	}
}

// ValidateTable validates a single table using checksums
func (cv *ChecksumValidator) ValidateTable(tableName string) (*ChecksumResult, error) {
	result := &ChecksumResult{
		TableName: tableName,
		Timestamp: time.Now(),
	}

	sourceConn, err := cv.connMgr.GetSourceConnection()
	if err != nil {
		result.Error = fmt.Errorf("source connection error: %w", err)
		return result, result.Error
	}

	targetConn, err := cv.connMgr.GetTargetConnection()
	if err != nil {
		result.Error = fmt.Errorf("target connection error: %w", err)
		return result, result.Error
	}

	// Calculate checksum for source table
	sourceChecksum, err := cv.calculateChecksum(sourceConn, tableName)
	if err != nil {
		result.Error = fmt.Errorf("source checksum error: %w", err)
		return result, result.Error
	}
	result.SourceChecksum = sourceChecksum

	// Calculate checksum for target table
	targetChecksum, err := cv.calculateChecksum(targetConn, tableName)
	if err != nil {
		result.Error = fmt.Errorf("target checksum error: %w", err)
		return result, result.Error
	}
	result.TargetChecksum = targetChecksum

	// Compare checksums
	result.Match = (sourceChecksum == targetChecksum)

	return result, nil
}

// ValidateAllTables validates multiple tables
func (cv *ChecksumValidator) ValidateAllTables(tables []string) ([]*ChecksumResult, error) {
	results := make([]*ChecksumResult, 0, len(tables))

	for _, table := range tables {
		result, err := cv.ValidateTable(table)
		if err != nil {
			// Continue with other tables even if one fails
			results = append(results, result)
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// calculateChecksum calculates checksum for a table
func (cv *ChecksumValidator) calculateChecksum(conn interface{ Query(string, ...interface{}) (*sql.Rows, error) }, tableName string) (string, error) {
	query := fmt.Sprintf("CHECKSUM TABLE `%s`", tableName)
	rows, err := conn.Query(query)
	if err != nil {
		return "", fmt.Errorf("checksum query failed: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", fmt.Errorf("no checksum result returned")
	}

	var table string
	var checksum interface{}
	if err := rows.Scan(&table, &checksum); err != nil {
		return "", fmt.Errorf("failed to scan checksum result: %w", err)
	}

	if checksum == nil {
		return "", fmt.Errorf("checksum is NULL (table may not exist)")
	}

	return fmt.Sprintf("%v", checksum), nil
}
