package partition_operations

import (
	"fmt"
	"strconv"
	"strings"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/operations/auth"
	ext2 "disk.simulator.com/m/v2/internal/disk/types/structures/ext"
	"disk.simulator.com/m/v2/utils"
)

// Función auxiliar para copiar un directorio y sus archivos con manejo explícito
func copyDirectoryWithFiles(superBlock ext2.SuperBlock, partitionPath string,
	sourcePath string, destPath string, uid, gid int32) error {

	fmt.Printf("Copiando directorio '%s' a '%s'\n", sourcePath, destPath)

	// Obtener componentes de las rutas
	sourceComponents := strings.Split(strings.Trim(sourcePath, "/"), "/")
	destComponents := strings.Split(strings.Trim(destPath, "/"), "/")

	// Asegurar que el directorio destino exista
	currentDestPath := []string{}
	for _, dir := range destComponents {
		// Para cada directorio en la ruta de destino
		if len(currentDestPath) > 0 {
			// Verificar si el directorio actual existe
			parentPath := currentDestPath[:len(currentDestPath)-1]
			dirName := currentDestPath[len(currentDestPath)-1]

			exists, _ := superBlock.FolderExists(partitionPath, parentPath, dirName)
			if !exists {
				// Crear directorio si no existe
				fmt.Printf("Creando directorio '%s' en ruta '%v'\n", dirName, parentPath)
				err := superBlock.CreateFolder(partitionPath, parentPath, dirName, false, uid, gid)
				if err != nil && !strings.Contains(err.Error(), "ya existe") {
					return fmt.Errorf("error al crear directorio destino '%s': %v", dirName, err)
				}
			}
		}

		// Añadir directorio actual a la ruta
		currentDestPath = append(currentDestPath, dir)
	}

	// Verificar que el último componente del destino exista
	destParent := destComponents[:len(destComponents)-1]
	lastDest := destComponents[len(destComponents)-1]
	err := superBlock.CreateFolder(partitionPath, destParent, lastDest, true, uid, gid)
	if err != nil && !strings.Contains(err.Error(), "ya existe") {
		return fmt.Errorf("error al crear directorio final '%s': %v", lastDest, err)
	}

	// Verificar que el directorio destino existe
	finalExists, err := superBlock.FolderExists(partitionPath, destParent, lastDest)
	if err != nil || !finalExists {
		return fmt.Errorf("error: no se pudo crear o verificar el directorio destino '%s'", lastDest)
	}

	// 1. Listar archivos en el directorio origen
	sourceParentPath := sourceComponents[:len(sourceComponents)-1]
	sourceName := sourceComponents[len(sourceComponents)-1]

	// Obtener inodo del directorio origen
	sourceInodeIndex, err := superBlock.FindFileInode(partitionPath, sourceParentPath, sourceName)
	if err != nil {
		return fmt.Errorf("error al buscar el directorio origen: %v", err)
	}

	// Leer inodo
	sourceInode := &ext2.INode{}
	err = sourceInode.Deserialize(partitionPath, int64(superBlock.SInodeStart+(sourceInodeIndex*superBlock.SInodeS)))
	if err != nil {
		return fmt.Errorf("error al leer inodo del directorio origen: %v", err)
	}

	// Recorrer bloques de directorio
	for _, blockIndex := range sourceInode.IBlock {
		if blockIndex == -1 {
			continue
		}

		// Leer bloque
		dirBlock := &ext2.DirBlock{}
		err = dirBlock.Deserialize(partitionPath, int64(superBlock.SBlockStart+(blockIndex*superBlock.SBlockS)))
		if err != nil {
			return fmt.Errorf("error al leer bloque de directorio: %v", err)
		}

		// Recorrer entradas
		for _, entry := range dirBlock.BContent {
			if entry.BInodo == -1 {
				continue
			}

			entryName := strings.Trim(string(entry.BName[:]), "\x00")
			// Ignorar "." y ".."
			if entryName == "." || entryName == ".." {
				continue
			}

			// Leer inodo de entrada
			entryInode := &ext2.INode{}
			err = entryInode.Deserialize(partitionPath, int64(superBlock.SInodeStart+(entry.BInodo*superBlock.SInodeS)))
			if err != nil {
				return fmt.Errorf("error al leer inodo de entrada: %v", err)
			}

			// Comprobar tipo
			if entryInode.IType[0] == '1' {
				// Es archivo - copiar directamente
				fmt.Printf("Copiando archivo: %s\n", entryName)
				content, err := superBlock.ReadFile(partitionPath, append(sourceParentPath, sourceName), entryName)
				if err != nil {
					return fmt.Errorf("error al leer archivo origen: %v", err)
				}

				destFullPath := append(destParent, lastDest)
				err = superBlock.CreateFile(partitionPath, destFullPath, entryName, 0, content, true, uid, gid)
				if err != nil {
					return fmt.Errorf("error al crear archivo destino: %v", err)
				}
			} else if entryInode.IType[0] == '0' {
				// Es directorio - crear primero y luego copiar contenido
				fmt.Printf("Copiando subdirectorio: %s\n", entryName)

				// Crear subdirectorio en destino
				destFullPath := append(destParent, lastDest)

				// Verificar si ya existe
				subdirExists, _ := superBlock.FolderExists(partitionPath, destFullPath, entryName)
				if !subdirExists {
					err = superBlock.CreateFolder(partitionPath, destFullPath, entryName, true, uid, gid)
					if err != nil && !strings.Contains(err.Error(), "ya existe") {
						return fmt.Errorf("error al crear subdirectorio: %v", err)
					}
				}

				// Verificar que se creó
				subdirExists, err = superBlock.FolderExists(partitionPath, destFullPath, entryName)
				if err != nil || !subdirExists {
					// Intentar crearlo una vez más si falló
					err = superBlock.CreateFolder(partitionPath, destFullPath, entryName, true, uid, gid)
					if err != nil && !strings.Contains(err.Error(), "ya existe") {
						return fmt.Errorf("error al crear subdirectorio (segundo intento): %v", err)
					}

					// Verificar nuevamente
					subdirExists, err = superBlock.FolderExists(partitionPath, destFullPath, entryName)
					if err != nil || !subdirExists {
						return fmt.Errorf("error: no se pudo crear subdirectorio '%s'", entryName)
					}
				}

				// Copiar contenido del subdirectorio recursivamente
				newSourcePath := sourcePath + "/" + entryName
				newDestPath := destPath + "/" + entryName
				err = copyDirectoryWithFiles(superBlock, partitionPath, newSourcePath, newDestPath, uid, gid)
				if err != nil {
					return fmt.Errorf("error al copiar subdirectorio '%s': %v", entryName, err)
				}
			}
		}
	}

	return nil
}

