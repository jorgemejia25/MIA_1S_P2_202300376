package disk_operations

import (
	"encoding/json"

	partition_operations "disk.simulator.com/m/v2/internal/disk/operations/partitions"
)

// PartitionData contiene información sobre las particiones de un disco
type PartitionData struct {
	Partitions        []partition_operations.PartitionInfo        `json:"partitions"`
	LogicalPartitions []partition_operations.LogicalPartitionInfo `json:"logicalPartitions"`
}

// ListPartitions obtiene todas las particiones de un disco dado su path
func ListPartitions(diskPath string) (string, error) {
	// Llamar a la función existente de ListPartitions en el paquete partitions
	partitions, logicalPartitions, err := partition_operations.ListPartitions(diskPath)
	if err != nil {
		return "", err
	}

	// Crear una estructura que contenga ambos tipos de particiones
	data := PartitionData{
		Partitions:        partitions,
		LogicalPartitions: logicalPartitions,
	}

	// Convertir la estructura a JSON para devolverla
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// GetPartitionsFormatted devuelve un string con formato para visualizar las particiones
func GetPartitionsFormatted(diskPath string) (string, error) {
	// Utiliza la función existente para obtener la información formateada
	return partition_operations.GetPartitionsInfo(diskPath)
}
