package auth

import (
	"fmt"
	"strconv"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/types/structures/ext2"
	"disk.simulator.com/m/v2/utils"
)

func CreateGroup(
	name string,
) error {

	userData := GetInstance()

	if userData.User == nil {
		return fmt.Errorf("error al crear grupo: no hay un usuario loggeado")
	}

	if userData.User.Group != "root" {
		return fmt.Errorf("error al crear grupo: no tienes permisos para realizar esta acción")
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

	groupExists, _ := utils.FindGroupInFile(content, name)

	if groupExists != nil {
		return fmt.Errorf("error al crear grupo: el grupo %s ya existe", name)
	}

	lastGroup := utils.FindLastGroupInFile(content)

	gid, err := strconv.Atoi(lastGroup.GID)

	if err != nil {
		return fmt.Errorf("error al convertir GID a entero: %v", err)
	}

	content += fmt.Sprintf("%d,G,%s\n", gid+1, name)

	err = superBlock.UpdateFile(partitionPath, []string{}, "users.txt", content)

	if err != nil {
		return err
	}

	// Re-serializar SuperBlock
	err = superBlock.SerializeSuperBlock(partition.Path, partition.Partition.Part_start)
	if err != nil {
		return fmt.Errorf("error al guardar SuperBlock: %v", err)
	}

	return nil
}