func CopyFileOrDirectory(sourcePath string, destPath string) error {
	instance := auth.GetInstance()

	if instance.User == nil {
		return fmt.Errorf("error al copiar: no hay un usuario loggeado")
	}

	id := instance.ID

	partition, partitionPath, err := memory.GetInstance().GetMountedPartition(id)
	if err != nil {
		return fmt.Errorf("error al obtener la partición: %v", err)
	}

	superBlock := ext2.SuperBlock{}
	err = superBlock.DeserializeSuperBlock(partitionPath, partition.Partition.Part_start)
	if err != nil {
		return fmt.Errorf("error al leer el superbloque: %v", err)
	}

	uidInt, _ := strconv.ParseInt(instance.User.UID, 10, 32)
	gidInt, _ := strconv.ParseInt(instance.GID, 10, 32)

	fmt.Printf("Copiando '%s' a '%s'\n", sourcePath, destPath)

	// Limpiar las rutas para evitar problemas con slashes repetidos o finales
	sourcePath = cleanPath(sourcePath)
	destPath = cleanPath(destPath)

	// Verificar si el origen es un directorio o un archivo
	sourceParents, sourceName := utils.GetParentDirectories(sourcePath)
	destParents, destName := utils.GetParentDirectories(destPath)

	// Asegurar que existan los directorios padres del destino
	err = createParentDirectories(superBlock, partitionPath, destParents, int32(uidInt), int32(gidInt))
	if err != nil {
		return fmt.Errorf("error al crear directorios destino: %v", err)
	}

	// Verificar si el origen es un archivo o un directorio
	isDirectory := false
	isFile := false

	// Primero intentamos verificar si es un directorio
	_, err = superBlock.FolderExists(partitionPath, sourceParents, sourceName)
	if err == nil {
		isDirectory = true
	} else {
		// Si no es directorio, verificamos si es un archivo
		_, err = superBlock.FindFileInode(partitionPath, sourceParents, sourceName)
		if err == nil {
			isFile = true
		} else {
			return fmt.Errorf("origen no encontrado: %v", err)
		}
	}

	// Verificar si el destino ya existe y qué tipo es
	destIsDir := false
	destExists, _ := superBlock.FolderExists(partitionPath, destParents, destName)
	if destExists {
		destIsDir = true
	}

	fmt.Printf("Tipo de origen - Directorio: %v, Archivo: %v\n", isDirectory, isFile)
	fmt.Printf("Destino existe como directorio: %v\n", destIsDir)

	// Determinar destino final
	finalDestParents := destParents
	finalDestName := destName

	if destIsDir && destName != "" {
		// Si el destino existe y es un directorio, el nombre final será el mismo que el origen
		finalDestParents = append(destParents, destName)
		finalDestName = sourceName
	} else if destName == "" {
		// Si no se especificó un nombre de destino (termina en slash), usar el nombre de origen
		finalDestName = sourceName
	}

	// Realizar la copia según el tipo de origen
	if isFile {
		// Es un archivo - copiar directamente
		return copyFile(superBlock, partitionPath, sourceParents, sourceName, finalDestParents, finalDestName, int32(uidInt), int32(gidInt))
	} else if isDirectory {
		// Es un directorio - crear directorio destino y copiar su contenido
		err = superBlock.CreateFolder(partitionPath, finalDestParents, finalDestName, true, int32(uidInt), int32(gidInt))
		if err != nil && !strings.Contains(err.Error(), "ya existe") {
			return fmt.Errorf("error al crear directorio destino: %v", err)
		}

		destFullPath := append(finalDestParents, finalDestName)
		return copyDirectoryContents(superBlock, partitionPath, sourceParents, sourceName, destFullPath, int32(uidInt), int32(gidInt))
	}

	return fmt.Errorf("no se pudo determinar el tipo de origen")
}

