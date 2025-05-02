package ext2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

const (
	INodeSize = 88
)

type INode struct {
	IUid   int32     // UID del usuario propietario del archivo o carpeta
	IGid   int32     // GID del grupo al que pertenece el archivo o carpeta
	ISize  int32     // Tamaño del archivo en bytes
	IAtime float32   // Última fecha en que se leyó el inodo sin modificarlo
	ICtime float32   // Fecha en la que se creó el inodo
	IMtime float32   // Última fecha en la que se modifica el inodo
	IBlock [15]int32 // Array de bloques (12 directos, 1 simple indirecto, 1 doble indirecto, 1 triple indirecto)
	IType  [1]byte   // Indica si es archivo o carpeta (1 = Archivo, 0 = Carpeta)
	IPerm  [3]byte   // Permisos del archivo o carpeta en forma octal (UGO)
}

// Serialize escribe la estructura Inode en un archivo binario en la posición especificada
func (inode *INode) Serialize(path string, offset int64) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Serializar la estructura Inode directamente en el archivo
	err = binary.Write(file, binary.LittleEndian, inode)
	if err != nil {
		return err
	}

	return nil
}

// Deserialize lee la estructura Inode desde un archivo binario en la posición especificada
func (inode *INode) Deserialize(path string, offset int64) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Obtener el tamaño de la estructura Inode
	inodeSize := binary.Size(inode)
	if inodeSize <= 0 {
		return fmt.Errorf("invalid Inode size: %d", inodeSize)
	}

	// Leer solo la cantidad de bytes que corresponden al tamaño de la estructura Inode
	buffer := make([]byte, inodeSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Deserializar los bytes leídos en la estructura Inode
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, inode)
	if err != nil {
		return err
	}

	return nil
}

// Print imprime los atributos del inodo
func (inode *INode) Print() {
	atime := time.Unix(int64(inode.IAtime), 0)
	ctime := time.Unix(int64(inode.ICtime), 0)
	mtime := time.Unix(int64(inode.IMtime), 0)

	fmt.Printf("I_uid: %d\n", inode.IGid)
	fmt.Printf("I_gid: %d\n", inode.IUid)
	fmt.Printf("I_size: %d\n", inode.ISize)
	fmt.Printf("I_atime: %s\n", atime.Format(time.RFC3339))
	fmt.Printf("I_ctime: %s\n", ctime.Format(time.RFC3339))
	fmt.Printf("I_mtime: %s\n", mtime.Format(time.RFC3339))
	fmt.Printf("I_block: %v\n", inode.IBlock)
	fmt.Printf("I_type: %s\n", string(inode.IType[:]))
	fmt.Printf("I_perm: %s\n", string(inode.IPerm[:]))
}
