package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"mariadb-encryption-monitor/internal/alert"
	"mariadb-encryption-monitor/internal/config"
	"mariadb-encryption-monitor/internal/storage"
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// WebServer serves the web interface and API
type WebServer struct {
	config    *config.Config
	storage   *storage.MetricsStorage
	alertMgr  *alert.AlertManager
	router    *http.ServeMux
	wsClients map[*websocket.Conn]bool
	mu        sync.RWMutex
	upgrader  websocket.Upgrader
}

// NewWebServer creates a new web server
func NewWebServer(cfg *config.Config, store *storage.MetricsStorage, alertMgr *alert.AlertManager) *WebServer {
	ws := &WebServer{
		config:    cfg,
		storage:   store,
		alertMgr:  alertMgr,
		router:    http.NewServeMux(),
		wsClients: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for simplicity
			},
		},
	}

	ws.setupRoutes()
	return ws
}

// setupRoutes configures HTTP routes
func (ws *WebServer) setupRoutes() {
	ws.router.HandleFunc("/", ws.handleIndex)
	ws.router.HandleFunc("/ws", ws.handleWebSocket)
	ws.router.HandleFunc("/api/metrics", ws.handleMetrics)
	ws.router.HandleFunc("/api/alerts", ws.handleAlerts)
	ws.router.HandleFunc("/api/health", ws.handleHealth)
}

// Start starts the web server
func (ws *WebServer) Start() error {
	addr := fmt.Sprintf(":%d", ws.config.WebServerPort)
	log.Printf("Starting web server on %s", addr)

	// Start broadcast loop
	go ws.broadcastLoop()

	return http.ListenAndServe(addr, ws.router)
}

// handleIndex serves the main HTML page
func (ws *WebServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(indexHTML))
}

// handleWebSocket handles WebSocket connections
func (ws *WebServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	ws.mu.Lock()
	ws.wsClients[conn] = true
	ws.mu.Unlock()

	log.Printf("New WebSocket client connected (total: %d)", len(ws.wsClients))

	// Send initial data
	metrics := ws.storage.GetCurrentMetrics()
	ws.sendToClient(conn, WSMessage{
		Type:      "metrics_update",
		Timestamp: time.Now(),
		Data:      metrics,
	})

	// Handle client disconnection
	go func() {
		defer func() {
			ws.mu.Lock()
			delete(ws.wsClients, conn)
			ws.mu.Unlock()
			conn.Close()
			log.Printf("WebSocket client disconnected (total: %d)", len(ws.wsClients))
		}()

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}

// handleMetrics handles the metrics API endpoint
func (ws *WebServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := ws.storage.GetCurrentMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// handleAlerts handles the alerts API endpoint
func (ws *WebServer) handleAlerts(w http.ResponseWriter, r *http.Request) {
	alerts := ws.alertMgr.GetAlertHistory()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

// handleHealth handles the health check endpoint
func (ws *WebServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	metrics := ws.storage.GetCurrentMetrics()
	
	// Count connected database pairs
	totalPairs := len(metrics.ConnectionStatus)
	connectedPairs := 0
	for _, status := range metrics.ConnectionStatus {
		if status.SourceConnected && status.TargetConnected {
			connectedPairs++
		}
	}
	
	health := map[string]interface{}{
		"status":            "ok",
		"total_pairs":       totalPairs,
		"connected_pairs":   connectedPairs,
		"connection_status": metrics.ConnectionStatus,
		"last_updated":      metrics.LastUpdated,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// broadcastLoop periodically broadcasts updates to all connected clients
func (ws *WebServer) broadcastLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		metrics := ws.storage.GetCurrentMetrics()
		ws.BroadcastUpdate(WSMessage{
			Type:      "metrics_update",
			Timestamp: time.Now(),
			Data:      metrics,
		})
	}
}

// BroadcastUpdate sends an update to all connected WebSocket clients
func (ws *WebServer) BroadcastUpdate(msg WSMessage) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	for conn := range ws.wsClients {
		ws.sendToClient(conn, msg)
	}
}

// sendToClient sends a message to a specific client
func (ws *WebServer) sendToClient(conn *websocket.Conn, msg WSMessage) {
	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Error sending to WebSocket client: %v", err)
	}
}
