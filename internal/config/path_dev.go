//go:build !release

package config

func GetConfigDir() (string, error) {
	return ".", nil
}
