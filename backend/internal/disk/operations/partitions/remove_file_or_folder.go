package partition_operations

import (
	"fmt"
	"strconv"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/operations/auth"
	ext2 "disk.simulator.com/m/v2/internal/disk/types/structures/ext"
	"disk.simulator.com/m/v2/utils"
)

func RemoveFileOrDirectory(path string) error {
	instance := auth.GetInstance()

	if instance.User == nil {
		return fmt.Errorf("error al eliminar archivo o directorio: no hay un usuario loggeado")
	}

	id := instance.ID

	partition, partitionPath, err := memory.GetInstance().GetMountedPartition(id)
	if err != nil {
		return fmt.Errorf("error al obtener la partición: %v", err)
	}

	parentDirs, destFile := utils.GetParentDirectories(path)

	superBlock := ext2.SuperBlock{}
	err = superBlock.DeserializeSuperBlock(partitionPath, partition.Partition.Part_start)
	if err != nil {
		return fmt.Errorf("error al leer el superbloque: %v", err)
	}

	uidInt, _ := strconv.ParseInt(instance.User.UID, 10, 32)
	gidInt, _ := strconv.ParseInt(instance.GID, 10, 32)

	return superBlock.RemoveFileOrDirectory(
		partitionPath,
		parentDirs,
		destFile,
		int32(uidInt),
		int32(gidInt),
	)
}
