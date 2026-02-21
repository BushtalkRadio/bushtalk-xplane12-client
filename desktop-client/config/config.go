package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

// Config holds application configuration
type Config struct {
	Username    string `json:"username,omitempty"`
	ApiToken    string `json:"api_token,omitempty"`
	ApiURL      string `json:"api_url"`
	XPlanePort  int    `json:"xplane_port"`
	ShowConsole bool   `json:"show_console"`
}

// DefaultConfig returns configuration with default values
func DefaultConfig() *Config {
	return &Config{
		ApiURL:     "https://bushtalkradio.com",
		XPlanePort: 8086,
	}
}

// configDir returns the appropriate config directory for the OS
func configDir() (string, error) {
	var dir string

	switch runtime.GOOS {
	case "windows":
		// Use %APPDATA%\BushtalkRadio
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
		dir = filepath.Join(appData, "BushtalkRadio")
	case "darwin":
		// macOS: ~/Library/Application Support/BushtalkRadio
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, "Library", "Application Support", "BushtalkRadio")
	default:
		// Linux: ~/.config/bushtalkradio
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config", "bushtalkradio")
	}

	return dir, nil
}

// configPath returns the path to config.json
func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load reads configuration from config.json
// Returns default config if file doesn't exist
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save writes configuration to config.json
func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// HasCredentials returns true if username and token are saved
func (c *Config) HasCredentials() bool {
	return c.Username != "" && c.ApiToken != ""
}

// ClearCredentials removes saved credentials
func (c *Config) ClearCredentials() {
	c.Username = ""
	c.ApiToken = ""
}
