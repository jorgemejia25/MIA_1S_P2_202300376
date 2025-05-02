package ext2

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Move realiza el movimiento de un archivo o directorio de una ubicación a otra
// Esto se implementa como una operación de copia seguida de eliminación del origen
func (sb *SuperBlock) Move(
	path string,
	sourceParentDirs []string,
	sourceName string,
	destParentDirs []string,
	destName string,
	uid int32,
	gid int32,
) error {
	// Buscar el inodo del origen
	sourceInodeIndex, err := sb.FindFileInode(path, sourceParentDirs, sourceName)
	if err != nil {
		return fmt.Errorf("error al encontrar el origen '%s': %v", sourceName, err)
	}

	// Leer el inodo de origen
	sourceInode := &INode{}
	err = sourceInode.Deserialize(path, int64(sb.SInodeStart+(sourceInodeIndex*sb.SInodeS)))
	if err != nil {
		return fmt.Errorf("error al leer inodo de origen: %v", err)
	}

	// Verificar permisos de lectura en el origen
	if !sb.userHasReadPermission(sourceInode, uid, gid) {
		return fmt.Errorf("error: no tienes permisos de lectura en el origen")
	}

	// Verificar permisos de escritura en el directorio padre del origen (necesario para eliminar)
	sourceParentInodeIndex := int32(0) // La raíz por defecto
	if len(sourceParentDirs) > 0 {
		// Si el origen no está en la raíz, buscar el directorio padre
		lastParentIndex := len(sourceParentDirs) - 1
		var parentDirs []string
		if lastParentIndex > 0 {
			parentDirs = sourceParentDirs[:lastParentIndex]
		} else {
			parentDirs = []string{}
		}

		var err error
		sourceParentInodeIndex, err = sb.FindFileInode(path, parentDirs, sourceParentDirs[lastParentIndex])
		if err != nil {
			return fmt.Errorf("error al encontrar directorio padre del origen: %v", err)
		}
	}

	sourceParentInode := &INode{}
	err = sourceParentInode.Deserialize(path, int64(sb.SInodeStart+(sourceParentInodeIndex*sb.SInodeS)))
	if err != nil {
		return fmt.Errorf("error al leer inodo del directorio padre origen: %v", err)
	}

	if !sb.userHasWritePermission(sourceParentInode, uid, gid) {
		return fmt.Errorf("error: no tienes permisos para eliminar el origen")
	}

	// Si no se proporciona un nombre destino específico, usar el nombre original
	if destName == "" {
		destName = sourceName
	}

	// Verificar si el último directorio en destParentDirs es realmente un directorio
	destParentInodeIndex, err := sb.FindFileInode(path, destParentDirs[:len(destParentDirs)-1], destParentDirs[len(destParentDirs)-1])
	if err != nil {
		return fmt.Errorf("error al encontrar directorio destino: %v", err)
	}

	// Leer el inodo del directorio destino
	destParentInode := &INode{}
	err = destParentInode.Deserialize(path, int64(sb.SInodeStart+(destParentInodeIndex*sb.SInodeS)))
	if err != nil {
		return fmt.Errorf("error al leer inodo de destino: %v", err)
	}

	// Verificar que el destino sea un directorio
	if destParentInode.IType[0] != '0' {
		return fmt.Errorf("error: el destino no es un directorio")
	}

	// Verificar permisos de escritura en el destino
	if !sb.userHasWritePermission(destParentInode, uid, gid) {
		return fmt.Errorf("error: no tienes permisos de escritura en el destino")
	}

	// Verificar si ya existe un elemento con el mismo nombre en el destino
	exists, err := sb.fileExistsInDirectory(path, destParentInodeIndex, destName)
	if err != nil {
		return fmt.Errorf("error al verificar si existe '%s' en el destino: %v", destName, err)
	}

	// Si ya existe un elemento con el mismo nombre en el destino
	if exists {
		destElementInodeIndex, err := sb.findInodeInDirectory(path, destParentInodeIndex, destName)
		if err != nil {
			return fmt.Errorf("error al obtener inodo del elemento destino: %v", err)
		}

		destElementInode := &INode{}
		err = destElementInode.Deserialize(path, int64(sb.SInodeStart+(destElementInodeIndex*sb.SInodeS)))
		if err != nil {
			return fmt.Errorf("error al leer inodo del elemento destino: %v", err)
		}

		// Si el elemento destino no es un directorio, es un error
		if destElementInode.IType[0] != '0' {
			return fmt.Errorf("ya existe un archivo con el nombre '%s' en el directorio destino", destName)
		}

		// Verificar permisos de escritura en el directorio destino
		if !sb.userHasWritePermission(destElementInode, uid, gid) {
			return fmt.Errorf("error: no tienes permisos de escritura en el directorio destino '%s'", destName)
		}

		// Ahora vamos a mover (copiar+eliminar) el contenido del elemento origen dentro del directorio destino existente
		switch sourceInode.IType[0] {
		case '0': // Directorio
			// Copiar el contenido recursivamente al directorio destino existente
			err = sb.copyDirectoryContents(path, sourceInodeIndex, destElementInodeIndex, uid, gid)
			if err != nil {
				return fmt.Errorf("error al copiar contenido del directorio: %v", err)
			}

			// Eliminar el directorio original y su contenido
			err = sb.removeDirectoryContents(path, sourceInodeIndex)
			if err != nil {
				return fmt.Errorf("error al eliminar directorio origen después de copiar: %v", err)
			}

			// Eliminar la entrada del directorio en el padre
			err = sb.removeDirectoryEntry(path, sourceParentInodeIndex, sourceName)
			if err != nil {
				return fmt.Errorf("error al eliminar entrada del directorio en el padre: %v", err)
			}

			fmt.Printf("Contenido del directorio '%s' movido exitosamente a '%s'\n", sourceName, destName)
			return nil

		case '1': // Archivo - este caso no debería ocurrir si el destino ya es un directorio
			return fmt.Errorf("ya existe un directorio con el nombre '%s' en el destino", destName)
		default:
			return fmt.Errorf("tipo de inodo no reconocido")
		}
	}

	// Basado en el tipo de inodo, mover archivo o directorio

	switch sourceInode.IType[0] {
	case '0': // Directorio
		// Copiar el directorio
		err = sb.copyDirectory(path, sourceInodeIndex, sourceInode, destParentInodeIndex, destName, uid, gid)
		if err != nil {
			return fmt.Errorf("error al copiar directorio: %v", err)
		}

		// Obtener el inodo del nuevo directorio creado
		if err != nil {
			return fmt.Errorf("error al encontrar directorio copiado: %v", err)
		}

	case '1': // Archivo
		// Copiar el archivo
		err = sb.copyFile(path, sourceInodeIndex, sourceInode, destParentInodeIndex, destName, uid, gid)
		if err != nil {
			return fmt.Errorf("error al copiar archivo: %v", err)
		}

		// Obtener el inodo del nuevo archivo creado
		if err != nil {
			return fmt.Errorf("error al encontrar archivo copiado: %v", err)
		}

	default:
		return fmt.Errorf("tipo de inodo no reconocido")
	}

	// Actualizar la fecha de modificación del directorio destino
	destParentInode.IMtime = float32(time.Now().Unix())
	err = destParentInode.Serialize(path, int64(sb.SInodeStart+(destParentInodeIndex*sb.SInodeS)))
	if err != nil {
		return fmt.Errorf("error al actualizar tiempo de modificación del destino: %v", err)
	}

	// Ahora eliminar el archivo o directorio original
	err = sb.removeFileOrDirectory(path, sourceParentInodeIndex, sourceInodeIndex, sourceName)
	if err != nil {
		return fmt.Errorf("error al eliminar elemento origen después de copiarlo: %v", err)
	}

	fmt.Printf("'%s' movido exitosamente a '%s'\n", sourceName, destName)
	return nil
}

