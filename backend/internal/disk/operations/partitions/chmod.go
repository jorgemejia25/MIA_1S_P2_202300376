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

// ChangePermissions cambia los permisos de un archivo o directorio especificado por path
func ChangePermissions(path string, ugo string, r bool) error {
	instance := auth.GetInstance()

	if instance.User == nil {
		return fmt.Errorf("error: no hay un usuario loggeado")
	}

	// Verificar que el usuario sea root
	if instance.User.Group != "root" {
		return fmt.Errorf("error: solo el usuario root puede cambiar permisos")
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

	// Validar el formato de permisos UGO
	if len(ugo) != 3 {
		return fmt.Errorf("error: el formato de permisos debe ser 3 números (UGO)")
	}

	// Convertir cada carácter a un número y validar que esté en el rango [0-7]
	newPerms := [3]byte{}
	for i, c := range ugo {
		num, err := strconv.Atoi(string(c))
		if err != nil || num < 0 || num > 7 {
			return fmt.Errorf("error: los permisos deben ser números del 0 al 7")
		}
		newPerms[i] = byte(c)
	}

	// Obtener el UID del usuario loggeado
	currentUIDInt, _ := strconv.ParseInt(instance.User.UID, 10, 32)

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

	// Cambiar los permisos del archivo/directorio
	targetInode.IPerm = newPerms
	targetInode.IMtime = float32(time.Now().Unix())

	// Escribir el inodo actualizado
	err = targetInode.Serialize(partitionPath, int64(superBlock.SInodeStart+(targetInodeIndex*superBlock.SInodeS)))
	if err != nil {
		return fmt.Errorf("error al actualizar el inodo: %v", err)
	}

	// Si es un directorio y se especificó la opción recursiva, cambiar permisos recursivamente
	if targetInode.IType[0] == '0' && r {
		// Implementar cambio recursivo para todos los archivos y carpetas dentro del directorio
		// que pertenezcan al usuario actual
		err = changePermissionsRecursive(superBlock, partitionPath, parentDirs, targetName, newPerms, int32(currentUIDInt))
		if err != nil {
			return fmt.Errorf("error al cambiar permisos recursivamente: %v", err)
		}
	}

	// Actualizar el superbloque para guardar cambios
	err = superBlock.SerializeSuperBlock(partitionPath, partition.Partition.Part_start)
	if err != nil {
		return fmt.Errorf("error al actualizar el superbloque: %v", err)
	}

	fmt.Printf("Se han cambiado los permisos de '%s' a %s\n", path, ugo)
	return nil
}

// changePermissionsRecursive cambia los permisos de todos los archivos y directorios dentro de un directorio recursivamente
// Solo modifica los archivos que pertenecen al usuario actual
func changePermissionsRecursive(
	superBlock ext2.SuperBlock,
	partitionPath string,
	parentDirs []string,
	dirName string,
	newPerms [3]byte,
	currentUID int32,
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

			// Cambiar los permisos solo si el archivo pertenece al usuario actual (o es root)
			if entryInode.IUid == currentUID || currentUID == 1 {
				entryInode.IPerm = newPerms
				entryInode.IMtime = float32(time.Now().Unix())

				// Guardar el inodo modificado
				err = entryInode.Serialize(partitionPath, int64(superBlock.SInodeStart+(entry.BInodo*superBlock.SInodeS)))
				if err != nil {
					return fmt.Errorf("error al actualizar inodo de '%s': %v", entryName, err)
				}
			}

			// Si es un directorio, procesar recursivamente
			if entryInode.IType[0] == '0' {
				err = changePermissionsRecursive(superBlock, partitionPath, fullPath, entryName, newPerms, currentUID)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
