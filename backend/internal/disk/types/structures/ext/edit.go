package ext2

import (
	"fmt"
	"os"
	"time"
)

func (sb *SuperBlock) EditFile(partitionPath string, parentDirs []string, fileName string, contentPath string, uid int32, gid int32) error {
	// 1. Buscar inodo del archivo
	inodeIndex, err := sb.FindFileInode(partitionPath, parentDirs, fileName)
	if err != nil {
		return fmt.Errorf("error al buscar el archivo: %v", err)
	}

	// 2. Obtener información del archivo
	fileInode := &INode{}
	err = fileInode.Deserialize(partitionPath, int64(sb.SInodeStart+(inodeIndex*sb.SInodeS)))
	if err != nil {
		return fmt.Errorf("error al leer el inodo del archivo: %v", err)
	}

	// 3. Verificar que sea un archivo (tipo '1')
	if fileInode.IType[0] != '1' {
		return fmt.Errorf("'%s' no es un archivo", fileName)
	}

	// 4. Verificar permisos de escritura
	if !sb.userHasWritePermission(fileInode, uid, gid) {
		return fmt.Errorf("error: no tienes permisos de escritura sobre este archivo")
	}

	// 5. Leer el contenido del archivo local
	newContent, err := os.ReadFile(contentPath)
	if err != nil {
		return fmt.Errorf("error al leer el archivo de contenido '%s': %v", contentPath, err)
	}

	// 6. Verificar que el nuevo contenido no exceda el tamaño original
	if len(newContent) > int(fileInode.ISize) {
		return fmt.Errorf("el nuevo contenido excede el tamaño original del archivo (%d bytes vs %d bytes)", len(newContent), fileInode.ISize)
	}

	// 7. Escribir el nuevo contenido en los bloques existentes
	contentOffset := 0
	contentSize := len(newContent)

	// 7.1 Escribir en bloques directos (0-11)
	for i := 0; i < 12 && fileInode.IBlock[i] != -1 && contentOffset < contentSize; i++ {
		blockIndex := fileInode.IBlock[i]
		fileBlock := &FileBlock{
			BContent: [FileBlockSize]byte{}, // Inicializar con ceros
		}

		// Obtener el bloque existente (por si necesitamos mantener parte de su contenido)
		err = fileBlock.Deserialize(partitionPath, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
		if err != nil {
			return fmt.Errorf("error al leer bloque de archivo %d: %v", blockIndex, err)
		}

		// Calcular cuántos bytes copiar en este bloque
		remainingBytes := contentSize - contentOffset
		bytesToCopy := FileBlockSize
		if remainingBytes < FileBlockSize {
			bytesToCopy = remainingBytes
		}

		// Si hay menos contenido que antes, llenar el resto con ceros
		for j := 0; j < FileBlockSize; j++ {
			if j < bytesToCopy {
				fileBlock.BContent[j] = newContent[contentOffset+j]
			} else {
				fileBlock.BContent[j] = 0
			}
		}

		// Escribir el bloque actualizado
		err = fileBlock.Serialize(partitionPath, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
		if err != nil {
			return fmt.Errorf("error al escribir bloque de archivo %d: %v", blockIndex, err)
		}

		contentOffset += bytesToCopy
		fmt.Printf("Bloque directo #%d actualizado con %d bytes\n", blockIndex, bytesToCopy)
	}

	// 7.2 Escribir en bloques indirectos simples (si es necesario)
	if contentOffset < contentSize && fileInode.IBlock[12] != -1 {
		// Obtener el bloque de punteros indirectos
		pointerBlock := &PointerBlock{}
		err = pointerBlock.Deserialize(partitionPath, int64(sb.SBlockStart+(fileInode.IBlock[12]*sb.SBlockS)))
		if err != nil {
			return fmt.Errorf("error al leer bloque de punteros: %v", err)
		}

		// Iterar sobre los punteros
		for i := 0; i < len(pointerBlock.PContent) && pointerBlock.PContent[i] != -1 && contentOffset < contentSize; i++ {
			blockIndex := pointerBlock.PContent[i]
			fileBlock := &FileBlock{
				BContent: [FileBlockSize]byte{}, // Inicializar con ceros
			}

			// Obtener el bloque existente
			err = fileBlock.Deserialize(partitionPath, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
			if err != nil {
				return fmt.Errorf("error al leer bloque indirecto %d: %v", blockIndex, err)
			}

			// Calcular cuántos bytes copiar en este bloque
			remainingBytes := contentSize - contentOffset
			bytesToCopy := FileBlockSize
			if remainingBytes < FileBlockSize {
				bytesToCopy = remainingBytes
			}

			// Si hay menos contenido que antes, llenar el resto con ceros
			for j := 0; j < FileBlockSize; j++ {
				if j < bytesToCopy {
					fileBlock.BContent[j] = newContent[contentOffset+j]
				} else {
					fileBlock.BContent[j] = 0
				}
			}

			// Escribir el bloque actualizado
			err = fileBlock.Serialize(partitionPath, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
			if err != nil {
				return fmt.Errorf("error al escribir bloque indirecto %d: %v", blockIndex, err)
			}

			contentOffset += bytesToCopy
			fmt.Printf("Bloque indirecto simple #%d actualizado con %d bytes\n", blockIndex, bytesToCopy)
		}
	}

	// 7.3 Si hay bloques indirectos dobles, procesarlos (si es necesario)
	if contentOffset < contentSize && fileInode.IBlock[13] != -1 {
		// Obtener el bloque de punteros dobles
		doublePointerBlock := &PointerBlock{}
		err = doublePointerBlock.Deserialize(partitionPath, int64(sb.SBlockStart+(fileInode.IBlock[13]*sb.SBlockS)))
		if err != nil {
			return fmt.Errorf("error al leer bloque de punteros dobles: %v", err)
		}

		// Iterar sobre los bloques de punteros simples
		for i := 0; i < len(doublePointerBlock.PContent) && doublePointerBlock.PContent[i] != -1 && contentOffset < contentSize; i++ {
			simplePointerBlockIndex := doublePointerBlock.PContent[i]
			simplePointerBlock := &PointerBlock{}

			// Obtener el bloque de punteros simples
			err = simplePointerBlock.Deserialize(partitionPath, int64(sb.SBlockStart+(simplePointerBlockIndex*sb.SBlockS)))
			if err != nil {
				return fmt.Errorf("error al leer bloque de punteros simples dentro de dobles: %v", err)
			}

			// Iterar sobre los bloques de datos
			for j := 0; j < len(simplePointerBlock.PContent) && simplePointerBlock.PContent[j] != -1 && contentOffset < contentSize; j++ {
				blockIndex := simplePointerBlock.PContent[j]
				fileBlock := &FileBlock{
					BContent: [FileBlockSize]byte{}, // Inicializar con ceros
				}

				// Obtener el bloque existente
				err = fileBlock.Deserialize(partitionPath, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
				if err != nil {
					return fmt.Errorf("error al leer bloque indirecto doble %d: %v", blockIndex, err)
				}

				// Calcular cuántos bytes copiar en este bloque
				remainingBytes := contentSize - contentOffset
				bytesToCopy := FileBlockSize
				if remainingBytes < FileBlockSize {
					bytesToCopy = remainingBytes
				}

				// Si hay menos contenido que antes, llenar el resto con ceros
				for k := 0; k < FileBlockSize; k++ {
					if k < bytesToCopy {
						fileBlock.BContent[k] = newContent[contentOffset+k]
					} else {
						fileBlock.BContent[k] = 0
					}
				}

				// Escribir el bloque actualizado
				err = fileBlock.Serialize(partitionPath, int64(sb.SBlockStart+(blockIndex*sb.SBlockS)))
				if err != nil {
					return fmt.Errorf("error al escribir bloque indirecto doble %d: %v", blockIndex, err)
				}

				contentOffset += bytesToCopy
				fmt.Printf("Bloque indirecto doble #%d actualizado con %d bytes\n", blockIndex, bytesToCopy)
			}
		}
	}

	// 7.4 Si hay bloques indirectos triples, procesarlos (si es necesario)
	if contentOffset < contentSize && fileInode.IBlock[14] != -1 {
		// Lógica similar a los bloques indirectos dobles, pero con un nivel más de indirección
		// Esta parte se implementaría si fuera necesario (raramente se usarían bloques triples para archivos pequeños)
		fmt.Println("Advertencia: Edición en bloques indirectos triples no implementada")
	}

	// 8. Actualizar el tamaño del archivo y timestamp de modificación
	fileInode.ISize = int32(len(newContent))
	fileInode.IMtime = float32(time.Now().Unix())

	// 9. Guardar el inodo actualizado
	err = fileInode.Serialize(partitionPath, int64(sb.SInodeStart+(inodeIndex*sb.SInodeS)))
	if err != nil {
		return fmt.Errorf("error al actualizar el inodo del archivo: %v", err)
	}

	fmt.Printf("Archivo '%s' editado exitosamente (nuevo tamaño: %d bytes)\n", fileName, len(newContent))
	return nil
}
