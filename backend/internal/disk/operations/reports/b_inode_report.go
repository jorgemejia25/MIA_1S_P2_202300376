package reports

import (
	"fmt"
	"os"
	"strings"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/types/structures/ext"
	"disk.simulator.com/m/v2/utils"
)

func BInodeReport(
	outputPath string,
	id string,

) error {
	partition, diskPath, err := memory.GetInstance().GetMountedPartition(id)

	if err != nil {
		return err
	}

	superBlock := ext2.SuperBlock{}
	superBlock.DeserializeSuperBlock(diskPath, partition.Partition.Part_start)
	superBlock.Print()

	err = utils.CreateParentDirs(outputPath)

	if err != nil {
		return err
	}

	file, err := os.Open(diskPath)
	if err != nil {
		return err
	}

	totalInodes := superBlock.SInodesCount + superBlock.SFreeInodesCount

	var bitmapContent strings.Builder

	for i := int32(0); i < totalInodes; i++ {
		_, err := file.Seek(int64(superBlock.SBmInodeStart+i), 0)

		if err != nil {
			return err
		}

		char := make([]byte, 1)
		_, err = file.Read(char)
		if err != nil {
			return fmt.Errorf("error al leer el byte del archivo: %v", err)
		}

		// Agregar el carácter al contenido del bitmap
		bitmapContent.WriteByte(char[0])

		// Agregar un carácter de nueva línea cada 20 caracteres (20 inodos)
		if (i+1)%20 == 0 {
			bitmapContent.WriteString("\n")
		}
	}

	txtFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error al crear el archivo TXT: %v", err)
	}
	defer txtFile.Close()

	// Escribir el contenido del bitmap en el archivo TXT
	_, err = txtFile.WriteString(bitmapContent.String())
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo TXT: %v", err)
	}

	fmt.Println("Archivo del bitmap de inodos generado:", outputPath)

	return nil
}
