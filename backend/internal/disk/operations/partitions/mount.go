package partition_operations

import (
	"fmt"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/types/structures"
)

// MountPartition monta una partición en el sistema y le asigna un identificador único.
// Una partición montada puede ser utilizada para operaciones de sistema de archivos.
//
// Parámetros:
//   - name: nombre de la partición a montar
//   - path: ruta del archivo de disco
//
// Retorna un error si la partición no existe o si hay problemas durante el montaje
func MountPartition(name string, path string) error {
	// Leer el MBR del disco
	mbr := structures.MBR{}
	err := mbr.DeserializeMBR(path)

	if err != nil {
		return err
	}

	// Buscar la partición con el nombre especificado
	partition, index, err := FindPartition(name, path)

	if err != nil {
		return err
	}

	if index == -1 {
		return fmt.Errorf("partition not found")
	}

	// Montar la partición
	partition.Part_mount = 1

	// Agregar la partición al almacenamiento en memoria o actualizar fecha
	storage := memory.GetInstance()

	// No necesitamos verificar si ya está montada aquí, Storage.MountPartition lo maneja
	id, err := storage.MountPartition(name, path, partition)
	if err != nil {
		return err
	}

	fmt.Printf("Partition mounted successfully with ID: %s\n", id)

	return nil
}

// UnmountPartition desmonta una partición montada identificada por su ID.
//
// Parámetros:
//   - id: identificador único de la partición a desmontar
//
// Retorna un error si la partición no está montada o si hay problemas durante el desmontaje
func UnmountPartition(id string) error {
	// Obtener la instancia de almacenamiento en memoria
	storage := memory.GetInstance()

	// Intentar desmontar la partición
	err := storage.UnmountPartition(id)
	if err != nil {
		return fmt.Errorf("error al desmontar la partición con ID %s: %v", id, err)
	}

	fmt.Printf("Partition with ID %s unmounted successfully\n", id)
	return nil
}
