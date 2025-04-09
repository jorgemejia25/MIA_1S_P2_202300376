package partition_operations

import (
	"fmt"
	"os"
)

// ReadFile lee el contenido de un archivo en la ruta especificada
func ReadFile(path string) (string, error) {
	// Aquí iría la lógica real para leer el archivo del sistema de archivos simulado
	// Por ahora, implementamos una versión simple que intenta leer un archivo real

	// Validar que la ruta no esté vacía
	if path == "" {
		return "", fmt.Errorf("la ruta del archivo es requerida")
	}

	// Si es una ruta simulada, podríamos procesarla aquí
	// Por simplicidad, intentamos leer un archivo real
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error al leer el archivo: %w", err)
	}

	return string(content), nil
}
