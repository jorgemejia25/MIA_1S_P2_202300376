package partition_operations

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/operations/auth"
	ext2 "disk.simulator.com/m/v2/internal/disk/types/structures/ext"
	"disk.simulator.com/m/v2/utils"
)

// ChangeOwner cambia el propietario de un archivo o directorio especificado por path
func ChangeOwner(path string, usuario string, r bool) error {
	instance := auth.GetInstance()

	if instance.User == nil {
		return fmt.Errorf("error: no hay un usuario loggeado")
	}

	id := instance.ID

	// Obtener la partición montada
	partition, partitionPath, err := memory.GetInstance().GetMountedPartition(id)
	if err != nil {
		return fmt.Errorf("error al obtener la partición: %v", err)
	}

	// Leer el superbloque de la partición
	superBlock := ext2.SuperBlock{}
	err = superBlock.DeserializeSuperBlock(partitionPath, partition.Partition.Part_start)
	if err != nil {
		return fmt.Errorf("error al leer el superbloque: %v", err)
	}

	// Extraer directorios padre y nombre del archivo/directorio
	parentDirs, targetName := utils.GetParentDirectories(path)

	// Verificar que el usuario existe en el sistema
	content, err := superBlock.ReadFile(partitionPath, []string{}, "users.txt")
	if err != nil {
		return fmt.Errorf("error al leer users.txt: %v", err)
	}

	userFound, _ := utils.FindUserInFile(content, usuario)
	if userFound == nil {
		return fmt.Errorf("error: el usuario '%s' no existe", usuario)
	}

	// Obtener el UID del usuario a asignar como propietario
	newUID, err := strconv.ParseInt(userFound.UID, 10, 32)
	if err != nil {
		return fmt.Errorf("error al convertir UID a entero: %v", err)
	}

	// Obtener el UID y GID del usuario loggeado
	currentUIDInt, _ := strconv.ParseInt(instance.User.UID, 10, 32)
	currentGIDInt, _ := strconv.ParseInt(instance.GID, 10, 32)

	// Verificar que el archivo existe
	targetInodeIndex, err := superBlock.FindFileInode(partitionPath, parentDirs, targetName)
	if err != nil {
		return fmt.Errorf("error: el archivo o carpeta '%s' no existe: %v", path, err)
	}

	// Obtener el inodo del archivo o directorio
	targetInode := &ext2.INode{}
	err = targetInode.Deserialize(partitionPath, int64(superBlock.SInodeStart+(targetInodeIndex*superBlock.SInodeS)))
	if err != nil {
		return fmt.Errorf("error al leer el inodo: %v", err)
	}

	// Verificar permisos: solo el root o el propietario pueden cambiar el dueño
	if instance.User.Group != "root" && targetInode.IUid != int32(currentUIDInt) {
		return fmt.Errorf("error: no tienes permisos para cambiar el propietario de este archivo o carpeta")
	}

	// Cambiar el propietario del archivo/directorio
	targetInode.IUid = int32(newUID)
	targetInode.IMtime = float32(time.Now().Unix())

	// Escribir el inodo actualizado
	err = targetInode.Serialize(partitionPath, int64(superBlock.SInodeStart+(targetInodeIndex*superBlock.SInodeS)))
	if err != nil {
		return fmt.Errorf("error al actualizar el inodo: %v", err)
	}

	// Si es un directorio y se especificó la opción recursiva, cambiar propietario recursivamente
	if targetInode.IType[0] == '0' && r {
		// Implementar cambio recursivo para todos los archivos y carpetas dentro del directorio
		err = changeOwnerRecursive(superBlock, partitionPath, parentDirs, targetName, int32(newUID), int32(currentUIDInt), int32(currentGIDInt))
		if err != nil {
			return fmt.Errorf("error al cambiar propietario recursivamente: %v", err)
		}
	}

	// Actualizar el superbloque para guardar cambios
	err = superBlock.SerializeSuperBlock(partitionPath, partition.Partition.Part_start)
	if err != nil {
		return fmt.Errorf("error al actualizar el superbloque: %v", err)
	}

	fmt.Printf("Se ha cambiado el propietario de '%s' al usuario '%s'\n", path, usuario)
	return nil
}

// changeOwnerRecursive cambia el propietario de todos los archivos y directorios dentro de un directorio recursivamente
func changeOwnerRecursive(
	superBlock ext2.SuperBlock,
	partitionPath string,
	parentDirs []string,
	dirName string,
	newUID int32,
	currentUID int32,
	currentGID int32,
) error {
	// Ruta completa del directorio
	fullPath := append(parentDirs, dirName)

	// Obtener el inodo del directorio
	dirInodeIndex, err := superBlock.FindFileInode(partitionPath, parentDirs, dirName)
	if err != nil {
		return fmt.Errorf("error al buscar directorio: %v", err)
	}

	// Leer el inodo del directorio
	dirInode := &ext2.INode{}
	err = dirInode.Deserialize(partitionPath, int64(superBlock.SInodeStart+(dirInodeIndex*superBlock.SInodeS)))
	if err != nil {
		return fmt.Errorf("error al leer inodo del directorio: %v", err)
	}

	// Recorrer todos los bloques del directorio
	for _, blockIndex := range dirInode.IBlock {
		if blockIndex == -1 {
			continue
		}

		// Leer el bloque de directorio
		dirBlock := &ext2.DirBlock{}
		err = dirBlock.Deserialize(partitionPath, int64(superBlock.SBlockStart+(blockIndex*superBlock.SBlockS)))
		if err != nil {
			return fmt.Errorf("error al leer bloque de directorio: %v", err)
		}

		// Procesar cada entrada en el bloque
		for _, entry := range dirBlock.BContent {
			if entry.BInodo == -1 {
				continue
			}

			entryName := strings.Trim(string(entry.BName[:]), "\x00")
			// Ignorar entradas "." y ".."
			if entryName == "." || entryName == ".." {
				continue
			}

			// Obtener el inodo de la entrada
			entryInode := &ext2.INode{}
			err := entryInode.Deserialize(partitionPath, int64(superBlock.SInodeStart+(entry.BInodo*superBlock.SInodeS)))
			if err != nil {
				return fmt.Errorf("error al leer inodo de entrada: %v", err)
			}

			// Cambiar el propietario si es el propietario actual o es root
			if currentUID == 1 || entryInode.IUid == currentUID {
				entryInode.IUid = newUID
				entryInode.IMtime = float32(time.Now().Unix())

				// Guardar el inodo modificado
				err = entryInode.Serialize(partitionPath, int64(superBlock.SInodeStart+(entry.BInodo*superBlock.SInodeS)))
				if err != nil {
					return fmt.Errorf("error al actualizar inodo de '%s': %v", entryName, err)
				}
			}

			// Si es un directorio, procesar recursivamente
			if entryInode.IType[0] == '0' {
				err = changeOwnerRecursive(superBlock, partitionPath, fullPath, entryName, newUID, currentUID, currentGID)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