// removeFileOrDirectory elimina un archivo o directorio según su tipo
func (sb *SuperBlock) removeFileOrDirectory(path string, parentInodeIndex, inodeIndex int32, name string) error {
	// Leer el inodo a eliminar
	inode := &INode{}
	err := inode.Deserialize(path, int64(sb.SInodeStart+(inodeIndex*sb.SInodeS)))
	if err != nil {
		return err
	}

	// Eliminar según el tipo
	switch inode.IType[0] {
	case '0': // Directorio
		// Eliminar contenido del directorio
		err = sb.removeDirectoryContents(path, inodeIndex)
		if err != nil {
			return err
		}
	case '1': // Archivo
		// Liberar bloques del archivo
		err = sb.freeFileBlocks(path, inode)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("tipo de inodo no reconocido")
	}

	// Liberar el inodo
	err = sb.freeInode(path, inodeIndex)
	if err != nil {
		return err
	}

	// Eliminar la entrada del directorio en el directorio padre
	return sb.removeDirectoryEntry(path, parentInodeIndex, name)
}

// freeFileBlocks libera los bloques ocupados por un archivo
func (sb *SuperBlock) freeFileBlocks(path string, inode *INode) error {
	// Liberar bloques directos
	for i := 0; i < 12; i++ {
		if inode.IBlock[i] != -1 {
			err := sb.freeBlock(path, inode.IBlock[i])
			if err != nil {
				return err
			}
		}
	}

	// Liberar bloques indirectos simples
	if inode.IBlock[12] != -1 {
		pointerBlock := &PointerBlock{}
		err := pointerBlock.Deserialize(path, int64(sb.SBlockStart+(inode.IBlock[12]*sb.SBlockS)))
		if err != nil {
			return err
		}

		for _, ptr := range pointerBlock.PContent {
			if ptr != -1 {
				err := sb.freeBlock(path, ptr)
				if err != nil {
					return err
				}
			}
		}

		// Liberar el bloque de punteros
		err = sb.freeBlock(path, inode.IBlock[12])
		if err != nil {
			return err
		}
	}

	// Aquí se podrían agregar bloques indirectos dobles y triples si fuera necesario

	return nil
}

