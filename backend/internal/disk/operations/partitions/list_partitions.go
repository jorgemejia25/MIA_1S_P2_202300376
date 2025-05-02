// filepath: /home/jorgis/Documents/USAC/archivos/proyecto2/backend/internal/disk/operations/partitions/list_partitions.go
package partition_operations

import (
	"fmt"
	"os"
	"strings"

	"disk.simulator.com/m/v2/internal/disk/types/structures"
)

// PartitionInfo contiene información de una partición para mostrar al usuario
type PartitionInfo struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Fit       string `json:"fit"`
	Start     int32  `json:"start"`
	Size      int32  `json:"size"`
	IsMounted bool   `json:"isMounted"`
	MountID   string `json:"mountId"`
}

// LogicalPartitionInfo contiene información de una partición lógica
type LogicalPartitionInfo struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Fit       string `json:"fit"`
	Start     int32  `json:"start"`
	Size      int32  `json:"size"`
	Next      int32  `json:"next"`
	IsMounted bool   `json:"isMounted"`
	MountID   string `json:"mountId"`
}

// ListPartitions obtiene una lista de todas las particiones en un disco
func ListPartitions(path string) ([]PartitionInfo, []LogicalPartitionInfo, error) {
	// Verificar que el archivo exista
	_, err := os.Stat(path)
	if err != nil {
		return nil, nil, fmt.Errorf("el disco no existe en la ruta: %s", path)
	}

	// Leer el MBR del disco
	var mbr structures.MBR
	err = mbr.DeserializeMBR(path)
	if err != nil {
		return nil, nil, fmt.Errorf("error al deserializar el MBR: %v", err)
	}

	var partitions []PartitionInfo
	var logicalPartitions []LogicalPartitionInfo

	// Procesar las particiones primarias y extendidas
	for _, partition := range mbr.Mbr_partitions {
		// Verificar si la partición está activa (no eliminada)
		if partition.Part_status != 'N' && partition.Part_size > 0 {
			// Determinar el tipo de partición
			partType := ""
			if partition.Part_type == 'P' {
				partType = "Primaria"
			} else if partition.Part_type == 'E' {
				partType = "Extendida"
			} else {
				partType = "Desconocido"
			}

			// Determinar el estado de la partición
			partStatus := ""
			if partition.Part_status == '1' {
				partStatus = "Activa"
			} else {
				partStatus = "Inactiva"
			}

			// Determinar el tipo de ajuste
			partFit := ""
			switch partition.Part_fit {
			case 'F':
				partFit = "First Fit"
			case 'B':
				partFit = "Best Fit"
			case 'W':
				partFit = "Worst Fit"
			default:
				partFit = "Desconocido"
			}

			// Determinar si está montada
			isMounted := partition.Part_mount == '1'

			// Crear el ID de montaje si está montada
			mountID := ""
			if isMounted {
				mountID = string(partition.Part_id[:])
				mountID = strings.Trim(mountID, "\x00")
			}

			// Obtener el nombre limpio de la partición
			name := string(partition.Part_name[:])
			name = strings.Trim(name, "\x00")

			// Agregar la información de la partición a la lista
			partInfo := PartitionInfo{
				Name:      name,
				Type:      partType,
				Status:    partStatus,
				Fit:       partFit,
				Start:     partition.Part_start,
				Size:      partition.Part_size,
				IsMounted: isMounted,
				MountID:   mountID,
			}
			partitions = append(partitions, partInfo)

			// Si es una partición extendida, buscar particiones lógicas
			if partition.Part_type == 'E' {
				logicalParts, err := getLogicalPartitions(path, partition.Part_start)
				if err != nil {
					// No detener la ejecución si hay un error con las particiones lógicas
					// Solo registrar el error y continuar
					fmt.Printf("Error al leer particiones lógicas: %v\n", err)
				} else {
					logicalPartitions = append(logicalPartitions, logicalParts...)
				}
			}
		}
	}

	return partitions, logicalPartitions, nil
}

