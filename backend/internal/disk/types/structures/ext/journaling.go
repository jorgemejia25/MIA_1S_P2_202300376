package ext2

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

type Journal struct {
	J_count   int32       // 4 bytes
	J_content Information // 10 + 100 + 200 + 4 = 314 bytes
	// Total: 318 bytes
}

type Information struct {
	I_operation [10]byte  // 10 bytes - Suficiente para operaciones como 'mkdir', 'mkfile', etc.
	I_path      [100]byte // 100 bytes - Aumentado para rutas más largas
	I_content   [200]byte // 200 bytes - Aumentado para contenidos más extensos
	I_date      float32   // 4 bytes
	// Total: 314 bytes
}

// SerializeJournal escribe la estructura Journal en un archivo binario
func (journal *Journal) Serialize(path string, journauling_start int64) error {
	// Calcular la posición en el archivo
	offset := journauling_start + (int64(binary.Size(Journal{})) * int64(journal.J_count))

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

	// Serializar la estructura Journal directamente en el archivo
	err = binary.Write(file, binary.LittleEndian, journal)
	if err != nil {
		return err
	}

	return nil
}

// DeserializeJournal lee la estructura Journal desde un archivo binario
func (journal *Journal) Deserialize(path string, offset int64) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Deserializar la estructura Journal directamente desde el archivo
	err = binary.Read(file, binary.LittleEndian, journal)
	if err != nil {
		return err
	}

	return nil
}

// PrintJournal imprime en consola la estructura Journal
func (journal *Journal) Print() {
	// Convertir el tiempo de montaje a una fecha
	date := time.Unix(int64(journal.J_content.I_date), 0)

	fmt.Println("Journal:")
	fmt.Printf("J_count: %d", journal.J_count)
	fmt.Println("Information:")
	fmt.Printf("I_operation: %s", string(journal.J_content.I_operation[:]))
	fmt.Printf("I_path: %s", string(journal.J_content.I_path[:]))
	fmt.Printf("I_content: %s", string(journal.J_content.I_content[:]))
	fmt.Printf("I_date: %s", date.Format(time.RFC3339))
}

func AddJournal(path string, partitionStart int64, journalCount int32, operation, filePath, content string) error {
	// Leer el SuperBlock para verificar si es ext3
	sb := &SuperBlock{}
	err := sb.DeserializeSuperBlock(path, int32(partitionStart))
	if err != nil {
		return fmt.Errorf("error al leer el SuperBlock: %v", err)
	}

	// Verificar si el journaling está habilitado (ext3)
	if sb.SFilesystemType != 3 {
		fmt.Println("El sistema de archivos no es ext3, no se creará el journal.")
		return nil
	}

	// El inicio del journaling es justo después del SuperBlock
	journalingStart := partitionStart + int64(binary.Size(SuperBlock{}))

	// Encontrar el siguiente índice de journal disponible
	nextIndex := int32(0)
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer file.Close()

	// Buscar la primera entrada vacía o el final del journaling
	for i := int32(0); i < sb.SFreeInodesCount; i++ { // Usar SFreeInodesCount como límite para la búsqueda
		offset := journalingStart + (int64(binary.Size(Journal{})) * int64(i))
		journal := Journal{}

		file.Seek(offset, 0)
		err := binary.Read(file, binary.LittleEndian, &journal)

		// Si hay error de lectura o la entrada está vacía, usar este índice
		if err != nil || isEmptyJournal(journal) {
			nextIndex = i
			break
		}

		// También comprobamos si hemos llegado al final del journaling
		if i == sb.SFreeInodesCount-1 {
			// Si llegamos al final, volver al principio (journaling circular)
			nextIndex = 0
			fmt.Println("El journaling está lleno, se sobrescribirá desde el principio.")
			break
		}
	}

	fmt.Printf("Usando índice de journaling: %d\n", nextIndex)

	// Crear una nueva entrada de Journal
	journal := Journal{
		J_count: nextIndex,
		J_content: Information{
			I_operation: [10]byte{},
			I_path:      [100]byte{},
			I_content:   [200]byte{},
			I_date:      float32(time.Now().Unix()),
		},
	}

	// Copiar los valores a los arrays de bytes
	copy(journal.J_content.I_operation[:], operation)
	copy(journal.J_content.I_path[:], filePath)

	// Si el contenido no está vacío, copiarlo
	if content != "" {
		copy(journal.J_content.I_content[:], content)
	}

	// Serializar la entrada de Journal en el archivo
	writeFile, err := os.OpenFile(path, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo para escritura: %v", err)
	}
	defer writeFile.Close()

	offset := journalingStart + (int64(binary.Size(Journal{})) * int64(nextIndex))
	writeFile.Seek(offset, 0)

	err = binary.Write(writeFile, binary.LittleEndian, &journal)
	if err != nil {
		return fmt.Errorf("error al escribir el journal: %v", err)
	}

	fmt.Printf("Journal agregado exitosamente con índice %d\n", nextIndex)
	return nil
}

func GetJournaling(path string, journalingStart int64, journalCount int32) ([]Journal, error) {
	var journals []Journal

	// Abrir el archivo en modo lectura
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer file.Close()

	// Leer cada entrada de Journal
	for i := int32(0); i < journalCount; i++ {
		offset := journalingStart + (int64(binary.Size(Journal{})) * int64(i))
		journal := Journal{}

		// Mover el puntero del archivo a la posición especificada
		_, err := file.Seek(offset, 0)
		if err != nil {
			return nil, fmt.Errorf("error al mover el puntero del archivo: %v", err)
		}

		// Deserializar la estructura Journal
		err = binary.Read(file, binary.LittleEndian, &journal)
		if err != nil {
			return nil, fmt.Errorf("error al leer el journal: %v", err)
		}

		// Solo añadir journales que tengan operaciones registradas
		// (para filtrar entradas vacías o no inicializadas)
		if !isEmptyJournal(journal) {
			journals = append(journals, journal)
		}
	}

	return journals, nil
}

// isEmptyJournal verifica si un journal está vacío (no tiene operación registrada)
func isEmptyJournal(journal Journal) bool {
	// Verificar si todos los bytes de la operación son nulos o espacios
	for _, b := range journal.J_content.I_operation {
		if b != 0 && b != 32 { // 32 es el espacio en ASCII
			return false
		}
	}
	return true
}
