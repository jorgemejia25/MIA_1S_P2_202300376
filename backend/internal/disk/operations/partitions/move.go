package partition_operations

import (
	"fmt"
	"strconv"
	"strings"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/operations/auth"
	ext2 "disk.simulator.com/m/v2/internal/disk/types/structures/ext"
	"disk.simulator.com/m/v2/utils"
)

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

	uidInt, _ := strconv.ParseInt(instance.User.UID, 10, 32)
	gidInt, _ := strconv.ParseInt(instance.GID, 10, 32)

	// Limpiar las rutas para evitar problemas con slashes repetidos o finales
	sourcePath = cleanPath(sourcePath)
	destPath = cleanPath(destPath)

	// Obtener los componentes de las rutas
	sourceParents, sourceName := utils.GetParentDirectories(sourcePath)
	destParents, destName := utils.GetParentDirectories(destPath)

	// Si el destino termina en slash o está vacío, usar el nombre del origen
	if destName == "" {
		destName = sourceName
	}

	return superBlock.Move(
		partitionPath,
		sourceParents,
		sourceName,
		destParents,
		destName,
		int32(uidInt),
		int32(gidInt),
	)
}

// Función auxiliar para limpiar la ruta, eliminar slashes repetidos y normalizar
func cleanPath(path string) string {
	// Eliminar slashes repetidos
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}

	// Asegurar que comience con /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return path
}