// Función auxiliar para crear directorios padres
func createParentDirectories(superBlock ext2.SuperBlock, partitionPath string, directories []string, uid, gid int32) error {
	if len(directories) == 0 {
		return nil
	}

	currentPath := []string{}

	for i, dir := range directories {
		if i > 0 {
			// Verificar si el directorio actual existe
			parentPath := currentPath
			dirExists, _ := superBlock.FolderExists(partitionPath, parentPath, dir)

			if !dirExists {
				// Crear si no existe
				fmt.Printf("Creando directorio '%s' en '%v'\n", dir, parentPath)
				err := superBlock.CreateFolder(partitionPath, parentPath, dir, false, uid, gid)
				if err != nil && !strings.Contains(err.Error(), "ya existe") {
					return fmt.Errorf("error al crear directorio '%s': %v", dir, err)
				}
			}
		}

		currentPath = append(currentPath, dir)
	}

	return nil
}

// Función auxiliar para copiar un archivo
func copyFile(superBlock ext2.SuperBlock, partitionPath string,
	sourceParents []string, sourceName string,
	destParents []string, destName string,
	uid, gid int32) error {

	fmt.Printf("Copiando archivo de '%v/%s' a '%v/%s'\n", sourceParents, sourceName, destParents, destName)

	// Leer contenido del archivo origen
	content, err := superBlock.ReadFile(partitionPath, sourceParents, sourceName)
	if err != nil {
		return fmt.Errorf("error al leer archivo origen: %v", err)
	}

	// Crear archivo en destino
	err = superBlock.CreateFile(partitionPath, destParents, destName, 0, content, true, uid, gid)
	if err != nil {
		return fmt.Errorf("error al crear archivo destino: %v", err)
	}

	return nil
}

