package reports

import (
	"fmt"
	"os"
	"strings"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/types/structures/ext2"
	"disk.simulator.com/m/v2/utils"
)

func BBlockReport(
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

	var bitmapContent strings.Builder
	for i := int32(0); i < (superBlock.SBlocksCount + superBlock.SFreeBlocksCount); i++ {
		_, err := file.Seek(int64(superBlock.SBmBlockStart+i), 0)
		if err != nil {
			return err
		}

		b := make([]byte, 1)
		_, err = file.Read(b)
		if err != nil {
			return err
		}

		if b[0] == 'X' {
			bitmapContent.WriteRune('X')
		} else {
			bitmapContent.WriteRune('0')
		}
		if (i+1)%20 == 0 {
			bitmapContent.WriteString("\n")
		}
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.WriteString(bitmapContent.String())
	if err != nil {
		return err
	}

	fmt.Println("Archivo del bitmap de bloques generado:", outputPath)
	return nil
}
