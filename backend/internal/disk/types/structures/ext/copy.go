package ext2

import (
	"fmt"
	"strings"
)

// Copy realiza la copia de un archivo o directorio desde una ubicación de origen a una de destino
func (sb *SuperBlock) Copy(
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
		// Para este caso, vamos a copiar el contenido directamente dentro del directorio existente
		// en lugar de intentar crear un nuevo directorio con el mismo nombre

		// Primero verificamos que el elemento destino sea un directorio
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

		// Ahora vamos a copiar el contenido del directorio origen dentro del directorio destino existente
		switch sourceInode.IType[0] {
		case '0': // Directorio
			// Copiar el contenido recursivamente al directorio destino existente
			err = sb.copyDirectoryContents(path, sourceInodeIndex, destElementInodeIndex, uid, gid)
			if err != nil {
				return fmt.Errorf("error al copiar contenido del directorio: %v", err)
			}
			fmt.Printf("Contenido del directorio '%s' copiado exitosamente a '%s'\n", sourceName, destName)
			return nil
		case '1': // Archivo - este caso no debería ocurrir para directorios existentes
			return fmt.Errorf("ya existe un directorio con el nombre '%s' en el destino", destName)
		default:
			return fmt.Errorf("tipo de inodo no reconocido")
		}
	}

	// Basado en el tipo de inodo, copiar archivo o directorio
	switch sourceInode.IType[0] {
	case '0': // Directorio
		return sb.copyDirectory(path, sourceInodeIndex, sourceInode, destParentInodeIndex, destName, uid, gid)
	case '1': // Archivo
		return sb.copyFile(path, sourceInodeIndex, sourceInode, destParentInodeIndex, destName, uid, gid)
	default:
		return fmt.Errorf("tipo de inodo no reconocido")
	}
}

// fileExistsInDirectory verifica si ya existe un archivo/directorio con ese nombre en el directorio
func (sb *SuperBlock) fileExistsInDirectory(path string, dirInodeIndex int32, name string) (bool, error) {
	dirInode := &INode{}
	err := dirInode.Deserialize(path, int64(sb.SInodeStart+(dirInodeIndex*sb.SInodeS)))
	if err != nil {
		return false, err
	}

	// Buscar en los bloques directos del directorio
	for i := 0; i < 12; i++ {
		blockIndex := dirInode.IBlock[i]
		if blockIndex == -1 {
			continue
		}

		dirBlock := &DirBlock{}
		err := dirBlock.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
		if err != nil {
			return false, err
		}

		for _, entry := range dirBlock.BContent {
			if entry.BInodo == -1 {
				continue
			}

			entryName := strings.Trim(string(entry.BName[:]), "\x00")
			if strings.EqualFold(entryName, name) {
				return true, nil
			}
		}
	}

	// Buscar en bloques indirectos si existen
	if dirInode.IBlock[12] != -1 {
		// TODO: Implementar búsqueda en bloques indirectos si es necesario
	}

	return false, nil
}

