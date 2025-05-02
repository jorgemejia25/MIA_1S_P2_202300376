package ext2

import (
	"fmt"
	"os"
	"strings"
)

func (sb *SuperBlock) RemoveFileOrDirectory(path string, parentDirs []string, targetName string, uid int32, gid int32) error {
	// Buscar el inodo del elemento a eliminar
	targetInodeIndex, err := sb.FindFileInode(path, parentDirs, targetName)
	if err != nil {
		return fmt.Errorf("elemento no encontrado: %v", err)
	}

	// Obtener el inodo del elemento
	targetInode := &INode{}
	err = targetInode.Deserialize(path, int64(sb.SInodeStart+(targetInodeIndex*sb.SInodeS)))
	if err != nil {
		return err
	}

	// Verificar permisos del usuario
	if !sb.userHasWritePermission(targetInode, uid, gid) {
		return fmt.Errorf("permisos insuficientes para eliminar el elemento")
	}
	// Obtener inodo del directorio padre
	parentInodeIndex, err := sb.FindFileInode(path, parentDirs[:len(parentDirs)-1], parentDirs[len(parentDirs)-1])
	if err != nil {
		return fmt.Errorf("error al encontrar directorio padre: %v", err)
	}

	// Eliminar la entrada del directorio padre
	if err := sb.removeFromParentDirectory(path, parentInodeIndex, targetName); err != nil {
		return err
	}

	// Eliminar recursivamente si es directorio
	if targetInode.IType[0] == '0' {
		if err := sb.deleteDirectoryContents(path, targetInodeIndex, uid, gid); err != nil {
			return fmt.Errorf("error al eliminar contenido del directorio: %v", err)
		}
	}

	// Liberar inodo y bloques
	if err := sb.freeInodeAndBlocks(path, targetInodeIndex, targetInode); err != nil {
		return fmt.Errorf("error al liberar recursos: %v", err)
	}

	return nil
}

func (sb *SuperBlock) removeFromParentDirectory(path string, parentInodeIndex int32, targetName string) error {
	parentInode := &INode{}
	if err := parentInode.Deserialize(path, int64(sb.SInodeStart+(parentInodeIndex*sb.SInodeS))); err != nil {
		return err
	}

	// Buscar en todos los bloques del directorio padre
	for i := 0; i < 12; i++ {
		blockIndex := parentInode.IBlock[i]
		if blockIndex == -1 {
			continue
		}

		dirBlock := &DirBlock{}
		if err := dirBlock.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS))); err != nil {
			return err
		}

		for j, entry := range dirBlock.BContent {
			entryName := strings.Trim(string(entry.BName[:]), "\x00")
			if entryName == targetName {
				// Marcar la entrada como libre
				dirBlock.BContent[j].BInodo = -1
				copy(dirBlock.BContent[j].BName[:], []byte{'-'})

				if err := dirBlock.Serialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS))); err != nil {
					return err
				}
				return nil
			}
		}
	}
	return fmt.Errorf("entrada no encontrada en directorio padre")
}

func (sb *SuperBlock) deleteDirectoryContents(path string, dirInodeIndex int32, uid int32, gid int32) error {
	dirInode := &INode{}
	if err := dirInode.Deserialize(path, int64(sb.SInodeStart+(dirInodeIndex*sb.SInodeS))); err != nil {
		return err
	}

	// Primera fase: Verificar permisos de todo el contenido
	if err := sb.verifyDirectoryDeletion(path, dirInodeIndex, uid, gid); err != nil {
		return err
	}

	// Segunda fase: Eliminar todo el contenido si la verificación fue exitosa
	return sb.forceDeleteDirectoryContents(path, dirInodeIndex, uid, gid)
}

