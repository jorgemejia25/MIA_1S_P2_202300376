package auth

import (
	"fmt"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/types/structures/ext2"
	"disk.simulator.com/m/v2/utils"
)

func RemoveUser(name string) error {
	userData := GetInstance()

	if userData.User == nil {
		return fmt.Errorf("error al eliminar usuario: no hay un usuario loggeado")
	}

	if userData.User.Group != "root" {
		return fmt.Errorf("error al eliminar usuario: no tienes permisos para realizar esta acción")
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

	userFound, index := utils.FindUserInFile(content, name)
	if userFound == nil {
		return fmt.Errorf("error al eliminar usuario: el usuario %s no existe", name)
	}

	userFound.UID = "0" // Marcarlo como eliminado
	newLine := fmt.Sprintf("%s,U,%s,%s,%s\n", userFound.UID, userFound.Group, userFound.Username, userFound.Password)
	content = utils.ReplaceLine(content, index, newLine)

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
