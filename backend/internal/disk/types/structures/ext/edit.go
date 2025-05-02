package ext2

import (
	"fmt"
	"math"
	"os"
	"time"
)

func (sb *SuperBlock) EditFile(partitionPath string, parentDirs []string, fileName string, newContent string, uid int32, gid int32) error {
	// Buscar el inodo del archivo
	inodeIndex, err := sb.FindFileInode(partitionPath, parentDirs, fileName)
	if err != nil {
		return fmt.Errorf("error al encontrar archivo: %v", err)
	}

	// Leer el inodo actual
	fileInode := &INode{}
	err = fileInode.Deserialize(partitionPath, int64(sb.SInodeStart+(inodeIndex*sb.SInodeS)))
	if err != nil {
		return err
	}

	// Verificar permisos de lectura y escritura
	if !sb.userHasReadPermission(fileInode, uid, gid) || !sb.userHasWritePermission(fileInode, uid, gid) {
		return fmt.Errorf("permisos insuficientes para editar el archivo")
	}

	// Calcular nuevos requerimientos de bloques
	contentSize := len(newContent)
	blocksNeeded := int(math.Ceil(float64(contentSize) / float64(FileBlockSize)))

	// Liberar bloques existentes
	if err := sb.freeFileBlocks(partitionPath, fileInode); err != nil {
		return fmt.Errorf("error al liberar bloques: %v", err)
	}

	// Asignar nuevos bloques y escribir contenido
	contentOffset := 0
	blocksAssigned := 0

	// Asignar bloques directos (0-11)
	for i := 0; i < 12 && blocksAssigned < blocksNeeded; i++ {
		blockIndex := (sb.SFirstBlo - sb.SBlockStart) / sb.SBlockS
		bytesToCopy := FileBlockSize
		if contentSize-contentOffset < FileBlockSize {
			bytesToCopy = contentSize - contentOffset
		}

		fileBlock := &FileBlock{}
		copy(fileBlock.BContent[:bytesToCopy], newContent[contentOffset:contentOffset+bytesToCopy])

		if err := fileBlock.Serialize(partitionPath, int64(sb.SBlockStart+(blockIndex*sb.SBlockS))); err != nil {
			return err
		}

		fileInode.IBlock[i] = blockIndex
		contentOffset += bytesToCopy
		blocksAssigned++

		// Actualizar bitmap y contadores
		if err := sb.UpdateBitmapBlock(partitionPath); err != nil {
			return err
		}
		sb.SBlocksCount++
		sb.SFreeBlocksCount--
		sb.SFirstBlo += sb.SBlockS
	}

	// Actualizar metadatos del inodo
	fileInode.ISize = int32(contentSize)
	fileInode.IMtime = float32(time.Now().Unix())
	fileInode.IAtime = float32(time.Now().Unix())

	// Escribir inodo actualizado
	if err := fileInode.Serialize(partitionPath, int64(sb.SInodeStart+(inodeIndex*sb.SInodeS))); err != nil {
		return fmt.Errorf("error al actualizar inodo: %v", err)
	}

	return nil
}

// Función para liberar bloques de un archivo
func (sb *SuperBlock) freeFileBlocks(partitionPath string, inode *INode) error {
	for i, blockIndex := range inode.IBlock {
		if blockIndex == -1 {
			continue
		}

		// Marcar bloque como libre
		file, err := os.OpenFile(partitionPath, os.O_WRONLY, 0666)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = file.Seek(int64(sb.SBmBlockStart+blockIndex), 0)
		if err != nil {
			return err
		}
		if _, err := file.Write([]byte{0}); err != nil {
			return err
		}

		// Actualizar contadores
		sb.SBlocksCount--
		sb.SFreeBlocksCount++
		inode.IBlock[i] = -1
	}
	return nil
}
