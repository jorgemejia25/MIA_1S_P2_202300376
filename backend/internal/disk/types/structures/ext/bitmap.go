package ext2

import (
	"encoding/binary"
	"os"
)

// CreateBitMaps crea los Bitmaps de inodos y bloques en el archivo especificado
func (sb *SuperBlock) CreateBitMaps(path string) error {
	// Escribir Bitmaps
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Bitmap de inodos
	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(int64(sb.SBmInodeStart), 0)
	if err != nil {
		return err
	}

	// Crear un buffer de n '0'
	buffer := make([]byte, sb.SFreeInodesCount)
	for i := range buffer {
		buffer[i] = '0'
	}

	// Escribir el buffer en el archivo
	err = binary.Write(file, binary.LittleEndian, buffer)
	if err != nil {
		return err
	}

	// Bitmap de bloques
	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(int64(sb.SBlockStart), 0)
	if err != nil {
		return err
	}

	// Crear un buffer de n 'O'
	buffer = make([]byte, sb.SFreeBlocksCount)
	for i := range buffer {
		buffer[i] = 'O'
	}

	// Escribir el buffer en el archivo
	err = binary.Write(file, binary.LittleEndian, buffer)
	if err != nil {
		return err
	}

	return nil
}

// Actualizar Bitmap de inodos
func (sb *SuperBlock) UpdateBitmapInode(path string) error {
	// Abrir el archivo
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición del bitmap de inodos
	_, err = file.Seek(int64(sb.SBmInodeStart)+int64(sb.SInodesCount), 0)
	if err != nil {
		return err
	}

	// Escribir el bit en el archivo
	_, err = file.Write([]byte{'1'})
	if err != nil {
		return err
	}

	return nil
}

// Actualizar Bitmap de bloques
func (sb *SuperBlock) UpdateBitmapBlock(path string) error {
	// Abrir el archivo
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición del bitmap de bloques
	_, err = file.Seek(int64(sb.SBmBlockStart)+int64(sb.SBlocksCount), 0)
	if err != nil {
		return err
	}

	// Escribir el bit en el archivo
	_, err = file.Write([]byte{'X'})
	if err != nil {
		return err
	}

	return nil
}
