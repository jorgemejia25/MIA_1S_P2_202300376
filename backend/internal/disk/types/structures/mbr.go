package structures

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

const (
	MBRSize = 153 // 4 + 4 + 4 + 1 + (35 * 4)
)

type MBR struct {
	Mbr_size           int32        // Tamaño del MBR en bytes
	Mbr_creation_date  float32      // Fecha y hora de creación del MBR
	Mbr_disk_signature int32        // Firma del disco
	Mbr_disk_fit       [1]byte      // Tipo de ajuste
	Mbr_partitions     [4]Partition // Particiones del MBR
}

// SerializeMBR escribe la estructura MBR al inicio de un archivo binario
func (mbr *MBR) SerializeMBR(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Serializar la estructura MBR directamente en el archivo
	err = binary.Write(file, binary.LittleEndian, mbr)
	if err != nil {
		return err
	}

	return nil
}

// DeserializeMBR lee la estructura MBR desde el inicio de un archivo binario
func (mbr *MBR) DeserializeMBR(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Obtener el tamaño de la estructura MBR
	mbrSize := binary.Size(mbr)
	if mbrSize <= 0 {
		return fmt.Errorf("invalid MBR size: %d", mbrSize)
	}

	// Leer solo la cantidad de bytes que corresponden al tamaño de la estructura MBR
	buffer := make([]byte, mbrSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Deserializar los bytes leídos en la estructura MBR
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, mbr)
	if err != nil {
		return err
	}

	return nil
}

// DeletePartitionFull elimina completamente una partición del disco
func (mbr *MBR) DeletePartitionFull(path string, partitionName string) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer file.Close()

	// Buscar la partición por nombre
	for i, partition := range mbr.Mbr_partitions {
		if string(partition.Part_name[:]) == partitionName {
			// Calcular el tamaño y la posición de la partición
			partitionSize := partition.Part_size
			partitionStart := partition.Part_start

			// Sobrescribir el espacio de la partición con \0
			_, err := file.Seek(int64(partitionStart), 0)
			if err != nil {
				return fmt.Errorf("error al buscar la posición de la partición: %v", err)
			}

			zeroData := make([]byte, partitionSize)
			_, err = file.Write(zeroData)
			if err != nil {
				return fmt.Errorf("error al sobrescribir la partición: %v", err)
			}

			// Marcar la partición como eliminada en el MBR
			mbr.Mbr_partitions[i].Part_size = 0
			mbr.Mbr_partitions[i].Part_start = 0
			mbr.Mbr_partitions[i].Part_name = [16]byte{}
			mbr.Mbr_partitions[i].Part_status = '0'
			mbr.Mbr_partitions[i].Part_type = '0'
			mbr.Mbr_partitions[i].Part_fit = '0'

			// Serializar el MBR actualizado
			err = mbr.SerializeMBR(path)
			if err != nil {
				return fmt.Errorf("error al actualizar el MBR: %v", err)
			}

			return nil
		}
	}

	return fmt.Errorf("la partición '%s' no existe", partitionName)
}
