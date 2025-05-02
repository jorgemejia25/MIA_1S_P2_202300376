package partition_operations

import (
	"bytes"
	"fmt"
	"strings"

	"disk.simulator.com/m/v2/internal/disk/types"
	"disk.simulator.com/m/v2/internal/disk/types/structures"
)

func AddSpacePartition(
	params types.FDisk,
) error {
	// Leer el MBR del disco
	var mbr structures.MBR
	err := mbr.DeserializeMBR(params.Path)
	if err != nil {
		return fmt.Errorf("error al leer el MBR: %v", err)
	}

	// Buscar la partición por nombre
	partitionFound := false
	for i, partition := range mbr.Mbr_partitions {
		// Verificar que la partición esté activa
		if partition.Part_status != '1' && partition.Part_size <= 0 {
			continue
		}

		// Comparamos los nombres eliminando los bytes nulos al final
		partName := string(bytes.Trim(partition.Part_name[:], "\x00"))
		if strings.TrimSpace(partName) == strings.TrimSpace(params.Name) {
			partitionFound = true

			// Calcular el tamaño a agregar o quitar
			var sizeChange int64
			switch params.Unit {
			case "B":
				sizeChange = int64(params.Add)
			case "K":
				sizeChange = int64(params.Add) * 1024
			case "M":
				sizeChange = int64(params.Add) * 1024 * 1024
			default:
				return fmt.Errorf("unidad desconocida: %s", params.Unit)
			}

			newSize := int64(partition.Part_size) + sizeChange

			// Verificar que no quede espacio negativo
			if newSize <= 0 {
				return fmt.Errorf("el tamaño resultante de la partición sería negativo o cero")
			}

			// Verificar que haya espacio libre si se está agregando espacio
			if sizeChange > 0 {
				partitionEnd := int64(partition.Part_start) + int64(partition.Part_size)
				availableSpace := int64(mbr.Mbr_size) - partitionEnd
				if sizeChange > availableSpace {
					return fmt.Errorf("no hay suficiente espacio libre para expandir la partición")
				}
			}

			// Actualizar el tamaño de la partición
			mbr.Mbr_partitions[i].Part_size = int32(newSize)

			// Asegurar que la partición se mantenga activa
			mbr.Mbr_partitions[i].Part_status = '1'

			// Serializar el MBR actualizado
			err = mbr.SerializeMBR(params.Path)
			if err != nil {
				return fmt.Errorf("error al actualizar el MBR: %v", err)
			}

			fmt.Printf("Partición '%s' modificada exitosamente. Nuevo tamaño: %d bytes\n", params.Name, newSize)
			return nil
		}
	}

	if !partitionFound {
		return fmt.Errorf("la partición '%s' no existe", params.Name)
	}

	return nil
}
