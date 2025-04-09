package partition_operations

import (
	"fmt"
	"strconv"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/operations/auth"
	"disk.simulator.com/m/v2/internal/disk/types/structures/ext2"
	"disk.simulator.com/m/v2/utils"
)

func CreateDirectory(dirPath string, p bool) error {

	instance := auth.GetInstance()

	if instance.User == nil {
		return fmt.Errorf("error al crear directorio: no hay un usuario loggeado")
	}

	id := instance.ID

	// Aquí iría la lógica para crear el directorio
	partition, partitionPath, err := memory.GetInstance().GetMountedPartition(id)

	if err != nil {
		fmt.Println(id)
		return err
	}

	parentDirs, destDir := utils.GetParentDirectories(dirPath)

	superBlock := ext2.SuperBlock{}
	superBlock.DeserializeSuperBlock(partition.Path, partition.Partition.Part_start)

	// Convertir uid y gid de string a int32
	uidInt, _ := strconv.ParseInt(instance.User.UID, 10, 32)
	gidInt, _ := strconv.ParseInt(instance.GID, 10, 32)

	err = superBlock.CreateFolder(partitionPath, parentDirs, destDir, p, int32(uidInt), int32(gidInt))

	if err != nil {
		return err
	}
	// Serializar el superbloque
	err = superBlock.SerializeSuperBlock(partition.Path, partition.Partition.Part_start)

	if err != nil {
		return err
	}

	fmt.Printf("Directory %s created\n", dirPath)

	return nil
}
