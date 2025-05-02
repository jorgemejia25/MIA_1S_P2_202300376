package auth

import (
	"fmt"
	"strconv"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/types/structures/ext"
	"disk.simulator.com/m/v2/utils"
)

func CreateUser(
	username string,
	password string,
	group string,
) error {
	userData := GetInstance()

	if userData.User == nil {
		return fmt.Errorf("error al crear usuario no hay usuario loggeado")
	}

	if userData.User.Group != "root" {
		return fmt.Errorf("error al crear usuario: no tienes permisos para realizar esta acción")
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

	userExists, _ := utils.FindUserInFile(content, username)

	if userExists != nil {
		return fmt.Errorf("error al crear usuario: el usuario %s ya existe", username)
	}

	groupExists, _ := utils.FindGroupInFile(content, group)

	if groupExists == nil {
		return fmt.Errorf("error al crear usuario: el grupo %s no existe", group)
	}

	lastUser, _ := utils.FindLastUserInFile(content)

	uid, err := strconv.Atoi(lastUser.UID)

	if err != nil {
		return fmt.Errorf("error al convertir UID a entero: %v", err)
	}

	content += fmt.Sprintf("%d,U,%s,%s,%s\n", uid+1, group, username, password)

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