func (sb *SuperBlock) verifyDirectoryDeletion(path string, dirInodeIndex int32, uid int32, gid int32) error {
	dirInode := &INode{}
	if err := dirInode.Deserialize(path, int64(sb.SInodeStart+(dirInodeIndex*sb.SInodeS))); err != nil {
		return err
	}

	for i := 0; i < 12; i++ {
		blockIndex := dirInode.IBlock[i]
		if blockIndex == -1 {
			continue
		}

		dirBlock := &DirBlock{}
		if err := dirBlock.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS))); err != nil {
			return err
		}

		for _, entry := range dirBlock.BContent {
			if entry.BInodo == -1 || string(entry.BName[:2]) == ".." || string(entry.BName[:1]) == "." {
				continue
			}

			entryName := strings.Trim(string(entry.BName[:]), "\x00")
			targetInode := &INode{}
			if err := targetInode.Deserialize(path, int64(sb.SInodeStart+(entry.BInodo*sb.SInodeS))); err != nil {
				return err
			}

			// Verificar permisos recursivamente
			if !sb.userHasWritePermission(targetInode, uid, gid) {
				return fmt.Errorf("permisos denegados para eliminar '%s'", entryName)
			}

			// Si es directorio, verificar su contenido recursivamente
			if targetInode.IType[0] == '0' {
				if err := sb.verifyDirectoryDeletion(path, entry.BInodo, uid, gid); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (sb *SuperBlock) forceDeleteDirectoryContents(path string, dirInodeIndex int32, uid int32, gid int32) error {
	dirInode := &INode{}
	if err := dirInode.Deserialize(path, int64(sb.SInodeStart+(dirInodeIndex*sb.SInodeS))); err != nil {
		return err
	}

	for i := 0; i < 12; i++ {
		blockIndex := dirInode.IBlock[i]
		if blockIndex == -1 {
			continue
		}

		dirBlock := &DirBlock{}
		if err := dirBlock.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS))); err != nil {
			return err
		}

		for _, entry := range dirBlock.BContent {
			if entry.BInodo == -1 || string(entry.BName[:2]) == ".." || string(entry.BName[:1]) == "." {
				continue
			}

			// Obtener inodo directamente desde la entrada del directorio
			targetInodeIndex := entry.BInodo

			// Eliminar usando el inodo directamente
			if err := sb.deleteByInode(path, targetInodeIndex, uid, gid); err != nil {
				return err
			}
		}
	}
	return nil
}

// Nueva función para eliminar por inodo
func (sb *SuperBlock) deleteByInode(path string, inodeIndex int32, uid int32, gid int32) error {
	targetInode := &INode{}
	if err := targetInode.Deserialize(path, int64(sb.SInodeStart+(inodeIndex*sb.SInodeS))); err != nil {
		return err
	}

	// Verificar permisos
	if !sb.userHasWritePermission(targetInode, uid, gid) {
		return fmt.Errorf("permisos denegados")
	}

	// Eliminar recursivamente si es directorio
	if targetInode.IType[0] == '0' {
		if err := sb.deleteDirectoryContents(path, inodeIndex, uid, gid); err != nil {
			return err
		}
	}

	// Liberar recursos
	return sb.freeInodeAndBlocks(path, inodeIndex, targetInode)
}

func (sb *SuperBlock) freeInodeAndBlocks(path string, inodeIndex int32, inode *INode) error {
	// Marcar inodo como libre
	file, err := os.OpenFile(path, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// Actualizar bitmap de inodos
	_, err = file.Seek(int64(sb.SBmInodeStart+inodeIndex), 0)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte{0})
	if err != nil {
		return err
	}

	// Actualizar contadores
	sb.SFreeInodesCount++
	sb.SInodesCount--

	// Liberar bloques de datos
	for _, blockIndex := range inode.IBlock {
		if blockIndex == -1 {
			continue
		}

		// Marcar bloque como libre
		_, err = file.Seek(int64(sb.SBmBlockStart+blockIndex), 0)
		if err != nil {
			return err
		}
		_, err = file.Write([]byte{0})
		if err != nil {
			return err
		}

		sb.SFreeBlocksCount++
		sb.SBlocksCount--
	}

	return nil
}
