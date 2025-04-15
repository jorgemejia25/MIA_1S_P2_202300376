package ext2

import (
	"fmt"
	"strings"
	"time"
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
	// Si el nombre de destino está vacío, usar el mismo que el origen
	if destName == "" {
		destName = sourceName
	}

	// Buscar el inodo de origen
	sourceInodeIndex, err := sb.FindFileInode(path, sourceParentDirs, sourceName)
	if err != nil {
		return fmt.Errorf("error al buscar el origen: %v", err)
	}

	// Leer el inodo de origen
	sourceInode := &INode{}
	err = sourceInode.Deserialize(path, int64(sb.SInodeStart+(sourceInodeIndex*sb.SInodeS)))
	if err != nil {
		return fmt.Errorf("error al leer el inodo origen: %v", err)
	}

	// Verificar si existe el directorio destino
	destExists, err := sb.FolderExists(path, destParentDirs, "")
	if err != nil {
		return fmt.Errorf("error al verificar directorio destino: %v", err)
	}

	if !destExists {
		// El directorio destino no existe, crearlo
		err = sb.createParentFolders(path, destParentDirs, uid, gid)
		if err != nil {
			return fmt.Errorf("error al crear directorio destino: %v", err)
		}
		fmt.Println("Directorio destino creado con éxito")
	}

	fmt.Printf("Verificando tipo de origen: %c\n", sourceInode.IType[0])

	// Verificar tipo de origen (archivo o directorio)
	if sourceInode.IType[0] == '1' {
		// Es un archivo - copiar archivo
		fmt.Printf("Copiando archivo de '%v/%s' a '%v/%s'\n", sourceParentDirs, sourceName, destParentDirs, destName)
		return sb.copyFile(path, sourceParentDirs, sourceName, destParentDirs, destName, uid, gid)
	} else if sourceInode.IType[0] == '0' {
		// Es un directorio - copiar directorio recursivamente
		fmt.Printf("Copiando directorio de '%v/%s' a '%v/%s'\n", sourceParentDirs, sourceName, destParentDirs, destName)

		// Cuando copiamos directamente a la raíz o a un nivel principal
		if len(destParentDirs) <= 1 {
			fmt.Printf("Copiando a nivel raíz o principal\n")

			// Crear el directorio destino con el mismo nombre que el origen
			err = sb.CreateFolder(path, destParentDirs, sourceName, true, uid, gid)
			if err != nil && !strings.Contains(err.Error(), "ya existe") {
				return fmt.Errorf("error al crear directorio destino: %v", err)
			}

			// Verificar que el directorio se haya creado correctamente
			dirExists, err := sb.FolderExists(path, destParentDirs, sourceName)
			if err != nil {
				return fmt.Errorf("error al verificar directorio destino: %v", err)
			}
			if !dirExists {
				return fmt.Errorf("error: no se pudo verificar la creación del directorio destino '%s'", sourceName)
			}

			// Siempre usar copyDirectory para garantizar que se copie todo el contenido correctamente
			return sb.copyDirectory(path, sourceParentDirs, sourceName, destParentDirs, sourceName, uid, gid)
		} else {
			// Si estamos copiando a un nivel más profundo, usar copyDirectory
			return sb.copyDirectory(path, sourceParentDirs, sourceName, destParentDirs, destName, uid, gid)
		}
	}

	return fmt.Errorf("tipo de inodo no reconocido")
}

// createParentFolders crea los directorios padres recursivamente
func (sb *SuperBlock) createParentFolders(path string, folders []string, uid int32, gid int32) error {
	if len(folders) == 0 {
		return nil
	}

	// Construir la ruta paso a paso
	currentPath := []string{}

	for i, folder := range folders {
		// Verificar si este nivel ya existe
		exists, err := sb.FolderExists(path, currentPath, folder)
		if err != nil {
			return fmt.Errorf("error al verificar existencia del directorio '%s': %v", folder, err)
		}

		if !exists {
			// Si no existe, crear este directorio
			fmt.Printf("Creando directorio: %v - %s\n", currentPath, folder)
			err := sb.CreateFolder(path, currentPath, folder, false, uid, gid)
			if err != nil {
				return fmt.Errorf("error al crear directorio '%s': %v", folder, err)
			}
		}

		// Agregar este directorio a la ruta actual
		if i < len(folders)-1 {
			currentPath = append(currentPath, folder)
		}
	}

	return nil
}

