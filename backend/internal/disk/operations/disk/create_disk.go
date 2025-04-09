package disk_operations

import (
	"fmt"
	"os"
	"path/filepath"

	mbr_operations "disk.simulator.com/m/v2/internal/disk/operations/mbr"
	"disk.simulator.com/m/v2/internal/disk/types"
	"disk.simulator.com/m/v2/utils"
)

func CreateDisk(params types.MkDisk) error {
	// Crear el directorio si no existe
	err := os.MkdirAll(filepath.Dir(params.Path), os.ModePerm)
	if err != nil {
		return fmt.Errorf("error al crear el directorio: %v", err)
	}

	// Crear el archivo
	file, err := os.Create(params.Path)
	if err != nil {
		return fmt.Errorf("error al crear el disco: %v", err)
	}
	defer file.Close()

	// Convertir el tamaño a bytes
	sizeInBytes, _ := utils.ConvertToBytes(params.Size, params.Unit)

	// Crear buffer de 1MB para escribir más eficientemente
	buffer := make([]byte, 1024*1024) // 1MB

	// Escribir ceros en el archivo hasta alcanzar el tamaño deseado
	var remaining int64 = sizeInBytes
	for remaining > 0 {
		writeSize := int64(len(buffer))
		if remaining < writeSize {
			writeSize = remaining
		}

		_, err := file.Write(buffer[:writeSize])
		if err != nil {
			return fmt.Errorf("error al escribir en el disco: %v", err)
		}

		remaining -= writeSize
	}

	err = mbr_operations.CreateMBR(params, int32(sizeInBytes))

	return nil
}

