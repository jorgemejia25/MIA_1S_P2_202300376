package auth

import (
	"fmt"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/types/structures/ext2"
	"disk.simulator.com/m/v2/utils"
)

func Login(user, password, id string) error {
	partition, partitionPath, err := memory.GetInstance().GetMountedPartition(id)
	if err != nil {
		return err
	}
	sb := &ext2.SuperBlock{}
	sb.DeserializeSuperBlock(partition.Path, partition.Partition.Part_start)
	content, err := sb.ReadFile(partitionPath, []string{}, "users.txt")
	if err != nil {
		return err
	}

	userData, _ := utils.FindUserInFile(content, user)
	groupData, _ := utils.FindGroupInFile(content, userData.Group)
	fmt.Println("Usuario encontrado:", userData)

	if userData == nil {
		return fmt.Errorf("error al iniciar sesión: usuario %s no encontrado", user)
	}

	if userData.Password != password {
		return fmt.Errorf("error al iniciar sesión: contraseña incorrecta")
	}

	loggedUser := GetInstance()

	// check if user is already logged in
	if loggedUser.User != nil {
		return fmt.Errorf("error al iniciar sesión: ya hay un usuario loggeado")
	}

	loggedUser.SetLoggedUser(id, groupData.GID, userData)

	fmt.Println("Sesión iniciada correctamente en el grupo", userData.Group)

	return nil
}
