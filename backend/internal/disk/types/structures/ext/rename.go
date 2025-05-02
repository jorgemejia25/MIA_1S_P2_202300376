package ext2

import (
	"fmt"
	"strings"
	"time"
)

func (sb *SuperBlock) Rename(partitionPath string, parentDirs []string, oldName string, newName string, uid int32, gid int32) error {
	// Buscar el inodo del elemento a renombrar
	targetInodeIndex, err := sb.FindFileInode(partitionPath, parentDirs, oldName)
	if err != nil {
		return fmt.Errorf("elemento no encontrado: %v", err)
	}

	// Verificar permisos sobre el elemento
	targetInode := &INode{}
	if err := targetInode.Deserialize(partitionPath, int64(sb.SInodeStart+(targetInodeIndex*sb.SInodeS))); err != nil {
		return err
	}

	if !sb.userHasWritePermission(targetInode, uid, gid) {
		return fmt.Errorf("permisos insuficientes para renombrar el elemento")
	}

	// Buscar el directorio padre
	parentInodeIndex, err := sb.FindFileInode(partitionPath, parentDirs[:len(parentDirs)-1], parentDirs[len(parentDirs)-1])
	if err != nil {
		return fmt.Errorf("error al encontrar directorio padre: %v", err)
	}

	// Verificar que el nuevo nombre no exista
	if exists, _ := sb.fileExistsInDirectory(partitionPath, parentInodeIndex, newName); exists {
		return fmt.Errorf("ya existe un elemento con el nombre '%s'", newName)
	}

	// Actualizar la entrada en el directorio padre
	if err := sb.updateDirectoryEntry(partitionPath, parentInodeIndex, oldName, newName); err != nil {
		return err
	}

	// Actualizar tiempo de modificación del directorio padre
	parentInode := &INode{}
	if err := parentInode.Deserialize(partitionPath, int64(sb.SInodeStart+(parentInodeIndex*sb.SInodeS))); err != nil {
		return err
	}

	parentInode.IMtime = float32(time.Now().Unix())
	return parentInode.Serialize(partitionPath, int64(sb.SInodeStart+(parentInodeIndex*sb.SInodeS)))
}


func (sb *SuperBlock) updateDirectoryEntry(partitionPath string, parentInodeIndex int32, oldName string, newName string) error {
	parentInode := &INode{}
	if err := parentInode.Deserialize(partitionPath, int64(sb.SInodeStart+(parentInodeIndex*sb.SInodeS))); err != nil {
		return err
	}

	var targetBlockIndex int32 = -1
	var targetEntryIndex int = -1
	found := false

	// Buscar en todos los bloques
	for i := 0; i < 12 && !found; i++ {
		blockIndex := parentInode.IBlock[i]
		if blockIndex == -1 {
			continue
		}

		dirBlock := &DirBlock{}
		if err := dirBlock.Deserialize(partitionPath, int64(sb.SBlockStart+(blockIndex*sb.SBlockS))); err != nil {
			return err
		}

		for j, entry := range dirBlock.BContent {
			entryName := strings.Trim(string(entry.BName[:]), "\x00")
			if entryName == oldName {
				targetBlockIndex = blockIndex
				targetEntryIndex = j
				found = true
				break
			}
		}
	}

	if !found {
		return fmt.Errorf("entrada no encontrada en directorio padre")
	}

	// Actualizar bloque específico
	dirBlock := &DirBlock{}
	if err := dirBlock.Deserialize(partitionPath, int64(sb.SBlockStart+(targetBlockIndex*sb.SBlockS))); err != nil {
		return err
	}

	// Copiar nuevo nombre truncando a 12 bytes
	var newNameBytes [12]byte
	copy(newNameBytes[:], []byte(newName))
	dirBlock.BContent[targetEntryIndex].BName = newNameBytes

	// Escribir bloque actualizado
	if err := dirBlock.Serialize(partitionPath, int64(sb.SBlockStart+(targetBlockIndex*sb.SBlockS))); err != nil {
		return err
	}

	// Actualizar journal
	return AddJournal(partitionPath, int64(sb.SBlockStart+(targetBlockIndex*sb.SBlockS)), sb.SInodesCount,
		"rename",
		oldName,
		newName,
	)
}
