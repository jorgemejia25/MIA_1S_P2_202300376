package partition_operations

import (
	"fmt"
	"strconv"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/operations/auth"
	ext2 "disk.simulator.com/m/v2/internal/disk/types/structures/ext"
	"disk.simulator.com/m/v2/utils"
)

// MoveFileOrDirectory mueve un archivo o directorio a una ubicación de destino
func MoveFileOrDirectory(sourcePath string, destPath string) error {
	instance := auth.GetInstance()

	if instance.User == nil {
		return fmt.Errorf("error al mover: no hay un usuario loggeado")
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

	// Obtener directorios padres y nombre del origen
	sourceParentDirs, sourceName := utils.GetParentDirectories(sourcePath)

	// Verificar si destino es un directorio y debe mantener el nombre de origen
	destParentDirs, destDirName := utils.GetParentDirectories(destPath)

	exists, err := superBlock.FolderExists(partitionPath, destParentDirs, destDirName)
	if err != nil {
		return fmt.Errorf("error al verificar el destino: %v", err)
	}

	var destName string

	if exists {
		// Si el destino es un directorio existente, moveremos dentro manteniendo el nombre original
		destParentDirs = append(destParentDirs, destDirName)
		destName = sourceName
	} else {
		// Si no existe o no es un directorio, usamos el nombre especificado en la ruta destino
		destName = destDirName
	}

	// Convertir uid y gid de string a int32
	uidInt, _ := strconv.ParseInt(instance.User.UID, 10, 32)
	gidInt, _ := strconv.ParseInt(instance.GID, 10, 32)

	// Ejecutar la operación de movimiento
	err = superBlock.Move(
		partitionPath,
		sourceParentDirs,
		sourceName,
		destParentDirs,
		destName,
		int32(uidInt),
		int32(gidInt),
	)
	if err != nil {
		return fmt.Errorf("error al mover: %v", err)
	}

	// Actualizar el superbloque con los cambios
	err = superBlock.SerializeSuperBlock(partitionPath, partition.Partition.Part_start)
	if err != nil {
		return fmt.Errorf("error al actualizar el superbloque: %v", err)
	}

	// Si el sistema de archivos es ext3, registrar la operación en el journaling
	if superBlock.SFilesystemType == 3 {
		// Registrar la operación en el journal
		err = ext2.AddJournal(
			partitionPath,
			int64(partition.Partition.Part_start),
			0, // Este parámetro es ignorado ahora
			"move",
			sourcePath+" to "+destPath,
			"",
		)

		if err != nil {
			fmt.Printf("Advertencia: No se pudo registrar la operación en el journaling: %v\n", err)
			// No retornar error, ya que el movimiento fue exitoso
		} else {
			fmt.Println("Operación registrada en el journaling")
		}
	}

	fmt.Printf("'%s' fue movido exitosamente a '%s'\n", sourcePath, destPath)
	return nil
}
