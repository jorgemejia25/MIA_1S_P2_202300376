package ext2

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"disk.simulator.com/m/v2/utils"
)

func (sb *SuperBlock) CreateUsersFile(path string) error {

	// Crear el inodo raíz (inodo #0)
	rootInode := &INode{
		IUid:   1,
		IGid:   1,
		ISize:  0,
		IAtime: float32(time.Now().Unix()),
		ICtime: float32(time.Now().Unix()),
		IMtime: float32(time.Now().Unix()),
		IBlock: [15]int32{0, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, // Primer bloque es 0
		IType:  [1]byte{'0'},                                                         // Tipo directorio
		IPerm:  [3]byte{'7', '7', '7'},
	}

	// Serializar el inodo raíz en la posición inicial de la tabla de inodos
	err := rootInode.Serialize(path, int64(sb.SInodeStart))
	if err != nil {
		return err
	}

	// Actualizar el bitmap de inodos para marcar el primer inodo como usado
	file, err := os.OpenFile(path, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// Marcar como usado el primer inodo en el bitmap
	_, err = file.Seek(int64(sb.SBmInodeStart), 0)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte{1}) // 1 = usado
	if err != nil {
		return err
	}

	// Actualizar contadores
	sb.SInodesCount = 1
	sb.SFreeInodesCount--
	sb.SFirstIno = sb.SInodeStart + sb.SInodeS // Siguiente inodo libre

	// Crear el bloque para la carpeta raíz (bloque #0)
	rootBlock := &DirBlock{
		BContent: [4]DirContent{
			{BName: [12]byte{'.'}, BInodo: 0},                                         // Referencia a sí mismo
			{BName: [12]byte{'.', '.'}, BInodo: 0},                                    // Referencia a padre (es el mismo)
			{BName: [12]byte{'u', 's', 'e', 'r', 's', '.', 't', 'x', 't'}, BInodo: 1}, // Apuntará al inodo de users.txt
			{BName: [12]byte{'-'}, BInodo: -1},                                        // Entrada libre
		},
	}

	// Serializar el bloque de directorio raíz en la posición inicial de la tabla de bloques
	err = rootBlock.Serialize(path, int64(sb.SBlockStart))
	if err != nil {
		return err
	}

	// Marcar como usado el primer bloque en el bitmap
	_, err = file.Seek(int64(sb.SBmBlockStart), 0)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte{1}) // 1 = usado
	if err != nil {
		return err
	}

	// Actualizar bitmap de bloques
	err = sb.UpdateBitmapBlock(path)
	if err != nil {
		return err
	}

	// Actualizar contadores
	sb.SBlocksCount = 1
	sb.SFreeBlocksCount--
	sb.SFirstBlo = sb.SBlockStart + sb.SBlockS // Siguiente bloque libre

	// Serializar el journal para la carpeta raíz
	err = AddJournal(path, int64(sb.SBlockStart), sb.SInodesCount,
		"mkdir",
		"/",
		"",
	)
	if err != nil {
		fmt.Printf("Advertencia: No se pudo registrar la creación de la carpeta raíz en el journaling: %v\n", err)
		// No retornar error, ya que la carpeta fue creada exitosamente
	} else {
		fmt.Println("Creación de carpeta raíz registrada en el journaling")
	}

	// ----------- Creamos /users.txt -----------
	usersText := "1,G,root\n1,U,root,root,123\n"

	// Crear el inodo para users.txt (inodo #1)
	usersInode := &INode{
		IUid:   1,
		IGid:   1,
		ISize:  int32(len(usersText)),
		IAtime: float32(time.Now().Unix()),
		ICtime: float32(time.Now().Unix()),
		IMtime: float32(time.Now().Unix()),
		IBlock: [15]int32{1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, // Apunta al bloque #1
		IType:  [1]byte{'1'},                                                         // Tipo archivo
		IPerm:  [3]byte{'6', '6', '4'},                                               // Permisos rw-rw-r--
	}

	// Serializar el inodo de users.txt
	err = usersInode.Serialize(path, int64(sb.SFirstIno))
	if err != nil {
		return err
	}

	// Marcar como usado el segundo inodo en el bitmap
	_, err = file.Seek(int64(sb.SBmInodeStart+1), 0)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte{1}) // 1 = usado
	if err != nil {
		return err
	}

	// Actualizar contadores
	sb.SInodesCount++
	sb.SFreeInodesCount--
	sb.SFirstIno += sb.SInodeS // Siguiente inodo libre

	// Serializar el journal para el archivo users.txt
	err = AddJournal(path, int64(sb.SBlockStart), sb.SInodesCount,
		"mkfile",
		"/users.txt",
		usersText,
	)
	if err != nil {
		fmt.Printf("Advertencia: No se pudo registrar la creación del archivo users.txt en el journaling: %v\n", err)
		// No retornar error, ya que el archivo fue creado exitosamente
	} else {
		fmt.Println("Creación de users.txt registrada en el journaling")
	}

	// Crear el bloque para el archivo users.txt
	usersBlock := &FileBlock{
		BContent: [64]byte{},
	}
	copy(usersBlock.BContent[:], usersText)

	// Serializar el bloque de users.txt
	err = usersBlock.Serialize(path, int64(sb.SFirstBlo))
	if err != nil {
		return err
	}

	// Marcar como usado el segundo bloque en el bitmap
	_, err = file.Seek(int64(sb.SBmBlockStart+1), 0)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte{1}) // 1 = usado
	if err != nil {
		return err
	}

	// Actualizar bitmap de bloques
	err = sb.UpdateBitmapBlock(path)
	if err != nil {
		return err
	}

	// Actualizar contadores
	sb.SBlocksCount++
	sb.SFreeBlocksCount--
	sb.SFirstBlo += sb.SBlockS // Siguiente bloque libre

	fmt.Println("\nInodos y bloques creados con éxito:")
	fmt.Printf("Inodo raíz #0 creado\n")
	fmt.Printf("Bloque directorio raíz #0 creado\n")
	fmt.Printf("Inodo users.txt #1 creado\n")
	fmt.Printf("Bloque archivo users.txt #1 creado\n")

	return nil
}

