package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

// Define structs that match the JSON structure.
// The `json:"..."` tags tell Go which JSON field maps to which variable.

type Config struct {
	Server  ServerConfig   `json:"server"`
	Cameras []CameraConfig `json:"cameras"`
}

type ServerConfig struct {
	Port          int    `json:"port"`
	DBPath   string `json:"db_path"`
	StoragePath   string `json:"storage_path"`
	RetentionDays int    `json:"retention_days"`
}

type CameraConfig struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

// LoadConfig reads the file at the given path and returns a filled Config struct
func LoadConfig(path string) (*Config, error) {
	// Read the file content
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config json: %w", err)
	}

	// (Optional)
	if cfg.Server.DBPath == "" {
		cfg.Server.DBPath = "./"
	}

	// (Optional) Validate basic settings
	if cfg.Server.StoragePath == "" {
		cfg.Server.StoragePath = "./recordings" // Default fallback
	}

	return &cfg, nil
}
