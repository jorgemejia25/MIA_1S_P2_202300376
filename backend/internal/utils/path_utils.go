package utils

import (
	"path/filepath"
	"strings"
)

func GetParentDirectories(fullPath string) ([]string, string) {
	cleaned := filepath.Clean(fullPath)
	if cleaned == "/" {
		return []string{}, ""
	}

	// Manejar rutas que empiezan y terminan con /
	cleaned = strings.Trim(cleaned, "/")
	parts := strings.Split(cleaned, "/")

	if len(parts) == 0 {
		return []string{}, ""
	}

	if len(parts) == 1 {
		return []string{}, parts[0]
	}

	return parts[:len(parts)-1], parts[len(parts)-1]
}
