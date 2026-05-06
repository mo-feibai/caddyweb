package handlers

import (
	"log"
	"net/http"

	"github.com/caddyweb/caddyapi/internal/caddy"
	"github.com/gin-gonic/gin"
)

// 全局 Caddy 客户端
var caddyClient *caddy.Client

func NewRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Initialize global Caddy client
	caddyClient = caddy.NewRemoteClient("http://localhost:2019")

	// API routes
	api := router.Group("/api")
	{
		// Health check
		api.GET("/health", func(c *gin.Context) {
			Success(c, gin.H{
				"status":  "ok",
				"service": "CaddyAPI",
			})
		})

		// Caddy version
		api.GET("/version", func(c *gin.Context) {
			version, err := caddyClient.GetCaddyVersion()
			if err != nil {
				log.Printf("[ERROR] Failed to get version: %v", err)
				InternalServerError(c, "Failed to get version")
				return
			}
			Success(c, version)
		})

		// Config endpoints
		api.GET("/config", func(c *gin.Context) {
			config, err := caddyClient.GetConfig()
			if err != nil {
				log.Printf("[ERROR] Failed to get config: %v", err)
				InternalServerError(c, "Failed to get config")
				return
			}
			Success(c, config)
		})

		api.PUT("/config", func(c *gin.Context) {
			var config caddy.Config
			if err := c.ShouldBindJSON(&config); err != nil {
				log.Printf("[WARN] Invalid config JSON: %v", err)
				BadRequest(c, "Invalid request body")
				return
			}
			if err := caddyClient.PutConfig(&config); err != nil {
				log.Printf("[ERROR] Failed to put config: %v", err)
				InternalServerError(c, "Failed to update config")
				return
			}
			SuccessWithMessage(c, "config updated", nil)
		})

		// Config path operations
		api.GET("/config/*path", func(c *gin.Context) {
			path := c.Param("path")
			data, err := caddyClient.GetConfigPath(path, false)
			if err != nil {
				log.Printf("[ERROR] Failed to get config path %s: %v", path, err)
				InternalServerError(c, "Failed to get config")
				return
			}
			Success(c, data)
		})

		api.PUT("/config/*path", func(c *gin.Context) {
			path := c.Param("path")
			var data interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				log.Printf("[WARN] Invalid data JSON for config path %s: %v", path, err)
				BadRequest(c, "Invalid request body")
				return
			}
			if err := caddyClient.PutConfigPath(path, false, data); err != nil {
				log.Printf("[ERROR] Failed to put config path %s: %v", path, err)
				InternalServerError(c, "Failed to update config")
				return
			}
			SuccessWithMessage(c, "config path updated", nil)
		})

		api.POST("/config/*path", func(c *gin.Context) {
			path := c.Param("path")
			var data interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				log.Printf("[WARN] Invalid data JSON for config path %s: %v", path, err)
				BadRequest(c, "Invalid request body")
				return
			}
			if err := caddyClient.PostConfigPath(path, false, data); err != nil {
				log.Printf("[ERROR] Failed to patch config path %s: %v", path, err)
				InternalServerError(c, "Failed to update config")
				return
			}
			SuccessWithMessage(c, "config path updated", nil)
		})

		api.DELETE("/config/*path", func(c *gin.Context) {
			path := c.Param("path")
			if err := caddyClient.DeleteConfigPath(path, false); err != nil {
				log.Printf("[ERROR] Failed to delete config path %s: %v", path, err)
				InternalServerError(c, "Failed to delete config")
				return
			}
			SuccessWithMessage(c, "config path deleted", nil)
		})

		// Connection check
		api.GET("/ping", func(c *gin.Context) {
			if err := caddyClient.CheckConnection(); err != nil {
				log.Printf("[ERROR] Caddy connection check failed: %v", err)
				ServiceUnavailable(c, "Caddy is not available")
				return
			}
			Success(c, gin.H{"status": "available"})
		})

		// Settings endpoints
		api.GET("/settings", GetSettings)
		api.POST("/settings", SaveSettings)
		api.PUT("/settings", UpdateSettings)
		api.DELETE("/settings", ResetSettings)

		// Caddy management endpoints
		api.GET("/caddy/detect", DetectCaddy)
		api.POST("/caddy/install", InstallCaddy)
		api.GET("/caddy/status", GetCaddyStatus)
		api.GET("/caddy/check-install", CheckCaddyInstallStatus)
		api.POST("/caddy/init", InitCaddy)

		// SSE endpoint for real-time updates
		api.GET("/sse", SSEHandler)

		api.GET("/caddy/servers", GetServers)

		// Domain management
		api.GET("/domains", ListDomains)
		api.POST("/domains", CreateDomain)
		api.DELETE("/domains/:server_id/domains/:domain_id", DeleteDomain)

		// Site management - get sites for domain
		api.GET("/domain-sites/:domain_id", GetDomainSites)
		api.GET("/sites", ListSites)

		// Domain operations (must be after specific routes)
		api.GET("/domains/:name", GetDomain)
		api.POST("/domains/:id/sites", CreateSite)
		api.PUT("/sites/:id", UpdateSite)

		// Site delete
		api.DELETE("/sites/:domain_id/:site_id", DeleteSite)

		// Certificate management
		api.GET("/certs", GetCertificates)
		api.POST("/certs", AddCertificate)
	}

	return router
}
