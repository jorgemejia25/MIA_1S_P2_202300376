package partition_operations

import (
	"fmt"
	"os"
	"strings"

	"disk.simulator.com/m/v2/internal/disk/memory"
	ext2 "disk.simulator.com/m/v2/internal/disk/types/structures/ext"
)

// SimulateSystemLoss simula un fallo en el sistema formateando áreas críticas con caracteres nulos
func SimulateSystemLoss(id string) (string, error) {
	var output strings.Builder

	// Obtener la partición montada
	partition, path, err := memory.GetInstance().GetMountedPartition(id)
	if err != nil {
		return "", fmt.Errorf("error al obtener la partición: %v", err)
	}

	// Leer el superbloque
	superBlock := ext2.SuperBlock{}
	err = superBlock.DeserializeSuperBlock(path, partition.Partition.Part_start)
	if err != nil {
		return "", fmt.Errorf("error al leer el superbloque: %v", err)
	}

	// Abrir el archivo en modo escritura
	file, err := os.OpenFile(path, os.O_WRONLY, 0666)
	if err != nil {
		return "", fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer file.Close()

	output.WriteString(fmt.Sprintf("Simulando pérdida de sistema de archivos en la partición %s (ID: %s)\n", partition.Name, id))

	// 1. Limpiar bloque de bitmap de Inodos
	output.WriteString("1. Limpiando bitmap de inodos...\n")
	bitmapInodeSize := superBlock.SFreeInodesCount // Tamaño del bitmap de inodos
	err = cleanArea(file, int64(superBlock.SBmInodeStart), int64(bitmapInodeSize))
	if err != nil {
		return output.String(), fmt.Errorf("error al limpiar bitmap de inodos: %v", err)
	}

	// 2. Limpiar bloque de bitmap de Bloques
	output.WriteString("2. Limpiando bitmap de bloques...\n")
	bitmapBlockSize := superBlock.SFreeBlocksCount // Tamaño del bitmap de bloques
	err = cleanArea(file, int64(superBlock.SBmBlockStart), int64(bitmapBlockSize))
	if err != nil {
		return output.String(), fmt.Errorf("error al limpiar bitmap de bloques: %v", err)
	}

	// 3. Limpiar área de Inodos
	output.WriteString("3. Limpiando área de inodos...\n")
	inodeAreaSize := superBlock.SInodesCount * superBlock.SInodeS // Número de inodos * tamaño de un inodo
	err = cleanArea(file, int64(superBlock.SInodeStart), int64(inodeAreaSize))
	if err != nil {
		return output.String(), fmt.Errorf("error al limpiar área de inodos: %v", err)
	}

	// 4. Limpiar área de Bloques
	output.WriteString("4. Limpiando área de bloques...\n")
	blockAreaSize := superBlock.SBlocksCount * superBlock.SBlockS // Número de bloques * tamaño de un bloque
	err = cleanArea(file, int64(superBlock.SBlockStart), int64(blockAreaSize))
	if err != nil {
		return output.String(), fmt.Errorf("error al limpiar área de bloques: %v", err)
	}

	output.WriteString("Simulación de pérdida de sistema completada exitosamente.\n")
	output.WriteString("Utilice el comando 'recovery -id=" + id + "' para recuperar los datos desde el journaling.\n")

	return output.String(), nil
}

// cleanArea sobrescribe un área del archivo con caracteres nulos
func cleanArea(file *os.File, offset int64, size int64) error {
	// Si el tamaño es muy grande, podría dividirse en chunks para evitar problemas de memoria
	bufferSize := 8192 // 8KB chunks
	buffer := make([]byte, bufferSize)

	// Posicionar el cursor en el offset indicado
	_, err := file.Seek(offset, 0)
	if err != nil {
		return fmt.Errorf("error al posicionarse en offset %d: %v", offset, err)
	}

	// Escribir bytes nulos en bloques
	remaining := size
	for remaining > 0 {
		// Determinar el tamaño a escribir en esta iteración
		writeSize := int64(bufferSize)
		if remaining < int64(bufferSize) {
			writeSize = remaining
			buffer = make([]byte, writeSize) // Redimensionar el buffer
		}

		// Escribir el buffer lleno de ceros
		_, err = file.Write(buffer)
		if err != nil {
			return fmt.Errorf("error al escribir bytes nulos: %v", err)
		}

		remaining -= writeSize
	}

	return nil
}
