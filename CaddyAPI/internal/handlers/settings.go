package handlers

import (
	"log"

	"github.com/caddyweb/caddyapi/internal/config"
	"github.com/gin-gonic/gin"
)

// GetSettings 获取设置
func GetSettings(c *gin.Context) {
	cfg := config.GetManager().Get()
	Success(c, cfg)
}

// SaveSettings 保存设置
func SaveSettings(c *gin.Context) {
	var cfg config.Config
	if err := c.ShouldBindJSON(&cfg); err != nil {
		log.Printf("[WARN] Invalid settings JSON: %v", err)
		BadRequest(c, "Invalid request body")
		return
	}

	if err := config.GetManager().Save(&cfg); err != nil {
		log.Printf("[ERROR] Failed to save settings: %v", err)
		InternalServerError(c, "Failed to save settings")
		return
	}

	SuccessWithMessage(c, "settings saved", nil)
}

// UpdateSettings 更新部分设置
func UpdateSettings(c *gin.Context) {
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		log.Printf("[WARN] Invalid settings updates JSON: %v", err)
		BadRequest(c, "Invalid request body")
		return
	}

	cfg := config.GetManager().Get()

	// 应用更新
	if theme, ok := updates["theme"].(string); ok {
		cfg.Theme = theme
	}
	if language, ok := updates["language"].(string); ok {
		cfg.Language = language
	}
	if caddySettings, ok := updates["caddy"].(map[string]interface{}); ok {
		if unixSocket, ok := caddySettings["unixSocket"].(string); ok {
			cfg.CaddySettings.UnixSocket = unixSocket
		}
		if adminPort, ok := caddySettings["adminPort"].(float64); ok {
			cfg.CaddySettings.AdminPort = int(adminPort)
		}
	}

	if err := config.GetManager().Save(cfg); err != nil {
		log.Printf("[ERROR] Failed to update settings: %v", err)
		InternalServerError(c, "Failed to update settings")
		return
	}

	SuccessWithMessage(c, "settings updated", nil)
}

// ResetSettings 重置设置
func ResetSettings(c *gin.Context) {
	if err := config.GetManager().Reset(); err != nil {
		log.Printf("[ERROR] Failed to reset settings: %v", err)
		InternalServerError(c, "Failed to reset settings")
		return
	}

	SuccessWithMessage(c, "settings reset", nil)
}

// CheckCaddyInstallStatus 检测 Caddy 安装状态
func CheckCaddyInstallStatus(c *gin.Context) {
	installed, version := caddyClient.IsInstalled()
	running, _ := caddyClient.IsRunning()

	Success(c, gin.H{
		"installed":  installed,
		"version":    version,
		"running":    running,
		"unixSocket": "/var/run/caddy/caddy.sock",
		"adminPort":  2019,
	})
}

