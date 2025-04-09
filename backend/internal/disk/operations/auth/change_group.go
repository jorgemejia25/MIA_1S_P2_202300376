package auth

import (
	"fmt"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/types/structures/ext2"
	"disk.simulator.com/m/v2/utils"
)

func ChangeGroup(username string, groupname string) error {
	userData := GetInstance()

	if userData.User == nil {
		return fmt.Errorf("error al cambiar grupo: no hay un usuario loggeado")
	}

	if userData.User.Group != "root" {
		return fmt.Errorf("error al cambiar grupo: no tienes permisos para realizar esta acción")
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

	// Verificar que el grupo existe
	groupFound, _ := utils.FindGroupInFile(content, groupname)
	if groupFound == nil {
		return fmt.Errorf("error al cambiar grupo: el grupo %s no existe", groupname)
	}

	// Buscar al usuario que se quiere modificar
	userFound, index := utils.FindUserInFile(content, username)
	if userFound == nil {
		return fmt.Errorf("error al cambiar grupo: el usuario %s no existe", username)
	}

	// Cambiar el grupo del usuario
	userFound.Group = groupname
	newLine := fmt.Sprintf("%s,U,%s,%s,%s\n", userFound.UID, userFound.Group, userFound.Username, userFound.Password)
	content = utils.ReplaceLine(content, index, newLine)

	err = superBlock.UpdateFile(partitionPath, []string{}, "users.txt", content)
	if err != nil {
		return err
	}

	// Si el usuario modificado es el usuario actual, actualizar también en memoria
	if username == userData.User.Username {
		userData.User.Group = groupname
	}

	err = superBlock.SerializeSuperBlock(partition.Path, partition.Partition.Part_start)
	if err != nil {
		return fmt.Errorf("error al guardar SuperBlock: %v", err)
	}

	return nil
}
