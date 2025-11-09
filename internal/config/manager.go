package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/alejandro-bustamante/flick/internal/models"
	"github.com/pelletier/go-toml/v2"
)

func ensureDefaultConfig(path string, defaultContent []byte) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("could not check config file %s: %w", path, err)
	}

	log.Printf("Config file not found. Creating default config at: %s", path)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("could not create config directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, defaultContent, 0644); err != nil {
		return fmt.Errorf("could not write default config file %s: %w", path, err)
	}

	return nil
}

// GetConfigPaths determines the correct paths for settings.toml and patterns.toml.
// It prioritizes the customConfigPath (from the --config flag).
// If no custom path is given, it uses the default (dev or prod) directory
// and ensures the default config files exist.
func GetConfigPaths(customConfigPath string) (string, string, error) {
	var settingsPath, patternsPath string

	if customConfigPath != "" {
		settingsPath = customConfigPath
		patternsPath = filepath.Join(filepath.Dir(customConfigPath), "patterns.toml")
	} else {
		configDir, err := GetConfigDir()
		if err != nil {
			return "", "", fmt.Errorf("could not determine config directory: %w", err)
		}

		settingsPath = filepath.Join(configDir, "settings.toml")
		patternsPath = filepath.Join(configDir, "patterns.toml")

		if err := ensureDefaultConfig(settingsPath, []byte(DefaultSettings)); err != nil {
			log.Printf("Warning: could not create default settings.toml: %v", err)
			log.Println("Please create it manually at:", settingsPath)
		}
		if err := ensureDefaultConfig(patternsPath, DefaultPatterns); err != nil {
			log.Printf("Warning: could not create default patterns.toml: %v", err)
		}
	}

	return settingsPath, patternsPath, nil
}

func LoadPatterns(path string) (*models.Config, error) {
	patterns, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg models.Config
	if err := toml.Unmarshal(patterns, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func LoadSettings(path string) (*models.UserSettings, error) {
	settings, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg models.UserSettings
	if err := toml.Unmarshal(settings, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
