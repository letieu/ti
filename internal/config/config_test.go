package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	tempDir := t.TempDir()

	manager, err := NewManager(tempDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	if manager.configDir != tempDir {
		t.Errorf("Expected configDir %s, got %s", tempDir, manager.configDir)
	}

	// Verify directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("Config directory was not created")
	}
}

func TestLoadCreateDefault(t *testing.T) {
	tempDir := t.TempDir()
	manager, _ := NewManager(tempDir)

	config, err := manager.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if config == nil {
		t.Fatal("Expected default config, got nil")
	}

	if config.Data == nil {
		t.Error("Expected initialized settings map")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	manager, _ := NewManager(tempDir)

	config := &Config{
		Data: map[string]string{
			"theme": "dark",
			"lang":  "en",
		},
	}

	// Save
	if err := manager.Save(config); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load
	loaded, err := manager.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Data["theme"] != "dark" {
		t.Errorf("Expected theme 'dark', got '%s'", loaded.Data["theme"])
	}

	if loaded.Data["lang"] != "en" {
		t.Errorf("Expected lang 'en', got '%s'", loaded.Data["lang"])
	}
}

func TestSettings(t *testing.T) {
	config := &Config{
		Data: make(map[string]string),
	}

	// Test SetSetting
	config.SetSetting("key1", "value1")
	config.SetSetting("key2", "value2")

	// Test GetSetting
	val, ok := config.GetSetting("key1")
	if !ok || val != "value1" {
		t.Errorf("Expected 'value1', got '%s' (ok=%v)", val, ok)
	}

	// Test missing setting
	_, ok = config.GetSetting("nonexistent")
	if ok {
		t.Error("Expected ok=false for nonexistent key")
	}

	// Test DeleteSetting
	config.DeleteSetting("key1")
	_, ok = config.GetSetting("key1")
	if ok {
		t.Error("Expected key1 to be deleted")
	}
}

func TestDelete(t *testing.T) {
	tempDir := t.TempDir()
	manager, _ := NewManager(tempDir)

	config := &Config{
		Data: make(map[string]string),
	}

	// Save a config
	if err := manager.Save(config); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify it exists
	if _, err := os.Stat(manager.ConfigPath()); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Delete it
	if err := manager.Delete(); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	if _, err := os.Stat(manager.ConfigPath()); !os.IsNotExist(err) {
		t.Error("Config file still exists after delete")
	}

	// Deleting again should not error
	if err := manager.Delete(); err != nil {
		t.Error("Delete on non-existent file should not error")
	}
}

func TestConfigPath(t *testing.T) {
	tempDir := t.TempDir()
	manager, _ := NewManager(tempDir)

	expectedPath := filepath.Join(tempDir, "config.json")
	if manager.ConfigPath() != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, manager.ConfigPath())
	}
}

func TestFilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	manager, _ := NewManager(tempDir)

	config := &Config{
		Data: make(map[string]string),
	}

	if err := manager.Save(config); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	info, err := os.Stat(manager.ConfigPath())
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	// Check that file is only readable/writable by owner (0600)
	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("Expected file permissions 0600, got %o", mode)
	}
}