// copyFile copia un archivo de origen a destino
func (sb *SuperBlock) copyFile(
	path string,
	sourceInodeIndex int32,
	sourceInode *INode,
	destDirInodeIndex int32,
	destName string,
	uid int32,
	gid int32,
) error {
	// Leer el contenido del archivo original
	content := make([]byte, sourceInode.ISize)
	contentOffset := 0

	// Leer los bloques directos (0-11)
	for i := 0; i < 12 && sourceInode.IBlock[i] != -1 && contentOffset < int(sourceInode.ISize); i++ {
		blockIndex := sourceInode.IBlock[i]
		fileBlock := &FileBlock{}
		err := fileBlock.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
		if err != nil {
			return fmt.Errorf("error al leer bloque de archivo: %v", err)
		}

		// Copiar el contenido al buffer
		remainingSize := int(sourceInode.ISize) - contentOffset
		bytesToCopy := FileBlockSize
		if remainingSize < FileBlockSize {
			bytesToCopy = remainingSize
		}

		copy(content[contentOffset:contentOffset+bytesToCopy], fileBlock.BContent[:bytesToCopy])
		contentOffset += bytesToCopy
	}

	// Leer los bloques indirectos si es necesario
	if contentOffset < int(sourceInode.ISize) && sourceInode.IBlock[12] != -1 {
		err := sb.readIndirectBlocks(path, sourceInode, content, &contentOffset)
		if err != nil {
			return fmt.Errorf("error al leer bloques indirectos: %v", err)
		}
	}

	// Crear el nuevo archivo en el destino
	err := sb.createFileInInode(
		path,
		destDirInodeIndex,
		[]string{}, // No se necesitan directorios padres ya que ya estamos en el directorio destino
		destName,
		string(content),
		uid,
		gid,
	)

	if err != nil {
		return fmt.Errorf("error al crear archivo copiado: %v", err)
	}

	fmt.Printf("Archivo '%s' copiado exitosamente a '%s'\n", sourceInode.GetName(), destName)
	return nil
}

// readIndirectBlocks lee los bloques indirectos de un archivo
func (sb *SuperBlock) readIndirectBlocks(
	path string,
	sourceInode *INode,
	content []byte,
	contentOffset *int,
) error {
	// Leer bloques indirectos simples (bloque 12)
	if sourceInode.IBlock[12] != -1 {
		pointerBlock := &PointerBlock{}
		err := pointerBlock.Deserialize(path, int64(sb.SBlockStart+(sourceInode.IBlock[12]*sb.SBlockS)))
		if err != nil {
			return err
		}

		for i := 0; i < len(pointerBlock.PContent) && pointerBlock.PContent[i] != -1 && *contentOffset < int(sourceInode.ISize); i++ {
			blockIndex := pointerBlock.PContent[i]
			fileBlock := &FileBlock{}
			err := fileBlock.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
			if err != nil {
				return err
			}

			// Copiar contenido al buffer
			remainingSize := int(sourceInode.ISize) - *contentOffset
			bytesToCopy := FileBlockSize
			if remainingSize < FileBlockSize {
				bytesToCopy = remainingSize
			}

			copy(content[*contentOffset:*contentOffset+bytesToCopy], fileBlock.BContent[:bytesToCopy])
			*contentOffset += bytesToCopy
		}
	}

	// Leer bloques indirectos dobles y triples si fuera necesario
	// Esto se implementaría de manera similar para los bloques 13 y 14
	// Se omite por brevedad, pero se agregaría para archivos muy grandes

	return nil
}

// copyDirectory copia un directorio y todo su contenido recursivamente
func (sb *SuperBlock) copyDirectory(
	path string,
	sourceInodeIndex int32,
	sourceInode *INode,
	destDirInodeIndex int32,
	destName string,
	uid int32,
	gid int32,
) error {
	// Crear el directorio destino
	err := sb.createFolderInInode(
		path,
		destDirInodeIndex,
		[]string{}, // No se necesitan directorios padres ya que ya estamos en el directorio destino
		destName,
		false,
		uid,
		gid,
	)
	if err != nil {
		return fmt.Errorf("error al crear directorio destino: %v", err)
	}

	// Obtener el inodo del nuevo directorio creado
	newDirInodeIndex, err := sb.findInodeInDirectory(path, destDirInodeIndex, destName)
	if err != nil {
		return fmt.Errorf("error al encontrar nuevo directorio creado: %v", err)
	}

	// Copiar el contenido del directorio origen al destino
	err = sb.copyDirectoryContents(path, sourceInodeIndex, newDirInodeIndex, uid, gid)
	if err != nil {
		return fmt.Errorf("error al copiar contenido del directorio: %v", err)
	}

	fmt.Printf("Directorio '%s' copiado exitosamente a '%s'\n", sourceInode.GetName(), destName)
	return nil
}

