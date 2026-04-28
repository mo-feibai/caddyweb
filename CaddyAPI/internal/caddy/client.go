package caddy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	unixSocket string
	adminPort  int
	adminURL   string
	httpClient *http.Client
}

type Config struct {
	Apps map[string]json.RawMessage `json:"apps,omitempty"`
}

type HTTPAppConfig struct {
	Servers map[string]Server `json:"servers,omitempty"`
}

type Server struct {
	Listen      []string         `json:"listen,omitempty"`
	NamedRoutes map[string]Route `json:"named_routes,omitempty"`
	TLSPolicies []TLSPolicy      `json:"tls_connection_policies,omitempty"`
	Errors      *Routes          `json:"errors,omitempty"`
	Routes      []Route          `json:"routes,omitempty"`
}

type TLSPolicy struct {
	Certificate    string `json:"certificate,omitempty"`
	CertificateKey string `json:"certificate_key,omitempty"`
}

type Route struct {
	Id       string   `json:"@id,omitempty"`
	Match    []Match  `json:"match,omitempty"`
	Handle   []Handle `json:"handle,omitempty"`
	Terminal bool     `json:"terminal,omitempty"`
}

type Match struct {
	Host []string `json:"host,omitempty"`
	Path []string `json:"path,omitempty"`
}

type Handle struct {
	Handler      string        `json:"handler,omitempty"`
	Id           string        `json:"@id,omitempty"`
	Root         string        `json:"root,omitempty"`
	IndexNames   []string      `json:"index_names,omitempty"`
	Upstreams    []Upstream    `json:"upstreams,omitempty"`
	Routes       []Route       `json:"routes,omitempty"`
	HealthChecks *HealthChecks `json:"health_checks,omitempty"`
}

type Upstream struct {
	Dial string `json:"dial,omitempty"`
}

type HealthChecks struct {
	Active *ActiveHealthCheck `json:"active,omitempty"`
}

type ActiveHealthCheck struct {
	Path     string `json:"path,omitempty"`
	Interval string `json:"interval,omitempty"`
	Timeout  string `json:"timeout,omitempty"`
}

type Routes []Route

func (r Routes) MarshalJSON() ([]byte, error) {
	type Alias Routes
	return json.Marshal(struct{ Alias }{Alias(r)})
}

type Version struct {
	Version  string `json:"version"`
	Runtime  string `json:"runtime"`
	Revision string `json:"revision"`
}

func NewClient(unixSocket string, adminPort int) *Client {
	return &Client{
		unixSocket: unixSocket,
		adminPort:  adminPort,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func NewRemoteClient(adminURL string) *Client {
	return &Client{
		adminURL: adminURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	var reqURL string
	if c.adminURL != "" {
		reqURL = c.adminURL + path
	} else {
		reqURL = fmt.Sprintf("http://localhost:%d%s", c.adminPort, path)
	}

	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *Client) GetConfig() (*Config, error) {
	data, err := c.doRequest("GET", "/config", nil)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func (c *Client) GetHTTPConfig() (*HTTPAppConfig, error) {
	data, err := c.doRequest("GET", "/config/apps/http", nil)
	if err != nil {
		return nil, err
	}

	var httpConfig HTTPAppConfig
	if err := json.Unmarshal(data, &httpConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal http config: %w", err)
	}

	return &httpConfig, nil
}

func (c *Client) buildPath(path string, withId bool) string {
	path = strings.TrimPrefix(path, "/")
	if withId {
		path = "/id/" + path
	} else {
		path = "/config/" + path
	}
	return path
}

func (c *Client) doConfig(method, path string, withId bool, data interface{}) ([]byte, error) {
	return c.doRequest(method, c.buildPath(path, withId), data)
}

func (c *Client) GetConfigPath(path string, withId bool) (json.RawMessage, error) {
	return c.doConfig("GET", path, withId, nil)
}

func (c *Client) config(method, path string, withId bool, data interface{}) error {
	_, err := c.doConfig(method, path, withId, data)
	return err
}

func (c *Client) PutConfigPath(path string, withId bool, data interface{}) error {
	return c.config("PUT", path, withId, data)
}

func (c *Client) PutConfig(config *Config) error {
	_, err := c.doRequest("PUT", "/config", config)
	return err
}

func (c *Client) PostConfigPath(path string, withId bool, data interface{}) error {
	return c.config("POST", path, withId, data)
}

func (c *Client) DeleteConfigPath(path string, withId bool) error {
	return c.config("DELETE", path, withId, nil)
}

func (c *Client) PatchConfigPath(path string, withId bool, data interface{}) error {
	return c.config("PATCH", path, withId, data)
}

func (c *Client) LoadConfig() error {
	_, err := c.doRequest("POST", "/load", nil)
	return err
}

func (c *Client) CheckConnection() error {
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

func (c *Client) IsInstalled() (bool, string) {
	version, err := c.GetCaddyVersion()
	if err != nil {
		return false, ""
	}
	return true, version
}

func (c *Client) IsRunning() (bool, string) {
	if err := c.CheckConnection(); err != nil {
		return false, ""
	}
	version, _ := c.GetCaddyVersion()
	return true, version
}

func (c *Client) GetCaddyVersion() (string, error) {
	cmd := exec.Command("caddy", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get caddy version: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func (c *Client) Install() (bool, string) {
	os := runtime.GOOS

	var installCmd *exec.Cmd

	switch os {
	case "linux":
		installCmd = exec.Command("bash", "-c", "curl -s https://getcaddy.com | bash")
	case "darwin":
		installCmd = exec.Command("brew", "install", "caddy")
	default:
		return false, fmt.Sprintf("Unsupported OS: %s", os)
	}

	output, err := installCmd.CombinedOutput()
	if err != nil {
		return false, fmt.Sprintf("Installation failed: %v, output: %s", err, string(output))
	}

	version, _ := c.GetCaddyVersion()
	return true, fmt.Sprintf("Caddy installed successfully. Version: %s", version)
}

func (c *Client) Start() error {
	var cmd *exec.Cmd

	os := runtime.GOOS
	if os == "linux" {
		cmd = exec.Command("systemctl", "start", "caddy")
		if err := cmd.Run(); err != nil {
			cmd = exec.Command("caddy", "run", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile")
			return cmd.Start()
		}
		return nil
	}

	cmd = exec.Command("caddy", "run")
	return cmd.Start()
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

func StreamLogs(wsConn *websocket.Conn, follow bool) error {
	return nil
}
