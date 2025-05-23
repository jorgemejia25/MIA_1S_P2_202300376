package partition_operations

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	mbr_operations "disk.simulator.com/m/v2/internal/disk/operations/mbr"
	"disk.simulator.com/m/v2/internal/disk/types"
	"disk.simulator.com/m/v2/internal/disk/types/structures"
)

// CreatePartition crea una nueva partición en el disco según los parámetros especificados.
// Puede crear particiones primarias, extendidas o lógicas.
//
// Parámetros:
//   - params: estructura con los parámetros de la partición (nombre, tamaño, tipo, etc.)
//
// Retorna un error si hay problemas durante la creación de la partición
func CreatePartition(params types.FDisk) error {
	// Obtener el tamaño del disco en bytes
	fileInfo, err := os.Stat(params.Path)
	if err != nil {
		return fmt.Errorf("error al obtener el tamaño del disco: %v", err)
	}
	diskSize := fileInfo.Size()
	fmt.Printf("Tamaño del disco: %d bytes\n", diskSize)

	// Leer el MBR del disco
	var mbr structures.MBR
	err = mbr.DeserializeMBR(params.Path)
	if err != nil {
		return fmt.Errorf("error al leer el MBR: %v", err)
	}

	// Verificar si ya existe una partición con el mismo nombre
	for _, part := range mbr.Mbr_partitions {
		if part.Part_size > 0 {
			partName := string(bytes.Trim(part.Part_name[:], "\x00"))
			if strings.TrimSpace(partName) == strings.TrimSpace(params.Name) {
				return fmt.Errorf("ya existe una partición con el nombre '%s'", params.Name)
			}
		}
	}

	// Si es una partición lógica, verificar en las particiones extendidas también
	if params.Type == "L" {
		extended, _, err := mbr_operations.FindExtendedPartition(params.Path)
		if err == nil {
			currentEBR := structures.EBR{}
			currentPos := extended.Part_start

			for {
				err = currentEBR.DeserializeEBR(params.Path, currentPos)
				if err != nil {
					break
				}

				if currentEBR.Part_size > 0 {
					ebrName := string(bytes.Trim(currentEBR.Part_name[:], "\x00"))
					if strings.TrimSpace(ebrName) == strings.TrimSpace(params.Name) {
						return fmt.Errorf("ya existe una partición lógica con el nombre '%s'", params.Name)
					}
				}

				if currentEBR.Part_next <= 0 {
					break
				}
				currentPos = currentEBR.Part_next
			}
		}
	}

	// Verificar que no haya más de 4 particiones en el MBR
	partitionCount := 0
	for _, partition := range mbr.Mbr_partitions {
		if partition.Part_size > 0 {
			partitionCount++
		}
	}
	if partitionCount >= 4 {
		return fmt.Errorf("no se pueden crear más de 4 particiones en el MBR")
	}

	// Calcular el espacio ocupado por las particiones existentes
	usedSpace := int64(structures.MBRSize) // Espacio ocupado por el MBR
	for _, partition := range mbr.Mbr_partitions {
		if partition.Part_size > 0 {
			usedSpace += int64(partition.Part_size)
		}
	}

	// Calcular el tamaño de la nueva partición en bytes
	var partitionSize int64
	switch params.Unit {
	case "B":
		partitionSize = int64(params.Size)
	case "K":
		partitionSize = int64(params.Size) * 1024
	case "M":
		partitionSize = int64(params.Size) * 1024 * 1024
	default:
		return fmt.Errorf("unidad desconocida: %s", params.Unit)
	}

	// Buscar espacio disponible para la nueva partición
	availableStart := int64(structures.MBRSize)
	for _, partition := range mbr.Mbr_partitions {
		if partition.Part_size > 0 {
			partitionEnd := int64(partition.Part_start) + int64(partition.Part_size)
			if partitionEnd > availableStart {
				availableStart = partitionEnd
			}
		}
	}

	// Verificar si hay espacio suficiente después del último bloque ocupado
	if availableStart+partitionSize > diskSize {
		return fmt.Errorf("no hay suficiente espacio contiguo en el disco para crear la partición. Espacio disponible desde %d bytes, espacio requerido: %d bytes",
			availableStart, partitionSize)
	}

	// Asignar el inicio de la nueva partición
	params.Start = int(availableStart)

	// Continuar con la creación de la partición
	if params.Type == "L" {
		// Buscar la partición extendida
		extended, _, err := mbr_operations.FindExtendedPartition(params.Path)
		if err != nil {
			return fmt.Errorf("no se encontró una partición extendida")
		}

		fmt.Printf("Partición extendida encontrada en %d\n", extended.Part_start)

		// Crear la partición lógica dentro de la partición extendida
		logicalPartition, err := CreateLogicalPartition(params, extended.Part_start)
		if err != nil {
			return fmt.Errorf("error al crear la partición lógica: %v", err)
		}

		fmt.Printf("Partición lógica creada: %v\n", logicalPartition.Part_start)
		return nil
	}

	// Crear partición primaria o extendida
	partition, err := mbr_operations.CreateMBRPartition(params)
	if err != nil {
		return fmt.Errorf("error al crear la partición: %v", err)
	}

	fmt.Printf("Partición creada: %v\n", partition.Part_start)

	if partition.Part_type == 'E' {
		// Crear el EBR inicial para la partición extendida
		ebr := structures.EBR{
			Part_mount: 'N',
			Part_fit:   'N',
			Part_start: partition.Part_start,
			Part_size:  -1,
			Part_next:  -1,
			Part_name:  [16]byte{'N'},
		}

		err = ebr.SerializeEBR(params.Path, partition.Part_start)
		if err != nil {
			return fmt.Errorf("error al crear el EBR: %v", err)
		}

		fmt.Printf("EBR creado en %d\n", partition.Part_start)
	}

	return nil
}
