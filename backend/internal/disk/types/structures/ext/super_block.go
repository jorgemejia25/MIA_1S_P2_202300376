package ext2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

const (
	SuperBlockSize = 68
)

type SuperBlock struct {
	SFilesystemType  int32   // Guarda el número que identifica el sistema de archivos utilizado
	SInodesCount     int32   // Guarda el número total de inodos
	SBlocksCount     int32   // Guarda el número total de bloques
	SFreeBlocksCount int32   // Contiene el número de bloques libres
	SFreeInodesCount int32   // Contiene el número de inodos libres
	SMtime           float32 // Última fecha en el que el sistema fue montado
	SUmTime          float32 // Última fecha en que el sistema fue desmontado
	SMntCount        int32   // Indica cuantas veces se ha montado el sistema
	SMagic           int32   // Valor que identifica al sistema de archivos, tendrá el valor 0xEF53
	SInodeS          int32   // Tamaño del inodo
	SBlockS          int32   // Tamaño del bloque
	SFirstIno        int32   // Primer inodo libre (dirección del inodo)
	SFirstBlo        int32   // Primer bloque libre (dirección del inodo)
	SBmInodeStart    int32   // Guardará el inicio del bitmap de inodos
	SBmBlockStart    int32   // Guardará el inicio del bitmap de bloques
	SInodeStart      int32   // Guardará el inicio de la tabla de inodos
	SBlockStart      int32   // Guardará el inicio de la tabla de bloques
}

// SerializeSuperBlock escribe la estructura SuperBlock en su representación binaria en un archivo
func (sb *SuperBlock) SerializeSuperBlock(path string, start int32) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer file.Close()

	// Posicionarse en el inicio del SuperBlock
	_, err = file.Seek(int64(start), 0)
	if err != nil {
		return fmt.Errorf("error al posicionarse en el disco: %v", err)
	}

	// Crear un buffer para almacenar los datos
	buf := new(bytes.Buffer)

	// Escribir todos los campos del SuperBlock
	binary.Write(buf, binary.LittleEndian, sb.SFilesystemType)
	binary.Write(buf, binary.LittleEndian, sb.SInodesCount)
	binary.Write(buf, binary.LittleEndian, sb.SBlocksCount)
	binary.Write(buf, binary.LittleEndian, sb.SFreeBlocksCount)
	binary.Write(buf, binary.LittleEndian, sb.SFreeInodesCount)
	binary.Write(buf, binary.LittleEndian, sb.SMtime)
	binary.Write(buf, binary.LittleEndian, sb.SUmTime)
	binary.Write(buf, binary.LittleEndian, sb.SMntCount)
	binary.Write(buf, binary.LittleEndian, sb.SMagic)
	binary.Write(buf, binary.LittleEndian, sb.SInodeS)
	binary.Write(buf, binary.LittleEndian, sb.SBlockS)
	binary.Write(buf, binary.LittleEndian, sb.SFirstIno)
	binary.Write(buf, binary.LittleEndian, sb.SFirstBlo)
	binary.Write(buf, binary.LittleEndian, sb.SBmInodeStart)
	binary.Write(buf, binary.LittleEndian, sb.SBmBlockStart)
	binary.Write(buf, binary.LittleEndian, sb.SInodeStart)
	binary.Write(buf, binary.LittleEndian, sb.SBlockStart)

	// Escribir el buffer en el archivo
	_, err = file.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("error al escribir el SuperBlock: %v", err)
	}

	return nil
}