// InstallCaddy 安装 Caddy
func InstallCaddy(c *gin.Context) {
	var req struct {
		InstallType string `json:"installType"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.InstallType = "auto"
	}

	if req.InstallType == "useExisting" {
		installed, version := caddyClient.IsInstalled()
		if !installed {
			SuccessWithMessage(c, "Caddy is not installed. Please install Caddy first.", gin.H{
				"success": false,
				"action":  "install",
			})
			return
		}

		if err := caddyClient.Start(); err != nil {
			log.Printf("[ERROR] Failed to start Caddy: %v", err)
			SuccessWithMessage(c, "Caddy is installed but failed to start. Please start it manually.", gin.H{
				"success": false,
				"version": version,
				"action":  "manual_start",
			})
			return
		}

		SuccessWithMessage(c, "Caddy is ready to use", gin.H{
			"success": true,
			"version": version,
			"action":  "ready",
		})
		return
	}

	success, message := caddyClient.Install()

	if success {
		if err := caddyClient.Start(); err != nil {
			log.Printf("[WARN] Caddy installed but failed to auto-start: %v", err)
			message = "Caddy installed. Please start it manually."
		}
	}

	SuccessWithMessage(c, message, gin.H{
		"success": success,
	})
}

// GetCaddyStatus 获取 Caddy 运行状态
func GetCaddyStatus(c *gin.Context) {
	running, version := caddyClient.IsRunning()

	if running {
		SuccessWithMessage(c, "Caddy2 is running", gin.H{
			"status":  "running",
			"version": version,
		})
		return
	}

	SuccessWithMessage(c, "Caddy2 is not running", gin.H{
		"status":  "stopped",
		"version": "",
	})
}

// DetectCaddy 检测 Caddy 连接
func DetectCaddy(c *gin.Context) {
	running, version := caddyClient.IsRunning()

	if running {
		SuccessWithMessage(c, "Caddy2 is running", gin.H{
			"success": true,
			"version": version,
		})
		return
	}

	SuccessWithMessage(c, "Caddy2 is not running", gin.H{
		"success": false,
	})
}

type ServerInfo struct {
	ID     string   `json:"id"`
	Listen []string `json:"listen"`
}

// GetServers 获取服务器列表
func GetServers(c *gin.Context) {
	running, _ := caddyClient.IsRunning()

	if !running {
		SuccessWithMessage(c, "Caddy2 is not running", gin.H{
			"success": false,
		})
		return
	}

	httpConfig, err := caddyClient.GetHTTPConfig()
	if err != nil {
		log.Printf("[ERROR] Failed to get HTTP config: %v", err)
		SuccessWithMessage(c, "Failed to get servers", gin.H{
			"success": false,
		})
		return
	}

	servers := []ServerInfo{}
	if httpConfig != nil && httpConfig.Servers != nil {
		for id, server := range httpConfig.Servers {
			servers = append(servers, ServerInfo{
				ID:     id,
				Listen: server.Listen,
			})
		}
	}

	Success(c, servers)
}

type InitCaddyRequest struct {
	Force bool `json:"force"`
}

// InitCaddy 初始化 Caddy 配置
func InitCaddy(c *gin.Context) {
	running, version := caddyClient.IsRunning()

	if !running {
		SuccessWithMessage(c, "Caddy2 is not running", gin.H{
			"success": false,
		})
		return
	}

	httpConfig, err := caddyClient.GetHTTPConfig()
	if err != nil {
		log.Printf("[ERROR] Failed to get HTTP config: %v", err)
		SuccessWithMessage(c, "Failed to get Caddy config", gin.H{
			"success": false,
		})
		return
	}

	existingServers := make(map[string]bool)
	if httpConfig != nil && httpConfig.Servers != nil {
		for name := range httpConfig.Servers {
			existingServers[name] = true
		}
	}

	hasExistingConfig := len(existingServers) > 0

	var req InitCaddyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Force = false
	}

	if hasExistingConfig && !req.Force {
		SuccessWithMessage(c, "Existing Caddy configuration detected. This will replace existing server configurations.", gin.H{
			"success":        false,
			"warning":        true,
			"existing_count": len(existingServers),
		})
		return
	}

	secureServerID := "caddyweb_https"
	httpServerID := "caddyweb_http"

	secureServerConfig := map[string]interface{}{
		"@id":          secureServerID,
		"listen":       []string{":443"},
		"named_routes": map[string]interface{}{},
	}

	httpServerConfig := map[string]interface{}{
		"@id":          httpServerID,
		"listen":       []string{":80"},
		"named_routes": map[string]interface{}{},
	}

	caddyClient.DeleteConfigPath("apps/http/servers", false)

	if err := caddyClient.PutConfigPath("apps/http/servers/"+secureServerID, false, secureServerConfig); err != nil {
		log.Printf("[ERROR] Failed to initialize HTTPS server: %v", err)
		SuccessWithMessage(c, "Failed to initialize Caddy", gin.H{
			"success": false,
		})
		return
	}

	if err := caddyClient.PutConfigPath("apps/http/servers/"+httpServerID, false, httpServerConfig); err != nil {
		log.Printf("[ERROR] Failed to initialize HTTP server: %v", err)
		SuccessWithMessage(c, "Failed to initialize Caddy", gin.H{
			"success": false,
		})
		return
	}

	log.Printf("[INFO] Caddy initialized successfully with server IDs: %s (443), %s (80)", secureServerID, httpServerID)
	SuccessWithMessage(c, "Caddy initialized successfully", gin.H{
		"success":   true,
		"version":   version,
		"server_id": secureServerID,
		"ports":     []string{":80", ":443"},
	})
}
