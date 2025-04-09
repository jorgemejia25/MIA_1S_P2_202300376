package mbr_operations

import (
	"fmt"

	"disk.simulator.com/m/v2/internal/disk/types/structures"
)

// FindExtendedPartition busca y retorna la información de la partición extendida
// Retorna la partición y su posición en el arreglo, o un error si no se encuentra
func FindExtendedPartition(path string) (structures.Partition, int, error) {
	mbr := structures.MBR{}
	err := mbr.DeserializeMBR(path)
	if err != nil {
		return structures.Partition{}, -1, fmt.Errorf("error al leer el MBR: %v", err)
	}

	for i, part := range mbr.Mbr_partitions {
		if part.Part_status == '1' && part.Part_type == 'E' {
			return part, i, nil
		}
	}

	return structures.Partition{}, -1, fmt.Errorf("no se encontró una partición extendida")
}
