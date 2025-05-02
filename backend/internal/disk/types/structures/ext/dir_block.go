package ext2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

const (
	DirBlockSize = 64
)

type DirBlock struct {
	BContent [4]DirContent // Entradas de directorio
}

type DirContent struct {
	BName  [12]byte // Nombre de la carpeta o archivo
	BInodo int32    // Número de inodo
}

// Serialize escribe la estructura FolderBlock en un archivo binario en la posición especificada
func (fb *DirBlock) Serialize(path string, offset int64) error {
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

	// Serializar la estructura FolderBlock directamente en el archivo
	err = binary.Write(file, binary.LittleEndian, fb)
	if err != nil {
		return err
	}

	return nil
}

// Deserialize lee la estructura FolderBlock desde un archivo binario en la posición especificada
func (fb *DirBlock) Deserialize(path string, offset int64) error {
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

	// Obtener el tamaño de la estructura FolderBlock
	fbSize := binary.Size(fb)
	if fbSize <= 0 {
		return fmt.Errorf("invalid FolderBlock size: %d", fbSize)
	}

	// Leer solo la cantidad de bytes que corresponden al tamaño de la estructura FolderBlock
	buffer := make([]byte, fbSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Deserializar los bytes leídos en la estructura FolderBlock
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, fb)
	if err != nil {
		return err
	}

	return nil
}

// Print imprime los atributos del bloque de carpeta
func (fb *DirBlock) Print() {
	for i, content := range fb.BContent {
		name := string(content.BName[:])
		fmt.Printf("Content %d:\n", i+1)
		fmt.Printf("  B_name: %s\n", name)
		fmt.Printf("  B_inodo: %d\n", content.BInodo)
	}
}
