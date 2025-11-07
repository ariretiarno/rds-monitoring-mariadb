package monitor

import (
	"log"
	"sync"
	"time"

	"mariadb-encryption-monitor/internal/alert"
	"mariadb-encryption-monitor/internal/config"
	"mariadb-encryption-monitor/internal/database"
	"mariadb-encryption-monitor/internal/storage"
)

// DatabasePairMonitor monitors a single database pair
type DatabasePairMonitor struct {
	pairName           string
	tables             []string
	connMgr            *database.ConnectionManager
	replicaLagMonitor  *ReplicaLagMonitor
	checksumValidator  *ChecksumValidator
	consistencyChecker *ConsistencyChecker
}

// MonitoringEngine orchestrates all monitoring operations
type MonitoringEngine struct {
	config       *config.Config
	pairMonitors []*DatabasePairMonitor
	storage      *storage.MetricsStorage
	alertMgr     *alert.AlertManager
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

// NewMonitoringEngine creates a new monitoring engine
func NewMonitoringEngine(cfg *config.Config, store *storage.MetricsStorage, alertMgr *alert.AlertManager) *MonitoringEngine {
	// Create monitors for each database pair
	pairMonitors := make([]*DatabasePairMonitor, 0, len(cfg.DatabasePairs))
	
	for _, pair := range cfg.DatabasePairs {
		connMgr := database.NewConnectionManager(&pair.SourceDB, &pair.TargetDB, pair.Name)
		
		pairMonitor := &DatabasePairMonitor{
			pairName:           pair.Name,
			tables:             pair.TablesToMonitor,
			connMgr:            connMgr,
			replicaLagMonitor:  NewReplicaLagMonitor(connMgr),
			checksumValidator:  NewChecksumValidator(connMgr),
			consistencyChecker: NewConsistencyChecker(connMgr),
		}
		
		pairMonitors = append(pairMonitors, pairMonitor)
	}

	return &MonitoringEngine{
		config:       cfg,
		pairMonitors: pairMonitors,
		storage:      store,
		alertMgr:     alertMgr,
		stopChan:     make(chan struct{}),
	}
}

// Start starts the monitoring engine
func (me *MonitoringEngine) Start() error {
	log.Printf("Starting monitoring engine for %d database pair(s)...", len(me.pairMonitors))

	// Connect to all database pairs
	for _, pairMonitor := range me.pairMonitors {
		log.Printf("Connecting to database pair: %s", pairMonitor.pairName)
		
		if err := pairMonitor.connMgr.ConnectSource(); err != nil {
			log.Printf("Warning: Failed to connect to source database for pair '%s': %v", pairMonitor.pairName, err)
		}

		if err := pairMonitor.connMgr.ConnectTarget(); err != nil {
			log.Printf("Warning: Failed to connect to target database for pair '%s': %v", pairMonitor.pairName, err)
		}

		// Update initial connection status
		sourceOK, targetOK := pairMonitor.connMgr.HealthCheck()
		me.storage.UpdateConnectionStatus(pairMonitor.pairName, storage.ConnectionStatus{
			SourceConnected: sourceOK,
			TargetConnected: targetOK,
			LastChecked:     time.Now(),
		})
	}

	// Start monitoring loop
	me.wg.Add(1)
	go me.monitoringLoop()

	log.Println("Monitoring engine started")
	return nil
}

// Stop stops the monitoring engine
func (me *MonitoringEngine) Stop() {
	log.Println("Stopping monitoring engine...")
	close(me.stopChan)
	me.wg.Wait()
	
	// Close all database connections
	for _, pairMonitor := range me.pairMonitors {
		pairMonitor.connMgr.Close()
	}
	
	log.Println("Monitoring engine stopped")
}

// monitoringLoop runs the monitoring cycle at configured intervals
func (me *MonitoringEngine) monitoringLoop() {
	defer me.wg.Done()

	ticker := time.NewTicker(me.config.MonitoringInterval)
	defer ticker.Stop()

	// Run initial cycle immediately
	me.runMonitoringCycle()

	for {
		select {
		case <-ticker.C:
			me.runMonitoringCycle()
		case <-me.stopChan:
			return
		}
	}
}

// runMonitoringCycle executes a single monitoring cycle
func (me *MonitoringEngine) runMonitoringCycle() {
	log.Println("Running monitoring cycle...")

	var wg sync.WaitGroup

	// Monitor each database pair
	for _, pairMonitor := range me.pairMonitors {
		wg.Add(1)
		go func(pm *DatabasePairMonitor) {
			defer wg.Done()
			me.monitorDatabasePair(pm)
		}(pairMonitor)
	}

	wg.Wait()
	log.Println("Monitoring cycle completed")
}

// monitorDatabasePair monitors a single database pair
func (me *MonitoringEngine) monitorDatabasePair(pm *DatabasePairMonitor) {
	// Update connection status
	sourceOK, targetOK := pm.connMgr.HealthCheck()
	me.storage.UpdateConnectionStatus(pm.pairName, storage.ConnectionStatus{
		SourceConnected: sourceOK,
		TargetConnected: targetOK,
		LastChecked:     time.Now(),
	})

	var wg sync.WaitGroup

	// Run replica lag monitoring
	wg.Add(1)
	go func() {
		defer wg.Done()
		if targetOK {
			metric, err := pm.replicaLagMonitor.MeasureLag()
			if err != nil {
				log.Printf("[%s] Replica lag monitoring error: %v", pm.pairName, err)
			}
			if metric != nil {
				// Convert to storage type
				storageMetric := &storage.ReplicaLagMetric{
					DatabasePair: pm.pairName,
					Timestamp:    metric.Timestamp,
					LagSeconds:   metric.LagSeconds,
					Status:       metric.Status,
					Error:        metric.Error,
				}
				me.storage.StoreReplicaLag(storageMetric)
				// Convert to alert type
				alertMetric := &alert.ReplicaLagMetric{
					LagSeconds: metric.LagSeconds,
					Status:     metric.Status,
					Error:      metric.Error,
				}
				me.alertMgr.EvaluateReplicaLag(pm.pairName, alertMetric)
			}
		} else {
			log.Printf("[%s] Skipping replica lag check: target database not connected", pm.pairName)
		}
	}()

	// Run checksum validation
	if len(pm.tables) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if sourceOK && targetOK {
				results, err := pm.checksumValidator.ValidateAllTables(pm.tables)
				if err != nil {
					log.Printf("[%s] Checksum validation error: %v", pm.pairName, err)
				}
				for _, result := range results {
					// Convert to storage type
					storageResult := &storage.ChecksumResult{
						DatabasePair:   pm.pairName,
						TableName:      result.TableName,
						SourceChecksum: result.SourceChecksum,
						TargetChecksum: result.TargetChecksum,
						Match:          result.Match,
						Timestamp:      result.Timestamp,
						Error:          result.Error,
					}
					me.storage.StoreChecksumResult(storageResult)
					// Convert to alert type
					alertResult := &alert.ChecksumResult{
						TableName:      result.TableName,
						SourceChecksum: result.SourceChecksum,
						TargetChecksum: result.TargetChecksum,
						Match:          result.Match,
						Error:          result.Error,
					}
					me.alertMgr.EvaluateChecksum(pm.pairName, alertResult)
				}
			} else {
				log.Printf("[%s] Skipping checksum validation: databases not connected", pm.pairName)
			}
		}()

		// Run consistency checking
		wg.Add(1)
		go func() {
			defer wg.Done()
			if sourceOK && targetOK {
				results, err := pm.consistencyChecker.CheckAllTables(pm.tables)
				if err != nil {
					log.Printf("[%s] Consistency check error: %v", pm.pairName, err)
				}
				for _, result := range results {
					// Convert to storage type
					storageResult := &storage.ConsistencyResult{
						DatabasePair:   pm.pairName,
						TableName:      result.TableName,
						SourceRowCount: result.SourceRowCount,
						TargetRowCount: result.TargetRowCount,
						Consistent:     result.Consistent,
						Timestamp:      result.Timestamp,
						Error:          result.Error,
					}
					me.storage.StoreConsistencyResult(storageResult)
					// Convert to alert type
					alertResult := &alert.ConsistencyResult{
						TableName:      result.TableName,
						SourceRowCount: result.SourceRowCount,
						TargetRowCount: result.TargetRowCount,
						Consistent:     result.Consistent,
						Error:          result.Error,
					}
					me.alertMgr.EvaluateConsistency(pm.pairName, alertResult)
				}
			} else {
				log.Printf("[%s] Skipping consistency check: databases not connected", pm.pairName)
			}
		}()
	}

	wg.Wait()
}
