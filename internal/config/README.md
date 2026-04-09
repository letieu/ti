# Config Module

A simple configuration persistence module for application settings.

## Features

- **File-based persistence**: Stores configuration in JSON format
- **Settings storage**: Key-value store for application settings
- **Secure file permissions**: Config files are created with 0600 permissions (read/write for owner only)
- **Default location**: Uses `~/.config/ti/config.json` by default
- **Type-safe**: Strongly typed configuration structures

## Usage

### Basic Setup

```go
import "github.com/letieu/ti/internal/config"

// Create a manager with default location (~/.config/ti)
manager, err := config.NewManager("")
if err != nil {
    log.Fatal(err)
}

// Load config (creates default if not exists)
cfg, err := manager.Load()
if err != nil {
    log.Fatal(err)
}
```

### Settings Management

```go
// Set settings
cfg.SetSetting("theme", "dark")
cfg.SetSetting("language", "en")
cfg.SetSetting("editor", "vim")

// Save to disk
if err := manager.Save(cfg); err != nil {
    log.Fatal(err)
}

// Get settings
if theme, ok := cfg.GetSetting("theme"); ok {
    fmt.Println("Theme:", theme)
}

// Delete settings
cfg.DeleteSetting("theme")
manager.Save(cfg)
```

### Custom Config Location

```go
// Use a custom directory
manager, err := config.NewManager("/path/to/custom/dir")
```

### Complete Example

See `examples/usage.go` for a complete working example showing both config and auth usage.

## Data Structure

The configuration is stored as JSON with the following structure:

```json
{
  "settings": {
    "theme": "dark",
    "language": "en",
    "editor": "vim"
  }
}
```

## Security

- Config files are created with `0600` permissions (owner read/write only)
- Configuration is stored in the user's config directory
- No sensitive data should be stored here (use `authstore` for credentials)

## API Reference

### Manager

- `NewManager(configDir string) (*Manager, error)` - Creates a new config manager
- `Load() (*Config, error)` - Loads config from disk or creates default
- `Save(config *Config) error` - Saves config to disk
- `Delete() error` - Deletes the config file
- `ConfigPath() string` - Returns the config file path

### Config

- `GetSetting(key string) (string, bool)` - Gets a setting value
- `SetSetting(key, value string)` - Sets a setting value
- `DeleteSetting(key string)` - Deletes a setting

## Authentication

For authentication credentials, use the separate `authstore` module which supports multiple OAuth providers. See `internal/authstore/README.md` for details.

## Testing

Run the tests with:

```bash
go test ./internal/config -v
```

All functions are fully tested with comprehensive test coverage.
