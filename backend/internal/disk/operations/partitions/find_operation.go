package partition_operations

import (
	"fmt"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/operations/auth"
	ext2 "disk.simulator.com/m/v2/internal/disk/types/structures/ext"
	"disk.simulator.com/m/v2/utils"
)

func FindFileOrFolderTree(
	path string,
	name string,
) (string, error) {
	instance := auth.GetInstance()

	if instance.User == nil {
		return "", fmt.Errorf("error al buscar: no hay un usuario loggeado")
	}

	id := instance.ID

	partition, partitionPath, err := memory.GetInstance().GetMountedPartition(id)
	if err != nil {
		return "", fmt.Errorf("error al obtener la partición: %v", err)
	}

	superBlock := ext2.SuperBlock{}
	err = superBlock.DeserializeSuperBlock(partitionPath, partition.Partition.Part_start)

	if err != nil {
		return "", fmt.Errorf("error al leer el superbloque: %v", err)
	}

	// Separar la ruta inicial en directorios padres
	parentDirs, destinyDir := utils.GetParentDirectories(path)

	parentDirs = append(parentDirs, destinyDir)

	// Buscar el archivo o carpeta y generar el árbol de búsqueda
	tree, err := superBlock.FindFileOrFolderByName(partitionPath, parentDirs, name)
	if err != nil {
		return "", fmt.Errorf("error al buscar '%s' a partir de '%s': %v", name, path, err)
	}

	return tree, nil
}
