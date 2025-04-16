package ext2

import (
	"fmt"
	"strings"
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
	// Si el nombre de destino está vacío, usar el mismo que el origen
	if destName == "" {
		destName = sourceName
	}

	// Verificar que el origen existe
	sourceInodeIndex, err := sb.FindFileInode(path, sourceParentDirs, sourceName)
	if err != nil {
		return fmt.Errorf("origen no encontrado: %v", err)
	}

	// Leer el inodo de origen
	sourceInode := &INode{}
	err = sourceInode.Deserialize(path, int64(sb.SInodeStart+(sourceInodeIndex*sb.SInodeS)))
	if err != nil {
		return fmt.Errorf("error al leer el inodo origen: %v", err)
	}

	// Verificar permisos en el origen
	if !sb.userHasWritePermission(sourceInode, uid, gid) {
		return fmt.Errorf("permisos insuficientes para mover el elemento de origen")
	}

	// Verificar si existe el directorio destino padre
	var destParentExists bool
	if len(destParentDirs) == 0 {
		// Si destParentDirs está vacío, el destino es la raíz, que siempre existe
		destParentExists = true
	} else if len(destParentDirs) == 1 {
		// Si solo hay un elemento, verificar si ese directorio existe en la raíz
		destParentExists, err = sb.FolderExists(path, []string{}, destParentDirs[0])
	} else {
		// Caso normal: verificar si el directorio padre existe
		destParentExists, err = sb.FolderExists(path, destParentDirs[:len(destParentDirs)-1], destParentDirs[len(destParentDirs)-1])
	}

	if err != nil || !destParentExists {
		// El directorio destino no existe, crearlo
		err = sb.createParentFolders(path, destParentDirs, uid, gid)
		if err != nil {
			return fmt.Errorf("error al crear directorio destino: %v", err)
		}
	}

	// Verificar si el destino ya existe
	// Primero, comprobamos si el destino es un directorio
	destIsDir := false
	destInodeIndex, err := sb.FindFileInode(path, destParentDirs, destName)
	if err == nil {
		// El destino existe, veamos si es un directorio
		destInode := &INode{}
		err = destInode.Deserialize(path, int64(sb.SInodeStart+(destInodeIndex*sb.SInodeS)))
		if err != nil {
			return fmt.Errorf("error al leer el inodo destino: %v", err)
		}

		if destInode.IType[0] == '0' {
			// Es un directorio, podemos mover dentro de él
			destIsDir = true
			// Ajustar el destino para mover dentro del directorio
			destParentDirs = append(destParentDirs, destName)
			destName = sourceName
		} else {
			// Es un archivo, no podemos sobrescribir
			return fmt.Errorf("el destino ya existe y no es un directorio")
		}
	}

	// Si no es un directorio y existe, verificar si existe con el nombre final
	if !destIsDir {
		_, err = sb.FindFileInode(path, destParentDirs, destName)
		if err == nil {
			return fmt.Errorf("el destino ya existe, no se puede sobrescribir")
		}
	}

	// Determinar tipo de origen y realizar la operación
	if sourceInode.IType[0] == '1' {
		// Es un archivo - copiar y luego eliminar
		err = sb.copyFile(path, sourceParentDirs, sourceName, destParentDirs, destName, uid, gid)
		if err != nil {
			return fmt.Errorf("error al copiar archivo: %v", err)
		}
	} else if sourceInode.IType[0] == '0' {
		// Es un directorio - copiar recursivamente y luego eliminar
		err = sb.copyDirectory(path, sourceParentDirs, sourceName, destParentDirs, destName, uid, gid)
		if err != nil {
			return fmt.Errorf("error al copiar directorio: %v", err)
		}
	} else {
		return fmt.Errorf("tipo de elemento no reconocido")
	}

	// Una vez copiado exitosamente, eliminar el origen
	err = sb.RemoveFileOrDirectory(path, sourceParentDirs, sourceName, uid, gid)
	if err != nil {
		return fmt.Errorf("error al eliminar origen después de copiar: %v", err)
	}

	// Registrar la operación en el journal
	sourceFullPath := append(sourceParentDirs, sourceName)
	destFullPath := append(destParentDirs, destName)
	err = AddJournal(path, int64(sb.SBlockStart), sb.SInodesCount,
		"move",
		strings.Join(sourceFullPath, "/"),
		strings.Join(destFullPath, "/"),
	)
	if err != nil {
		return fmt.Errorf("error al registrar en el journal: %v", err)
	}

	return nil
}
