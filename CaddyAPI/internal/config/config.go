package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type Config struct {
	DeployedMode  string        `json:"deployed_mode"`
	APIBaseURL    string        `json:"api_base_url"`
	WSBaseURL     string        `json:"ws_base_url"`
	CaddySettings CaddySettings `json:"caddy"`
	Theme         string        `json:"theme"`
	Language      string        `json:"language"`
	FirstLaunch   bool          `json:"first_launch"`
	ReloadMode    string        `json:"reload_mode"`
}

type CaddySettings struct {
	APIURL     string `json:"api_url"`
	UnixSocket string `json:"unix_socket"`
	AdminPort  int    `json:"admin_port"`
}

type ConfigManager struct {
	configPath string
	config     *Config
	mu         sync.RWMutex
}

var (
	defaultConfig = &Config{
		DeployedMode: "local",
		APIBaseURL:   "http://localhost:8081",
		WSBaseURL:    "http://localhost:8081",
		CaddySettings: CaddySettings{
			UnixSocket: "/var/run/caddy/caddy.sock",
			AdminPort:  2019,
		},
		Theme:       "light",
		Language:    "zh-CN",
		FirstLaunch: true,
		ReloadMode:  "auto",
	}
	manager     *ConfigManager
	managerOnce sync.Once
)

func GetManager() *ConfigManager {
	managerOnce.Do(func() {
		configDir, err := os.UserConfigDir()
		if err != nil {
			configDir = "/etc"
		}
		configPath := filepath.Join(configDir, "caddyweb", "config.json")
		manager = &ConfigManager{
			configPath: configPath,
			config:     defaultConfig,
		}
		if err := manager.Load(); err != nil {
			log.Printf("[ERROR] Failed to load config: %v", err)
		}
	})
	return manager
}

func (m *ConfigManager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.config = defaultConfig
			return nil
		}
		log.Printf("[ERROR] Failed to read config from %s: %v", m.configPath, err)
		return fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("[ERROR] Failed to parse config from %s: %v", m.configPath, err)
		return fmt.Errorf("failed to parse config: %w", err)
	}

	m.config = &cfg
	return nil
}

func (m *ConfigManager) Save(cfg *Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Printf("[ERROR] Failed to marshal config: %v", err)
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("[ERROR] Failed to create config directory %s: %v", dir, err)
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		log.Printf("[ERROR] Failed to write config to %s: %v", m.configPath, err)
		return fmt.Errorf("failed to write config: %w", err)
	}

	m.config = cfg
	return nil
}

func (m *ConfigManager) Get() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

func (m *ConfigManager) Update(cfg *Config) error {
	return m.Save(cfg)
}

func (m *ConfigManager) Reset() error {
	return m.Save(defaultConfig)
}
