package partition_operations

import (
	"bytes"
	"fmt"
	"strings"

	"disk.simulator.com/m/v2/internal/disk/types/structures"
)

// FindPartition busca una partición por nombre en el disco especificado.
// La búsqueda se realiza tanto en particiones primarias como en lógicas.
//
// Parámetros:
//   - name: nombre de la partición a buscar
//   - path: ruta del archivo de disco
//
// Retorna:
//   - structures.Partition: la partición encontrada
//   - error: error si ocurre algún problema durante la búsqueda
//   - int: índice de la partición (-1 si no se encuentra)
func FindPartition(name string, path string) (structures.Partition, int, error) {
	// Leer el MBR del disco
	mbr := structures.MBR{}
	err := mbr.DeserializeMBR(path)

	if err != nil {
		return structures.Partition{}, -1, err
	}

	// Buscar la partición con el nombre especificado
	name = strings.TrimSpace(name)

	// Primero buscar en particiones primarias y extendidas
	for i, part := range mbr.Mbr_partitions {
		endIndex := bytes.IndexByte(part.Part_name[:], 0)
		if endIndex == -1 {
			endIndex = len(part.Part_name)
		}

		partName := strings.TrimSpace(string(part.Part_name[:endIndex]))

		if partName == name {
			fmt.Printf("Partición encontrada en índice %d\n", i)
			return part, i, nil
		}

		// Si es una partición extendida, buscar en las particiones lógicas
		if part.Part_type == 'E' {
			// Buscar particiones lógicas dentro de la extendida
			currentEBR := structures.EBR{}
			currentPos := part.Part_start

			for {
				err := currentEBR.DeserializeEBR(path, currentPos)
				if err != nil {
					break
				}

				if currentEBR.Part_size != -1 { // Si es una partición válida
					ebrName := strings.TrimSpace(string(bytes.Trim(currentEBR.Part_name[:], "\x00")))
					if ebrName == name {
						// Convertir EBR a Partition para mantener la consistencia
						logicalPart := structures.Partition{
							Part_status: '1',
							Part_type:   'L',
							Part_fit:    currentEBR.Part_fit,
							Part_start:  currentEBR.Part_start,
							Part_size:   currentEBR.Part_size,
							Part_name:   currentEBR.Part_name,
						}
						return logicalPart, i, nil
					}
				}

				if currentEBR.Part_next == -1 {
					break
				}
				currentPos = currentEBR.Part_next
			}
		}
	}

	return structures.Partition{}, -1, nil
}