// Función auxiliar para copiar el contenido de un directorio recursivamente
func copyDirectoryContents(superBlock ext2.SuperBlock, partitionPath string,
	sourceParents []string, sourceName string,
	destPath []string, uid, gid int32) error {

	fmt.Printf("Copiando contenido del directorio '%v/%s' a '%v'\n", sourceParents, sourceName, destPath)

	// Ruta completa del origen
	sourceFullPath := append(sourceParents, sourceName)

	// Obtener inodo del directorio origen
	sourceInodeIndex, err := superBlock.FindFileInode(partitionPath, sourceParents, sourceName)
	if err != nil {
		return fmt.Errorf("error al buscar directorio origen: %v", err)
	}

	// Leer inodo
	sourceInode := &ext2.INode{}
	err = sourceInode.Deserialize(partitionPath, int64(superBlock.SInodeStart+(sourceInodeIndex*superBlock.SInodeS)))
	if err != nil {
		return fmt.Errorf("error al leer inodo de directorio origen: %v", err)
	}

	// Primero procesamos todos los archivos en este directorio
	for _, blockIndex := range sourceInode.IBlock {
		if blockIndex == -1 {
			continue
		}

		// Leer bloque de directorio
		dirBlock := &ext2.DirBlock{}
		err = dirBlock.Deserialize(partitionPath, int64(superBlock.SBlockStart+(blockIndex*superBlock.SBlockS)))
		if err != nil {
			return fmt.Errorf("error al leer bloque de directorio: %v", err)
		}

		// Procesar cada entrada
		for _, entry := range dirBlock.BContent {
			if entry.BInodo == -1 {
				continue
			}

			entryName := strings.Trim(string(entry.BName[:]), "\x00")
			// Ignorar "." y ".."
			if entryName == "." || entryName == ".." {
				continue
			}

			// Leer inodo de la entrada
			entryInode := &ext2.INode{}
			err = entryInode.Deserialize(partitionPath, int64(superBlock.SInodeStart+(entry.BInodo*superBlock.SInodeS)))
			if err != nil {
				return fmt.Errorf("error al leer inodo de entrada: %v", err)
			}

			// Copiar según el tipo
			if entryInode.IType[0] == '1' {
				// Es archivo
				fmt.Printf("Copiando archivo '%s' de '%v' a '%v'\n", entryName, sourceFullPath, destPath)

				// Leer contenido
				content, err := superBlock.ReadFile(partitionPath, sourceFullPath, entryName)
				if err != nil {
					return fmt.Errorf("error al leer archivo '%s': %v", entryName, err)
				}

				// Crear en destino
				err = superBlock.CreateFile(partitionPath, destPath, entryName, 0, content, true, uid, gid)
				if err != nil {
					return fmt.Errorf("error al crear archivo destino '%s': %v", entryName, err)
				}
			} else if entryInode.IType[0] == '0' {
				// Es directorio - crear subdirectorio en destino
				fmt.Printf("Procesando subdirectorio '%s'\n", entryName)

				// Verificar si ya existe
				subdirExists, _ := superBlock.FolderExists(partitionPath, destPath, entryName)
				if !subdirExists {
					// Crear subdirectorio
					fmt.Printf("Creando subdirectorio '%s' en '%v'\n", entryName, destPath)
					err = superBlock.CreateFolder(partitionPath, destPath, entryName, true, uid, gid)
					if err != nil && !strings.Contains(err.Error(), "ya existe") {
						return fmt.Errorf("error al crear subdirectorio '%s': %v", entryName, err)
					}
				}

				// Verificar que se haya creado correctamente
				subdirExists, err = superBlock.FolderExists(partitionPath, destPath, entryName)
				if err != nil || !subdirExists {
					// Intentar crear nuevamente
					err = superBlock.CreateFolder(partitionPath, destPath, entryName, true, uid, gid)
					if err != nil && !strings.Contains(err.Error(), "ya existe") {
						return fmt.Errorf("error al crear subdirectorio '%s' (segundo intento): %v", entryName, err)
					}

					// Verificar nuevamente
					subdirExists, _ = superBlock.FolderExists(partitionPath, destPath, entryName)
					if !subdirExists {
						return fmt.Errorf("error: no se pudo crear subdirectorio '%s'", entryName)
					}
				}

				// Copiar contenido del subdirectorio recursivamente
				newDestPath := append(destPath, entryName)
				err = copyDirectoryContents(superBlock, partitionPath, sourceFullPath, entryName, newDestPath, uid, gid)
				if err != nil {
					return fmt.Errorf("error al copiar contenido de subdirectorio '%s': %v", entryName, err)
				}
			}
		}
	}

	return nil
}
