package handlers

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"

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
	installed, version := checkCaddyInstalled()
	running, _ := checkCaddyRunning()

	Success(c, gin.H{
		"installed":  installed,
		"version":    version,
		"running":    running,
		"unixSocket": "/var/run/caddy/caddy.sock",
		"adminPort":  2019,
	})
}

// checkCaddyInstalled 检查 Caddy 是否已安装
func checkCaddyInstalled() (bool, string) {
	version, err := getCaddyVersion()
	if err != nil {
		return false, ""
	}
	return true, version
}

// checkCaddyRunning 检查 Caddy 是否正在运行
func checkCaddyRunning() (bool, string) {
	client := NewCaddyClient()
	if err := client.CheckConnection(); err != nil {
		return false, ""
	}
	version, _ := client.GetVersion()
	return true, version
}

// getCaddyVersion 获取 Caddy 版本号
func getCaddyVersion() (string, error) {
	cmd := exec.Command("caddy", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// InstallCaddy 安装 Caddy
func InstallCaddy(c *gin.Context) {
	var req struct {
		InstallType string `json:"installType"` // "auto" or "useExisting"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// 尝试自动安装
		req.InstallType = "auto"
	}

	if req.InstallType == "useExisting" {
		// 用户选择使用已安装的 Caddy
		installed, version := checkCaddyInstalled()
		if !installed {
			SuccessWithMessage(c, "Caddy is not installed. Please install Caddy first.", gin.H{
				"success": false,
				"action":  "install",
			})
			return
		}

		// 启动 Caddy
		if err := startCaddy(); err != nil {
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

	// 自动安装 Caddy
	success, message := performCaddyInstall()

	if success {
		// 安装成功后尝试启动
		if err := startCaddy(); err != nil {
			log.Printf("[WARN] Caddy installed but failed to auto-start: %v", err)
			message = "Caddy installed. Please start it manually."
		}
	}

	SuccessWithMessage(c, message, gin.H{
		"success": success,
	})
}

// performCaddyInstall 执行 Caddy 安装
func performCaddyInstall() (bool, string) {
	os := runtime.GOOS

	var installCmd *exec.Cmd

	switch os {
	case "linux":
		// 使用官方安装脚本
		installCmd = exec.Command("bash", "-c", "curl -s https://getcaddy.com | bash")
	case "darwin":
		installCmd = exec.Command("brew", "install", "caddy")
	default:
		return false, fmt.Sprintf("Unsupported OS: %s", os)
	}

	output, err := installCmd.CombinedOutput()
	if err != nil {
		log.Printf("[ERROR] Caddy installation failed: %v, output: %s", err, string(output))
		return false, fmt.Sprintf("Installation failed: %v", err)
	}

	version, _ := getCaddyVersion()
	return true, fmt.Sprintf("Caddy installed successfully. Version: %s", version)
}

// startCaddy 启动 Caddy
func startCaddy() error {
	var cmd *exec.Cmd

	os := runtime.GOOS
	if os == "linux" {
		// 尝试使用 systemctl 启动
		cmd = exec.Command("systemctl", "start", "caddy")
		if err := cmd.Run(); err != nil {
			// 如果 systemctl 失败，尝试直接启动
			cmd = exec.Command("caddy", "run", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile")
			return cmd.Start()
		}
		return nil
	}

	// 直接启动
	cmd = exec.Command("caddy", "run")
	return cmd.Start()
}

// GetCaddyStatus 获取 Caddy 运行状态
func GetCaddyStatus(c *gin.Context) {
	running, version := checkCaddyRunning()

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
	running, version := checkCaddyRunning()

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
	running, _ := checkCaddyRunning()

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
	running, version := checkCaddyRunning()

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

// ReloadCaddy 重新加载 Caddy 配置
func ReloadCaddy(c *gin.Context) {
	if err := caddyClient.CheckConnection(); err != nil {
		log.Printf("[ERROR] Caddy is not running: %v", err)
		SuccessWithMessage(c, "Caddy is not running", gin.H{"success": false})
		return
	}

	if err := caddyClient.LoadConfig(); err != nil {
		log.Printf("[ERROR] Failed to reload Caddy config: %v", err)
		SuccessWithMessage(c, "Failed to reload Caddy", gin.H{"success": false})
		return
	}

	log.Printf("[INFO] Caddy config reloaded successfully")
	SuccessWithMessage(c, "Caddy config reloaded", gin.H{"success": true})
}

// CaddyClient 简化的 Caddy 客户端
type CaddyClient struct {
	adminURL string
}

func NewCaddyClient() *CaddyClient {
	return &CaddyClient{
		adminURL: "http://localhost:2019",
	}
}

func (c *CaddyClient) CheckConnection() error {
	// 使用 systemctl 检查 Caddy 服务状态
	cmd := exec.Command("systemctl", "is-active", "caddy")
	err := cmd.Run()
	if err == nil {
		return nil // 服务正在运行
	}

	// 如果 systemctl 失败，尝试使用 ps 检查
	cmd = exec.Command("bash", "-c", "pgrep -x caddy > /dev/null 2>&1")
	err = cmd.Run()
	if err == nil {
		return nil // 进程存在
	}

	return fmt.Errorf("caddy is not running")
}

func (c *CaddyClient) GetVersion() (string, error) {
	// 使用 caddy version 命令获取版本
	cmd := exec.Command("caddy", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get caddy version: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}
