package ext2

import (
	"fmt"
	"strings"
)

// matchPattern verifica si un nombre coincide con un patrón que puede contener comodines
// * - coincide con cero o más caracteres
// ? - coincide con exactamente un carácter
func matchPattern(pattern, name string) bool {
	// Caso base: si el patrón está vacío, el nombre también debe estarlo
	if pattern == "" {
		return name == ""
	}

	// Si el patrón empieza con '*'
	if pattern[0] == '*' {
		// Intentar todas las posibles longitudes del comodín '*'
		// '*' puede coincidir con 0 o más caracteres
		for i := 0; i <= len(name); i++ {
			// Verificar si el resto del patrón coincide con el resto del nombre
			if matchPattern(pattern[1:], name[i:]) {
				return true
			}
		}
		return false
	}

	// Si el patrón empieza con '?' o si el primer carácter coincide
	if len(name) > 0 && (pattern[0] == '?' || pattern[0] == name[0]) {
		// Continuar con el siguiente carácter en ambos
		return matchPattern(pattern[1:], name[1:])
	}

	// Si llegamos aquí, no hay coincidencia
	return false
}

// GenerateFileSystemTree genera una representación gráfica en texto del sistema de archivos
func (sb *SuperBlock) GenerateFileSystemTree(path string) (string, error) {
	var output strings.Builder
	output.WriteString("# Arbol Actual\n")

	// Empezar desde la raíz (inodo 0)
	rootInode, err := sb.GetInodeByNumber(path, 0)
	if err != nil {
		return "", fmt.Errorf("error al leer el inodo raíz: %v", err)
	}

	// Agregar la raíz al árbol
	output.WriteString("# /\n")

	// Recorrer recursivamente los archivos y carpetas
	err = sb.traverseDirectory(path, rootInode, 0, "", &output, 0, make(map[int32]bool))
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

// FindFileOrFolderByName busca un archivo o carpeta con el nombre especificado a partir de un directorio
// y genera un árbol de búsqueda mostrando solo las rutas donde se encuentra
func (sb *SuperBlock) FindFileOrFolderByName(
	diskPath string,
	startDirs []string,
	name string,
) (string, error) {
	var output strings.Builder
	output.WriteString("# Arbol de Búsqueda: " + name + "\n")

	// Primero, encontrar el inodo del directorio de inicio
	var startInodeIndex int32 = 0 // Por defecto, comenzar desde la raíz
	var currentDirName string = "/"

	// Si hay una ruta de inicio, navegamos a ese directorio
	if len(startDirs) > 0 {
		// Encontrar el inodo del directorio de inicio
		for i, dir := range startDirs {
			if i == 0 && dir == "" {
				// Si la ruta comienza con "/", comenzamos desde la raíz
				continue
			}

			// Buscar el directorio en el directorio actual
			dirInodeIndex, err := sb.findInodeInDirectory(diskPath, startInodeIndex, dir)
			if err != nil {
				return "", fmt.Errorf("error al buscar directorio '%s': %v", dir, err)
			}

			startInodeIndex = dirInodeIndex
			currentDirName = dir
		}
	}

	// Leer el inodo del directorio de inicio
	startInode, err := sb.GetInodeByNumber(diskPath, startInodeIndex)
	if err != nil {
		return "", fmt.Errorf("error al leer inodo de inicio: %v", err)
	}

	// Inicializar el árbol de búsqueda con la raíz
	output.WriteString("# " + currentDirName + "\n")

	// Estructuras para rastrear la búsqueda
	foundPaths := make([]string, 0)
	foundInodes := make([]int32, 0)

	// Realizar la búsqueda recursiva
	err = sb.searchInDirectory(
		diskPath,
		startInode,
		startInodeIndex,
		name,
		[]string{currentDirName},
		"",
		&foundPaths,
		&foundInodes,
		make(map[int32]bool),
	)

	if err != nil {
		return "", err
	}

	// Si no se encontró nada, devolver un mensaje
	if len(foundPaths) == 0 {
		return fmt.Sprintf("# No se encontró '%s' a partir de '%s'", name, currentDirName), nil
	}

	// Generar el árbol con los resultados encontrados
	for i, path := range foundPaths {
		inodeIndex := foundInodes[i]
		inode, err := sb.GetInodeByNumber(diskPath, inodeIndex)
		if err != nil {
			continue
		}

		// Formatear los permisos para mostrarlos
		perms := fmt.Sprintf("#%c%c%c", inode.IPerm[0], inode.IPerm[1], inode.IPerm[2])

		// Generar la ruta completa con el formato de árbol
		pathComponents := strings.Split(path, "/")

		// Ignorar componentes vacíos
		var validComponents []string
		for _, comp := range pathComponents {
			if comp != "" {
				validComponents = append(validComponents, comp)
			}
		}

		// Generar el árbol para esta ruta
		for j, comp := range validComponents {
			prefix := strings.Repeat("|  ", j)
			if j == len(validComponents)-1 {
				// El último componente tiene los permisos
				output.WriteString(fmt.Sprintf("%s|_ %s %s\n", prefix, comp, perms))
			} else {
				output.WriteString(fmt.Sprintf("%s|_ %s\n", prefix, comp))
			}
		}
	}

	return output.String(), nil
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

	// Buscar en bloques indirectos
	if dirInode.IBlock[12] != -1 {
		inodeIndex, found, err := sb.findInIndirectBlocks(diskPath, dirInode.IBlock[12], name)
		if err == nil && found {
			return inodeIndex, nil
		}
	}

	return -1, fmt.Errorf("no se encontró '%s' en el directorio", name)
}

// findInIndirectBlocks busca un nombre en bloques indirectos
func (sb *SuperBlock) findInIndirectBlocks(diskPath string, blockIndex int32, name string) (int32, bool, error) {
	pointerBlock := &PointerBlock{}
	err := pointerBlock.Deserialize(diskPath, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
	if err != nil {
		return -1, false, err
	}

	for _, ptr := range pointerBlock.PContent {
		if ptr == -1 {
			continue
		}

		dirBlock := &DirBlock{}
		err := dirBlock.Deserialize(diskPath, int64(sb.SBlockStart+(ptr*sb.SBlockS)))
		if err != nil {
			continue
		}

		for _, entry := range dirBlock.BContent {
			if entry.BInodo == -1 {
				continue
			}

			entryName := strings.Trim(string(entry.BName[:]), "\x00")
			if entryName == name {
				return entry.BInodo, true, nil
			}
		}
	}

	return -1, false, nil
}

// searchInDirectory busca recursivamente un archivo o carpeta por nombre en el sistema de archivos
func (sb *SuperBlock) searchInDirectory(
	diskPath string,
	dirInode *INode,
	inodeIndex int32,
	nameToFind string,
	currentPath []string,
	prefix string,
	foundPaths *[]string,
	foundInodes *[]int32,
	visited map[int32]bool,
) error {
	// Evitar ciclos
	if visited[inodeIndex] {
		return nil
	}
	visited[inodeIndex] = true

	// Procesar bloques directos
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
			if entryName == "." || entryName == ".." {
				continue
			}

			// Verificar si este es el archivo o carpeta que estamos buscando
			// usando matchPattern en lugar de coincidencia exacta
			if matchPattern(nameToFind, entryName) {
				// Construir la ruta completa
				fullPath := strings.Join(currentPath, "/") + "/" + entryName
				*foundPaths = append(*foundPaths, fullPath)
				*foundInodes = append(*foundInodes, entry.BInodo)
			}

			// Si es un directorio, buscar recursivamente
			entryInode := &INode{}
			err := entryInode.Deserialize(diskPath, int64(sb.SInodeStart+(entry.BInodo*sb.SInodeS)))
			if err != nil {
				continue
			}

			// Solo continuar la búsqueda en directorios
			if entryInode.IType[0] == '0' && !visited[entry.BInodo] {
				// Actualizar la ruta actual
				newPath := append([]string{}, currentPath...)
				newPath = append(newPath, entryName)

				// Continuar la búsqueda en este directorio
				err = sb.searchInDirectory(
					diskPath,
					entryInode,
					entry.BInodo,
					nameToFind,
					newPath,
					prefix+"|  ",
					foundPaths,
					foundInodes,
					visited,
				)
				if err != nil {
					return err
				}
			}
		}
	}

	// Procesar bloques indirectos si es necesario
	if dirInode.IBlock[12] != -1 {
		err := sb.searchInIndirectBlocks(
			diskPath,
			dirInode.IBlock[12],
			nameToFind,
			currentPath,
			prefix,
			foundPaths,
			foundInodes,
			visited,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// searchInIndirectBlocks busca en bloques indirectos
func (sb *SuperBlock) searchInIndirectBlocks(
	diskPath string,
	blockIndex int32,
	nameToFind string,
	currentPath []string,
	prefix string,
	foundPaths *[]string,
	foundInodes *[]int32,
	visited map[int32]bool,
) error {
	pointerBlock := &PointerBlock{}
	err := pointerBlock.Deserialize(diskPath, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
	if err != nil {
		return err
	}

	for _, ptr := range pointerBlock.PContent {
		if ptr == -1 {
			continue
		}

		dirBlock := &DirBlock{}
		err := dirBlock.Deserialize(diskPath, int64(sb.SBlockStart+(ptr*sb.SBlockS)))
		if err != nil {
			continue
		}

		for _, entry := range dirBlock.BContent {
			if entry.BInodo == -1 {
				continue
			}

			entryName := strings.Trim(string(entry.BName[:]), "\x00")
			if entryName == "." || entryName == ".." {
				continue
			}

			// Verificar si este es el archivo o carpeta que estamos buscando
			// usando matchPattern en lugar de coincidencia exacta
			if matchPattern(nameToFind, entryName) {
				// Construir la ruta completa
				fullPath := strings.Join(currentPath, "/") + "/" + entryName
				*foundPaths = append(*foundPaths, fullPath)
				*foundInodes = append(*foundInodes, entry.BInodo)
			}

			// Si es un directorio, buscar recursivamente
			entryInode := &INode{}
			err := entryInode.Deserialize(diskPath, int64(sb.SInodeStart+(entry.BInodo*sb.SInodeS)))
			if err != nil {
				continue
			}

			// Solo continuar la búsqueda en directorios
			if entryInode.IType[0] == '0' && !visited[entry.BInodo] {
				// Actualizar la ruta actual
				newPath := append([]string{}, currentPath...)
				newPath = append(newPath, entryName)

				// Continuar la búsqueda en este directorio
				err = sb.searchInDirectory(
					diskPath,
					entryInode,
					entry.BInodo,
					nameToFind,
					newPath,
					prefix+"|  ",
					foundPaths,
					foundInodes,
					visited,
				)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// traverseDirectory recorre recursivamente un directorio y sus contenidos
func (sb *SuperBlock) traverseDirectory(
	diskPath string, // Ruta del disco
	dirInode *INode, // Inodo del directorio actual
	inodeIndex int32, // Índice del inodo actual
	prefix string, // Prefijo para la indentación
	output *strings.Builder, // Buffer donde se escribe la salida
	depth int, // Profundidad actual en el árbol
	visited map[int32]bool, // Mapa para evitar ciclos
) error {
	// Marcar este inodo como visitado para evitar ciclos
	visited[inodeIndex] = true

	// Procesar cada bloque de directorios en el inodo
	for i := 0; i < 12; i++ {
		blockIndex := dirInode.IBlock[i]
		if blockIndex == -1 {
			continue
		}

		dirBlock := &DirBlock{}
		err := dirBlock.Deserialize(diskPath, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
		if err != nil {
			return fmt.Errorf("error al leer bloque de directorio: %v", err)
		}

		// Procesar cada entrada en el bloque de directorio
		for _, entry := range dirBlock.BContent {
			if entry.BInodo == -1 {
				continue
			}

			// Obtener nombre de la entrada
			entryName := strings.Trim(string(entry.BName[:]), "\x00")

			// Ignorar "." y ".."
			if entryName == "." || entryName == ".." {
				continue
			}

			// Evitar ciclos
			if visited[entry.BInodo] {
				continue
			}

			// Leer el inodo de la entrada
			entryInode := &INode{}
			err := entryInode.Deserialize(diskPath, int64(sb.SInodeStart+(entry.BInodo*sb.SInodeS)))
			if err != nil {
				return fmt.Errorf("error al leer inodo %d: %v", entry.BInodo, err)
			}

			// Formatear los permisos para mostrarlos en octal
			perms := fmt.Sprintf("#%c%c%c", entryInode.IPerm[0], entryInode.IPerm[1], entryInode.IPerm[2])

			// Mostrar la entrada en el árbol
			output.WriteString(fmt.Sprintf("%s|_ %s %s\n", prefix, entryName, perms))

			// Si es un directorio, recorrer recursivamente
			if entryInode.IType[0] == '0' {
				nextPrefix := prefix + "|  "
				err = sb.traverseDirectory(diskPath, entryInode, entry.BInodo, nextPrefix, output, depth+1, visited)
				if err != nil {
					return err
				}
			}
		}
	}

	// Procesar bloques indirectos si es necesario
	if dirInode.IBlock[12] != -1 {
		err := sb.traverseIndirectBlocks(diskPath, dirInode.IBlock[12], prefix, output, depth, visited)
		if err != nil {
			return err
		}
	}

	return nil
}

// traverseIndirectBlocks procesa los bloques indirectos para encontrar más entradas de directorio
func (sb *SuperBlock) traverseIndirectBlocks(
	diskPath string,
	blockIndex int32,
	prefix string,
	output *strings.Builder,
	depth int,
	visited map[int32]bool,
) error {
	pointerBlock := &PointerBlock{}
	err := pointerBlock.Deserialize(diskPath, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
	if err != nil {
		return fmt.Errorf("error al leer bloque indirecto: %v", err)
	}

	for _, ptr := range pointerBlock.PContent {
		if ptr == -1 {
			continue
		}

		dirBlock := &DirBlock{}
		err := dirBlock.Deserialize(diskPath, int64(sb.SBlockStart+(ptr*sb.SBlockS)))
		if err != nil {
			continue
		}

		// Procesar cada entrada en el bloque de directorio indirecto
		for _, entry := range dirBlock.BContent {
			if entry.BInodo == -1 {
				continue
			}

			// Obtener nombre de la entrada
			entryName := strings.Trim(string(entry.BName[:]), "\x00")

			// Ignorar "." y ".."
			if entryName == "." || entryName == ".." {
				continue
			}

			// Evitar ciclos
			if visited[entry.BInodo] {
				continue
			}

			// Leer el inodo de la entrada
			entryInode := &INode{}
			err := entryInode.Deserialize(diskPath, int64(sb.SInodeStart+(entry.BInodo*sb.SInodeS)))
			if err != nil {
				return fmt.Errorf("error al leer inodo %d: %v", entry.BInodo, err)
			}

			// Formatear los permisos para mostrarlos en octal
			perms := fmt.Sprintf("#%c%c%c", entryInode.IPerm[0], entryInode.IPerm[1], entryInode.IPerm[2])

			// Mostrar la entrada en el árbol
			output.WriteString(fmt.Sprintf("%s|_ %s %s\n", prefix, entryName, perms))

			// Si es un directorio, recorrer recursivamente
			if entryInode.IType[0] == '0' {
				nextPrefix := prefix + "|  "
				err = sb.traverseDirectory(diskPath, entryInode, entry.BInodo, nextPrefix, output, depth+1, visited)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
