package reports

import (
	"fmt"
	"strings"
	"time"

	"disk.simulator.com/m/v2/internal/disk/memory"
	ext2 "disk.simulator.com/m/v2/internal/disk/types/structures/ext"
)

// JournalingReport genera un reporte en formato texto que muestra todas las transacciones realizadas en el sistema de archivos
func JournalingReport(outputPath string, id string) (string, error) {
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

	// Verificar si es ext3 (tiene journaling)
	journalStart := partition.Partition.Part_start + ext2.SuperBlockSize

	// Simplemente verificamos el tipo de sistema de archivos
	if superBlock.SFilesystemType != 3 {
		return "", fmt.Errorf("la partición no tiene journaling (no es ext3)")
	}

	// Obtener todas las entradas del journal
	journals, err := ext2.GetJournaling(path, int64(journalStart), superBlock.SFreeInodesCount)
	if err != nil {
		return "", fmt.Errorf("error al obtener el journaling: %v", err)
	}

	// Construir el reporte en formato texto
	var reportBuilder strings.Builder

	reportBuilder.WriteString("=============================================\n")
	reportBuilder.WriteString("            REPORTE DE JOURNALING            \n")
	reportBuilder.WriteString("=============================================\n")
	reportBuilder.WriteString(fmt.Sprintf("Partición: %s (ID: %s)\n", partition.Name, id))
	reportBuilder.WriteString(fmt.Sprintf("Tipo de sistema de archivos: EXT%d\n", superBlock.SFilesystemType))
	reportBuilder.WriteString(fmt.Sprintf("Número total de transacciones: %d\n", len(journals)))
	reportBuilder.WriteString("=============================================\n\n")

	for i, journal := range journals {
		// Convertir el tiempo a formato legible
		date := time.Unix(int64(journal.J_content.I_date), 0)
		formattedDate := date.Format("02/01/2006 15:04:05")

		// Extraer información de la transacción
		operation := string(journal.J_content.I_operation[:])
		filePath := string(journal.J_content.I_path[:])
		content := string(journal.J_content.I_content[:])

		// Limpiar strings eliminando caracteres nulos
		operation = cleanString(operation)
		filePath = cleanString(filePath)
		content = cleanString(content)

		reportBuilder.WriteString(fmt.Sprintf("TRANSACCIÓN #%d\n", i))
		reportBuilder.WriteString(fmt.Sprintf("- Operación:  %s\n", operation))
		reportBuilder.WriteString(fmt.Sprintf("- Ruta:       %s\n", filePath))
		reportBuilder.WriteString(fmt.Sprintf("- Contenido:  %s\n", content))
		reportBuilder.WriteString(fmt.Sprintf("- Fecha/hora: %s\n", formattedDate))
		reportBuilder.WriteString("---------------------------------------------\n\n")
	}

	if len(journals) == 0 {
		reportBuilder.WriteString("No se encontraron transacciones en el journaling.\n")
	}

	reportText := reportBuilder.String()
	return reportText, nil
}

// Función auxiliar para limpiar strings eliminando caracteres nulos
func cleanString(s string) string {
	return strings.TrimRight(s, "\x00")
}