// DeserializeSuperBlock lee una estructura SuperBlock desde su representación binaria en un archivo
func (sb *SuperBlock) DeserializeSuperBlock(path string, start int32) error {
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer file.Close()

	// Posicionarse en el inicio del SuperBlock
	_, err = file.Seek(int64(start), 0)
	if err != nil {
		return fmt.Errorf("error al posicionarse en el disco: %v", err)
	}

	// Leer los campos del SuperBlock
	err = binary.Read(file, binary.LittleEndian, &sb.SFilesystemType)
	if err != nil {
		return fmt.Errorf("error al leer SFilesystemType: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &sb.SInodesCount)
	if err != nil {
		return fmt.Errorf("error al leer SInodesCount: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &sb.SBlocksCount)
	if err != nil {
		return fmt.Errorf("error al leer SBlocksCount: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &sb.SFreeBlocksCount)
	if err != nil {
		return fmt.Errorf("error al leer SFreeBlocksCount: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &sb.SFreeInodesCount)
	if err != nil {
		return fmt.Errorf("error al leer SFreeInodesCount: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &sb.SMtime)
	if err != nil {
		return fmt.Errorf("error al leer SMtime: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &sb.SUmTime)
	if err != nil {
		return fmt.Errorf("error al leer SUmTime: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &sb.SMntCount)
	if err != nil {
		return fmt.Errorf("error al leer SMntCount: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &sb.SMagic)
	if err != nil {
		return fmt.Errorf("error al leer SMagic: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &sb.SInodeS)
	if err != nil {
		return fmt.Errorf("error al leer SInodeS: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &sb.SBlockS)
	if err != nil {
		return fmt.Errorf("error al leer SBlockS: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &sb.SFirstIno)
	if err != nil {
		return fmt.Errorf("error al leer SFirstIno: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &sb.SFirstBlo)
	if err != nil {
		return fmt.Errorf("error al leer SFirstBlo: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &sb.SBmInodeStart)
	if err != nil {
		return fmt.Errorf("error al leer SBmInodeStart: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &sb.SBmBlockStart)
	if err != nil {
		return fmt.Errorf("error al leer SBmBlockStart: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &sb.SInodeStart)
	if err != nil {
		return fmt.Errorf("error al leer SInodeStart: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &sb.SBlockStart)
	if err != nil {
		return fmt.Errorf("error al leer SBlockStart: %v", err)
	}

	return nil
}

func (sb *SuperBlock) Print() {
	// Convertir el tiempo de montaje a una fecha
	mountTime := time.Unix(int64(sb.SMtime), 0)
	// Convertir el tiempo de desmontaje a una fecha
	unmountTime := time.Unix(int64(sb.SUmTime), 0)

	fmt.Printf("Filesystem Type: %d\n", sb.SFilesystemType)
	fmt.Printf("Inodes Count: %d\n", sb.SInodesCount)
	fmt.Printf("Blocks Count: %d\n", sb.SBlocksCount)
	fmt.Printf("Free Inodes Count: %d\n", sb.SFreeInodesCount)
	fmt.Printf("Free Blocks Count: %d\n", sb.SFreeBlocksCount)
	fmt.Printf("Mount Time: %s\n", mountTime.Format(time.RFC3339))
	fmt.Printf("Unmount Time: %s\n", unmountTime.Format(time.RFC3339))
	fmt.Printf("Mount Count: %d\n", sb.SMntCount)
	fmt.Printf("Magic: %d\n", sb.SMagic)
	fmt.Printf("Inode Size: %d\n", sb.SInodeS)
	fmt.Printf("Block Size: %d\n", sb.SBlockS)
	fmt.Printf("First Inode: %d\n", sb.SFirstIno)
	fmt.Printf("First Block: %d\n", sb.SFirstBlo)
	fmt.Printf("Bitmap Inode Start: %d\n", sb.SBmInodeStart)
	fmt.Printf("Bitmap Block Start: %d\n", sb.SBmBlockStart)
	fmt.Printf("Inode Start: %d\n", sb.SInodeStart)
	fmt.Printf("Block Start: %d\n", sb.SBlockStart)
}

func (sb *SuperBlock) PrintInodes(path string) error {
	// Imprimir inodos
	fmt.Println("\nInodos\n----------------")
	// Iterar sobre cada inodo
	for i := int32(0); i < sb.SInodesCount; i++ {
		inode := &INode{}
		// Deserializar el inodo
		err := inode.Deserialize(path, int64(sb.SInodeStart+(i*sb.SInodeS)))
		if err != nil {
			return err
		}
		// Imprimir el inodo
		fmt.Printf("\nInodo %d:\n", i)
		inode.Print()
	}

	return nil
}

func (sb *SuperBlock) PrintBlocks(path string) error {
	// Imprimir bloques
	fmt.Println("\nBloques\n----------------")
	// Iterar sobre cada inodo
	for i := int32(0); i < sb.SInodesCount; i++ {
		inode := &INode{}
		// Deserializar el inodo
		err := inode.Deserialize(path, int64(sb.SInodeStart+(i*sb.SInodeS)))
		if err != nil {
			return err
		}
		// Iterar sobre cada bloque del inodo (apuntadores)
		for _, blockIndex := range inode.IBlock {
			// Si el bloque no existe, salir
			if blockIndex == -1 {
				break
			}
			// Si el inodo es de tipo carpeta
			if inode.IType[0] == '0' {
				block := &DirBlock{}
				// Deserializar el bloque
				err := block.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS))) // 64 porque es el tamaño de un bloque
				if err != nil {
					return err
				}
				// Imprimir el bloque
				fmt.Printf("\nBloque %d:\n", blockIndex)
				block.Print()
				continue

				// Si el inodo es de tipo archivo
			} else if inode.IType[0] == '1' {
				block := &FileBlock{}
				// Deserializar el bloque
				err := block.Deserialize(path, int64(sb.SBlockStart+(blockIndex*sb.SBlockS))) // 64 porque es el tamaño de un bloque
				if err != nil {
					return err
				}
				// Imprimir el bloque
				fmt.Printf("\nBloque %d:\n", blockIndex)
				block.Print()
				continue
			}

		}
	}

	return nil
}

func (sb *SuperBlock) GetBlockByNumber(path string, blockNumber int32) (interface{}, error) {
	offset := int64(sb.SBlockStart + (blockNumber * sb.SBlockS))

	dirBlock := &DirBlock{}
	if err := dirBlock.Deserialize(path, offset); err == nil {
		return dirBlock, nil
	}

	fileBlock := &FileBlock{}
	if err := fileBlock.Deserialize(path, offset); err == nil {
		return fileBlock, nil
	}

	pointerBlock := &PointerBlock{}
	if err := pointerBlock.Deserialize(path, offset); err == nil {
		return pointerBlock, nil
	}

	return nil, fmt.Errorf("no se pudo determinar el tipo del bloque %d", blockNumber)
}