// copyDirectoryContents copia todos los archivos y subdirectorios de un directorio a otro
func (sb *SuperBlock) copyDirectoryContents(
	path string,
	sourceDirInodeIndex int32,
	destDirInodeIndex int32,
	uid int32,
	gid int32,
) error {
	// Obtener el inodo del directorio origen
	sourceDirInode := &INode{}
	err := sourceDirInode.Deserialize(path, int64(sb.SInodeStart+(sourceDirInodeIndex*sb.SInodeS)))
	if err != nil {
		return fmt.Errorf("error al leer inodo del directorio origen: %v", err)
	}

	// Procesar todos los bloques directos del directorio origen
	for i := 0; i < 12; i++ {
		blockIndex := sourceDirInode.IBlock[i]
		if blockIndex == -1 {
			continue
		}

		// Leer bloque de directorio
		dirBlock := &DirBlock{}
		err := dirBlock.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
		if err != nil {
			return fmt.Errorf("error al leer bloque de directorio: %v", err)
		}

		// Procesar cada entrada en el bloque de directorio
		for _, entry := range dirBlock.BContent {
			if entry.BInodo == -1 {
				continue
			}

			entryName := strings.Trim(string(entry.BName[:]), "\x00")
			// Ignorar las entradas "." y ".." que son referencias a sí mismo y al padre
			if entryName == "." || entryName == ".." {
				continue
			}

			// Leer el inodo de esta entrada
			entryInode := &INode{}
			err := entryInode.Deserialize(path, int64(sb.SInodeStart+(entry.BInodo*sb.SInodeS)))
			if err != nil {
				return fmt.Errorf("error al leer inodo de entrada: %v", err)
			}

			// Verificar si ya existe un elemento con el mismo nombre en el destino
			exists, err := sb.fileExistsInDirectory(path, destDirInodeIndex, entryName)
			if err != nil {
				return fmt.Errorf("error al verificar si existe '%s' en el destino: %v", entryName, err)
			}

			if exists {
				// Si ya existe, verificamos si es un directorio para poder copiar dentro
				existingInodeIndex, err := sb.findInodeInDirectory(path, destDirInodeIndex, entryName)
				if err != nil {
					return fmt.Errorf("error al obtener inodo existente: %v", err)
				}

				existingInode := &INode{}
				err = existingInode.Deserialize(path, int64(sb.SInodeStart+(existingInodeIndex*sb.SInodeS)))
				if err != nil {
					return fmt.Errorf("error al leer inodo existente: %v", err)
				}

				// Si ambos son directorios, podemos copiar dentro
				if entryInode.IType[0] == '0' && existingInode.IType[0] == '0' {
					// Copiar recursivamente el contenido del subdirectorio
					err = sb.copyDirectoryContents(path, entry.BInodo, existingInodeIndex, uid, gid)
					if err != nil {
						return fmt.Errorf("error al copiar contenido del subdirectorio '%s': %v", entryName, err)
					}
					continue
				} else {
					// Si uno es archivo u otro directorio, saltamos esta entrada
					fmt.Printf("Saltando '%s': ya existe en el destino\n", entryName)
					continue
				}
			}

			// Copiar la entrada según su tipo
			switch entryInode.IType[0] {
			case '0': // Directorio
				// Crear el subdirectorio en el destino
				err = sb.createFolderInInode(path, destDirInodeIndex, []string{}, entryName, false, uid, gid)
				if err != nil {
					return fmt.Errorf("error al crear subdirectorio '%s': %v", entryName, err)
				}

				// Obtener el inodo del nuevo subdirectorio creado
				newSubDirInodeIndex, err := sb.findInodeInDirectory(path, destDirInodeIndex, entryName)
				if err != nil {
					return fmt.Errorf("error al encontrar nuevo subdirectorio '%s': %v", entryName, err)
				}

				// Copiar recursivamente el contenido del subdirectorio
				err = sb.copyDirectoryContents(path, entry.BInodo, newSubDirInodeIndex, uid, gid)
				if err != nil {
					return fmt.Errorf("error al copiar contenido del subdirectorio '%s': %v", entryName, err)
				}

			case '1': // Archivo
				// Leer el contenido del archivo
				content := make([]byte, entryInode.ISize)
				contentOffset := 0

				// Leer los bloques directos (0-11)
				for j := 0; j < 12 && entryInode.IBlock[j] != -1 && contentOffset < int(entryInode.ISize); j++ {
					fileBlockIndex := entryInode.IBlock[j]
					fileBlock := &FileBlock{}
					err := fileBlock.Deserialize(path, int64(sb.SBlockStart+(fileBlockIndex*sb.SBlockS)))
					if err != nil {
						return fmt.Errorf("error al leer bloque de archivo: %v", err)
					}

					// Copiar el contenido al buffer
					remainingSize := int(entryInode.ISize) - contentOffset
					bytesToCopy := FileBlockSize
					if remainingSize < FileBlockSize {
						bytesToCopy = remainingSize
					}

					copy(content[contentOffset:contentOffset+bytesToCopy], fileBlock.BContent[:bytesToCopy])
					contentOffset += bytesToCopy
				}

				// Leer los bloques indirectos si es necesario
				if contentOffset < int(entryInode.ISize) && entryInode.IBlock[12] != -1 {
					err := sb.readIndirectBlocks(path, entryInode, content, &contentOffset)
					if err != nil {
						return fmt.Errorf("error al leer bloques indirectos: %v", err)
					}
				}

				// Crear el archivo en el directorio destino
				err = sb.createFileInInode(path, destDirInodeIndex, []string{}, entryName, string(content), uid, gid)
				if err != nil {
					return fmt.Errorf("error al crear archivo copiado '%s': %v", entryName, err)
				}

			default:
				return fmt.Errorf("tipo de inodo no reconocido para '%s'", entryName)
			}
		}
	}

	// Procesar bloques indirectos si existen
	// Por simplicidad se omite pero sería similar a la búsqueda anterior

	return nil
}