// freeBlock marca un bloque como libre en el bitmap de bloques
func (sb *SuperBlock) freeBlock(path string, blockIndex int32) error {
	bitmapOffset := int64(sb.SBmBlockStart + blockIndex)

	file, err := sb.openPartition(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Marcar el bloque como libre (0)
	_, err = file.Seek(bitmapOffset, 0)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte{0})
	if err != nil {
		return err
	}

	sb.SFreeBlocksCount++
	return nil
}

// freeInode marca un inodo como libre en el bitmap de inodos
func (sb *SuperBlock) freeInode(path string, inodeIndex int32) error {
	bitmapOffset := int64(sb.SBmInodeStart + inodeIndex)

	file, err := sb.openPartition(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Marcar el inodo como libre (0)
	_, err = file.Seek(bitmapOffset, 0)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte{0})
	if err != nil {
		return err
	}

	sb.SFreeInodesCount++
	return nil
}

// removeDirectoryContents elimina recursivamente el contenido de un directorio
func (sb *SuperBlock) removeDirectoryContents(path string, dirInodeIndex int32) error {
	// Obtener el inodo del directorio
	dirInode := &INode{}
	err := dirInode.Deserialize(path, int64(sb.SInodeStart+(dirInodeIndex*sb.SInodeS)))
	if err != nil {
		return err
	}

	// Buscar todas las entradas en el directorio
	for i := 0; i < 12; i++ {
		if dirInode.IBlock[i] == -1 {
			continue
		}

		dirBlock := &DirBlock{}
		err := dirBlock.Deserialize(path, int64(sb.SBlockStart+(dirInode.IBlock[i]*sb.SBlockS)))
		if err != nil {
			return err
		}

		for j, entry := range dirBlock.BContent {
			if entry.BInodo == -1 || j < 2 {
				// Ignorar entradas vacías o las entradas "." y ".."
				continue
			}

			entryName := strings.Trim(string(entry.BName[:]), "\x00")
			if entryName == "." || entryName == ".." {
				continue
			}

			// Leer el inodo de esta entrada
			entryInode := &INode{}
			err := entryInode.Deserialize(path, int64(sb.SInodeStart+(entry.BInodo*sb.SInodeS)))
			if err != nil {
				return err
			}

			// Eliminar recursivamente según el tipo
			if entryInode.IType[0] == '0' { // Directorio
				// Eliminar contenido del subdirectorio
				err = sb.removeDirectoryContents(path, entry.BInodo)
				if err != nil {
					return err
				}
			} else if entryInode.IType[0] == '1' { // Archivo
				// Liberar bloques del archivo
				err = sb.freeFileBlocks(path, entryInode)
				if err != nil {
					return err
				}
			}

			// Liberar el inodo
			err = sb.freeInode(path, entry.BInodo)
			if err != nil {
				return err
			}

			// Marcar la entrada como libre
			dirBlock.BContent[j].BInodo = -1
			copy(dirBlock.BContent[j].BName[:], "-")
		}

		// Actualizar el bloque de directorio
		err = dirBlock.Serialize(path, int64(sb.SBlockStart+(dirInode.IBlock[i]*sb.SBlockS)))
		if err != nil {
			return err
		}

		// Liberar este bloque
		err = sb.freeBlock(path, dirInode.IBlock[i])
		if err != nil {
			return err
		}

		// Marcar el bloque como no usado en el inodo
		dirInode.IBlock[i] = -1
	}

	// También habría que procesar punteros indirectos si los hubiera

	// Actualizar el inodo del directorio
	return dirInode.Serialize(path, int64(sb.SInodeStart+(dirInodeIndex*sb.SInodeS)))
}