func (sb *SuperBlock) createFolderInInode(
	path string,
	inodeIndex int32,
	parentsDir []string,
	destDir string,
	p bool,
	uid int32,
	gid int32,
) error {
	inode := &INode{}

	err := inode.Deserialize(path, int64(sb.SInodeStart+(inodeIndex*sb.SInodeS)))
	if err != nil {
		return err
	}

	// Si el inodo es tipo archivo, no puede contener carpetas
	if inode.IType[0] == '1' {
		return nil
	}

	// Validar permiso de escritura
	if !sb.userHasWritePermission(inode, uid, gid) {
		return fmt.Errorf("error: no tienes permisos de escritura en la carpeta padre")
	}

	// Si las carpetas padre no están vacías, buscar la carpeta padre más cercana
	if len(parentsDir) != 0 {
		// Obtener la carpeta padre más cercana
		parentDir, err := utils.First(parentsDir)
		if err != nil {
			return err
		}

		// Buscar esta carpeta en el inodo actual
		found := false
		var childInodeIndex int32 = -1

		// Iterar sobre cada bloque del inodo (apuntadores)
		for _, blockIndex := range inode.IBlock {
			if blockIndex == -1 {
				continue
			}

			// Deserializar el bloque
			dirBlock := &DirBlock{}
			err := dirBlock.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
			if err != nil {
				return err
			}

			// Buscar la carpeta en las entradas del directorio
			for _, content := range dirBlock.BContent {
				if content.BInodo == -1 {
					continue
				}

				entryName := strings.Trim(string(content.BName[:]), "\x00")
				parentDirStr := strings.Trim(parentDir, "\x00")

				if strings.EqualFold(entryName, parentDirStr) {
					found = true
					childInodeIndex = content.BInodo
					break
				}
			}

			if found {
				break
			}
		}

		if !found {
			if p {
				// Crear el directorio padre si no existe
				fmt.Printf("Creando directorio padre faltante '%s' en inodo %d\n", parentDir, inodeIndex)
				err2 := sb.createFolderInInode(path, inodeIndex, []string{}, parentDir, false, uid, gid)
				if err2 != nil {
					return fmt.Errorf("error al crear directorio padre '%s': %v", parentDir, err2)
				}

				// Después de crear el directorio, debemos buscarlo para obtener su inodo
				found = false
				childInodeIndex = -1

				// Volver a cargar el inodo actual porque podría haber cambiado
				err = inode.Deserialize(path, int64(sb.SInodeStart+(inodeIndex*sb.SInodeS)))
				if err != nil {
					return err
				}

				// Buscar el directorio recién creado en cada bloque del inodo
				for _, blockIndex := range inode.IBlock {
					if blockIndex == -1 {
						continue
					}

					dirBlock := &DirBlock{}
					err := dirBlock.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
					if err != nil {
						return err
					}

					for _, content := range dirBlock.BContent {
						if content.BInodo == -1 {
							continue
						}

						entryName := strings.Trim(string(content.BName[:]), "\x00")
						parentDirStr := strings.Trim(parentDir, "\x00")

						if strings.EqualFold(entryName, parentDirStr) {
							found = true
							childInodeIndex = content.BInodo
							fmt.Printf("Directorio padre '%s' encontrado en inodo %d\n", parentDir, childInodeIndex)
							break
						}
					}

					if found {
						break
					}
				}

				if !found || childInodeIndex == -1 {
					return fmt.Errorf("no se pudo encontrar el directorio padre '%s' después de crearlo", parentDir)
				}
			} else {
				return fmt.Errorf("no se encontró el directorio padre '%s'", parentDir)
			}
		}

		if found && childInodeIndex != -1 {
			// Continuar con el siguiente nivel de directorio
			remainingDirs := utils.RemoveElement(parentsDir, 0)
			fmt.Printf("Continuando navegación hacia '%s' (quedan %d directorios): %v\n",
				destDir, len(remainingDirs), remainingDirs)
			return sb.createFolderInInode(
				path,
				childInodeIndex,
				remainingDirs,
				destDir,
				p,
				uid,
				gid,
			)
		} else {
			return fmt.Errorf("no se encontró el directorio padre '%s' o su inodo es inválido", parentDir)
		}
	} else {
		// Aquí creamos el directorio final en el inodo actual

		// Primero verificamos si el directorio ya existe
		for _, blockIndex := range inode.IBlock {
			if blockIndex == -1 {
				continue
			}

			dirBlock := &DirBlock{}
			err := dirBlock.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
			if err != nil {
				return err
			}

			for _, content := range dirBlock.BContent {
				if content.BInodo == -1 {
					continue
				}

				entryName := strings.Trim(string(content.BName[:]), "\x00")
				if strings.EqualFold(entryName, destDir) {
					// El directorio ya existe
					return fmt.Errorf("el directorio '%s' ya existe", destDir)
				}
			}
		}

		// Ahora buscamos espacio para la nueva entrada de directorio
		foundSpace := false

		// Buscar un bloque con espacio libre
		for i, blockIndex := range inode.IBlock {
			if blockIndex == -1 {
				// Encontramos un espacio para nuevo bloque
				// Crear nuevo bloque de directorio
				newBlockIndex := (sb.SFirstBlo - sb.SBlockStart) / sb.SBlockS

				// Actualizar puntero en el inodo actual
				inode.IBlock[i] = newBlockIndex
				err = inode.Serialize(path, int64(sb.SInodeStart+(inodeIndex*sb.SInodeS)))
				if err != nil {
					return err
				}

				// Crear nuevo inodo para la carpeta destino
				newDirInodeIndex := sb.SInodesCount

				// Crear bloque de directorio con la nueva entrada
				dirBlock := &DirBlock{
					BContent: [4]DirContent{
						{BName: [12]byte{'.'}, BInodo: inodeIndex},
						{BName: [12]byte{'.', '.'}, BInodo: inodeIndex},
						{BName: [12]byte{}, BInodo: newDirInodeIndex},
						{BName: [12]byte{'-'}, BInodo: -1},
					},
				}
				copy(dirBlock.BContent[2].BName[:], destDir)

				// Escribir el bloque de directorio
				err = dirBlock.Serialize(path, int64(sb.SBlockStart+(newBlockIndex*sb.SBlockS)))
				if err != nil {
					return err
				}

				// Actualizar bitmap de bloques
				err = sb.UpdateBitmapBlock(path)
				if err != nil {
					return err
				}

				// Actualizar contadores
				sb.SBlocksCount++
				sb.SFreeBlocksCount--
				sb.SFirstBlo += sb.SBlockS

				// Crear inodo para el nuevo directorio
				newDirBlockIndex := (sb.SFirstBlo - sb.SBlockStart) / sb.SBlockS

				newInode := &INode{
					IUid:   uid,
					IGid:   gid,
					ISize:  0,
					IAtime: float32(time.Now().Unix()),
					ICtime: float32(time.Now().Unix()),
					IMtime: float32(time.Now().Unix()),
					IBlock: [15]int32{newDirBlockIndex, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					IType:  [1]byte{'0'},
					IPerm:  [3]byte{'6', '6', '4'},
				}

				// Serializar el inodo del nuevo directorio
				err = newInode.Serialize(path, int64(sb.SFirstIno))
				if err != nil {
					return err
				}

				// Actualizar bitmap de inodos
				err = sb.UpdateBitmapInode(path)
				if err != nil {
					return err
				}

				// Actualizar contadores de inodos
				sb.SInodesCount++
				sb.SFreeInodesCount--
				sb.SFirstIno += sb.SInodeS

				// Crear el primer bloque para el nuevo directorio
				newDirBlock := &DirBlock{
					BContent: [4]DirContent{
						{BName: [12]byte{'.'}, BInodo: newDirInodeIndex},
						{BName: [12]byte{'.', '.'}, BInodo: inodeIndex},
						{BName: [12]byte{'-'}, BInodo: -1},
						{BName: [12]byte{'-'}, BInodo: -1},
					},
				}

				// Serializar el bloque del nuevo directorio
				err = newDirBlock.Serialize(path, int64(sb.SBlockStart+(newDirBlockIndex*sb.SBlockS)))
				if err != nil {
					return err
				}

				// Actualizar bitmap de bloques
				err = sb.UpdateBitmapBlock(path)
				if err != nil {
					return err
				}

				// Actualizar contadores de bloques
				sb.SBlocksCount++
				sb.SFreeBlocksCount--
				sb.SFirstBlo += sb.SBlockS

				fmt.Printf("Carpeta '%s' creada con éxito en inodo %d\n", destDir, newDirInodeIndex)
				foundSpace = true
				break
			} else {
				// Este bloque ya existe, revisar si tiene espacio libre
				dirBlock := &DirBlock{}
				err := dirBlock.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
				if err != nil {
					return err
				}

				// Buscar entrada libre
				for j := 2; j < len(dirBlock.BContent); j++ {
					if dirBlock.BContent[j].BInodo == -1 {
						// Encontramos espacio libre

						// Crear nuevo inodo para el directorio destino
						newDirInodeIndex := sb.SInodesCount

						// Actualizar la entrada en el bloque directorio
						copy(dirBlock.BContent[j].BName[:], destDir)
						dirBlock.BContent[j].BInodo = newDirInodeIndex

						// Escribir el bloque actualizado
						err = dirBlock.Serialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
						if err != nil {
							return err
						}

						// Crear el primer bloque para el nuevo directorio
						newDirBlockIndex := (sb.SFirstBlo - sb.SBlockStart) / sb.SBlockS

						// Crear inodo para el nuevo directorio
						newInode := &INode{
							IUid:   uid,
							IGid:   gid,
							ISize:  0,
							IAtime: float32(time.Now().Unix()),
							ICtime: float32(time.Now().Unix()),
							IMtime: float32(time.Now().Unix()),
							IBlock: [15]int32{newDirBlockIndex, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
							IType:  [1]byte{'0'},
							IPerm:  [3]byte{'6', '6', '4'},
						}

						// Serializar el inodo del nuevo directorio
						err = newInode.Serialize(path, int64(sb.SFirstIno))
						if err != nil {
							return err
						}

						// Actualizar bitmap de inodos
						err = sb.UpdateBitmapInode(path)
						if err != nil {
							return err
						}

						// Actualizar contadores de inodos
						sb.SInodesCount++
						sb.SFreeInodesCount--
						sb.SFirstIno += sb.SInodeS

						// Crear el bloque para el nuevo directorio
						newDirBlock := &DirBlock{
							BContent: [4]DirContent{
								{BName: [12]byte{'.'}, BInodo: newDirInodeIndex},
								{BName: [12]byte{'.', '.'}, BInodo: inodeIndex},
								{BName: [12]byte{'-'}, BInodo: -1},
								{BName: [12]byte{'-'}, BInodo: -1},
							},
						}

						// Serializar el bloque del nuevo directorio
						err = newDirBlock.Serialize(path, int64(sb.SBlockStart+(newDirBlockIndex*sb.SBlockS)))
						if err != nil {
							return err
						}

						// Actualizar bitmap de bloques
						err = sb.UpdateBitmapBlock(path)
						if err != nil {
							return err
						}

						// Actualizar contadores de bloques
						sb.SBlocksCount++
						sb.SFreeBlocksCount--
						sb.SFirstBlo += sb.SBlockS

						fmt.Printf("Carpeta '%s' creada con éxito en inodo %d\n", destDir, newDirInodeIndex)
						foundSpace = true
						break
					}
				}
			}

			if foundSpace {
				break
			}
		}

		if !foundSpace {
			return fmt.Errorf("no hay espacio disponible en el directorio para crear la carpeta '%s'", destDir)
		}

		return nil
	}
}

func (sb *SuperBlock) CreateFolder(
	path string,
	parentsDir []string,
	destDir string,
	p bool,
	uid int32,
	gid int32,
) error {
	fmt.Printf("Creando carpeta '%s' con padres %v (crear padres: %v)\n", destDir, parentsDir, p)
	created := false
	for i := 0; i < int(sb.SInodesCount); i++ {
		err := sb.createFolderInInode(path, int32(i), parentsDir, destDir, p, uid, gid)
		if err == nil {
			created = true
			break
		}
	}
	if !created {
		return fmt.Errorf("no se pudo crear la carpeta '%s'", destDir)
	}

	// Agregar al journal
	err := AddJournal(path, int64(sb.SBlockStart), sb.SInodesCount,
		"mkdir",
		utils.PrintPath(append(parentsDir, destDir)),
		"",
	)
	if err != nil {
		return fmt.Errorf("error al agregar al journal: %v", err)
	}

	return nil
}

func (sb *SuperBlock) CreateFile(
	path string,
	parentsDir []string,
	destFile string,
	size int,
	content string,
	r bool,
	uid int32,
	gid int32,
) error {
	fmt.Printf("Creando archivo '%s' con padres %v (crear padres: %v), tamaño: %d\n",
		destFile, parentsDir, r, size)

	// Eliminamos la validación que exigía size>0 o content no vacío
	// Permitimos ahora crear un archivo vacío cuando ambos están ausentes

	// Manejo del contenido según los parámetros proporcionados
	if content == "" {
		// Si no hay contenido, usar size (incluso si es 0)
		content = strings.Repeat("\x00", size)
	}
	// Si hay contenido, se utilizará tal cual (ignorando size)

	// Si no hay directorios padres, crear directamente en la raíz
	if len(parentsDir) == 0 {
		return sb.createFileInInode(path, 0, parentsDir, destFile, content, uid, gid)
	}

	// Si hay directorios padres, verificar si existen y crearlos si es necesario

	// Empezar siempre desde la raíz (inodo 0)
	var currentInodeIndex int32 = 0
	var currentParentPath []string

	// Recorrer cada directorio padre
	for i, parentDir := range parentsDir {
		// Buscar este directorio padre en el inodo actual
		found := false
		foundInodeIndex := int32(-1)

		// Buscar en todos los bloques del inodo actual
		inode := &INode{}
		err := inode.Deserialize(path, int64(sb.SInodeStart+(currentInodeIndex*sb.SInodeS)))
		if err != nil {
			return fmt.Errorf("error al leer inodo %d: %v", currentInodeIndex, err)
		}

		// Verificar que sea un directorio
		if inode.IType[0] != '0' {
			return fmt.Errorf("el inodo %d no es un directorio", currentInodeIndex)
		}

		// Buscar en cada bloque del inodo actual
		for _, blockIndex := range inode.IBlock {
			if blockIndex == -1 {
				continue
			}

			// Deserializar el bloque
			block := &DirBlock{}
			err := block.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
			if err != nil {
				return fmt.Errorf("error al leer bloque %d: %v", blockIndex, err)
			}

			// Buscar el directorio en las entradas del bloque
			for _, entry := range block.BContent {
				if entry.BInodo == -1 {
					continue
				}

				entryName := strings.Trim(string(entry.BName[:]), "\x00")
				fmt.Printf("Comparando entrada '%s' con '%s'\n", entryName, parentDir)

				if strings.EqualFold(entryName, parentDir) {
					found = true
					foundInodeIndex = entry.BInodo
					fmt.Printf("¡Encontrado directorio '%s' en inodo %d!\n", parentDir, foundInodeIndex)
					break
				}
			}

			if found {
				break
			}
		}

		// Si el directorio padre no existe
		if !found {
			// Si flag r no está activo, devolver error
			if !r {
				return fmt.Errorf("directorio padre '%s' no encontrado", parentDir)
			}

			// Si flag r está activo, crear este directorio padre
			fmt.Printf("Directorio padre '%s' no encontrado, creándolo automáticamente...\n", parentDir)

			// Obtener la ruta actual hasta este directorio padre
			pathToCreate := utils.PrintPath(currentParentPath)
			fmt.Printf("Creando directorio '%s' en '%s'\n", parentDir, pathToCreate)

			// Crear este directorio padre
			err := sb.createFolderInInode(path, currentInodeIndex, []string{}, parentDir, false, uid, gid)
			if err != nil {
				return fmt.Errorf("error al crear directorio padre '%s': %v", parentDir, err)
			}

			// Buscar el inodo del directorio recién creado
			foundInodeIndex = -1
			// Buscar nuevamente el directorio en el inodo actual
			for _, blockIndex := range inode.IBlock {
				if blockIndex == -1 {
					continue
				}

				block := &DirBlock{}
				err := block.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
				if err != nil {
					return err
				}

				for _, entry := range block.BContent {
					if entry.BInodo == -1 {
						continue
					}

					entryName := strings.Trim(string(entry.BName[:]), "\x00")
					if strings.EqualFold(entryName, parentDir) {
						foundInodeIndex = entry.BInodo
						fmt.Printf("¡Encontrado directorio recién creado '%s' en inodo %d!\n", parentDir, foundInodeIndex)
						break
					}
				}

				if foundInodeIndex != -1 {
					break
				}
			}

			if foundInodeIndex == -1 {
				return fmt.Errorf("error: no se pudo encontrar el directorio recién creado '%s'", parentDir)
			}
		}

		// Avanzar al siguiente inodo y actualizar la ruta actual
		currentInodeIndex = foundInodeIndex
		currentParentPath = append(currentParentPath, parentDir)

		// Si estamos en el último directorio padre, crear el archivo en este directorio
		if i == len(parentsDir)-1 {
			// Crear el archivo dentro del último directorio padre
			fmt.Printf("Todos los directorios padres existen o han sido creados. Creando '%s' en inodo %d\n", destFile, currentInodeIndex)
			return sb.createFileInInode(path, currentInodeIndex, []string{}, destFile, content, uid, gid)
		}
	}

	// Este punto no debería alcanzarse nunca
	return fmt.Errorf("error inesperado en la creación del archivo")

}

func (sb *SuperBlock) createFileInInode(
	path string,
	inodeIndex int32,
	parentsDir []string,
	destFile string,
	content string,
	uid int32,
	gid int32,
) error {
	// Obtener el inodo donde crearemos el archivo
	inode := &INode{}
	err := inode.Deserialize(path, int64(sb.SInodeStart+(inodeIndex*sb.SInodeS)))
	if err != nil {
		return err
	}

	// Si el inodo es tipo archivo, no puede contener archivos
	if inode.IType[0] == '1' {
		return fmt.Errorf("no se puede crear un archivo dentro de otro archivo")
	}

	// Validar permiso de escritura
	if !sb.userHasWritePermission(inode, uid, gid) {
		return fmt.Errorf("error: no tienes permisos de escritura en la carpeta padre")
	}

	// Primero verificamos si el archivo ya existe en el directorio actual
	for blockPos := 0; blockPos < 12; blockPos++ {
		blockIndex := inode.IBlock[blockPos]
		if blockIndex == -1 {
			continue
		}

		// Deserializar el bloque
		dirBlock := &DirBlock{}
		err := dirBlock.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
		if err != nil {
			return err
		}

		// Verificar si el archivo ya existe
		for _, entry := range dirBlock.BContent {
			if entry.BInodo == -1 {
				continue
			}

			entryName := strings.Trim(string(entry.BName[:]), "\x00")
			if strings.EqualFold(entryName, destFile) {
				return fmt.Errorf("el archivo '%s' ya existe en este directorio", destFile)
			}
		}
	}

	// Si llegamos aquí, el archivo no existe y debemos crearlo
	// Primero, crear el inodo para el nuevo archivo
	fileInode := &INode{
		IUid:   uid,
		IGid:   gid,
		ISize:  int32(len(content)),
		IAtime: float32(time.Now().Unix()),
		ICtime: float32(time.Now().Unix()),
		IMtime: float32(time.Now().Unix()),
		IBlock: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, // Todos los bloques vacíos inicialmente
		IType:  [1]byte{'1'},                                                          // Tipo archivo
		IPerm:  [3]byte{'6', '6', '4'},                                                // Permisos rw-rw-r--
	}

	// Calcular cuántos bloques necesitamos para el contenido
	contentSize := len(content)
	blocksNeeded := int(math.Ceil(float64(contentSize) / float64(FileBlockSize)))

	fmt.Printf("Creando archivo con contenido de %d bytes, necesitando %d bloques\n", contentSize, blocksNeeded)

	// Asignar bloques para el contenido
	contentOffset := 0
	blocksAssigned := 0

	// Primero asignamos los bloques directos (0-11)
	for i := 0; i < 12 && blocksAssigned < blocksNeeded; i++ {
		// Crear un nuevo bloque de archivo
		fileBlock := &FileBlock{
			BContent: [FileBlockSize]byte{},
		}

		// Copiar parte del contenido a este bloque
		remainingBytes := contentSize - contentOffset
		bytesToCopy := FileBlockSize
		if remainingBytes < FileBlockSize {
			bytesToCopy = remainingBytes
		}

		copy(fileBlock.BContent[:bytesToCopy], content[contentOffset:contentOffset+bytesToCopy])
		contentOffset += bytesToCopy

		// Asignar y escribir el bloque
		blockIndex := (sb.SFirstBlo - sb.SBlockStart) / sb.SBlockS
		fileInode.IBlock[i] = blockIndex

		err := fileBlock.Serialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
		if err != nil {
			return fmt.Errorf("error al serializar bloque de archivo %d: %v", blockIndex, err)
		}

		// Actualizar bitmap de bloques
		err = sb.UpdateBitmapBlock(path)
		if err != nil {
			return err
		}

		sb.SBlocksCount++
		sb.SFreeBlocksCount--
		sb.SFirstBlo += sb.SBlockS
		blocksAssigned++

		fmt.Printf("Bloque directo #%d asignado con %d bytes de contenido\n", blockIndex, bytesToCopy)
	}

	// Si necesitamos más bloques, usar el bloque indirecto simple (posición 12)
	if blocksAssigned < blocksNeeded {
		// Crear bloque de punteros indirectos
		pointerBlock := &PointerBlock{}
		for i := range pointerBlock.PContent {
			pointerBlock.PContent[i] = -1 // Inicializar como vacíos
		}

		// Asignar bloque para los punteros
		pointerBlockIndex := (sb.SFirstBlo - sb.SBlockStart) / sb.SBlockS
		fileInode.IBlock[12] = pointerBlockIndex

		// Escribir el bloque de punteros
		err := pointerBlock.Serialize(path, int64(sb.SBlockStart+(pointerBlockIndex*sb.SBlockS)))
		if err != nil {
			return fmt.Errorf("error al serializar bloque de punteros: %v", err)
		}

		// Actualizar bitmap de bloques
		err = sb.UpdateBitmapBlock(path)
		if err != nil {
			return err
		}

		sb.SBlocksCount++
		sb.SFreeBlocksCount--
		sb.SFirstBlo += sb.SBlockS

		fmt.Printf("Bloque de punteros indirectos simple #%d asignado\n", pointerBlockIndex)

		// Asignar bloques indirectos mediante el bloque de punteros
		for i := 0; i < len(pointerBlock.PContent) && blocksAssigned < blocksNeeded; i++ {
			// Crear un nuevo bloque de archivo
			fileBlock := &FileBlock{
				BContent: [FileBlockSize]byte{},
			}

			// Copiar parte del contenido a este bloque
			remainingBytes := contentSize - contentOffset
			bytesToCopy := FileBlockSize
			if remainingBytes < FileBlockSize {
				bytesToCopy = remainingBytes
			}

			copy(fileBlock.BContent[:bytesToCopy], content[contentOffset:contentOffset+bytesToCopy])
			contentOffset += bytesToCopy

			// Asignar y escribir el bloque
			blockIndex := (sb.SFirstBlo - sb.SBlockStart) / sb.SBlockS
			pointerBlock.PContent[i] = blockIndex

			err := fileBlock.Serialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
			if err != nil {
				return fmt.Errorf("error al serializar bloque indirecto %d: %v", blockIndex, err)
			}

			// Actualizar bitmap de bloques
			err = sb.UpdateBitmapBlock(path)
			if err != nil {
				return err
			}

			sb.SBlocksCount++
			sb.SFreeBlocksCount--
			sb.SFirstBlo += sb.SBlockS
			blocksAssigned++

			fmt.Printf("Bloque indirecto simple #%d asignado con %d bytes de contenido\n", blockIndex, bytesToCopy)
		}

		// Actualizar el bloque de punteros con los nuevos valores
		err = pointerBlock.Serialize(path, int64(sb.SBlockStart+(pointerBlockIndex*sb.SBlockS)))
		if err != nil {
			return fmt.Errorf("error al actualizar bloque de punteros: %v", err)
		}
	}

	// Si necesitamos más bloques, usar el bloque indirecto doble (posición 13)
	if blocksAssigned < blocksNeeded {
		// Crear bloque de punteros indirectos dobles
		doublePointerBlock := &PointerBlock{}
		for i := range doublePointerBlock.PContent {
			doublePointerBlock.PContent[i] = -1 // Inicializar como vacíos
		}

		// Asignar bloque para los punteros dobles
		doublePointerBlockIndex := (sb.SFirstBlo - sb.SBlockStart) / sb.SBlockS
		fileInode.IBlock[13] = doublePointerBlockIndex

		// Escribir el bloque de punteros dobles
		err := doublePointerBlock.Serialize(path, int64(sb.SBlockStart+(doublePointerBlockIndex*sb.SBlockS)))
		if err != nil {
			return fmt.Errorf("error al serializar bloque de punteros doble: %v", err)
		}

		// Actualizar bitmap de bloques
		err = sb.UpdateBitmapBlock(path)
		if err != nil {
			return err
		}

		sb.SBlocksCount++
		sb.SFreeBlocksCount--
		sb.SFirstBlo += sb.SBlockS

		fmt.Printf("Bloque de punteros indirectos dobles #%d asignado\n", doublePointerBlockIndex)

		// Para cada entrada en el bloque de punteros dobles, crear un bloque de punteros simples
		for i := 0; i < len(doublePointerBlock.PContent) && blocksAssigned < blocksNeeded; i++ {
			// Crear un bloque de punteros indirectos simples
			simplePointerBlock := &PointerBlock{}
			for j := range simplePointerBlock.PContent {
				simplePointerBlock.PContent[j] = -1 // Inicializar como vacíos
			}

			// Asignar bloque para los punteros simples
			simplePointerBlockIndex := (sb.SFirstBlo - sb.SBlockStart) / sb.SBlockS
			doublePointerBlock.PContent[i] = simplePointerBlockIndex

			// Escribir el bloque de punteros simples
			err := simplePointerBlock.Serialize(path, int64(sb.SBlockStart+(simplePointerBlockIndex*sb.SBlockS)))
			if err != nil {
				return fmt.Errorf("error al serializar bloque de punteros simples dentro de dobles: %v", err)
			}

			// Actualizar bitmap de bloques
			err = sb.UpdateBitmapBlock(path)
			if err != nil {
				return err
			}

			sb.SBlocksCount++
			sb.SFreeBlocksCount--
			sb.SFirstBlo += sb.SBlockS

			fmt.Printf("Bloque de punteros simple #%d dentro de doble asignado\n", simplePointerBlockIndex)

			// Asignar bloques de archivo mediante este bloque de punteros simples
			for j := 0; j < len(simplePointerBlock.PContent) && blocksAssigned < blocksNeeded; j++ {
				// Crear un nuevo bloque de archivo
				fileBlock := &FileBlock{
					BContent: [FileBlockSize]byte{},
				}

				// Copiar parte del contenido a este bloque
				remainingBytes := contentSize - contentOffset
				bytesToCopy := FileBlockSize
				if remainingBytes < FileBlockSize {
					bytesToCopy = remainingBytes
				}

				copy(fileBlock.BContent[:bytesToCopy], content[contentOffset:contentOffset+bytesToCopy])
				contentOffset += bytesToCopy

				// Asignar y escribir el bloque
				blockIndex := (sb.SFirstBlo - sb.SBlockStart) / sb.SBlockS
				simplePointerBlock.PContent[j] = blockIndex

				err := fileBlock.Serialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
				if err != nil {
					return fmt.Errorf("error al serializar bloque indirecto doble %d: %v", blockIndex, err)
				}

				// Actualizar bitmap de bloques
				err = sb.UpdateBitmapBlock(path)
				if err != nil {
					return err
				}

				sb.SBlocksCount++
				sb.SFreeBlocksCount--
				sb.SFirstBlo += sb.SBlockS
				blocksAssigned++

				fmt.Printf("Bloque indirecto doble #%d asignado con %d bytes de contenido\n", blockIndex, bytesToCopy)
			}

			// Actualizar el bloque de punteros simples con los nuevos valores
			err = simplePointerBlock.Serialize(path, int64(sb.SBlockStart+(simplePointerBlockIndex*sb.SBlockS)))
			if err != nil {
				return fmt.Errorf("error al actualizar bloque de punteros simples dentro de dobles: %v", err)
			}
		}

		// Actualizar el bloque de punteros dobles con los nuevos valores
		err = doublePointerBlock.Serialize(path, int64(sb.SBlockStart+(doublePointerBlockIndex*sb.SBlockS)))
		if err != nil {
			return fmt.Errorf("error al actualizar bloque de punteros dobles: %v", err)
		}
	}

	// Si necesitamos más bloques, usar el bloque indirecto triple (posición 14)
	if blocksAssigned < blocksNeeded {
		// Crear bloque de punteros indirectos triples
		triplePointerBlock := &PointerBlock{}
		for i := range triplePointerBlock.PContent {
			triplePointerBlock.PContent[i] = -1 // Inicializar como vacíos
		}

		// Asignar bloque para los punteros triples
		triplePointerBlockIndex := (sb.SFirstBlo - sb.SBlockStart) / sb.SBlockS
		fileInode.IBlock[14] = triplePointerBlockIndex

		// Escribir el bloque de punteros triples
		err := triplePointerBlock.Serialize(path, int64(sb.SBlockStart+(triplePointerBlockIndex*sb.SBlockS)))
		if err != nil {
			return fmt.Errorf("error al serializar bloque de punteros triple: %v", err)
		}

		// Actualizar bitmap de bloques
		err = sb.UpdateBitmapBlock(path)
		if err != nil {
			return err
		}

		sb.SBlocksCount++
		sb.SFreeBlocksCount--
		sb.SFirstBlo += sb.SBlockS

		fmt.Printf("Bloque de punteros indirectos triples #%d asignado\n", triplePointerBlockIndex)

		// Para cada entrada en el bloque de punteros triples, crear un bloque de punteros dobles
		for i := 0; i < len(triplePointerBlock.PContent) && blocksAssigned < blocksNeeded; i++ {
			// Crear un bloque de punteros indirectos dobles
			doublePointerBlock := &PointerBlock{}
			for j := range doublePointerBlock.PContent {
				doublePointerBlock.PContent[j] = -1 // Inicializar como vacíos
			}

			// Asignar bloque para los punteros dobles
			doublePointerBlockIndex := (sb.SFirstBlo - sb.SBlockStart) / sb.SBlockS
			triplePointerBlock.PContent[i] = doublePointerBlockIndex

			// Escribir el bloque de punteros dobles
			err := doublePointerBlock.Serialize(path, int64(sb.SBlockStart+(doublePointerBlockIndex*sb.SBlockS)))
			if err != nil {
				return fmt.Errorf("error al serializar bloque de punteros dobles dentro de triples: %v", err)
			}

			// Actualizar bitmap de bloques
			err = sb.UpdateBitmapBlock(path)
			if err != nil {
				return err
			}

			sb.SBlocksCount++
			sb.SFreeBlocksCount--
			sb.SFirstBlo += sb.SBlockS

			fmt.Printf("Bloque de punteros doble #%d dentro de triple asignado\n", doublePointerBlockIndex)

			// Para cada entrada en el bloque de punteros dobles, crear un bloque de punteros simples
			for j := 0; j < len(doublePointerBlock.PContent) && blocksAssigned < blocksNeeded; j++ {
				// Crear un bloque de punteros indirectos simples
				simplePointerBlock := &PointerBlock{}
				for k := range simplePointerBlock.PContent {
					simplePointerBlock.PContent[k] = -1 // Inicializar como vacíos
				}

				// Asignar bloque para los punteros simples
				simplePointerBlockIndex := (sb.SFirstBlo - sb.SBlockStart) / sb.SBlockS
				doublePointerBlock.PContent[j] = simplePointerBlockIndex

				// Escribir el bloque de punteros simples
				err := simplePointerBlock.Serialize(path, int64(sb.SBlockStart+(simplePointerBlockIndex*sb.SBlockS)))
				if err != nil {
					return fmt.Errorf("error al serializar bloque de punteros simples dentro de triples: %v", err)
				}

				// Actualizar bitmap de bloques
				err = sb.UpdateBitmapBlock(path)
				if err != nil {
					return err
				}

				sb.SBlocksCount++
				sb.SFreeBlocksCount--
				sb.SFirstBlo += sb.SBlockS

				fmt.Printf("Bloque de punteros simple #%d dentro de triple asignado\n", simplePointerBlockIndex)

				// Asignar bloques de archivo mediante este bloque de punteros simples
				for k := 0; k < len(simplePointerBlock.PContent) && blocksAssigned < blocksNeeded; k++ {
					// Crear un nuevo bloque de archivo
					fileBlock := &FileBlock{
						BContent: [FileBlockSize]byte{},
					}

					// Copiar parte del contenido a este bloque
					remainingBytes := contentSize - contentOffset
					bytesToCopy := FileBlockSize
					if remainingBytes < FileBlockSize {
						bytesToCopy = remainingBytes
					}

					copy(fileBlock.BContent[:bytesToCopy], content[contentOffset:contentOffset+bytesToCopy])
					contentOffset += bytesToCopy

					// Asignar y escribir el bloque
					blockIndex := (sb.SFirstBlo - sb.SBlockStart) / sb.SBlockS
					simplePointerBlock.PContent[k] = blockIndex

					err := fileBlock.Serialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
					if err != nil {
						return fmt.Errorf("error al serializar bloque indirecto triple %d: %v", blockIndex, err)
					}

					// Actualizar bitmap de bloques
					err = sb.UpdateBitmapBlock(path)
					if err != nil {
						return err
					}

					sb.SBlocksCount++
					sb.SFreeBlocksCount--
					sb.SFirstBlo += sb.SBlockS
					blocksAssigned++

					fmt.Printf("Bloque indirecto triple #%d asignado con %d bytes de contenido\n", blockIndex, bytesToCopy)
				}

				// Actualizar el bloque de punteros simples con los nuevos valores
				err = simplePointerBlock.Serialize(path, int64(sb.SBlockStart+(simplePointerBlockIndex*sb.SBlockS)))
				if err != nil {
					return fmt.Errorf("error al actualizar bloque de punteros simples dentro de triples: %v", err)
				}
			}

			// Actualizar el bloque de punteros dobles con los nuevos valores
			err = doublePointerBlock.Serialize(path, int64(sb.SBlockStart+(doublePointerBlockIndex*sb.SBlockS)))
			if err != nil {
				return fmt.Errorf("error al actualizar bloque de punteros dobles dentro de triples: %v", err)
			}
		}

		// Actualizar el bloque de punteros triples con los nuevos valores
		err = triplePointerBlock.Serialize(path, int64(sb.SBlockStart+(triplePointerBlockIndex*sb.SBlockS)))
		if err != nil {
			return fmt.Errorf("error al actualizar bloque de punteros triples: %v", err)
		}
	}

	// Escribir el inodo del archivo
	err = fileInode.Serialize(path, int64(sb.SFirstIno))
	if err != nil {
		return fmt.Errorf("error al serializar inodo de archivo: %v", err)
	}

	// Actualizar bitmap y contadores de inodos
	err = sb.UpdateBitmapInode(path)
	if err != nil {
		return err
	}

	fileInodeIndex := sb.SInodesCount
	sb.SInodesCount++
	sb.SFreeInodesCount--
	sb.SFirstIno += sb.SInodeS

	fmt.Printf("Inodo de archivo #%d creado\n", fileInodeIndex)

	// Ahora debemos agregar una entrada en el directorio para el nuevo archivo
	foundSpace := false

	for blockPos := 0; blockPos < 12 && !foundSpace; blockPos++ {
		blockIndex := inode.IBlock[blockPos]

		// Si encontramos un puntero vacío, crear un nuevo bloque
		if blockIndex == -1 {
			fmt.Printf("Creando nuevo bloque de directorio para la entrada de archivo\n")

			blockIndex = (sb.SFirstBlo - sb.SBlockStart) / sb.SBlockS
			newBlock := &DirBlock{
				BContent: [4]DirContent{
					{BName: [12]byte{'.'}, BInodo: inodeIndex},
					{BName: [12]byte{'.', '.'}, BInodo: inodeIndex},
					{BName: [12]byte{}, BInodo: fileInodeIndex}, // Entrada para nuestro nuevo archivo
					{BName: [12]byte{'-'}, BInodo: -1},
				},
			}

			// Copiar el nombre del archivo
			copy(newBlock.BContent[2].BName[:], destFile)

			// Escribir el bloque
			err := newBlock.Serialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
			if err != nil {
				return fmt.Errorf("error al serializar bloque de directorio: %v", err)
			}

			// Actualizar el inodo del directorio
			inode.IBlock[blockPos] = blockIndex
			err = inode.Serialize(path, int64(sb.SInodeStart+(inodeIndex*sb.SInodeS)))
			if err != nil {
				return fmt.Errorf("error al actualizar inodo directorio: %v", err)
			}

			// Actualizar bitmap y contadores
			err = sb.UpdateBitmapBlock(path)
			if err != nil {
				return err
			}

			sb.SBlocksCount++
			sb.SFreeBlocksCount--
			sb.SFirstBlo += sb.SBlockS

			foundSpace = true
			break
		}

		// Si el bloque existe, buscar una entrada libre
		dirBlock := &DirBlock{}
		err := dirBlock.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
		if err != nil {
			return err
		}

		// Buscar una entrada libre
		for i, entry := range dirBlock.BContent {
			if entry.BInodo == -1 {
				// Encontramos una entrada libre
				fmt.Printf("Usando entrada libre %d en bloque %d\n", i, blockIndex)

				// Actualizar la entrada con el archivo
				copy(dirBlock.BContent[i].BName[:], destFile)
				dirBlock.BContent[i].BInodo = fileInodeIndex

				// Escribir el bloque actualizado
				err := dirBlock.Serialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
				if err != nil {
					return fmt.Errorf("error al actualizar bloque directorio: %v", err)
				}

				foundSpace = true
				break
			}
		}
	}

	if !foundSpace {
		return fmt.Errorf("no se encontró espacio para agregar la entrada del archivo en el directorio")
	}

	fmt.Printf("Archivo '%s' creado exitosamente\n", destFile)
	return nil
}

// ReadFile lee el contenido de un archivo en la ruta especificada
func (sb *SuperBlock) ReadFile(
	path string,
	parentDirs []string,
	fileName string,
) (string, error) {
	// Encontrar el inodo del archivo
	inodeIndex, err := sb.FindFileInode(path, parentDirs, fileName)
	if err != nil {
		return "", err
	}

	// Leer el inodo
	inode := &INode{}
	err = inode.Deserialize(path, int64(sb.SInodeStart+(inodeIndex*sb.SInodeS)))
	if err != nil {
		return "", err
	}

	// Verificar que sea un archivo
	if inode.IType[0] != '1' {
		return "", fmt.Errorf("'%s' no es un archivo", fileName)
	}

	// Buffer para el contenido
	content := make([]byte, inode.ISize)
	offset := 0

	// Leer los bloques directos (0-11)
	for i := 0; i < 12 && inode.IBlock[i] != -1 && offset < int(inode.ISize); i++ {
		blockIndex := inode.IBlock[i]
		fileBlock := &FileBlock{}
		err := fileBlock.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
		if err != nil {
			return "", err
		}

		// Copiar el contenido al buffer
		remainingSize := int(inode.ISize) - offset
		bytesToCopy := FileBlockSize
		if remainingSize < FileBlockSize {
			bytesToCopy = remainingSize
		}

		copy(content[offset:offset+bytesToCopy], fileBlock.BContent[:bytesToCopy])
		offset += bytesToCopy
	}

	// Si hay más contenido, leer bloques indirectos simples (bloque 12)
	if offset < int(inode.ISize) && inode.IBlock[12] != -1 {
		pointerBlock := &PointerBlock{}
		err := pointerBlock.Deserialize(path, int64(sb.SBlockStart+(inode.IBlock[12]*sb.SBlockS)))
		if err != nil {
			return "", err
		}

		for i := 0; i < len(pointerBlock.PContent) && pointerBlock.PContent[i] != -1 && offset < int(inode.ISize); i++ {
			blockIndex := pointerBlock.PContent[i]
			fileBlock := &FileBlock{}
			err := fileBlock.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
			if err != nil {
				return "", err
			}

			// Copiar el contenido al buffer
			remainingSize := int(inode.ISize) - offset
			bytesToCopy := FileBlockSize
			if remainingSize < FileBlockSize {
				bytesToCopy = remainingSize
			}

			copy(content[offset:offset+bytesToCopy], fileBlock.BContent[:bytesToCopy])
			offset += bytesToCopy
		}
	}

	// Si hay más contenido, leer bloques indirectos dobles (bloque 13)
	if offset < int(inode.ISize) && inode.IBlock[13] != -1 {
		doublePointerBlock := &PointerBlock{}
		err := doublePointerBlock.Deserialize(path, int64(sb.SBlockStart+(inode.IBlock[13]*sb.SBlockS)))
		if err != nil {
			return "", err
		}

		for i := 0; i < len(doublePointerBlock.PContent) && doublePointerBlock.PContent[i] != -1 && offset < int(inode.ISize); i++ {
			simplePointerBlockIndex := doublePointerBlock.PContent[i]
			simplePointerBlock := &PointerBlock{}
			err := simplePointerBlock.Deserialize(path, int64(sb.SBlockStart+(simplePointerBlockIndex*sb.SBlockS)))
			if err != nil {
				return "", err
			}

			for j := 0; j < len(simplePointerBlock.PContent) && simplePointerBlock.PContent[j] != -1 && offset < int(inode.ISize); j++ {
				blockIndex := simplePointerBlock.PContent[j]
				fileBlock := &FileBlock{}
				err := fileBlock.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
				if err != nil {
					return "", err
				}

				// Copiar el contenido al buffer
				remainingSize := int(inode.ISize) - offset
				bytesToCopy := FileBlockSize
				if remainingSize < FileBlockSize {
					bytesToCopy = remainingSize
				}

				copy(content[offset:offset+bytesToCopy], fileBlock.BContent[:bytesToCopy])
				offset += bytesToCopy
			}
		}
	}

	// Si hay más contenido, leer bloques indirectos triples (bloque 14)
	if offset < int(inode.ISize) && inode.IBlock[14] != -1 {
		triplePointerBlock := &PointerBlock{}
		err := triplePointerBlock.Deserialize(path, int64(sb.SBlockStart+(inode.IBlock[14]*sb.SBlockS)))
		if err != nil {
			return "", err
		}

		for i := 0; i < len(triplePointerBlock.PContent) && triplePointerBlock.PContent[i] != -1 && offset < int(inode.ISize); i++ {
			doublePointerBlockIndex := triplePointerBlock.PContent[i]
			doublePointerBlock := &PointerBlock{}
			err := doublePointerBlock.Deserialize(path, int64(sb.SBlockStart+(doublePointerBlockIndex*sb.SBlockS)))
			if err != nil {
				return "", err
			}

			for j := 0; j < len(doublePointerBlock.PContent) && doublePointerBlock.PContent[j] != -1 && offset < int(inode.ISize); j++ {
				simplePointerBlockIndex := doublePointerBlock.PContent[j]
				simplePointerBlock := &PointerBlock{}
				err := simplePointerBlock.Deserialize(path, int64(sb.SBlockStart+(simplePointerBlockIndex*sb.SBlockS)))
				if err != nil {
					return "", err
				}

				for k := 0; k < len(simplePointerBlock.PContent) && simplePointerBlock.PContent[k] != -1 && offset < int(inode.ISize); k++ {
					blockIndex := simplePointerBlock.PContent[k]
					fileBlock := &FileBlock{}
					err := fileBlock.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
					if err != nil {
						return "", err
					}

					// Copiar el contenido al buffer
					remainingSize := int(inode.ISize) - offset
					bytesToCopy := FileBlockSize
					if remainingSize < FileBlockSize {
						bytesToCopy = remainingSize
					}

					copy(content[offset:offset+bytesToCopy], fileBlock.BContent[:bytesToCopy])
					offset += bytesToCopy
				}
			}
		}
	}

	// Actualizar el tiempo de acceso del archivo
	inode.IAtime = float32(time.Now().Unix())
	err = inode.Serialize(path, int64(sb.SInodeStart+(inodeIndex*sb.SInodeS)))
	if err != nil {
		return "", fmt.Errorf("error al actualizar el tiempo de acceso: %v", err)
	}

	return string(content), nil
}

// FindFileInode encuentra el inodo de un archivo en la ruta especificada
func (sb *SuperBlock) FindFileInode(
	path string,
	parentDirs []string,
	fileName string,
) (int32, error) {
	// Ignorar posible directorio vacío al inicio
	if len(parentDirs) > 0 && parentDirs[0] == "" {
		parentDirs = parentDirs[1:]
	}
	// Empezar desde la raíz
	currentInodeIndex := int32(0)

	// Recorrer la ruta de directorios padre
	for _, parentDir := range parentDirs {
		found := false

		inode := &INode{}
		err := inode.Deserialize(path, int64(sb.SInodeStart+(currentInodeIndex*sb.SInodeS)))
		if err != nil {
			return -1, err
		}

		// Verificar que sea un directorio
		if inode.IType[0] != '0' {
			return -1, fmt.Errorf("'%s' no es un directorio", parentDir)
		}

		// Buscar el directorio en los bloques directos (0-11)
		for i := 0; i < 12; i++ {
			blockIndex := inode.IBlock[i]
			if blockIndex == -1 {
				continue
			}

			dirBlock := &DirBlock{}
			err = dirBlock.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
			if err != nil {
				return -1, err
			}

			// Buscar la entrada del directorio
			for _, entry := range dirBlock.BContent {
				if entry.BInodo == -1 {
					continue
				}

				entryName := strings.Trim(string(entry.BName[:]), "\x00")
				if strings.EqualFold(entryName, parentDir) {
					currentInodeIndex = entry.BInodo
					found = true
					break
				}
			}

			if found {
				break
			}
		}

		// Si no lo encontramos en bloques directos, buscar en bloques indirectos
		if !found && inode.IBlock[12] != -1 {
			// Bloque indirecto simple
			pointerBlock := &PointerBlock{}
			err = pointerBlock.Deserialize(path, int64(sb.SBlockStart+(inode.IBlock[12]*sb.SBlockS)))
			if err != nil {
				return -1, err
			}

			for _, ptr := range pointerBlock.PContent {
				if ptr == -1 {
					continue
				}

				dirBlock := &DirBlock{}
				err = dirBlock.Deserialize(path, int64(sb.SBlockStart+(ptr*sb.SBlockS)))
				if err != nil {
					return -1, err
				}

				// Buscar la entrada del directorio
				for _, entry := range dirBlock.BContent {
					if entry.BInodo == -1 {
						continue
					}

					entryName := strings.Trim(string(entry.BName[:]), "\x00")
					if strings.EqualFold(entryName, parentDir) {
						currentInodeIndex = entry.BInodo
						found = true
						break
					}
				}

				if found {
					break
				}
			}
		}

		if !found {
			return -1, fmt.Errorf("directorio '%s' no encontrado", parentDir)
		}
	}

	// Ahora buscar el archivo en el último directorio
	inode := &INode{}
	err := inode.Deserialize(path, int64(sb.SInodeStart+(currentInodeIndex*sb.SInodeS)))
	if err != nil {
		return -1, err
	}

	// Verificar que sea un directorio
	if inode.IType[0] != '0' {
		return -1, fmt.Errorf("la ubicación no es un directorio")
	}

	// Buscar el archivo en los bloques directos
	for i := 0; i < 12; i++ {
		blockIndex := inode.IBlock[i]
		if blockIndex == -1 {
			continue
		}

		dirBlock := &DirBlock{}
		err = dirBlock.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
		if err != nil {
			return -1, err
		}

		// Buscar la entrada del archivo
		for _, entry := range dirBlock.BContent {
			if entry.BInodo == -1 {
				continue
			}

			entryName := strings.Trim(string(entry.BName[:]), "\x00")
			if strings.EqualFold(entryName, fileName) {
				return entry.BInodo, nil
			}
		}
	}

	// Si no lo encontramos en bloques directos, buscar en bloques indirectos
	if inode.IBlock[12] != -1 {
		// Bloque indirecto simple
		pointerBlock := &PointerBlock{}
		err = pointerBlock.Deserialize(path, int64(sb.SBlockStart+(inode.IBlock[12]*sb.SBlockS)))
		if err != nil {
			return -1, err
		}

		for _, ptr := range pointerBlock.PContent {
			if ptr == -1 {
				continue
			}

			dirBlock := &DirBlock{}
			err = dirBlock.Deserialize(path, int64(sb.SBlockStart+(ptr*sb.SBlockS)))
			if err != nil {
				return -1, err
			}

			// Buscar la entrada del archivo
			for _, entry := range dirBlock.BContent {
				if entry.BInodo == -1 {
					continue
				}

				entryName := strings.Trim(string(entry.BName[:]), "\x00")
				if strings.EqualFold(entryName, fileName) {
					return entry.BInodo, nil
				}
			}
		}
	}

	return -1, fmt.Errorf("archivo '%s' no encontrado", fileName)
}

func (sb *SuperBlock) UpdateFile(
	path string,
	parentDirs []string,
	fileName string,
	newContent string,
) error {
	// 1. Buscar inodo del archivo
	inodeIndex, err := sb.FindFileInode(path, parentDirs, fileName)
	if err != nil {
		return err
	}

	// 2. Verificar tipo archivo
	fileInode := &INode{}
	err = fileInode.Deserialize(path, int64(sb.SInodeStart+(inodeIndex*sb.SInodeS)))
	if err != nil {
		return err
	}
	if fileInode.IType[0] != '1' {
		return fmt.Errorf("'%s' no es un archivo", fileName)
	}

	// 3. Liberar todos los bloques asignados a este inodo
	for i := 0; i < 12; i++ {
		blockIndex := fileInode.IBlock[i]
		if blockIndex == -1 {
			continue
		}

		// Marcar el bloque como libre en el bitmap
		file, err := os.OpenFile(path, os.O_WRONLY, 0666)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = file.Seek(int64(sb.SBmBlockStart+(blockIndex*sb.SBlockS)), 0)
		if err != nil {
			return err
		}
		_, err = file.Write([]byte{0}) // 0 = libre
		if err != nil {
			return err
		}

		// Actualizar contadores
		sb.SBlocksCount--
		sb.SFreeBlocksCount++
	}

	// 4. Reescribir nuevo contenido y reasignar bloques (similar a createFile)
	fileInode.ISize = int32(len(newContent))
	fileInode.IMtime = float32(time.Now().Unix())
	fileInode.IAtime = float32(time.Now().Unix())

	// Calcular cuántos bloques necesitamos para el contenido
	contentSize := len(newContent)
	blocksNeeded := int(math.Ceil(float64(contentSize) / float64(FileBlockSize)))

	fmt.Printf("Actualizando archivo con contenido de %d bytes, necesitando %d bloques\n", contentSize, blocksNeeded)

	// Asignar bloques para el contenido
	contentOffset := 0

	// Primero asignamos los bloques directos (0-11)
	for i := 0; i < blocksNeeded && i < 12; i++ {
		// Crear un nuevo bloque de archivo
		fileBlock := &FileBlock{
			BContent: [FileBlockSize]byte{},
		}

		// Copiar parte del contenido a este bloque
		remainingBytes := contentSize - contentOffset
		bytesToCopy := FileBlockSize
		if remainingBytes < FileBlockSize {
			bytesToCopy = remainingBytes
		}

		copy(fileBlock.BContent[:bytesToCopy], newContent[contentOffset:contentOffset+bytesToCopy])
		contentOffset += bytesToCopy

		// Asignar y escribir el bloque
		blockIndex := (sb.SFirstBlo - sb.SBlockStart) / sb.SBlockS
		fileInode.IBlock[i] = blockIndex

		err := fileBlock.Serialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
		if err != nil {
			return fmt.Errorf("error al serializar bloque de archivo %d: %v", blockIndex, err)
		}

		// Actualizar bitmap de bloques y contadores
		err = sb.UpdateBitmapBlock(path)
		if err != nil {
			return err
		}

		sb.SBlocksCount++
		sb.SFreeBlocksCount--
		sb.SFirstBlo += sb.SBlockS

		fmt.Printf("Bloque #%d asignado con %d bytes de contenido\n", blockIndex, bytesToCopy)
	}

	// Si necesitamos más bloques, usar el bloque indirecto simple (posición 12)
	if blocksNeeded > 12 {
		// Crear bloque de punteros indirectos
		pointerBlock := &PointerBlock{}
		for i := range pointerBlock.PContent {
			pointerBlock.PContent[i] = -1 // Inicializar como vacíos
		}

		// Asignar bloque para los punteros
		pointerBlockIndex := (sb.SFirstBlo - sb.SBlockStart) / sb.SBlockS
		fileInode.IBlock[12] = pointerBlockIndex

		// Escribir el bloque de punteros
		err := pointerBlock.Serialize(path, int64(sb.SBlockStart+(pointerBlockIndex*sb.SBlockS)))
		if err != nil {
			return fmt.Errorf("error al serializar bloque de punteros: %v", err)
		}

		// Actualizar bitmap de bloques
		err = sb.UpdateBitmapBlock(path)
		if err != nil {
			return err
		}

		// Actualizar bitmap y contadores
		err = sb.UpdateBitmapBlock(path)
		if err != nil {
			return err
		}

		sb.SBlocksCount++
		sb.SFreeBlocksCount--
		sb.SFirstBlo += sb.SBlockS

		fmt.Printf("Bloque de punteros #%d asignado para bloques indirectos\n", pointerBlockIndex)

		// Asignar bloques indirectos mediante el bloque de punteros
		for i := 0; i < blocksNeeded-12 && i < len(pointerBlock.PContent); i++ {
			// Crear un nuevo bloque de archivo
			fileBlock := &FileBlock{
				BContent: [FileBlockSize]byte{},
			}

			// Copiar parte del contenido a este bloque
			remainingBytes := contentSize - contentOffset
			bytesToCopy := FileBlockSize
			if remainingBytes < FileBlockSize {
				bytesToCopy = remainingBytes
			}

			copy(fileBlock.BContent[:bytesToCopy], newContent[contentOffset:contentOffset+bytesToCopy])
			contentOffset += bytesToCopy

			// Asignar y escribir el bloque
			blockIndex := (sb.SFirstBlo - sb.SBlockStart) / sb.SBlockS
			pointerBlock.PContent[i] = blockIndex

			err := fileBlock.Serialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
			if err != nil {
				return fmt.Errorf("error al serializar bloque indirecto %d: %v", blockIndex, err)
			}

			// Actualizar bitmap de bloques
			err = sb.UpdateBitmapBlock(path)
			if err != nil {
				return err
			}

			// Actualizar bitmap y contadores
			err = sb.UpdateBitmapBlock(path)
			if err != nil {
				return err
			}

			sb.SBlocksCount++
			sb.SFreeBlocksCount--
			sb.SFirstBlo += sb.SBlockS

			fmt.Printf("Bloque indirecto #%d asignado con %d bytes de contenido\n", blockIndex, bytesToCopy)
		}

		// Actualizar el bloque de punteros con los nuevos valores
		err = pointerBlock.Serialize(path, int64(sb.SBlockStart+(pointerBlockIndex*sb.SBlockS)))
		if err != nil {
			return fmt.Errorf("error al actualizar bloque de punteros: %v", err)
		}
	}

	// Actualizar y serializar inodo
	err = fileInode.Serialize(path, int64(sb.SInodeStart+(inodeIndex*sb.SInodeS)))
	if err != nil {
		return err
	}

	return nil
}

func (sb *SuperBlock) userHasWritePermission(inode *INode, uid int32, gid int32) bool {
	// Implementar validación según permisos 664 (rwx => 7, rw- => 6, r-- => 4).
	// Por ejemplo, si inode.IUid == uid y inode.IPerm[0] >= '2' => tiene escritura (6 => 'rw-').
	// Si inode.IGid == gid y inode.IPerm[1] >= '2' => tiene escritura.
	// Si ninguno anterior y inode.IPerm[2] >= '2' => tiene escritura.
	// De lo contrario, retornar false.
	if inode.IUid == uid && inode.IPerm[0] >= '2' {
		return true
	}
	if inode.IGid == gid && inode.IPerm[1] >= '2' {
		return true
	}
	if inode.IPerm[2] >= '2' {
		return true
	}
	return false
}

// folderExists verifica si existe un directorio en la ruta especificada
func (sb *SuperBlock) FolderExists(path string, parentDirs []string, folderName string) (bool, error) {
	// Si folderName está vacío, verificamos la existencia del último directorio en parentDirs
	if folderName == "" {
		if len(parentDirs) == 0 {
			// La raíz siempre existe
			return true, nil
		}

		// Tomamos el último elemento como el nombre del directorio a verificar
		folderName = parentDirs[len(parentDirs)-1]
		if len(parentDirs) > 1 {
			parentDirs = parentDirs[:len(parentDirs)-1]
		} else {
			parentDirs = []string{}
		}
	}

	// Buscar el inodo del directorio
	inodeIndex, err := sb.FindFileInode(path, parentDirs, folderName)
	if err != nil {
		if strings.Contains(err.Error(), "no encontrado") {
			// El directorio no existe
			return false, nil
		}
		// Otro tipo de error
		return false, err
	}

	// El directorio existe, verificar que sea realmente un directorio
	inode := &INode{}
	err = inode.Deserialize(path, int64(sb.SInodeStart+(inodeIndex*sb.SInodeS)))
	if err != nil {
		return false, err
	}

	// Verificar que sea un directorio (tipo '0')
	return inode.IType[0] == '0', nil
}