// GetName obtiene el nombre asociado al inodo
func (inode *INode) GetName() string {
	return "archivo/directorio"
}

// userHasReadPermission verifica si un usuario tiene permisos de lectura en un inodo
func (sb *SuperBlock) userHasReadPermission(inode *INode, uid int32, gid int32) bool {
	// Usuario propietario
	if inode.IUid == uid && inode.IPerm[0] >= '4' {
		return true
	}
	// Grupo propietario
	if inode.IGid == gid && inode.IPerm[1] >= '4' {
		return true
	}
	// Otros usuarios
	if inode.IPerm[2] >= '4' {
		return true
	}
	return false
}

// findInodeInDirectory busca un inodo por nombre dentro de un directorio específico
func (sb *SuperBlock) findInodeInDirectory(diskPath string, dirInodeIndex int32, name string) (int32, error) {
	dirInode, err := sb.GetInodeByNumber(diskPath, dirInodeIndex)
	if err != nil {
		return -1, err
	}

	// Buscar en los bloques directos
	for i := 0; i < 12; i++ {
		blockIndex := dirInode.IBlock[i]
		if blockIndex == -1 {
			continue
		}

		dirBlock := &DirBlock{}
		err := dirBlock.Deserialize(diskPath, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
		if err != nil {
			continue
		}

		for _, entry := range dirBlock.BContent {
			if entry.BInodo == -1 {
				continue
			}

			entryName := strings.Trim(string(entry.BName[:]), "\x00")
			if entryName == name {
				return entry.BInodo, nil
			}
		}
	}

	// Buscar en bloques indirectos si es necesario
	if dirInode.IBlock[12] != -1 {
		// Por simplicidad se omite la búsqueda en bloques indirectos
	}

	return -1, fmt.Errorf("no se encontró '%s' en el directorio", name)
}
