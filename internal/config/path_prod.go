//go:build release

package config

import (
	"os"
	"path/filepath"
)

func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "flick"), nil
}
