// filepath: /home/jorgis/Documents/USAC/archivos/proyecto2/backend/internal/disk/operations/disk/list_disks.go
package disk_operations

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"disk.simulator.com/m/v2/internal/disk/types/structures"
)

// DiskInfo contiene información sobre un disco
type DiskInfo struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// ListDisks retorna una lista de los discos creados con mkdisk
// Primero busca en el registro de discos y luego complementa con búsqueda en el sistema de archivos
func ListDisks() ([]DiskInfo, error) {
	var disks []DiskInfo

	// 1. Obtener discos del registro (estos son los creados con mkdisk)
	registry := GetDiskRegistry()
	registeredDisks := registry.GetDisks()

	// Verificar que los discos registrados aún existen en el sistema de archivos
	// y actualizar sus metadatos
	diskRegistry := GetDiskRegistry()
	registryUpdated := false

	for i, disk := range registeredDisks {
		if fileInfo, err := os.Stat(disk.Path); err == nil && !fileInfo.IsDir() {
			// El disco existe, actualizamos la fecha de modificación
			registeredDisks[i].Modified = fileInfo.ModTime()
			registeredDisks[i].Size = fileInfo.Size()
			disks = append(disks, registeredDisks[i])
		} else {
			// El disco ya no existe, lo eliminamos del registro
			diskRegistry.UnregisterDisk(disk.Path)
			registryUpdated = true
		}
	}

	// Si actualizamos el registro, guardar los cambios
	if registryUpdated {
		saveRegistryToFile()
	}

	// 2. Si no se encontraron discos registrados, buscar en el sistema de archivos
	if len(disks) == 0 {
		// Mapa para rastrear discos únicos por nombre
		uniqueDisks := make(map[string]bool)

		// Directorio raíz donde buscar discos - agregar directorios comunes
		homeDir, _ := os.UserHomeDir()
		rootDirs := []string{
			"/tmp",
			os.TempDir(),
			"./",
			"../",
			homeDir,
			filepath.Join(homeDir, "Calificacion_MIA/Discos"), // Directorio específico para los ejemplos
		}

		// Extensiones que podrían tener los archivos de disco
		diskExtensions := []string{
			".mia", // Según lo observado en el código, se usan archivos .mia
			".disk",
			".dsk",
		}

		// Buscar archivos de disco en los directorios
		for _, dir := range rootDirs {
			err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err != nil || info == nil {
					return nil // Ignorar errores de acceso o archivos que no existen
				}

				// Verificar si es archivo y tiene una extensión de disco
				if !info.IsDir() {
					ext := filepath.Ext(path)
					for _, diskExt := range diskExtensions {
						if ext == diskExt {
							// Verificar si el archivo es realmente un disco MBR
							if isMBRDisk(path) {
								// Verificar si ya hemos agregado un disco con este nombre
								if _, exists := uniqueDisks[info.Name()]; !exists {
									// Crear información del disco solo si es un nombre único
									disk := DiskInfo{
										Name:     info.Name(),
										Path:     path,
										Size:     info.Size(),
										Created:  time.Now(), // No podemos obtener la fecha de creación directamente
										Modified: info.ModTime(),
									}

									// Registrar este disco para futuros usos
									diskRegistry.RegisterDisk(disk)

									disks = append(disks, disk)
									uniqueDisks[info.Name()] = true
								}
								break
							}
						}
					}
				}
				return nil
			})

			if err != nil {
				// Continuar con otros directorios si hay error en uno
				continue
			}
		}
	}

	// Si no se encontraron discos, devolver un error
	if len(disks) == 0 {
		return nil, fmt.Errorf("no se encontraron discos")
	}

	return disks, nil
}

// isMBRDisk verifica si un archivo es realmente un disco con formato MBR
func isMBRDisk(path string) bool {
	// Intentar leer el MBR del archivo
	mbr := structures.MBR{}
	err := mbr.DeserializeMBR(path)

	// Si podemos leer el MBR correctamente, es un disco válido
	return err == nil
}

// GetDisksInfo retorna un string formateado con información de los discos
func GetDisksInfo() (string, error) {
	disks, err := ListDisks()
	if err != nil {
		return "", err
	}

	// Formatear la salida
	output := "LISTADO DE DISCOS DISPONIBLES:\n\n"
	output += fmt.Sprintf("%-20s | %-30s | %-15s | %-20s\n", "NOMBRE", "RUTA", "TAMAÑO (bytes)", "ÚLTIMA MODIFICACIÓN")
	output += fmt.Sprintf("%s\n", "-----------------------------------------------------------------------------------------------")

	for _, disk := range disks {
		output += fmt.Sprintf("%-20s | %-30s | %-15d | %-20s\n",
			disk.Name,
			disk.Path,
			disk.Size,
			disk.Modified.Format("2006-01-02 15:04:05"))
	}

	return output, nil
}
