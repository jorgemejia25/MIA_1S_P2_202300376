package partition_operations

import (
	"fmt"
	"os"
	"strconv"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/operations/auth"
	"disk.simulator.com/m/v2/internal/disk/types/structures/ext"
	"disk.simulator.com/m/v2/utils"
)

func CreateFile(
	dirPath string,
	size int,
	contentPath string,
	r bool,
) error {
	instance := auth.GetInstance()

	if instance.User == nil {
		return fmt.Errorf("error al crear directorio: no hay un usuario loggeado")
	}

	id := instance.ID
	// Obtener la partición montada
	partition, partitionPath, err := memory.GetInstance().GetMountedPartition(id)

	if err != nil {
		return fmt.Errorf("error al obtener la partición: %v", err)
	}

	// Extraer directorios padre y nombre del archivo
	parentDirs, destFile := utils.GetParentDirectories(dirPath)

	fmt.Println("Directorios padre:", parentDirs)
	fmt.Println("Archivo destino:", destFile)

	// Leer el superbloque de la partición
	superBlock := ext2.SuperBlock{}
	err = superBlock.DeserializeSuperBlock(partitionPath, partition.Partition.Part_start)
	if err != nil {
		return fmt.Errorf("error al leer el superbloque: %v", err)
	}

	// Convertir uid y gid de string a int32
	uidInt, _ := strconv.ParseInt(instance.User.UID, 10, 32)
	gidInt, _ := strconv.ParseInt(instance.GID, 10, 32)

	// Leer el contenido del archivo de la ruta especificada en mi computadora
	// Read the content from the specified file path
	var content []byte
	if contentPath != "" {
		var err error
		content, err = os.ReadFile(contentPath)
		if err != nil {
			return fmt.Errorf("error al leer el archivo de contenido: %v", err)
		}
	} else {
		// If no content path is specified, use an empty byte slice
		content = []byte{}
	}

	// Crear el archivo usando el superbloque
	err = superBlock.CreateFile(partitionPath, parentDirs, destFile, size, string(content), r,
		int32(uidInt), int32(gidInt),
	)

	if err != nil {
		return fmt.Errorf("error al crear el archivo: %v", err)
	}

	// Actualizar el superbloque con los cambios
	err = superBlock.SerializeSuperBlock(partitionPath, partition.Partition.Part_start)
	if err != nil {
		return fmt.Errorf("error al actualizar el superbloque: %v", err)
	}

	fmt.Printf("Archivo '%s' creado exitosamente en '%s'\n", destFile, dirPath)
	return nil
}
