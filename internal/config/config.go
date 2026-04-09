package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

var (
	ErrConfigNotFound = errors.New("config file not found")
	ErrInvalidConfig  = errors.New("invalid config data")
)

// Manager handles configuration persistence
type Manager struct {
	configDir  string
	configFile string
}

// NewManager creates a new configuration manager
// If configDir is empty, it uses the default user config directory
func NewManager(configDir string, name string) (*Manager, error) {
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		configDir = filepath.Join(homeDir, ".config", "ti")
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	return &Manager{
		configDir:  configDir,
		configFile: filepath.Join(configDir, name),
	}, nil
}

// Load reads the configuration from disk
func (m *Manager) Load(v any) error {
	data, err := os.ReadFile(m.configFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

// Save writes the configuration to disk
func (m *Manager) Save(config any) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.configFile, data, 0600)
}

// Delete removes the configuration file
func (m *Manager) Delete() error {
	if err := os.Remove(m.configFile); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// ConfigPath returns the path to the config file
func (m *Manager) ConfigPath() string {
	return m.configFile
}
