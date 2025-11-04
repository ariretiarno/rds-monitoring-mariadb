package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"mariadb-encryption-monitor/internal/config"
)

// ConnectionManager manages database connections with retry logic
type ConnectionManager struct {
	sourceConn *sql.DB
	targetConn *sql.DB
	sourceConfig *config.DatabaseConfig
	targetConfig *config.DatabaseConfig
	pairName   string
}

// NewConnectionManager creates a new connection manager for a database pair
func NewConnectionManager(sourceDB, targetDB *config.DatabaseConfig, pairName string) *ConnectionManager {
	return &ConnectionManager{
		sourceConfig: sourceDB,
		targetConfig: targetDB,
		pairName:     pairName,
	}
}

// ConnectSource establishes connection to source database with retry logic
func (cm *ConnectionManager) ConnectSource() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cm.sourceConfig.Username,
		cm.sourceConfig.Password,
		cm.sourceConfig.Host,
		cm.sourceConfig.Port,
		cm.sourceConfig.Database,
	)

	return cm.connectWithRetry(&cm.sourceConn, dsn, fmt.Sprintf("source[%s]", cm.pairName))
}

// ConnectTarget establishes connection to target database with retry logic
func (cm *ConnectionManager) ConnectTarget() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cm.targetConfig.Username,
		cm.targetConfig.Password,
		cm.targetConfig.Host,
		cm.targetConfig.Port,
		cm.targetConfig.Database,
	)

	return cm.connectWithRetry(&cm.targetConn, dsn, fmt.Sprintf("target[%s]", cm.pairName))
}

// connectWithRetry attempts to connect with exponential backoff
func (cm *ConnectionManager) connectWithRetry(conn **sql.DB, dsn, dbType string) error {
	maxRetries := 3
	retryInterval := 5 * time.Second

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			lastErr = err
			log.Printf("Attempt %d/%d: Failed to open %s database connection: %v", attempt, maxRetries, dbType, err)
			if attempt < maxRetries {
				time.Sleep(retryInterval)
			}
			continue
		}

		// Test the connection
		if err := db.Ping(); err != nil {
			lastErr = err
			db.Close()
			log.Printf("Attempt %d/%d: Failed to ping %s database: %v", attempt, maxRetries, dbType, err)
			if attempt < maxRetries {
				time.Sleep(retryInterval)
			}
			continue
		}

		// Configure connection pool
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(time.Hour)

		*conn = db
		log.Printf("Successfully connected to %s database", dbType)
		return nil
	}

	return fmt.Errorf("failed to connect to %s database after %d attempts: %w", dbType, maxRetries, lastErr)
}

// GetSourceConnection returns the source database connection
func (cm *ConnectionManager) GetSourceConnection() (*sql.DB, error) {
	if cm.sourceConn == nil {
		return nil, fmt.Errorf("source database connection not established")
	}
	return cm.sourceConn, nil
}

// GetTargetConnection returns the target database connection
func (cm *ConnectionManager) GetTargetConnection() (*sql.DB, error) {
	if cm.targetConn == nil {
		return nil, fmt.Errorf("target database connection not established")
	}
	return cm.targetConn, nil
}

// HealthCheck verifies the health of both database connections
func (cm *ConnectionManager) HealthCheck() (sourceOK, targetOK bool) {
	sourceOK = false
	targetOK = false

	if cm.sourceConn != nil {
		if err := cm.sourceConn.Ping(); err == nil {
			sourceOK = true
		}
	}

	if cm.targetConn != nil {
		if err := cm.targetConn.Ping(); err == nil {
			targetOK = true
		}
	}

	return sourceOK, targetOK
}

// Close closes both database connections
func (cm *ConnectionManager) Close() {
	if cm.sourceConn != nil {
		cm.sourceConn.Close()
		log.Println("Closed source database connection")
	}
	if cm.targetConn != nil {
		cm.targetConn.Close()
		log.Println("Closed target database connection")
	}
}