// getLogicalPartitions obtiene las particiones lógicas dentro de una partición extendida
func getLogicalPartitions(path string, start int32) ([]LogicalPartitionInfo, error) {
	var logicalPartitions []LogicalPartitionInfo
	var currentEBR structures.EBR

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer file.Close()

	// Leer la primera EBR
	err = currentEBR.DeserializeEBR(path, start)
	if err != nil {
		return nil, fmt.Errorf("error al leer el EBR: %v", err)
	}

	// Recorrer la lista enlazada de EBRs
	for {
		// Verificar si la partición lógica está activa (tiene tamaño > 0)
		if currentEBR.Part_size > 0 {
			// Determinar el tipo de ajuste
			partFit := ""
			switch currentEBR.Part_fit {
			case 'F':
				partFit = "First Fit"
			case 'B':
				partFit = "Best Fit"
			case 'W':
				partFit = "Worst Fit"
			default:
				partFit = "Desconocido"
			}

			// Determinar si está montada
			isMounted := currentEBR.Part_mount == 1

			// Obtener el nombre limpio de la partición
			name := string(currentEBR.Part_name[:])
			name = strings.Trim(name, "\x00")

			// Agregar la información de la partición lógica a la lista
			partInfo := LogicalPartitionInfo{
				Name:      name,
				Status:    "Activa",
				Fit:       partFit,
				Start:     currentEBR.Part_start,
				Size:      currentEBR.Part_size,
				Next:      currentEBR.Part_next,
				IsMounted: isMounted,
				MountID:   "", // La estructura EBR no tiene un campo Part_id según el código actual
			}
			logicalPartitions = append(logicalPartitions, partInfo)
		}

		// Si no hay más EBRs, terminar el recorrido
		if currentEBR.Part_next <= 0 {
			break
		}

		// Leer el siguiente EBR
		err = currentEBR.DeserializeEBR(path, currentEBR.Part_next)
		if err != nil {
			return logicalPartitions, fmt.Errorf("error al leer el siguiente EBR: %v", err)
		}
	}

	return logicalPartitions, nil
}

// GetPartitionsInfo devuelve un string formateado con la información de las particiones
func GetPartitionsInfo(path string) (string, error) {
	partitions, logicalPartitions, err := ListPartitions(path)
	if err != nil {
		return "", err
	}

	// Formatear la salida
	output := fmt.Sprintf("PARTICIONES DEL DISCO: %s\n\n", path)

	// Mostrar particiones primarias y extendidas
	output += "PARTICIONES PRIMARIAS Y EXTENDIDAS:\n"
	output += fmt.Sprintf("%-20s | %-10s | %-10s | %-15s | %-15s | %-10s | %-10s\n",
		"NOMBRE", "TIPO", "ESTADO", "INICIO (bytes)", "TAMAÑO (bytes)", "MONTADA", "ID MONTAJE")
	output += fmt.Sprintf("%s\n", "----------------------------------------------------------------------------------------------------")

	if len(partitions) == 0 {
		output += "No hay particiones primarias o extendidas en este disco.\n"
	} else {
		for _, part := range partitions {
			mountStatus := "No"
			if part.IsMounted {
				mountStatus = "Sí"
			}

			output += fmt.Sprintf("%-20s | %-10s | %-10s | %-15d | %-15d | %-10s | %-10s\n",
				part.Name, part.Type, part.Status, part.Start, part.Size, mountStatus, part.MountID)
		}
	}

	// Mostrar particiones lógicas si existen
	output += "\nPARTICIONES LÓGICAS:\n"
	output += fmt.Sprintf("%-20s | %-10s | %-15s | %-15s | %-10s | %-10s\n",
		"NOMBRE", "ESTADO", "INICIO (bytes)", "TAMAÑO (bytes)", "MONTADA", "ID MONTAJE")
	output += fmt.Sprintf("%s\n", "---------------------------------------------------------------------------------")

	if len(logicalPartitions) == 0 {
		output += "No hay particiones lógicas en este disco.\n"
	} else {
		for _, part := range logicalPartitions {
			mountStatus := "No"
			if part.IsMounted {
				mountStatus = "Sí"
			}

			output += fmt.Sprintf("%-20s | %-10s | %-15d | %-15d | %-10s | %-10s\n",
				part.Name, part.Status, part.Start, part.Size, mountStatus, part.MountID)
		}
	}

	return output, nil
}
