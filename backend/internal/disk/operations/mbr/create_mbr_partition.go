package mbr_operations

import (
	"bytes"
	"fmt"

	"disk.simulator.com/m/v2/internal/disk/types"
	"disk.simulator.com/m/v2/internal/disk/types/structures"
	"disk.simulator.com/m/v2/utils"
)

// Cambiar la firma de la función para retornar también la partición
func CreateMBRPartition(params types.FDisk) (structures.Partition, error) {
	// Leer el MBR actual
	mbr := structures.MBR{}
	err := mbr.DeserializeMBR(params.Path)
	if err != nil {
		return structures.Partition{}, fmt.Errorf("error al leer el MBR: %v", err)
	}

	// Validar que no exista una partición con el mismo nombre
	for _, part := range mbr.Mbr_partitions {
		if part.Part_status != 'N' {
			partName := string(bytes.Trim(part.Part_name[:], "\x00"))
			if partName == params.Name {
				return structures.Partition{}, fmt.Errorf("ya existe una partición con el nombre '%s'", params.Name)
			}
		}
	}

	// Validación mejorada para particiones extendidas
	hasExtended := false
	for _, part := range mbr.Mbr_partitions {
		if part.Part_type == 'E' {
			hasExtended = true
			break
		}
	}

	// Validar según el tipo de partición
	switch params.Type {
	case "E":
		if hasExtended {
			return structures.Partition{}, fmt.Errorf("ya existe una partición extendida en el disco, solo se permite una")
		}
	case "L":
		if !hasExtended {
			return structures.Partition{}, fmt.Errorf("no se puede crear una partición lógica sin una partición extendida")
		}
	case "P":
		// Las particiones primarias no necesitan validación especial
	default:
		return structures.Partition{}, fmt.Errorf("tipo de partición no válido: %s", params.Type)
	}

	// Convertir el tamaño a bytes
	sizeInBytes, err := utils.ConvertToBytes(params.Size, params.Unit)
	if err != nil {
		return structures.Partition{}, fmt.Errorf("error al convertir el tamaño: %v", err)
	}

	// Calcular la posición de inicio para la nueva partición
	startByte := int32(structures.MBRSize) // Comenzar después del MBR (153 bytes)
	partitionIndex := -1

	// Encontrar el siguiente espacio disponible y calcular la posición de inicio
	for i, part := range mbr.Mbr_partitions {
		if part.Part_status == 'N' && partitionIndex == -1 {
			partitionIndex = i
			// Si no es la primera partición, necesitamos calcular el inicio correcto
			if i > 0 {
				// Buscar la última partición activa antes de este índice
				for j := i - 1; j >= 0; j-- {
					if mbr.Mbr_partitions[j].Part_status != 'N' {
						startByte = mbr.Mbr_partitions[j].Part_start + mbr.Mbr_partitions[j].Part_size
						break
					}
				}
			}
		}
	}

	if partitionIndex == -1 {
		return structures.Partition{}, fmt.Errorf("no hay espacios disponibles para más particiones")
	}

	// Crear la nueva partición con la posición de inicio calculada
	newPartition := structures.Partition{
		Part_status: '1', // Cambié a '1' para indicar que está activa
		Part_type:   params.Type[0],
		Part_fit:    params.Fit[0],
		Part_start:  startByte,
		Part_size:   int32(sizeInBytes),
	}

	// Copiar el nombre (máximo 16 bytes)
	copy(newPartition.Part_name[:], []byte(params.Name))

	// Asignar la nueva partición al MBR
	mbr.Mbr_partitions[partitionIndex] = newPartition

	// Escribir el MBR actualizado
	err = mbr.SerializeMBR(params.Path)
	if err != nil {
		return structures.Partition{}, fmt.Errorf("error al escribir el MBR: %v", err)
	}

	return newPartition, nil
}
