package partition_operations

import (
	"fmt"

	"disk.simulator.com/m/v2/internal/disk/types"
	"disk.simulator.com/m/v2/internal/disk/types/structures"
)

func DeletePartition(
	params types.FDisk,
) error {
	// Leer el MBR del disco
	var mbr structures.MBR

	err := mbr.DeserializeMBR(params.Path)
	if err != nil {
		return fmt.Errorf("error al leer el MBR: %v", err)
	}
	partitionFound := false
	// If delete == "Full", delete the partition from the disk replacing the space with \0
	if params.Del == "Full" {
		// Reemplazar el espacio de la partición con \0
		err = mbr.DeletePartitionFull(params.Path, params.Name)
		if err != nil {
			return fmt.Errorf("error al eliminar la partición: %v", err)
		}
		fmt.Printf("Partición '%s' eliminada completamente.\n", params.Name)
	} else {

		// Verificar que la partición a eliminar exista
		for i, partition := range mbr.Mbr_partitions {
			if string(partition.Part_name[:]) == params.Name {
				partitionFound = true
				// Marcar la partición como eliminada
				mbr.Mbr_partitions[i].Part_size = 0
				mbr.Mbr_partitions[i].Part_start = 0
				mbr.Mbr_partitions[i].Part_name = [16]byte{}
				mbr.Mbr_partitions[i].Part_status = '0'
				mbr.Mbr_partitions[i].Part_type = '0'
				mbr.Mbr_partitions[i].Part_fit = '0'
				break
			}
		}
	}

	if !partitionFound {
		return fmt.Errorf("la partición '%s' no existe", params.Name)
	}

	return nil
}
