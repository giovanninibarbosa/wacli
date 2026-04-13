package config

import (
	"os"
	"path/filepath"
	"strings"
)

func DefaultStoreDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ".wacli"
	}
	return filepath.Join(home, ".wacli")
}

func ReadOnlyEnabled() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("WACLI_READONLY")))
	switch value {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
