package structures

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

// EBRSize define el tamaño en bytes de la estructura EBR
const EBRSize = 30 // 1 + 1 + 4 + 4 + 4 + 16 = 30 bytes

// EBR (Extended Boot Record) representa la estructura de una partición lógica
// dentro de una partición extendida. Funciona como una lista enlazada de particiones.
type EBR struct {
	Part_mount byte     // Estado de montaje: 0 = no montada, 1 = montada
	Part_fit   byte     // Tipo de ajuste: B = Best, F = First, W = Worst
	Part_start int32    // Posición inicial de la partición en bytes
	Part_size  int32    // Tamaño total de la partición en bytes
	Part_next  int32    // Apuntador al siguiente EBR (-1 si es el último)
	Part_name  [16]byte // Nombre de la partición (máximo 16 caracteres)
}

// SerializeEBR guarda la estructura EBR en el disco en la posición especificada.
// Parámetros:
//   - path: ruta del archivo de disco
//   - start: posición inicial donde se escribirá el EBR
//
// Retorna un error si hay problemas al escribir en el archivo
func (ebr *EBR) SerializeEBR(path string, start int32) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer file.Close()

	// Posicionarse en el inicio de la partición extendida
	_, err = file.Seek(int64(start), 0)
	if err != nil {
		return fmt.Errorf("error al posicionarse en el inicio del EBR: %v", err)
	}

	// Crear un buffer para almacenar los datos
	buf := new(bytes.Buffer)

	// Escribir todos los campos del EBR
	binary.Write(buf, binary.LittleEndian, ebr.Part_mount)
	binary.Write(buf, binary.LittleEndian, ebr.Part_fit)
	binary.Write(buf, binary.LittleEndian, ebr.Part_start)
	binary.Write(buf, binary.LittleEndian, ebr.Part_size)
	binary.Write(buf, binary.LittleEndian, ebr.Part_next)
	binary.Write(buf, binary.LittleEndian, ebr.Part_name)

	// Escribir el buffer en el archivo
	_, err = file.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("error al escribir el EBR: %v", err)
	}

	return nil
}

// DeserializeEBR lee una estructura EBR desde el disco en la posición especificada.
// Parámetros:
//   - path: ruta del archivo de disco
//   - start: posición desde donde se leerá el EBR
//
// Retorna un error si hay problemas al leer del archivo
func (ebr *EBR) DeserializeEBR(path string, start int32) error {
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer file.Close()

	// Posicionarse en el inicio del EBR
	_, err = file.Seek(int64(start), 0)
	if err != nil {
		return fmt.Errorf("error al posicionarse en el inicio del EBR: %v", err)
	}

	// Leer cada campo del EBR
	err = binary.Read(file, binary.LittleEndian, &ebr.Part_mount)
	if err != nil {
		return fmt.Errorf("error al leer Part_mount: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &ebr.Part_fit)
	if err != nil {
		return fmt.Errorf("error al leer Part_fit: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &ebr.Part_start)
	if err != nil {
		return fmt.Errorf("error al leer Part_start: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &ebr.Part_size)
	if err != nil {
		return fmt.Errorf("error al leer Part_size: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &ebr.Part_next)
	if err != nil {
		return fmt.Errorf("error al leer Part_next: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &ebr.Part_name)
	if err != nil {
		return fmt.Errorf("error al leer Part_name: %v", err)
	}

	return nil
}
