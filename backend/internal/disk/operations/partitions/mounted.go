package partition_operations

import (
	"strings"

	"disk.simulator.com/m/v2/internal/disk/memory"
)

func GetMountedPartitions() string {
	storage := memory.GetInstance()
	mountedPartitions := storage.GetMountedPartitions()

	if len(mountedPartitions) == 0 {
		return "No hay particiones montadas"
	}

	var ids []string
	for _, partition := range mountedPartitions {
		ids = append(ids, partition.ID)
	}

	return strings.Join(ids, ", ")
}
