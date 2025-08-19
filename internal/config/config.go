package config

import (
	"os"

	"github.com/alejandro-bustamante/flick/internal/models"
	"github.com/pelletier/go-toml/v2"
)

func LoadData(path string) (*models.Config, error) {

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
	settings, err := os.ReadFile("./settings.toml")
	if err != nil {
		return nil, err
	}

	var cfg models.UserSettings
	if err := toml.Unmarshal(settings, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
