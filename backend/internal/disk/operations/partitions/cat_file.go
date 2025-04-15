package partition_operations

import (
	"fmt"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/types/structures/ext"
	"disk.simulator.com/m/v2/utils"
)

func CatFile(filePath string) (string, error) {
	id := "761A"
	// Obtener la partición montada
	partition, partitionPath, err := memory.GetInstance().GetMountedPartition(id)
	if err != nil {
		return "", fmt.Errorf("error al obtener la partición: %v", err)
	}

	// Extraer directorios padre y nombre del archivo
	parentDirs, destFile := utils.GetParentDirectories(filePath)

	fmt.Println("Leyendo archivo:", filePath)
	fmt.Println("Directorios padre:", parentDirs)
	fmt.Println("Archivo:", destFile)

	// Leer el superbloque de la partición
	superBlock := ext2.SuperBlock{}
	err = superBlock.DeserializeSuperBlock(partitionPath, partition.Partition.Part_start)
	if err != nil {
		return "", fmt.Errorf("error al leer el superbloque: %v", err)
	}

	// Leer el contenido del archivo usando el superbloque
	content, err := superBlock.ReadFile(partitionPath, parentDirs, destFile)
	if err != nil {
		return "", fmt.Errorf("error al leer el archivo: %v", err)
	}

	return content, nil
}
