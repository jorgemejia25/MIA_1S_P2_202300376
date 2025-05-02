package partition_operations

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"disk.simulator.com/m/v2/internal/disk/memory"
	ext2 "disk.simulator.com/m/v2/internal/disk/types/structures/ext"
	"disk.simulator.com/m/v2/utils"
)

// FileInfo contiene información acerca de un archivo o carpeta
type FileInfo struct {
	Name        string    `json:"name"`        // Nombre del archivo o carpeta
	Type        string    `json:"type"`        // "file" o "directory"
	Size        int32     `json:"size"`        // Tamaño en bytes
	Permissions string    `json:"permissions"` // Permisos en formato octal (ej. "777")
	Owner       int32     `json:"owner"`       // ID del propietario
	Group       int32     `json:"group"`       // ID del grupo
	ModTime     time.Time `json:"modTime"`     // Fecha de modificación
	InodeID     int32     `json:"inodeId"`     // ID del inodo
}

// DirectoryContent representa el contenido de un directorio
type DirectoryContent struct {
	Path     string     `json:"path"`     // Ruta del directorio listado
	Files    []FileInfo `json:"files"`    // Lista de archivos en el directorio
	Success  bool       `json:"success"`  // Indicador de éxito de la operación
	ErrorMsg string     `json:"errorMsg"` // Mensaje de error si ocurre alguno
}

// ListDirectory lista el contenido de un directorio en formato JSON
func ListDirectory(diskPath string, partitionName string, dirPath string) (string, error) {
	// Verificar que la partición existe
	partition, partitionIndex, err := FindPartition(partitionName, diskPath)
	if err != nil {
		return "", fmt.Errorf("error al buscar partición: %v", err)
	}

	if partitionIndex == -1 {
		return "", fmt.Errorf("la partición '%s' no existe en el disco '%s'", partitionName, diskPath)
	}

	// Montar la partición temporalmente si no está montada
	var id string
	var isTempMount bool
	storage := memory.GetInstance()
	isMounted, index := storage.IsPartitionMounted(partitionName, diskPath)

	if !isMounted {
		// Montar temporalmente
		var mountErr error
		id, mountErr = storage.MountPartition(partitionName, diskPath, partition)
		if mountErr != nil {
			return "", fmt.Errorf("error al montar temporalmente la partición: %v", mountErr)
		}
		isTempMount = true
	} else {
		id = storage.GetMountedPartitions()[index].ID
		isTempMount = false
	}

	// Limpiar la ruta del directorio
	dirPath = strings.TrimSpace(dirPath)
	if dirPath == "" {
		dirPath = "/"
	}

	// Separar ruta en directorios padres y nombre del directorio final
	parentDirs, dirName := utils.GetParentDirectories(dirPath)

	// Leer superbloque
	mountedPartition, partitionPath, err := storage.GetMountedPartition(id)
	if err != nil {
		return "", fmt.Errorf("error al obtener partición: %v", err)
	}

	superBlock := ext2.SuperBlock{}
	err = superBlock.DeserializeSuperBlock(partitionPath, mountedPartition.Partition.Part_start)
	if err != nil {
		return "", fmt.Errorf("error al leer superbloque: %v", err)
	}

	// Manejar caso especial para la raíz
	var dirInodeIndex int32

	if dirPath == "/" {
		dirInodeIndex = 0 // Inodo raíz
	} else {
		// Si dirName está vacío, significa que estamos buscando el último elemento en parentDirs
		if dirName == "" && len(parentDirs) > 0 {
			// Extraer el último elemento de parentDirs
			dirName = parentDirs[len(parentDirs)-1]
			parentDirs = parentDirs[:len(parentDirs)-1]
		}

		// Buscar el inodo del directorio
		var findErr error
		dirInodeIndex, findErr = superBlock.FindFileInode(partitionPath, parentDirs, dirName)
		if findErr != nil {
			return "", fmt.Errorf("error al buscar el directorio '%s': %v", dirPath, findErr)
		}
	}

	// Leer el inodo del directorio
	dirInode := &ext2.INode{}
	err = dirInode.Deserialize(partitionPath, int64(superBlock.SInodeStart+(dirInodeIndex*superBlock.SInodeS)))
	if err != nil {
		return "", fmt.Errorf("error al leer inodo del directorio: %v", err)
	}

	// Verificar que es un directorio
	if dirInode.IType[0] != '0' {
		return "", fmt.Errorf("'%s' no es un directorio", dirPath)
	}

	// Preparar respuesta
	response := DirectoryContent{
		Path:    dirPath,
		Files:   []FileInfo{},
		Success: true,
	}

	// Procesar cada bloque del directorio
	for _, blockIndex := range dirInode.IBlock {
		if blockIndex == -1 {
			continue
		}

		dirBlock := &ext2.DirBlock{}
		err := dirBlock.Deserialize(partitionPath, int64(superBlock.SBlockStart+(blockIndex*superBlock.SBlockS)))
		if err != nil {
			return "", fmt.Errorf("error al leer bloque %d: %v", blockIndex, err)
		}

		// Procesar cada entrada en el bloque
		for _, entry := range dirBlock.BContent {
			if entry.BInodo == -1 {
				continue
			}

			name := strings.TrimRight(string(entry.BName[:]), "\x00")
			if name != "." && name != ".." {
				// Leer el inodo de la entrada
				entryInode := &ext2.INode{}
				err := entryInode.Deserialize(partitionPath, int64(superBlock.SInodeStart+(entry.BInodo*superBlock.SInodeS)))
				if err != nil {
					continue // Ignorar entradas con error
				}

				// Determinar tipo de entrada
				fileType := "file"
				if entryInode.IType[0] == '0' {
					fileType = "directory"
				}

				// Formatear permisos
				perms := string(entryInode.IPerm[:])

				// Fecha de modificación
				modTime := time.Unix(int64(entryInode.IMtime), 0)

				// Agregar a la lista de archivos
				fileInfo := FileInfo{
					Name:        name,
					Type:        fileType,
					Size:        entryInode.ISize,
					Permissions: perms,
					Owner:       entryInode.IUid,
					Group:       entryInode.IGid,
					ModTime:     modTime,
					InodeID:     entry.BInodo,
				}

				response.Files = append(response.Files, fileInfo)
			}
		}
	}

	// Desmontar la partición si se montó temporalmente
	if isTempMount {
		_ = storage.UnmountPartition(id)
	}

	// Convertir la respuesta a JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("error al convertir a JSON: %v", err)
	}

	return string(jsonResponse), nil
}