// copyFile copia un archivo de origen a destino
func (sb *SuperBlock) copyFile(
	path string,
	sourceParentDirs []string,
	sourceName string,
	destParentDirs []string,
	destName string,
	uid int32,
	gid int32,
) error {
	fmt.Printf("Copiando archivo: %s a %v/%s\n", sourceName, destParentDirs, destName)

	// Leer el contenido del archivo de origen
	content, err := sb.ReadFile(path, sourceParentDirs, sourceName)
	if err != nil {
		return fmt.Errorf("error al leer el archivo origen: %v", err)
	}

	// Crear el archivo en la ubicación de destino
	err = sb.CreateFile(path, destParentDirs, destName, 0, content, true, uid, gid)
	if err != nil {
		return fmt.Errorf("error al crear el archivo destino: %v", err)
	}

	// Registrar la operación en el journal
	fullDestPath := append(destParentDirs, destName)
	err = AddJournal(path, int64(sb.SBlockStart), sb.SInodesCount,
		"copy",
		strings.Join(append(sourceParentDirs, sourceName), "/"),
		strings.Join(fullDestPath, "/"),
	)
	if err != nil {
		return fmt.Errorf("error al registrar en el journal: %v", err)
	}

	return nil
}

// copyDirectoryContents copia solo el contenido de un directorio a otro destino
func (sb *SuperBlock) copyDirectoryContents(
	path string,
	sourceParentDirs []string,
	sourceName string,
	destDirs []string,
	uid int32,
	gid int32,
) error {
	fmt.Printf("Copiando contenido del directorio '%s' a '%v'\n", sourceName, destDirs)

	// Construir la ruta completa del origen
	sourceFullPath := append(sourceParentDirs, sourceName)

	// Encontrar el inodo del directorio origen
	sourceInodeIndex, err := sb.FindFileInode(path, sourceParentDirs, sourceName)
	if err != nil {
		return fmt.Errorf("error al buscar el directorio origen: %v", err)
	}

	// Leer el inodo del directorio
	sourceInode := &INode{}
	err = sourceInode.Deserialize(path, int64(sb.SInodeStart+(sourceInodeIndex*sb.SInodeS)))
	if err != nil {
		return fmt.Errorf("error al leer el inodo del directorio origen: %v", err)
	}

	// Recorrer los bloques del directorio para encontrar todos sus contenidos
	for _, blockIndex := range sourceInode.IBlock {
		if blockIndex == -1 {
			continue
		}

		// Leer el bloque de directorio
		dirBlock := &DirBlock{}
		err := dirBlock.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
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
			entryInode := &INode{}
			err := entryInode.Deserialize(path, int64(sb.SInodeStart+(entry.BInodo*sb.SInodeS)))
			if err != nil {
				return fmt.Errorf("error al leer inodo de entrada: %v", err)
			}

			// Copiar recursivamente dependiendo del tipo
			if entryInode.IType[0] == '1' {
				// Es un archivo
				fmt.Printf("Copiando archivo '%s' en '%s' a '%v'\n", entryName, strings.Join(sourceFullPath, "/"), strings.Join(destDirs, "/"))
				err = sb.copyFile(path, sourceFullPath, entryName, destDirs, entryName, uid, gid)
				if err != nil {
					return fmt.Errorf("error al copiar archivo '%s': %v", entryName, err)
				}
			} else if entryInode.IType[0] == '0' {
				// Es un directorio - crear subdirectorio en destino y luego copiar contenido
				fmt.Printf("Procesando subdirectorio '%s' en '%s'\n", entryName, strings.Join(sourceFullPath, "/"))

				// Asegurar que el directorio destino existe antes de intentar crear subdirectorios
				destExists, err := sb.FolderExists(path, destDirs, "")
				if err != nil {
					return fmt.Errorf("error al verificar existencia del directorio destino: %v", err)
				}

				if !destExists {
					fmt.Printf("Creando directorio destino '%s'\n", strings.Join(destDirs, "/"))
					err = sb.createParentFolders(path, destDirs, uid, gid)
					if err != nil {
						return fmt.Errorf("error al crear directorio destino: %v", err)
					}
				}

				// Crear subdirectorio en el destino
				fmt.Printf("Creando subdirectorio '%s' en '%s'\n", entryName, strings.Join(destDirs, "/"))
				err = sb.CreateFolder(path, destDirs, entryName, true, uid, gid)
				if err != nil && !strings.Contains(err.Error(), "ya existe") {
					return fmt.Errorf("error al crear subdirectorio '%s': %v", entryName, err)
				}

				// Verificar que el subdirectorio se haya creado correctamente
				// Pero sin devolver error si la verificación es exitosa
				subDirExists, err := sb.FolderExists(path, destDirs, entryName)
				if err != nil {
					// En lugar de fallar, intentar continuar
					fmt.Printf("Advertencia: Error al verificar subdirectorio '%s': %v. Intentando continuar...\n", entryName, err)
				} else if !subDirExists {
					// Intentar una vez más crear el directorio si la verificación falló
					fmt.Printf("Subdirectorio '%s' no encontrado después de crear, intentando nuevamente...\n", entryName)
					err = sb.CreateFolder(path, destDirs, entryName, true, uid, gid)
					if err != nil && !strings.Contains(err.Error(), "ya existe") {
						return fmt.Errorf("error al crear subdirectorio '%s' (segundo intento): %v", entryName, err)
					}

					// Verificar nuevamente
					subDirExists, err = sb.FolderExists(path, destDirs, entryName)
					if err != nil {
						fmt.Printf("Advertencia: Error al verificar subdirectorio '%s' (segunda vez): %v. Intentando continuar...\n", entryName, err)
					} else if !subDirExists {
						// Si aún no existe, este es un error crítico
						return fmt.Errorf("error: no se pudo encontrar el subdirectorio '%s' después de crearlo (segundo intento)", entryName)
					}
				}

				// Preparar para copiar contenido recursivamente
				subSourcePath := append(sourceFullPath, entryName)
				subDestPath := append(destDirs, entryName)

				fmt.Printf("Copiando contenido de '%s' a '%s'\n",
					strings.Join(subSourcePath, "/"),
					strings.Join(subDestPath, "/"))

				// Copiar contenido recursivamente
				err = sb.copyDirectoryContents(path, sourceFullPath, entryName, subDestPath, uid, gid)
				if err != nil {
					return fmt.Errorf("error al copiar contenido del subdirectorio '%s': %v", entryName, err)
				}
			}
		}
	}

	// Registrar la operación en el journal
	err = AddJournal(path, int64(sb.SBlockStart), sb.SInodesCount,
		"copy",
		strings.Join(sourceFullPath, "/"),
		strings.Join(destDirs, "/"),
	)
	if err != nil {
		return fmt.Errorf("error al registrar en el journal: %v", err)
	}

	return nil
}

