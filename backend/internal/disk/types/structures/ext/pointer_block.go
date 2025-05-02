package ext2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

const (
	PointerBlockSize = 64
)

type PointerBlock struct {
	PContent [16]int32 // Array de punteros a bloques de datos
}

// Serialize escribe la estructura PointerBlock en un archivo binario en la posición especificada
func (pb *PointerBlock) Serialize(path string, offset int64) error {
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

	// Serializar la estructura PointerBlock directamente en el archivo
	err = binary.Write(file, binary.LittleEndian, pb)
	if err != nil {
		return err
	}

	return nil
}

// Deserialize lee la estructura PointerBlock desde un archivo binario en la posición especificada
func (pb *PointerBlock) Deserialize(path string, offset int64) error {
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

	// Obtener el tamaño de la estructura PointerBlock
	pbSize := binary.Size(pb)
	if pbSize <= 0 {
		return fmt.Errorf("invalid PointerBlock size: %d", pbSize)
	}

	// Leer solo la cantidad de bytes que corresponden al tamaño de la estructura
	buffer := make([]byte, pbSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Deserializar los bytes leídos en la estructura PointerBlock
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, pb)
	if err != nil {
		return err
	}

	return nil
}

// Print imprime los punteros no nulos del bloque
func (pb *PointerBlock) Print() {
	fmt.Println("Bloque de punteros:")
	for i, ptr := range pb.PContent {
		if ptr != -1 {
			fmt.Printf("  Puntero %d: Bloque #%d\n", i, ptr)
		}
	}
}
