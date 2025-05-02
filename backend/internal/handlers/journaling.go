package handlers

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"disk.simulator.com/m/v2/internal/disk/memory"
	ext2 "disk.simulator.com/m/v2/internal/disk/types/structures/ext"
)

// JournalEntry representa una entrada de journaling para la respuesta JSON
type JournalEntry struct {
	Operation string    `json:"operation"`
	Path      string    `json:"path"`
	Content   string    `json:"content"`
	Date      time.Time `json:"date"`
}

// JournalingResponse representa la respuesta del endpoint de journaling
type JournalingResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message,omitempty"`
	Entries []JournalEntry `json:"entries,omitempty"`
}

// GetJournaling maneja la solicitud para obtener las entradas del journaling
func GetJournaling(w http.ResponseWriter, r *http.Request) {
	// Configurar cabeceras para CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Manejar solicitud OPTIONS para CORS
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Verificar que la solicitud sea GET
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Obtener parámetros de la solicitud
	diskPath := r.URL.Query().Get("diskPath")
	partitionName := r.URL.Query().Get("partitionName")

	if diskPath == "" || partitionName == "" {
		respondWithError(w, "Se requiere especificar diskPath y partitionName", http.StatusBadRequest)
		return
	}

	fmt.Printf("Buscando journaling para disco: %s, partición: %s\n", diskPath, partitionName)

	// Obtener instancia del storage
	storage := memory.GetInstance()

	// Crear respuesta de error por defecto
	response := JournalingResponse{
		Success: false,
		Message: "Error al obtener el journaling",
	}

	partFound := false
	var partitionData memory.MountedPartition

	// Buscar la partición montada
	for _, partition := range storage.GetMountedPartitions() {
		// Limpiar y normalizar nombres para comparación
		storedPartName := strings.TrimSpace(string(partition.Name[:]))
		storedPartName = strings.Trim(storedPartName, "\x00") // Eliminar null bytes
		diskPathClean := strings.TrimSpace(partition.Path)

		fmt.Printf("Comparando con partición montada: '%s' en disco '%s'\n", storedPartName, diskPathClean)

		// Comparar nombres de particiones y rutas de disco
		if diskPathClean == diskPath && storedPartName == partitionName {
			partFound = true
			partitionData = partition
			fmt.Println("¡Partición encontrada!")
			break
		}
	}

	if !partFound {
		response.Message = fmt.Sprintf("Partición '%s' no encontrada en el disco '%s'", partitionName, diskPath)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Leer el SuperBlock para acceder al journaling
	sb := &ext2.SuperBlock{}
	err := sb.DeserializeSuperBlock(diskPath, partitionData.Partition.Part_start)
	if err != nil {
		response.Message = "Error al leer el SuperBlock: " + err.Error()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Mostrar información del SuperBlock para depuración
	fmt.Println("SuperBlock leído correctamente:")
	fmt.Printf("Type: %d, MntCount: %d\n", sb.SFilesystemType, sb.SMntCount)
	fmt.Printf("SBmInodeStart: %d, Part_start: %d\n", sb.SBmInodeStart, partitionData.Partition.Part_start)

	// Verificar si el filesystem es ext3 (tiene journaling)
	if sb.SFilesystemType != 3 {
		response.Message = "La partición no usa ext3 y no tiene journaling habilitado"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// El inicio del journaling es justo después del SuperBlock
	// Calculamos el tamaño del SuperBlock de manera directa
	superBlockSize := binary.Size(ext2.SuperBlock{})
	journalingStart := int64(partitionData.Partition.Part_start) + int64(superBlockSize)

	fmt.Printf("Tamaño del SuperBlock: %d bytes\n", superBlockSize)
	fmt.Printf("Inicio del journaling calculado: %d\n", journalingStart)

	// Obtener las entradas de journaling
	journals, err := ext2.GetJournaling(diskPath, journalingStart, sb.SFreeInodesCount)
	if err != nil {
		response.Message = "Error al leer el journaling: " + err.Error()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	fmt.Printf("Se encontraron %d entradas de journaling\n", len(journals))

	// Convertir las entradas al formato de respuesta
	entries := make([]JournalEntry, 0, len(journals))
	for _, journal := range journals {
		// Convertir tiempo Unix a time.Time
		date := time.Unix(int64(journal.J_content.I_date), 0)

		// Limpiar los strings (quitar caracteres nulos)
		operation := strings.TrimSpace(strings.Trim(string(journal.J_content.I_operation[:]), "\x00"))
		path := strings.TrimSpace(strings.Trim(string(journal.J_content.I_path[:]), "\x00"))
		content := strings.TrimSpace(strings.Trim(string(journal.J_content.I_content[:]), "\x00"))

		entries = append(entries, JournalEntry{
			Operation: operation,
			Path:      path,
			Content:   content,
			Date:      date,
		})

		fmt.Printf("Journaling entry: Operation=%s, Path=%s\n", operation, path)
	}

	// Actualizar respuesta con las entradas
	response.Success = true
	response.Message = ""
	response.Entries = entries

	// Enviar respuesta JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// respondWithError envía una respuesta de error en formato JSON
func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(JournalingResponse{
		Success: false,
		Message: message,
	})
}
