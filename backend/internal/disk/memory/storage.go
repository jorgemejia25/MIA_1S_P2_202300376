package memory

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"disk.simulator.com/m/v2/internal/disk/types/structures"
)

// MountedPartition representa una partición montada en memoria
type MountedPartition struct {
	ID          string
	Name        string
	Path        string
	Partition   structures.Partition
	MountTime   time.Time // Momento en que se montó la partición
	UnmountTime time.Time // Momento en que se desmontó la partición por última vez
	MountCount  int       // Contador de cuántas veces se ha montado la partición
}

// Storage es el singleton que maneja el almacenamiento en memoria
type Storage struct {
	mountedPartitions []MountedPartition
	diskLetters       map[string]byte // Mapea path del disco a letra
	partitionCounts   map[string]int  // Cuenta particiones por disco
	mutex             sync.Mutex
}

var (
	instance *Storage
	once     sync.Once
)

// GetInstance retorna la única instancia del almacenamiento
func GetInstance() *Storage {
	once.Do(func() {
		instance = &Storage{
			mountedPartitions: make([]MountedPartition, 0),
			diskLetters:       make(map[string]byte),
			partitionCounts:   make(map[string]int),
		}
	})
	return instance
}

// IsPartitionMounted verifica si una partición ya está montada y retorna su índice
func (s *Storage) IsPartitionMounted(name string, path string) (bool, int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i, partition := range s.mountedPartitions {
		if partition.Name == name && partition.Path == path {
			return true, i
		}
	}
	return false, -1
}

// MountPartition agrega una partición a la memoria o actualiza su fecha si ya existe
func (s *Storage) MountPartition(name string, path string, partition structures.Partition) (string, error) {
	// Verificar si la partición ya está montada
	mounted, index := s.IsPartitionMounted(name, path)

	if mounted {
		s.mutex.Lock()
		defer s.mutex.Unlock()

		// Actualizar la fecha de montaje y el contador
		s.mountedPartitions[index].MountTime = time.Now()
		s.mountedPartitions[index].MountCount++
		return s.mountedPartitions[index].ID, nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Obtener la letra del disco o asignar una nueva
	diskPath := filepath.Clean(path)
	diskLetter, exists := s.diskLetters[diskPath]
	if !exists {
		diskLetter = byte('A' + len(s.diskLetters))
		s.diskLetters[diskPath] = diskLetter
		s.partitionCounts[diskPath] = 0
	}

	// Incrementar el contador de particiones para este disco
	s.partitionCounts[diskPath]++
	partitionNumber := s.partitionCounts[diskPath]

	// Crear el ID con el formato número+letra
	id := fmt.Sprintf("76%d%c", partitionNumber, diskLetter)

	newMounted := MountedPartition{
		ID:         id,
		Name:       name,
		Path:       path,
		Partition:  partition,
		MountTime:  time.Now(),
		MountCount: 1, // Primera vez que se monta
	}

	s.mountedPartitions = append(s.mountedPartitions, newMounted)
	return id, nil
}

// Agregamos una función para desmontar la partición
func (s *Storage) UnmountPartition(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for i, partition := range s.mountedPartitions {
		if partition.ID == id {
			s.mountedPartitions[i].UnmountTime = time.Now()
			return nil
		}
	}

	return fmt.Errorf("partition not found")
}

// GetMountedPartition obtiene una partición montada por su ID y también retorna el path del disco
func (s *Storage) GetMountedPartition(id string) (MountedPartition, string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, partition := range s.mountedPartitions {
		if partition.ID == id {
			return partition, partition.Path, nil
		}
	}

	return MountedPartition{}, "", fmt.Errorf("partition not found")
}

// GetMountedPartitions retorna todas las particiones montadas
func (s *Storage) GetMountedPartitions() []MountedPartition {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.mountedPartitions
}
