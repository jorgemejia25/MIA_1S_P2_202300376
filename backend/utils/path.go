package utils

import "strings"

// GetParentDirectories separa una ruta en sus directorios padres y el destino
func GetParentDirectories(path string) ([]string, string) {
	// Eliminar la barra inicial si existe
	cleanPath := path
	if strings.HasPrefix(cleanPath, "/") {
		cleanPath = path[1:]
	}

	// Si la ruta está vacía, devolver arrays vacíos
	if cleanPath == "" {
		return []string{}, ""
	}

	// Dividir la ruta en componentes
	components := strings.Split(cleanPath, "/")

	// El último componente es el destino
	destDir := components[len(components)-1]

	// Los componentes anteriores son los directorios padres
	var parentsDir []string
	if len(components) > 1 {
		parentsDir = components[:len(components)-1]
	} else {
		parentsDir = []string{}
	}

	return parentsDir, destDir
}

// IsRoot verifica si la ruta es la raíz
func IsRoot(path string) bool {
	return path == "/" || path == ""
}

// GetRelativePath obtiene la ruta relativa a partir de una absoluta
func GetRelativePath(path string) string {
	if strings.HasPrefix(path, "/") {
		return path[1:]
	}
	return path
}

// PrintPath imprime una ruta de forma legible
func PrintPath(components []string) string {
	if len(components) == 0 {
		return "/"
	}
	return "/" + strings.Join(components, "/")
}
