package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// SSE event types
const (
	SSEEventCaddyStatus = "caddy_status"
	SSEEventConfig      = "config"
	SSEEventLog         = "log"
)

// CaddyStatus represents the Caddy connection status
type CaddyStatus struct {
	Status  string `json:"status"` // "running", "stopped", "checking"
	Message string `json:"message"`
	Version string `json:"version,omitempty"`
}

// SSEClient represents a client connected to SSE endpoint
type SSEClient struct {
	channel chan string
}

// Global SSE client manager
var sseClients = make(map[*SSEClient]bool)

func init() {
	// Start Caddy status checker
	go startCaddyStatusChecker()
}

// startCaddyStatusChecker periodically checks Caddy status and broadcasts to all SSE clients
func startCaddyStatusChecker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		status := checkCaddyRealTime()
		broadcastSSE(SSEEventCaddyStatus, status)
	}
}

// checkCaddyRealTime checks the real-time Caddy status
func checkCaddyRealTime() CaddyStatus {
	client := NewCaddyClient()
	if err := client.CheckConnection(); err != nil {
		return CaddyStatus{
			Status:  "stopped",
			Message: "Caddy2 is not running",
		}
	}

	version, _ := client.GetVersion()
	return CaddyStatus{
		Status:  "running",
		Message: "Caddy2 is running",
		Version: version,
	}
}

// broadcastSSE sends an SSE event to all connected clients
func broadcastSSE(event string, data interface{}) {
	message, err := json.Marshal(data)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal SSE data: %v", err)
		return
	}

	for client := range sseClients {
		select {
		case client.channel <- fmt.Sprintf("event: %s\ndata: %s\n\n", event, string(message)):
		default:
			// Channel full or closed, skip this client
			log.Printf("[WARN] SSE client channel full, removing client")
			delete(sseClients, client)
			close(client.channel)
		}
	}
}

// SSEHandler handles Server-Sent Events for real-time updates
func SSEHandler(c *gin.Context) {
	// Create SSE client
	client := &SSEClient{
		channel: make(chan string, 100),
	}
	sseClients[client] = true

	// Send initial Caddy status immediately
	initialStatus := checkCaddyRealTime()
	initialMessage, _ := json.Marshal(initialStatus)

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Flush headers for initial data
	c.Writer.Flush()

	// Send initial status
	c.Writer.WriteString(fmt.Sprintf("event: %s\ndata: %s\n\n", SSEEventCaddyStatus, string(initialMessage)))
	c.Writer.Flush()

	// Create a ticker for keep-alive
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		delete(sseClients, client)
		close(client.channel)
	}()

	// Stream events to client
	clientGone := c.Request.Context().Done()
	for {
		select {
		case <-clientGone:
			log.Printf("[INFO] SSE client disconnected")
			return
		case <-ticker.C:
			// Send keep-alive comment
			c.Writer.WriteString(": keep-alive\n\n")
			c.Writer.Flush()
		case message, ok := <-client.channel:
			if !ok {
				return
			}
			c.Writer.WriteString(message)
			c.Writer.Flush()
		}
	}
}

// CheckCaddyAndInstall checks if Caddy is installed, returns installation status
func CheckCaddyAndInstall(c *gin.Context) {
	status := checkCaddyRealTime()
	Success(c, gin.H{
		"installed": status.Status == "running",
		"status":    status,
	})
}

// GetCaddyVersion gets Caddy version via local command
func GetCaddyVersion(c *gin.Context) {
	version, err := getCaddyVersion()
	if err != nil {
		log.Printf("[WARN] Failed to get Caddy version: %v", err)
		Success(c, gin.H{
			"installed": false,
			"version":   "",
			"message":   "Caddy not found",
		})
		return
	}

	Success(c, gin.H{
		"installed": true,
		"version":   version,
		"message":   "Caddy is installed",
	})
}
