package disk_operations

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// DiskRegistry es un singleton que mantiene un registro de todos los discos creados
type DiskRegistry struct {
	Disks map[string]DiskInfo // Mapa de ruta del disco a información del disco
	mutex sync.RWMutex        // Mutex para proteger el acceso concurrente
}

var (
	registry *DiskRegistry
	once     sync.Once
)

// Ruta del archivo para persistencia
const registryFilePath = "./disk_registry.json"

// GetDiskRegistry devuelve la instancia única del registro de discos
func GetDiskRegistry() *DiskRegistry {
	once.Do(func() {
		registry = &DiskRegistry{
			Disks: make(map[string]DiskInfo),
		}
		// Intentar cargar el registro desde el archivo
		loadRegistryFromFile()
	})
	return registry
}

// RegisterDisk agrega un disco al registro y lo guarda en el archivo
func (r *DiskRegistry) RegisterDisk(disk DiskInfo) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.Disks[disk.Path] = disk
	// Guardar el registro actualizado
	saveRegistryToFile()
}

// UnregisterDisk elimina un disco del registro y actualiza el archivo
func (r *DiskRegistry) UnregisterDisk(path string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.Disks, path)
	// Guardar el registro actualizado
	saveRegistryToFile()
}

// GetDisks devuelve una copia de todos los discos registrados
func (r *DiskRegistry) GetDisks() []DiskInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	disks := make([]DiskInfo, 0, len(r.Disks))
	for _, disk := range r.Disks {
		disks = append(disks, disk)
	}
	return disks
}

// DiskExists verifica si un disco ya está registrado
func (r *DiskRegistry) DiskExists(path string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	_, exists := r.Disks[path]
	return exists
}

// saveRegistryToFile guarda el registro de discos en un archivo JSON
func saveRegistryToFile() {
	// Asegurar que exista el directorio
	dir := filepath.Dir(registryFilePath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		// Si no podemos crear el directorio, simplemente logueamos el error y continuamos
		return
	}

	// Serializamos el mapa de discos a JSON
	data, err := json.MarshalIndent(registry.Disks, "", "  ")
	if err != nil {
		// Si hay un error en la serialización, simplemente logueamos y continuamos
		return
	}

	// Escribimos el archivo
	err = os.WriteFile(registryFilePath, data, 0644)
	if err != nil {
		// Si hay un error al escribir, logueamos y continuamos
		return
	}
}

// loadRegistryFromFile carga el registro de discos desde un archivo JSON
func loadRegistryFromFile() {
	// Verificar si el archivo existe
	if _, err := os.Stat(registryFilePath); os.IsNotExist(err) {
		// Si no existe, no hacemos nada más
		return
	}

	// Leer el archivo
	data, err := os.ReadFile(registryFilePath)
	if err != nil {
		// Si hay un error de lectura, simplemente continuamos con un registro vacío
		return
	}

	// Deserializar el JSON
	var disks map[string]DiskInfo
	if err := json.Unmarshal(data, &disks); err != nil {
		// Si hay un error de deserialización, continuamos con un registro vacío
		return
	}

	// Verificar que los archivos aún existen
	for path, disk := range disks {
		if _, err := os.Stat(path); err == nil {
			// El archivo existe, lo agregamos al registro
			registry.Disks[path] = disk
		}
		// Si el archivo no existe, no lo agregamos al registro
	}
}
