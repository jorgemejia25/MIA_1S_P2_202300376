package partition_operations

import (
	"fmt"
	"strconv"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/operations/auth"
	ext2 "disk.simulator.com/m/v2/internal/disk/types/structures/ext"
	"disk.simulator.com/m/v2/utils"
)

func EditFile(path string, contenido string) error {
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

	return superBlock.EditFile(
		partitionPath, // Ruta física de la partición
		parentDirs,    // Directorios padre
		fileName,      // Nombre del archivo
		contenido,     // Nuevo contenido
		int32(uidInt), // UID
		int32(gidInt), // GID
	)
}
