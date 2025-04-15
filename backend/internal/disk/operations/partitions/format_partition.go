package partition_operations

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"

	"disk.simulator.com/m/v2/internal/disk/memory"
	"disk.simulator.com/m/v2/internal/disk/types/structures"
	"disk.simulator.com/m/v2/internal/disk/types/structures/ext"
	"disk.simulator.com/m/v2/utils"
)

func FormatPartition(id string, formatType string, ext3 bool) error {
	// Aquí iría la lógica para formatear la partición
	partition, path, err := memory.GetInstance().GetMountedPartition(id)

	if err != nil {
		fmt.Println(id)
		return err
	}

	fmt.Printf("Partition %s formatted with filesystem type %s\n", partition.Name, formatType)
	fmt.Printf("Path: %s\n", path)

	// Formatear la partición, del path: inicio hasta el final (tamaño) escribir 0s
	// Crear un buffer de bytes con el tamaño de la partición
	buf := make([]byte, partition.Partition.Part_size)

	// Crear un archivo en modo escritura
	file, err := os.OpenFile(path, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	// Posicionarse en el inicio de la partición
	_, err = file.Seek(int64(partition.Partition.Part_start), 0)
	if err != nil {
		return err
	}

	// Escribir los 0s en la partición
	_, err = file.Write(buf)

	if err != nil {
		return err
	}

	n := CalculateN(&partition.Partition)

	fmt.Println("N: ", n)
	journalSize := int32(binary.Size(ext2.Journal{}))

	JournalStart := partition.Partition.Part_start + int32(binary.Size(ext2.SuperBlock{}))

	var SBmInodeStart int32

	if ext3 {
		SBmInodeStart = JournalStart + (journalSize * n)
	} else {
		SBmInodeStart = partition.Partition.Part_start + int32(binary.Size(ext2.SuperBlock{}))
	}

	SBmBlockStart := SBmInodeStart + n
	SInodeStart := SBmBlockStart + (3 * n)
	SBlockStart := SInodeStart + (int32(ext2.INodeSize * n))

	// Crear el SuperBloque del sistema de archivos
	superBlock := ext2.SuperBlock{
		SFilesystemType:  2,
		SInodesCount:     0,
		SBlocksCount:     0,
		SFreeBlocksCount: int32(n * 3),
		SFreeInodesCount: int32(n),
		SMtime:           utils.FormatTime(partition.MountTime),
		SUmTime:          utils.FormatTime(partition.UnmountTime),
		SMntCount:        int32(partition.MountCount),
		SMagic:           0xEF53,
		SInodeS:          ext2.INodeSize,
		SBlockS:          64,
		SFirstIno:        SInodeStart, // Debe apuntar a donde inicia la tabla de inodos
		SFirstBlo:        SBlockStart, // Debe apuntar a donde inicia la tabla de bloques
		SBmInodeStart:    SBmInodeStart,
		SBmBlockStart:    SBmBlockStart,
		SInodeStart:      SInodeStart,
		SBlockStart:      SBlockStart,
	}

	// Serializar el SuperBloque
	err = superBlock.SerializeSuperBlock(path, partition.Partition.Part_start)

	if err != nil {
		fmt.Println("Error serializing superblock")
		return err
	}

	fmt.Println("SuperBlock created")
	fmt.Println(superBlock.SBmInodeStart)
	fmt.Println(superBlock.SBmBlockStart)
	fmt.Println(superBlock.SInodeStart)
	fmt.Println(superBlock.SBlockStart)

	if ext3 {
		err = superBlock.CreateBitMaps(path)
		if err != nil {
			return err
		}
	}

	// Crear el archivo users.txt
	err = superBlock.CreateUsersFile(path)

	if err != nil {
		return err
	}

	// Guardar el SuperBloque actualizado después de crear los archivos
	err = superBlock.SerializeSuperBlock(path, partition.Partition.Part_start)
	if err != nil {
		fmt.Println("Error al guardar el SuperBlock actualizado")
		return err
	}

	fmt.Println("\nSuperBlock actualizado:")
	superBlock.Print()

	return nil
}

func CalculateN(partition *structures.Partition) int32 {
	// Calcular el tamaño del SuperBlock
	superBlockSize := binary.Size(ext2.SuperBlock{})

	// Verificar que el tamaño de la partición sea válido
	if int(partition.Part_size) <= superBlockSize {
		return 0 // No hay espacio para inodos o bloques
	}

	// Calcular el numerador
	numerator := int(partition.Part_size) - superBlockSize

	// Calcular el denominador
	inodeSize := binary.Size(ext2.INode{})
	fileBlockSize := binary.Size(ext2.FileBlock{})
	denominator := 4 + inodeSize + 3*fileBlockSize

	// Calcular n
	n := math.Floor(float64(numerator) / float64(denominator))

	// Devolver n como int32
	return int32(n)
}
