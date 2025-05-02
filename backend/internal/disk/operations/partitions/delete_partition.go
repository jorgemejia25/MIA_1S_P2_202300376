package partition_operations

import (
	"bytes"
	"fmt"
	"strings"

	"disk.simulator.com/m/v2/internal/disk/types"
	"disk.simulator.com/m/v2/internal/disk/types/structures"
)

func DeletePartition(
	params types.FDisk,
) error {
	// No realizar ninguna acción si no se está eliminando una partición
	if params.Del == "" {
		return nil
	}

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
		partitionFound = true
	} else {
		// Verificar que la partición a eliminar exista
		for i, partition := range mbr.Mbr_partitions {
			// Comparamos los nombres eliminando los bytes nulos al final
			partName := string(bytes.Trim(partition.Part_name[:], "\x00"))
			if strings.TrimSpace(partName) == strings.TrimSpace(params.Name) &&
				partition.Part_status == '1' && partition.Part_size > 0 {
				partitionFound = true
				// Marcar la partición como eliminada
				mbr.Mbr_partitions[i].Part_size = 0
				mbr.Mbr_partitions[i].Part_start = 0
				mbr.Mbr_partitions[i].Part_name = [16]byte{}
				mbr.Mbr_partitions[i].Part_status = '0'
				mbr.Mbr_partitions[i].Part_type = '0'
				mbr.Mbr_partitions[i].Part_fit = '0'

				fmt.Printf("Partición '%s' marcada como eliminada.\n", params.Name)
				break
			}
		}
	}

	// Solo verificar si la partición existe cuando realmente estamos intentando eliminar una
	if !partitionFound && params.Del != "" {
		return fmt.Errorf("la partición '%s' no existe o ya fue eliminada", params.Name)
	}

	// Serializar el MBR actualizado
	err = mbr.SerializeMBR(params.Path)
	if err != nil {
		return fmt.Errorf("error al actualizar el MBR: %v", err)
	}

	fmt.Printf("MBR actualizado correctamente después de eliminar '%s'.\n", params.Name)
	return nil
}
