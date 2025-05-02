package reports

import (
	"fmt"
	"os"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/types/structures/ext"
	"disk.simulator.com/m/v2/utils"
)

func FileReport(path_file string, output_path string, id string) error {
	// Obtener partición e información
	partition, partitionPath, err := memory.GetInstance().GetMountedPartition(id)
	if err != nil {
		return fmt.Errorf("error al obtener partición: %v", err)
	}

	// Separar ruta en directorios padres y nombre de archivo
	parentDirs, fileName := utils.GetParentDirectories(path_file)

	// Leer superbloque
	superBlock := ext2.SuperBlock{}
	err = superBlock.DeserializeSuperBlock(partitionPath, partition.Partition.Part_start)
	if err != nil {
		return fmt.Errorf("error al leer superbloque: %v", err)
	}

	// Leer contenido desde ext2
	content, err := superBlock.ReadFile(partitionPath, parentDirs, fileName)
	if err != nil {
		return fmt.Errorf("error leyendo archivo: %v", err)
	}

	// Guardar en output_path
	err = os.WriteFile(output_path, []byte(content), 0664)
	if err != nil {
		return fmt.Errorf("error escribiendo reporte: %v", err)
	}

	fmt.Printf("Reporte de archivo generado con id '%s' en %s\n", id, output_path)
	return nil
}