// removeDirectoryEntry elimina una entrada de un directorio
func (sb *SuperBlock) removeDirectoryEntry(path string, dirInodeIndex int32, entryName string) error {
	// Obtener el inodo del directorio
	dirInode := &INode{}
	err := dirInode.Deserialize(path, int64(sb.SInodeStart+(dirInodeIndex*sb.SInodeS)))
	if err != nil {
		return err
	}

	// Buscar la entrada en los bloques del directorio
	for i := 0; i < 12; i++ {
		if dirInode.IBlock[i] == -1 {
			continue
		}

		dirBlock := &DirBlock{}
		err := dirBlock.Deserialize(path, int64(sb.SBlockStart+(dirInode.IBlock[i]*sb.SBlockS)))
		if err != nil {
			return err
		}

		// Buscar la entrada por nombre
		for j, entry := range dirBlock.BContent {
			if entry.BInodo == -1 {
				continue
			}

			currentName := strings.Trim(string(entry.BName[:]), "\x00")
			if strings.EqualFold(currentName, entryName) {
				// Marcar la entrada como libre
				dirBlock.BContent[j].BInodo = -1
				copy(dirBlock.BContent[j].BName[:], "-")

				// Actualizar el bloque de directorio
				err = dirBlock.Serialize(path, int64(sb.SBlockStart+(dirInode.IBlock[i]*sb.SBlockS)))
				if err != nil {
					return err
				}

				// Actualizar tiempo de modificación del directorio
				dirInode.IMtime = float32(time.Now().Unix())
				err = dirInode.Serialize(path, int64(sb.SInodeStart+(dirInodeIndex*sb.SInodeS)))
				if err != nil {
					return err
				}

				return nil
			}
		}
	}

	// No se encontró la entrada
	return fmt.Errorf("no se encontró la entrada '%s' en el directorio", entryName)
}

// GetInodeByNumber obtiene un inodo por su número
func (sb *SuperBlock) GetInodeByNumber(diskPath string, inodeIndex int32) (*INode, error) {
	inode := &INode{}
	err := inode.Deserialize(diskPath, int64(sb.SInodeStart+(inodeIndex*sb.SInodeS)))
	return inode, err
}

// openPartition abre el archivo de la partición para escritura
func (sb *SuperBlock) openPartition(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR, 0666)
}
