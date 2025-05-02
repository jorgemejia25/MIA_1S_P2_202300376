// filepath: /home/jorgis/Documents/USAC/archivos/proyecto2/backend/internal/handlers/disks.go
package handlers

import (
	"encoding/json"
	"fmt"

	disk_operations "disk.simulator.com/m/v2/internal/disk/operations/disk"
	"github.com/gin-gonic/gin"
)

type DiskResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Disks   interface{} `json:"disks,omitempty"`
}

type PartitionResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"msg"`
	Partitions interface{} `json:"partitions,omitempty"`
}

// HandleDisk maneja las peticiones para listar todos los discos creados
func HandleDisk(c *gin.Context) {
	// Obtener la información de los discos directamente como estructura de datos
	disks, err := disk_operations.ListDisks()

	if err != nil {
		c.JSON(400, DiskResponse{
			Success: false,
			Message: fmt.Sprintf("Error al obtener la lista de discos: %v", err),
		})
		return
	}

	// Devolver la información de los discos en formato JSON
	c.JSON(200, DiskResponse{
		Success: true,
		Message: "Discos obtenidos correctamente",
		Disks:   disks,
	})
}

// HandleDiskPartitions maneja las peticiones para listar todas las particiones de un disco específico
func HandleDiskPartitions(c *gin.Context) {
	// Obtener el path del disco desde los parámetros de consulta (query)
	diskPath := c.Query("path")

	if diskPath == "" {
		c.JSON(400, PartitionResponse{
			Success: false,
			Message: "No se proporcionó la ruta del disco",
		})
		return
	}

	// Obtener las particiones del disco
	partitionsJson, err := disk_operations.ListPartitions(diskPath)
	if err != nil {
		c.JSON(400, PartitionResponse{
			Success: false,
			Message: fmt.Sprintf("Error al obtener las particiones: %v", err),
		})
		return
	}

	// Convertir el string JSON a una estructura de datos
	var partitionsData interface{}
	if err := json.Unmarshal([]byte(partitionsJson), &partitionsData); err != nil {
		c.JSON(500, PartitionResponse{
			Success: false,
			Message: fmt.Sprintf("Error al procesar los datos de particiones: %v", err),
		})
		return
	}

	// Devolver la información de las particiones en formato JSON
	c.JSON(200, PartitionResponse{
		Success:    true,
		Message:    "Particiones obtenidas correctamente",
		Partitions: partitionsData,
	})
}
