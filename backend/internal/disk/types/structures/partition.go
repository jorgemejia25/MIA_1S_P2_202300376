package structures

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

// Partition representa una partición en un disco simulado.
// Contiene información sobre el estado, tipo, ajuste, inicio y tamaño de la partición.
type Partition struct {
	Part_status      byte     // Estado de la partición (0: inactiva, 1: activa)
	Part_type        byte     // Tipo ('P': primaria, 'E': extendida)
	Part_fit         byte     // Ajuste ('B': Best, 'F': First, 'W': Worst)
	Part_mount       byte     // Estado de montaje (0: no montada, 1: montada)
	Part_start       int32    // Byte de inicio de la partición en el disco
	Part_size        int32    // Tamaño de la partición en bytes
	Part_name        [16]byte // Nombre de la partición
	Part_correlative int32    // Correlativo de montaje (-1: no montada, >=1: montada)
	Part_id          [4]byte  // ID único asignado al montar la partición
}

// SerializePartition escribe la estructura Partition en su representación binaria en un archivo
func (part *Partition) SerializePartition(path string) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer file.Close()

	// Posicionarse en el inicio de la partición usando Part_start
	_, err = file.Seek(int64(part.Part_start), 0)
	if err != nil {
		return fmt.Errorf("error al posicionarse en el inicio de la partición: %v", err)
	}

	// Crear un buffer para almacenar los datos
	buf := new(bytes.Buffer)

	// Escribir todos los campos de la partición
	binary.Write(buf, binary.LittleEndian, part.Part_status)
	binary.Write(buf, binary.LittleEndian, part.Part_type)
	binary.Write(buf, binary.LittleEndian, part.Part_fit)
	binary.Write(buf, binary.LittleEndian, part.Part_mount)
	binary.Write(buf, binary.LittleEndian, part.Part_start)
	binary.Write(buf, binary.LittleEndian, part.Part_size)
	binary.Write(buf, binary.LittleEndian, part.Part_name)
	binary.Write(buf, binary.LittleEndian, part.Part_correlative)
	binary.Write(buf, binary.LittleEndian, part.Part_id)

	// Verificar que el tamaño del buffer no exceda Part_size
	if int32(buf.Len()) > part.Part_size {
		return fmt.Errorf("el tamaño de la partición es menor que los datos a escribir")
	}

	// Escribir el buffer en el archivo
	_, err = file.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("error al escribir la partición: %v", err)
	}

	return nil
}

// DeserializePartition lee una estructura Partition desde su representación binaria en un archivo
func (part *Partition) DeserializePartition(path string) error {
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer file.Close()

	// Posicionarse en el inicio de la partición usando Part_start
	_, err = file.Seek(int64(part.Part_start), 0)
	if err != nil {
		return fmt.Errorf("error al posicionarse en el inicio de la partición: %v", err)
	}

	// Leer los campos de la partición
	err = binary.Read(file, binary.LittleEndian, &part.Part_status)
	if err != nil {
		return fmt.Errorf("error al leer Part_status: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &part.Part_type)
	if err != nil {
		return fmt.Errorf("error al leer Part_type: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &part.Part_fit)
	if err != nil {
		return fmt.Errorf("error al leer Part_fit: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &part.Part_mount)
	if err != nil {
		return fmt.Errorf("error al leer Part_mount: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &part.Part_start)
	if err != nil {
		return fmt.Errorf("error al leer Part_start: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &part.Part_size)
	if err != nil {
		return fmt.Errorf("error al leer Part_size: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &part.Part_name)
	if err != nil {
		return fmt.Errorf("error al leer Part_name: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &part.Part_correlative)
	if err != nil {
		return fmt.Errorf("error al leer Part_correlative: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &part.Part_id)
	if err != nil {
		return fmt.Errorf("error al leer Part_id: %v", err)
	}

	return nil
}

// Constante para el tamaño de la partición
const PartitionSize = 33 // 1 + 1 + 1 + 1 + 4 + 4 + 16 + 4 + 4 = 33 bytes
