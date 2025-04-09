package partition_operations

import (
	"bytes"
	"fmt"

	"disk.simulator.com/m/v2/internal/disk/types"
	"disk.simulator.com/m/v2/internal/disk/types/structures"
	"disk.simulator.com/m/v2/utils"
)

// validatePartitionName verifica si ya existe una partición con el nombre especificado.
// Busca en particiones primarias, extendidas y lógicas.
//
// Parámetros:
//   - path: ruta del archivo de disco
//   - name: nombre a validar
//   - extendedStart: posición inicial de la partición extendida
//
// Retorna un error si ya existe una partición con el nombre especificado
func validatePartitionName(path string, name string, extendedStart int32) error {
	// Primero validar contra particiones primarias y extendida en el MBR
	mbr := structures.MBR{}
	err := mbr.DeserializeMBR(path)
	if err != nil {
		return fmt.Errorf("error al leer el MBR: %v", err)
	}

	for _, part := range mbr.Mbr_partitions {
		if part.Part_status == '1' {
			partName := string(bytes.Trim(part.Part_name[:], "\x00"))
			if partName == name {
				return fmt.Errorf("ya existe una partición con el nombre '%s'", name)
			}
		}
	}

	// Luego validar contra particiones lógicas
	currentEBR := structures.EBR{}
	currentPos := extendedStart

	for {
		err := currentEBR.DeserializeEBR(path, currentPos)
		if err != nil {
			break // Si hay error al leer, asumimos que llegamos al final
		}

		if currentEBR.Part_size != -1 { // Si es una partición válida
			ebrName := string(bytes.Trim(currentEBR.Part_name[:], "\x00"))
			if ebrName == name {
				return fmt.Errorf("ya existe una partición lógica con el nombre '%s'", name)
			}
		}

		if currentEBR.Part_next == -1 {
			break
		}
		currentPos = currentEBR.Part_next
	}

	return nil
}

// CreateLogicalPartition crea una nueva partición lógica dentro de una partición extendida.
// La partición lógica se crea al final de la lista de EBRs existentes.
//
// Parámetros:
//   - params: estructura con los parámetros de la partición (nombre, tamaño, ajuste, etc.)
//   - extendedStart: posición inicial de la partición extendida
//
// Retorna:
//   - structures.Partition: la partición lógica creada
//   - error: error si hay problemas durante la creación
func CreateLogicalPartition(params types.FDisk, extendedStart int32) (structures.Partition, error) {
	// Validar que no exista una partición con el mismo nombre
	err := validatePartitionName(params.Path, params.Name, extendedStart)
	if err != nil {
		return structures.Partition{}, err
	}

	// Leer el EBR inicial de la partición extendida
	ebr := structures.EBR{}
	err = ebr.DeserializeEBR(params.Path, extendedStart)
	if err != nil {
		return structures.Partition{}, fmt.Errorf("error al leer el EBR inicial: %v", err)
	}

	// Convertir el tamaño a bytes
	sizeInBytes, err := utils.ConvertToBytes(params.Size, params.Unit)
	if err != nil {
		return structures.Partition{}, fmt.Errorf("error al convertir el tamaño: %v", err)
	}

	// Buscar el último EBR en la lista enlazada
	currentEBRStart := extendedStart
	lastEBR := ebr

	for lastEBR.Part_next != -1 {
		currentEBRStart = lastEBR.Part_next
		err = lastEBR.DeserializeEBR(params.Path, currentEBRStart)
		if err != nil {
			return structures.Partition{}, fmt.Errorf("error al leer el EBR en %d: %v", currentEBRStart, err)
		}
	}

	// Imprimir información del último EBR
	fmt.Printf("Último EBR encontrado en %d\n", lastEBR.Part_start)

	lastEBR.Part_next = currentEBRStart + int32(sizeInBytes)
	lastEBR.Part_size = int32(sizeInBytes)
	lastEBR.Part_fit = params.Fit[0]
	lastEBR.Part_mount = 'N'
	// Renombrar el último EBR
	copy(lastEBR.Part_name[:], []byte(params.Name))

	fmt.Printf("Nuevo EBR en %d\n", lastEBR.Part_next)

	// Guardar el último EBR
	err = lastEBR.SerializeEBR(params.Path, lastEBR.Part_start)
	if err != nil {
		return structures.Partition{}, fmt.Errorf("error al guardar el EBR: %v", err)
	}

	// Crear el nuevo EBR
	newEBR := structures.EBR{
		Part_mount: 'N',
		Part_fit:   'N',
		Part_start: lastEBR.Part_next,
		Part_size:  -1,
		Part_next:  -1,
	}

	// Guardar el nuevo EBR
	err = newEBR.SerializeEBR(params.Path, newEBR.Part_start)
	if err != nil {
		return structures.Partition{}, fmt.Errorf("error al guardar el nuevo EBR: %v", err)
	}

	// Crear la nueva partición lógica
	newPartition := structures.Partition{
		Part_status: '1', // Cambié a '1' para indicar que está activa
		Part_type:   'L',
		Part_fit:    params.Fit[0],
		Part_start:  1,
		Part_size:   -1,
	}

	// Copiar el nombre (máximo 16 bytes)
	copy(newPartition.Part_name[:], []byte(params.Name))

	return newPartition, nil
}
