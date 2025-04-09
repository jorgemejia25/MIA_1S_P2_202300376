package auth

import (
	"fmt"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/types/structures/ext2"
	"disk.simulator.com/m/v2/utils"
)

func RemoveGroup(
	name string,
) error {

	userData := GetInstance()

	if userData.User == nil {
		return fmt.Errorf("error al eliminar grupo: no hay un usuario loggeado")
	}

	if userData.User.Group != "root" {
		return fmt.Errorf("error al eliminar grupo: no tienes permisos para realizar esta acción")
	}

	partition, partitionPath, err := memory.GetInstance().GetMountedPartition(userData.ID)

	if err != nil {
		return err
	}

	superBlock := &ext2.SuperBlock{}
	superBlock.DeserializeSuperBlock(partition.Path, partition.Partition.Part_start)

	content, err := superBlock.ReadFile(partitionPath, []string{}, "users.txt")

	if err != nil {
		return err
	}

	groupExists, index := utils.FindGroupInFile(content, name)

	if groupExists == nil {
		return fmt.Errorf("error al eliminar grupo: el grupo %s no existe", name)
	}

	// change the gid from the group to 0
	groupExists.GID = "0"

	// Change the line in the file
	content = utils.ReplaceLine(content, index, fmt.Sprintf("%s,G,%s\n", groupExists.GID, groupExists.Name))

	err = superBlock.UpdateFile(partitionPath, []string{}, "users.txt", content)

	if err != nil {
		return err
	}

	err = superBlock.SerializeSuperBlock(partition.Path, partition.Partition.Part_start)
	if err != nil {
		return fmt.Errorf("error al guardar SuperBlock: %v", err)
	}

	return nil
}
