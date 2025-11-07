package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"mariadb-encryption-monitor/internal/alert"
	"mariadb-encryption-monitor/internal/config"
	"mariadb-encryption-monitor/internal/monitor"
	"mariadb-encryption-monitor/internal/storage"
	"mariadb-encryption-monitor/internal/web"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	log.Println("Loading configuration...")
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded successfully")
	log.Printf("Monitoring interval: %v", cfg.MonitoringInterval)
	log.Printf("Replica lag threshold: %v", cfg.ReplicaLagThreshold)
	log.Printf("Web server port: %d", cfg.WebServerPort)
	log.Printf("Tables to monitor: %v", cfg.TablesToMonitor)

	// Initialize components
	metricsStorage := storage.NewMetricsStorage()
	alertManager := alert.NewAlertManager(cfg)
	monitoringEngine := monitor.NewMonitoringEngine(cfg, metricsStorage, alertManager)
	webServer := web.NewWebServer(cfg, metricsStorage, alertManager)

	// Start monitoring engine
	if err := monitoringEngine.Start(); err != nil {
		log.Fatalf("Failed to start monitoring engine: %v", err)
	}

	// Start web server in a goroutine
	go func() {
		log.Printf("Starting web server on port %d...", cfg.WebServerPort)
		if err := webServer.Start(); err != nil {
			log.Fatalf("Web server error: %v", err)
		}
	}()

	log.Println("MariaDB Encryption Migration Monitor is running")
	log.Printf("Access the web interface at http://localhost:%d", cfg.WebServerPort)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutdown signal received")
	monitoringEngine.Stop()
	log.Println("Shutdown complete")
}
