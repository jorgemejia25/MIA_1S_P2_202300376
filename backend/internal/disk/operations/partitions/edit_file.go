package partition_operations

import (
	"fmt"
	"strconv"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/operations/auth"
	ext2 "disk.simulator.com/m/v2/internal/disk/types/structures/ext"
	"disk.simulator.com/m/v2/utils"
)

func EditFile(path string, contentPath string) error {
	instance := auth.GetInstance()

	if instance.User == nil {
		return fmt.Errorf("error al editar archivo: no hay un usuario loggeado")
	}

	id := instance.ID

	partition, partitionPath, err := memory.GetInstance().GetMountedPartition(id)
	if err != nil {
		return fmt.Errorf("error al obtener la partición: %v", err)
	}

	superBlock := ext2.SuperBlock{}
	err = superBlock.DeserializeSuperBlock(partitionPath, partition.Partition.Part_start)
	if err != nil {
		return fmt.Errorf("error al leer el superbloque: %v", err)
	}

	uidInt, _ := strconv.ParseInt(instance.User.UID, 10, 32)
	gidInt, _ := strconv.ParseInt(instance.GID, 10, 32)

	// Obtener directorios padre y nombre de archivo
	parentDirs, fileName := utils.GetParentDirectories(path)

	// Llamar a la función EditFile con la ruta del archivo que contiene el contenido
	err = superBlock.EditFile(
		partitionPath, // Ruta física de la partición
		parentDirs,    // Directorios padre
		fileName,      // Nombre del archivo
		contentPath,   // Ruta al archivo que contiene el nuevo contenido
		int32(uidInt), // UID
		int32(gidInt), // GID
	)

	if err != nil {
		return fmt.Errorf("error al editar archivo: %v", err)
	}

	// Si el sistema de archivos es ext3, registrar la operación en el journaling
	if superBlock.SFilesystemType == 3 {
		// Registrar la operación en el journal
		err = ext2.AddJournal(
			partitionPath,
			int64(partition.Partition.Part_start),
			0, // Este parámetro es ignorado ahora
			"edit",
			path,
			"(contenido de archivo actualizado)",
		)

		if err != nil {
			fmt.Printf("Advertencia: No se pudo registrar la operación en el journaling: %v\n", err)
			// No retornar error, ya que el archivo fue editado exitosamente
		} else {
			fmt.Println("Operación registrada en el journaling")
		}
	}

	// Actualizar el superbloque con los cambios
	err = superBlock.SerializeSuperBlock(partitionPath, partition.Partition.Part_start)
	if err != nil {
		return fmt.Errorf("error al actualizar el superbloque: %v", err)
	}

	fmt.Printf("Archivo '%s' editado exitosamente\n", path)
	return nil
}