// copyDirectory copia un directorio y todo su contenido recursivamente
func (sb *SuperBlock) copyDirectory(
	path string,
	sourceParentDirs []string,
	sourceName string,
	destParentDirs []string,
	destName string,
	uid int32,
	gid int32,
) error {
	fmt.Printf("Copiando directorio: %s a %v/%s\n", sourceName, destParentDirs, destName)

	// Verificar si el directorio destino ya existe
	destExists, err := sb.FolderExists(path, destParentDirs, destName)
	if err != nil {
		return fmt.Errorf("error al verificar directorio destino: %v", err)
	}

	if !destExists {
		// Crear el directorio destino si no existe
		fmt.Printf("Creando directorio destino '%s' en '%v'\n", destName, destParentDirs)
		err := sb.CreateFolder(path, destParentDirs, destName, true, uid, gid)
		if err != nil && !strings.Contains(err.Error(), "ya existe") {
			return fmt.Errorf("error al crear directorio destino: %v", err)
		}

		// Verificar que se haya creado correctamente
		destExists, err = sb.FolderExists(path, destParentDirs, destName)
		if err != nil {
			fmt.Printf("Advertencia: Error al verificar la creación del directorio '%s': %v. Intentando continuar...\n", destName, err)
		} else if !destExists {
			// Intentar una vez más
			fmt.Printf("Directorio '%s' no encontrado después de crear, intentando nuevamente...\n", destName)
			err = sb.CreateFolder(path, destParentDirs, destName, true, uid, gid)
			if err != nil && !strings.Contains(err.Error(), "ya existe") {
				return fmt.Errorf("error al crear directorio destino (segundo intento): %v", err)
			}
		}
	}

	// Construir las rutas completas
	sourceFullPath := append(sourceParentDirs, sourceName)
	destFullPath := append(destParentDirs, destName)

	fmt.Printf("Procesando directorio origen: '%s'\n", strings.Join(sourceFullPath, "/"))
	fmt.Printf("Destino: '%s'\n", strings.Join(destFullPath, "/"))

	// Encontrar el inodo del directorio origen
	sourceInodeIndex, err := sb.FindFileInode(path, sourceParentDirs, sourceName)
	if err != nil {
		return fmt.Errorf("error al buscar el directorio origen: %v", err)
	}

	// Leer el inodo del directorio
	sourceInode := &INode{}
	err = sourceInode.Deserialize(path, int64(sb.SInodeStart+(sourceInodeIndex*sb.SInodeS)))
	if err != nil {
		return fmt.Errorf("error al leer el inodo del directorio origen: %v", err)
	}

	// Recorrer los bloques del directorio para encontrar todos sus contenidos
	for _, blockIndex := range sourceInode.IBlock {
		if blockIndex == -1 {
			continue
		}

		// Leer el bloque de directorio
		dirBlock := &DirBlock{}
		err := dirBlock.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
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
			entryInode := &INode{}
			err := entryInode.Deserialize(path, int64(sb.SInodeStart+(entry.BInodo*sb.SInodeS)))
			if err != nil {
				return fmt.Errorf("error al leer inodo de entrada: %v", err)
			}

			// Copiar recursivamente dependiendo del tipo
			if entryInode.IType[0] == '1' {
				// Es un archivo
				fmt.Printf("Copiando archivo '%s' de '%s' a '%s'\n",
					entryName,
					strings.Join(sourceFullPath, "/"),
					strings.Join(destFullPath, "/"))

				err = sb.copyFile(path, sourceFullPath, entryName, destFullPath, entryName, uid, gid)
				if err != nil {
					return fmt.Errorf("error al copiar archivo '%s': %v", entryName, err)
				}
			} else if entryInode.IType[0] == '0' {
				// Es un directorio - llamada recursiva
				fmt.Printf("Procesando subdirectorio '%s' en '%s'\n",
					entryName,
					strings.Join(sourceFullPath, "/"))

				// Asegurar que el directorio destino existe
				subDirDestExists, err := sb.FolderExists(path, destFullPath, "")
				if err != nil {
					return fmt.Errorf("error al verificar existencia del directorio destino '%s': %v",
						strings.Join(destFullPath, "/"), err)
				}

				if !subDirDestExists {
					// Crear destino si no existe
					fmt.Printf("Creando directorio destino '%s'\n", strings.Join(destFullPath, "/"))
					err = sb.createParentFolders(path, destFullPath, uid, gid)
					if err != nil {
						return fmt.Errorf("error al crear directorio destino '%s': %v",
							strings.Join(destFullPath, "/"), err)
					}
				}

				err = sb.copyDirectory(path, sourceFullPath, entryName, destFullPath, entryName, uid, gid)
				if err != nil {
					return fmt.Errorf("error al copiar directorio '%s': %v", entryName, err)
				}
			}
		}
	}

	// Actualizar tiempos de los inodos
	destInodeIndex, err := sb.FindFileInode(path, destParentDirs, destName)
	if err == nil {
		destInode := &INode{}
		err = destInode.Deserialize(path, int64(sb.SInodeStart+(destInodeIndex*sb.SInodeS)))
		if err == nil {
			destInode.IAtime = float32(time.Now().Unix())
			destInode.IMtime = float32(time.Now().Unix())
			destInode.Serialize(path, int64(sb.SInodeStart+(destInodeIndex*sb.SInodeS)))
		}
	}

	// Registrar la operación en el journal
	err = AddJournal(path, int64(sb.SBlockStart), sb.SInodesCount,
		"copy",
		strings.Join(sourceFullPath, "/"),
		strings.Join(destFullPath, "/"),
	)
	if err != nil {
		return fmt.Errorf("error al registrar en el journal: %v", err)
	}

	return nil
}
