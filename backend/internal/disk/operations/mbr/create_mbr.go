package mbr_operations

import (
	"fmt"
	"math/rand"
	"time"

	"disk.simulator.com/m/v2/internal/disk/types"
	"disk.simulator.com/m/v2/internal/disk/types/structures"
	"disk.simulator.com/m/v2/utils"
)

func CreateMBR(mkdisk types.MkDisk, size int32) error {

	var fitByte [1]byte

	switch mkdisk.Fit {
	case "FF":
		fitByte = [1]byte{'F'}
	case "BF":
		fitByte = [1]byte{'B'}
	case "WF":
		fitByte = [1]byte{'W'}
	default:
		fmt.Println("Invalid fit type")
		return nil
	}

	// Crear el MBR
	mbr := structures.MBR{
		Mbr_size:          size,
		Mbr_disk_fit:      fitByte,
		Mbr_creation_date: utils.FormatTime(time.Now()),
		Mbr_partitions: [4]structures.Partition{
			// Inicializó todos los char en N y los enteros en -1 para que se puedan apreciar en el archivo binario.
			{Part_status: 'N', Part_type: 'N', Part_fit: 'N', Part_start: -1, Part_size: -1, Part_name: [16]byte{'N'}, Part_correlative: -1, Part_id: [4]byte{'N'}},
			{Part_status: 'N', Part_type: 'N', Part_fit: 'N', Part_start: -1, Part_size: -1, Part_name: [16]byte{'N'}, Part_correlative: -1, Part_id: [4]byte{'N'}},
			{Part_status: 'N', Part_type: 'N', Part_fit: 'N', Part_start: -1, Part_size: -1, Part_name: [16]byte{'N'}, Part_correlative: -1, Part_id: [4]byte{'N'}},
			{Part_status: 'N', Part_type: 'N', Part_fit: 'N', Part_start: -1, Part_size: -1, Part_name: [16]byte{'N'}, Part_correlative: -1, Part_id: [4]byte{'N'}},
		},
		Mbr_disk_signature: rand.Int31(),
	}

	err := mbr.SerializeMBR(mkdisk.Path)

	if err != nil {
		return fmt.Errorf("Error al crear el MBR: %v", err)
	}

	return nil

}
