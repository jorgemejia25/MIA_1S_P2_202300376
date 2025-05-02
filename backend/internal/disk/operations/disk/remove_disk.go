package disk_operations

import (
	"fmt"
	"os"
)

func RemoveDisk(path string) error {
	// Verificar si el archivo existe
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("el disco en la ruta %s no existe", path)
	}

	// Eliminar el archivo
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("error al eliminar el disco: %v", err)
	}

	// Eliminar el disco del registro
	registry := GetDiskRegistry()
	registry.UnregisterDisk(path)

	return nil
}
